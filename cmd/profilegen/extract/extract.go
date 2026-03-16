// extract.go - extract fingerprint configuration from existing Go files and generate YAML
//
// this tool parses Go files under profiles/ directory, extracts ClientProfile definitions,
// and generates corresponding YAML configuration files.
//
// note: this is a simplified implementation, complete implementation requires full Go AST parsing.
// current version is for demonstration of migration process.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ProfileData holds configuration data extracted from Go code.
type ProfileData struct {
	Name         string
	VarName      string
	Client       string
	Version      string
	CipherSuites []string
	Extensions   []string
}

func main() {
	var (
		inputDir  = flag.String("input", "profiles", "input Go file directory")
		outputDir = flag.String("output", "profiles/specs", "output YAML directory")
	)
	flag.Parse()

	// create output directory
	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create directory: %v\n", err)
		os.Exit(1)
	}

	// parse Go files
	profiles, err := extractProfilesFromDir(*inputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to extract configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("extracted %d fingerprint configurations\n", len(profiles))

	// generate YAML files
	for _, profile := range profiles {
		outputPath := filepath.Join(*outputDir, profile.Name+".yaml")
		if err := generateYAML(profile, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "failed to generate YAML %s: %v\n", profile.Name, err)
			continue
		}
		fmt.Printf("✓ generated %s\n", outputPath)
	}
}

// extractProfilesFromDir extracts all fingerprint configurations from a directory.
func extractProfilesFromDir(dir string) ([]ProfileData, error) {
	var profiles []ProfileData

	// use simplified method: regex extraction
	// complete implementation should use go/ast for full parsing
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(file.Name(), "_test.go") {
			continue
		}

		path := filepath.Join(dir, file.Name())
		fileProfiles, err := extractProfilesFromFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to parse %s: %v\n", path, err)
			continue
		}
		profiles = append(profiles, fileProfiles...)
	}

	return profiles, nil
}

// extractProfilesFromFile extracts fingerprint configuration from a single Go file.
func extractProfilesFromFile(path string) ([]ProfileData, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var profiles []ProfileData

	// iterate all declarations
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for _, value := range valueSpec.Values {
				compositeLit, ok := value.(*ast.CompositeLit)
				if !ok {
					continue
				}

				// check if it is ClientProfile type
				if isClientProfileType(compositeLit.Type) {
					profile := extractProfileData(valueSpec.Names[0].Name, compositeLit)
					if profile.Name != "" {
						profiles = append(profiles, profile)
					}
				}
			}
		}
	}

	return profiles, nil
}

// isClientProfileType checks whether type is ClientProfile.
func isClientProfileType(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "ClientProfile"
}

// extractProfileData extracts configuration data from CompositeLit.
func extractProfileData(varName string, lit *ast.CompositeLit) ProfileData {
	profile := ProfileData{
		VarName: varName,
		Name:    toSnakeCase(varName),
	}

	for _, elt := range lit.Elts {
		kvExpr, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key := getIdentName(kvExpr.Key)
		switch key {
		case "clientHelloId":
			if clientInfo := extractClientHelloId(kvExpr.Value); clientInfo != nil {
				profile.Client = clientInfo["Client"]
				profile.Version = clientInfo["Version"]
			}
		}
	}

	return profile
}

// extractClientHelloId extracts ClientHelloID information.
func extractClientHelloId(expr ast.Expr) map[string]string {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}

	info := make(map[string]string)
	for _, elt := range lit.Elts {
		kvExpr, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key := getIdentName(kvExpr.Key)
		value := getBasicLitValue(kvExpr.Value)
		info[key] = value
	}

	return info
}

// getIdentName returns identifier name.
func getIdentName(expr ast.Expr) string {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return ""
	}
	return ident.Name
}

// getBasicLitValue get basic literal value
func getBasicLitValue(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.BasicLit:
		return strings.Trim(v.Value, `"`)
	case *ast.Ident:
		return v.Name
	default:
		return ""
	}
}

// toSnakeCase convert camelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// generateYAML generate YAML configuration file
func generateYAML(profile ProfileData, path string) error {
	// use template to generate YAML
	template := `# %s fingerprint configuration
# Auto-generated from %s

name: %s
var_name: %s
display_name: %s
client: %s
version: "%s"
random_extension_order: false

# TODO: Add cipher_suites, extensions, settings, etc.
# This is a template - manual completion required

cipher_suites: []
extensions: []
settings: {}
settings_order: []
pseudo_header_order: []
connection_flow: 0
`

	content := fmt.Sprintf(template,
		profile.Client,
		profile.VarName,
		profile.Name,
		profile.VarName,
		strings.ReplaceAll(profile.VarName, "_", " "),
		profile.Client,
		profile.Version,
	)

	return os.WriteFile(path, []byte(content), 0o600)
}

// extractWithRegex extract using regex (backup method)
func extractWithRegex(content []byte) []ProfileData {
	var profiles []ProfileData

	// match var Name = ClientProfile{...}
	varRegex := regexp.MustCompile(`var\s+(\w+)\s*=\s*ClientProfile\{`)
	matches := varRegex.FindAllSubmatchIndex(content, -1)

	for _, match := range matches {
		if len(match) >= 4 {
			varName := string(content[match[2]:match[3]])
			profile := ProfileData{
				VarName: varName,
				Name:    toSnakeCase(varName),
			}
			profiles = append(profiles, profile)
		}
	}

	return profiles
}
