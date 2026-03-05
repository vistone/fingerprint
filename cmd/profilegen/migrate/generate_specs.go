// generate_specs.go - 基于现有 profiles 生成 YAML spec 模板
//
// 用法: go run cmd/profilegen/generate_specs.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bogdanfinn/fhttp/http2"
	"github.com/vistone/fingerprint/modules/profiles/legacy"
	"gopkg.in/yaml.v3"
)

// SpecTemplate YAML 配置模板
type SpecTemplate struct {
	Name                 string            `yaml:"name"`
	VarName              string            `yaml:"var_name"`
	DisplayName          string            `yaml:"display_name"`
	Client               string            `yaml:"client"`
	Version              string            `yaml:"version"`
	RandomExtensionOrder bool              `yaml:"random_extension_order"`
	CipherSuites         []string          `yaml:"cipher_suites"`
	CompressionMethods   []string          `yaml:"compression_methods"`
	Extensions           []ExtensionDef    `yaml:"extensions"`
	Settings             map[string]uint32 `yaml:"settings"`
	SettingsOrder        []string          `yaml:"settings_order"`
	PseudoHeaderOrder    []string          `yaml:"pseudo_header_order"`
	ConnectionFlow       uint32            `yaml:"connection_flow"`
}

type ExtensionDef struct {
	Type   string                 `yaml:"type"`
	Params map[string]interface{} `yaml:"params,omitempty"`
}

func main() {
	specsDir := "profiles/specs"
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "创建目录失败: %v\n", err)
		os.Exit(1)
	}

	// 为每个 profile 生成 YAML 模板
	generated := 0
	for name, profile := range profiles.MappedTLSClients {
		if err := generateSpec(name, profile, specsDir); err != nil {
			fmt.Fprintf(os.Stderr, "生成 %s 失败: %v\n", name, err)
			continue
		}
		generated++
	}

	fmt.Printf("✓ 生成了 %d 个 YAML 配置文件到 %s/\n", generated, specsDir)
	fmt.Println("\n提示:")
	fmt.Println("1. 这些文件是模板，需要手动填写 cipher_suites, extensions 等详细信息")
	fmt.Println("2. 参考 chrome_133.yaml 完成其他配置")
	fmt.Println("3. 完成后运行: go run ./cmd/profilegen -input profiles/specs -output profiles/generated.go")
}

func generateSpec(name string, profile profiles.ClientProfile, outputDir string) error {
	helloID := profile.GetClientHelloId()

	// 从 profile 名称推断版本
	version := extractVersion(name)

	spec := SpecTemplate{
		Name:                 name,
		VarName:              toPascalCase(name),
		DisplayName:          strings.ReplaceAll(name, "_", " "),
		Client:               helloID.Client,
		Version:              version,
		RandomExtensionOrder: false,
		CipherSuites:         []string{},
		CompressionMethods:   []string{"tls.CompressionNone"},
		Extensions:           []ExtensionDef{},
		Settings:             convertSettings(profile.GetSettings()),
		SettingsOrder:        toSettingNames(profile.GetSettingsOrder()),
		PseudoHeaderOrder:    profile.GetPseudoHeaderOrder(),
		ConnectionFlow:       profile.GetConnectionFlow(),
	}

	// 序列化为 YAML
	data, err := yaml.Marshal(spec)
	if err != nil {
		return err
	}

	// 添加头部注释
	header := fmt.Sprintf("# %s fingerprint configuration\n# Generated from profiles.MappedTLSClients[%q]\n#\n# TODO: Fill in cipher_suites, extensions, compression_methods\n# Reference: profiles/specs/chrome_133.yaml\n\n", spec.DisplayName, name)

	outputPath := filepath.Join(outputDir, name+".yaml")
	content := header + string(data)

	// 如果文件已存在，不要覆盖（避免丢失手动填写的内容）
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("  ⚠ %s 已存在，跳过\n", outputPath)
		return nil
	}

	return os.WriteFile(outputPath, []byte(content), 0644)
}

func extractVersion(name string) string {
	// 尝试从名称中提取版本号
	// chrome_133 -> 133
	// firefox_120 -> 120
	parts := strings.Split(name, "_")
	for _, part := range parts {
		if isVersion(part) {
			return part
		}
	}
	return ""
}

func isVersion(s string) bool {
	// 简单判断是否为版本号（包含数字）
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func toPascalCase(s string) string {
	// chrome_133 -> Chrome_133
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "_")
}

func convertSettings(settings map[http2.SettingID]uint32) map[string]uint32 {
	result := make(map[string]uint32)
	for k, v := range settings {
		result[fmt.Sprintf("%d", k)] = v
	}
	return result
}

func toSettingNames(order []http2.SettingID) []string {
	// 将 uint32 设置 ID 转换为名称
	// 简化实现，实际需要映射表
	names := []string{
		"SettingHeaderTableSize",
		"SettingEnablePush",
		"SettingInitialWindowSize",
		"SettingMaxHeaderListSize",
	}
	result := make([]string, 0, len(order))
	for i := range order {
		if i < len(names) {
			result = append(result, names[i])
		}
	}
	return result
}
