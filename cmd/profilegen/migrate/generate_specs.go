// generate_specs.go - generate YAML spec templates based on existing profiles
//
// Usage: go run cmd/profilegen/generate_specs.go
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

// SpecTemplate YAML configurationtemplate
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
		fmt.Fprintf(os.Stderr, "createdirectoryfailed: %v\n", err)
		os.Exit(1)
	}

	// generate YAML template for each profile
	generated := 0
	for name, profile := range profiles.MappedTLSClients {
		if err := generateSpec(name, profile, specsDir); err != nil {
			fmt.Fprintf(os.Stderr, "generate %s failed: %v\n", name, err)
			continue
		}
		generated++
	}

	fmt.Printf("✓ generated %d YAML configuration files to %s/\n", generated, specsDir)
	fmt.Println("\nTips:")
	fmt.Println("1. These files are templates, need to manually fill in cipher_suites, extensions and other detailed info")
	fmt.Println("2. Refer to chrome_133.yaml to complete other configurations")
	fmt.Println("3. After completion, run: go run ./cmd/profilegen -input profiles/specs -output profiles/generated.go")
}

func generateSpec(name string, profile profiles.ClientProfile, outputDir string) error {
	helloID := profile.GetClientHelloId()

	// infer version from profile name
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

	// serialize to YAML
	data, err := yaml.Marshal(spec)
	if err != nil {
		return err
	}

	// add header comments
	header := fmt.Sprintf("# %s fingerprint configuration\n# Generated from profiles.MappedTLSClients[%q]\n#\n# TODO: Fill in cipher_suites, extensions, compression_methods\n# Reference: profiles/specs/chrome_133.yaml\n\n", spec.DisplayName, name)

	outputPath := filepath.Join(outputDir, name+".yaml")
	content := header + string(data)

	// if file already exists, don't overwrite (avoid losing manually filled content)
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("  ⚠ %s already exists, skipping\n", outputPath)
		return nil
	}

	return os.WriteFile(outputPath, []byte(content), 0644)
}

func extractVersion(name string) string {
	// try to extract version number from name
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
	// simple check whether it is a version number (contains digits)
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
	// convert uint32 setting ID to names
	// simplified implementation, actual needs mapping table
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
