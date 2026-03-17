package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type violation struct {
	Rule    string
	File    string
	Line    int
	Details string
}

type config struct {
	root          string
	includeTests  bool
	checkComments bool
	checkFileLen  bool
	checkFuncLen  bool
	checkParams   bool
	maxFileLines  int
	maxFuncLines  int
	maxParams     int
}

func main() {
	cfg := parseFlags()

	var files []string
	err := filepath.WalkDir(cfg.root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		if isGenerated(path) {
			return nil
		}
		if !cfg.includeTests && strings.HasSuffix(path, "_test.go") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan failed: %v\n", err)
		os.Exit(2)
	}

	vios := make([]violation, 0, 64)
	for _, file := range files {
		fileVios, scanErr := scanFile(file, cfg)
		if scanErr != nil {
			fmt.Fprintf(os.Stderr, "parse failed: %s: %v\n", file, scanErr)
			os.Exit(2)
		}
		vios = append(vios, fileVios...)
	}

	sort.Slice(vios, func(i, j int) bool {
		if vios[i].File == vios[j].File {
			if vios[i].Line == vios[j].Line {
				return vios[i].Rule < vios[j].Rule
			}
			return vios[i].Line < vios[j].Line
		}
		return vios[i].File < vios[j].File
	})

	for _, v := range vios {
		fmt.Printf("%s|%s|%d|%s\n", v.Rule, filepath.ToSlash(v.File), v.Line, v.Details)
	}

	if len(vios) > 0 {
		os.Exit(1)
	}
}

func parseFlags() config {
	checks := flag.String("checks", "comments,file-length,func-length,params", "comma-separated checks: comments,file-length,func-length,params")
	root := flag.String("root", ".", "scan root")
	includeTests := flag.Bool("include-tests", false, "include *_test.go files")
	maxFileLines := flag.Int("max-file-lines", 500, "max lines per file")
	maxFuncLines := flag.Int("max-func-lines", 80, "max lines per function")
	maxParams := flag.Int("max-params", 5, "max function parameters")
	flag.Parse()

	cfg := config{
		root:         *root,
		includeTests: *includeTests,
		maxFileLines: *maxFileLines,
		maxFuncLines: *maxFuncLines,
		maxParams:    *maxParams,
	}

	for _, c := range strings.Split(*checks, ",") {
		switch strings.TrimSpace(c) {
		case "comments":
			cfg.checkComments = true
		case "file-length":
			cfg.checkFileLen = true
		case "func-length":
			cfg.checkFuncLen = true
		case "params":
			cfg.checkParams = true
		case "":
		default:
			fmt.Fprintf(os.Stderr, "unknown check: %s\n", c)
			os.Exit(2)
		}
	}

	return cfg
}

func scanFile(path string, cfg config) ([]violation, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	mode := parser.ParseComments
	f, err := parser.ParseFile(fset, path, content, mode)
	if err != nil {
		return nil, err
	}

	vios := make([]violation, 0, 8)
	tokFile := fset.File(f.Pos())
	if tokFile == nil {
		return vios, nil
	}

	if cfg.checkFileLen {
		total := tokFile.LineCount()
		if total > cfg.maxFileLines {
			vios = append(vios, violation{
				Rule:    "file-length",
				File:    path,
				Line:    1,
				Details: fmt.Sprintf("file has %d lines (limit %d)", total, cfg.maxFileLines),
			})
		}
	}

	if cfg.checkComments {
		for _, cg := range f.Comments {
			for _, c := range cg.List {
				if hasHan(c.Text) {
					p := fset.Position(c.Pos())
					vios = append(vios, violation{
						Rule:    "comments",
						File:    path,
						Line:    p.Line,
						Details: "comment contains non-English (Han) characters",
					})
				}
			}
		}
	}

	if cfg.checkFuncLen || cfg.checkParams {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}

			start := fset.Position(fn.Pos()).Line
			end := fset.Position(fn.End()).Line
			funcName := fn.Name.Name

			if cfg.checkFuncLen {
				length := end - start + 1
				if length > cfg.maxFuncLines {
					vios = append(vios, violation{
						Rule:    "func-length",
						File:    path,
						Line:    start,
						Details: fmt.Sprintf("function %s has %d lines (limit %d)", funcName, length, cfg.maxFuncLines),
					})
				}
			}

			if cfg.checkParams && fn.Type.Params != nil {
				paramCount := 0
				for _, field := range fn.Type.Params.List {
					if len(field.Names) == 0 {
						paramCount++
						continue
					}
					paramCount += len(field.Names)
				}
				if paramCount > cfg.maxParams {
					vios = append(vios, violation{
						Rule:    "params",
						File:    path,
						Line:    start,
						Details: fmt.Sprintf("function %s has %d params (limit %d)", funcName, paramCount, cfg.maxParams),
					})
				}
			}
		}
	}

	return vios, nil
}

func isGenerated(path string) bool {
	base := filepath.Base(path)
	if strings.HasSuffix(base, ".pb.go") || strings.HasSuffix(base, "_generated.go") {
		return true
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	head := string(data)
	if len(head) > 4096 {
		head = head[:4096]
	}
	return strings.Contains(head, "Code generated") && strings.Contains(head, "DO NOT EDIT")
}

func hasHan(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}
