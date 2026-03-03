// extract.go - 从现有的 Go 文件提取指纹配置并生成 YAML
//
// 该工具解析 profiles/ 目录下的 Go 文件，提取 ClientProfile 定义，
// 并生成对应的 YAML 配置文件。
//
// 注意：这是一个简化实现，完整的实现需要完整的 Go AST 解析。
// 当前版本用于演示迁移流程。
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

// ProfileData 从 Go 代码提取的配置数据
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
		inputDir  = flag.String("input", "profiles", "输入 Go 文件目录")
		outputDir = flag.String("output", "profiles/specs", "输出 YAML 目录")
	)
	flag.Parse()

	// 创建输出目录
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "创建目录失败: %v\n", err)
		os.Exit(1)
	}

	// 解析 Go 文件
	profiles, err := extractProfilesFromDir(*inputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "提取配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("提取了 %d 个指纹配置\n", len(profiles))

	// 生成 YAML 文件
	for _, profile := range profiles {
		outputPath := filepath.Join(*outputDir, profile.Name+".yaml")
		if err := generateYAML(profile, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "生成 YAML 失败 %s: %v\n", profile.Name, err)
			continue
		}
		fmt.Printf("✓ 生成 %s\n", outputPath)
	}
}

// extractProfilesFromDir 从目录中提取所有指纹配置
func extractProfilesFromDir(dir string) ([]ProfileData, error) {
	var profiles []ProfileData

	// 使用简化方法：正则表达式提取
	// 完整的实现应该使用 go/ast 进行完整解析
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
			fmt.Fprintf(os.Stderr, "警告: 解析 %s 失败: %v\n", path, err)
			continue
		}
		profiles = append(profiles, fileProfiles...)
	}

	return profiles, nil
}

// extractProfilesFromFile 从单个 Go 文件提取指纹配置
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

	// 遍历所有声明
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

				// 检查是否是 ClientProfile 类型
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

// isClientProfileType 检查类型是否为 ClientProfile
func isClientProfileType(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "ClientProfile"
}

// extractProfileData 从 CompositeLit 提取配置数据
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

// extractClientHelloId 提取 ClientHelloID 信息
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

// getIdentName 获取标识符名称
func getIdentName(expr ast.Expr) string {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return ""
	}
	return ident.Name
}

// getBasicLitValue 获取基本字面量值
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

// toSnakeCase 将驼峰命名转换为下划线命名
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

// generateYAML 生成 YAML 配置文件
func generateYAML(profile ProfileData, path string) error {
	// 使用模板生成 YAML
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

	return os.WriteFile(path, []byte(content), 0644)
}

// extractWithRegex 使用正则表达式提取（备用方法）
func extractWithRegex(content []byte) []ProfileData {
	var profiles []ProfileData

	// 匹配 var Name = ClientProfile{...}
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
