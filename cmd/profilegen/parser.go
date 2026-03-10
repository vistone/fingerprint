//go:build profilegen
// +build profilegen

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAMLProfile is the root structure of YAML configuration file
type YAMLProfile struct {
	Name                 string             `yaml:"name"`
	VarName              string             `yaml:"var_name"`
	DisplayName          string             `yaml:"display_name"`
	Client               string             `yaml:"client"`
	Version              string             `yaml:"version"`
	RandomExtensionOrder bool               `yaml:"random_extension_order"`
	CipherSuites         []string           `yaml:"cipher_suites"`
	CompressionMethods   []string           `yaml:"compression_methods"`
	Extensions           []YAMLExtension    `yaml:"extensions"`
	Settings             map[string]uint32  `yaml:"settings"`
	SettingsOrder        []string           `yaml:"settings_order"`
	PseudoHeaderOrder    []string           `yaml:"pseudo_header_order"`
	ConnectionFlow       uint32             `yaml:"connection_flow"`
	HeaderPriority       *YAMLPriorityParam `yaml:"header_priority,omitempty"`
	Priorities           []YAMLPriority     `yaml:"priorities,omitempty"`
}

type YAMLExtension struct {
	Type   string                 `yaml:"type"`
	Params map[string]interface{} `yaml:"params"`
}

type YAMLPriorityParam struct {
	StreamDep uint32 `yaml:"stream_dep"`
	Exclusive bool   `yaml:"exclusive"`
	Weight    uint8  `yaml:"weight"`
}

type YAMLPriority struct {
	StreamID uint32 `yaml:"stream_id"`
	YAMLPriorityParam
}

// parseYAMLFile parses single YAML file with security verification
func parseYAMLFile(path string) (*YAMLProfile, error) {
	// security check 1: prevent path traversal attack (check first)
	if strings.Contains(path, "..") {
		return nil, fmt.Errorf("path contains illegal characters '..': %s", path)
	}
	
	// security check 2: verify if file path is within allowed directory
	allowedDirs := []string{"profiles/specs", "cmd/profilegen/extract"}
	isAllowed := false
	
	// cleanup path and convert to absolute path
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path %s: %w", path, err)
	}
	
	for _, allowedDir := range allowedDirs {
		absAllowedDir, err := filepath.Abs(allowedDir)
		if err != nil {
			continue
		}
		
		// check if path starts with allowed directory
		if strings.HasPrefix(absPath, absAllowedDir) {
			isAllowed = true
			break
		}
	}
	
	if !isAllowed {
		return nil, fmt.Errorf("path not in allowed range: %s (allowed directories: %v)", path, allowedDirs)
	}
	
	// security check 3: check file size (maximum 10MB)
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("getfileinfofailed %s: %w", path, err)
	}
	
	if fileInfo.Size() > maxFileSize {
		return nil, fmt.Errorf("file size exceeds limit (%d > %d bytes): %s", fileInfo.Size(), maxFileSize, path)
	}
	
	// read file content
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	
	// security check 4: verify read data size again
	if len(data) > maxFileSize {
		return nil, fmt.Errorf("file data size exceeds limit: %s", path)
	}

	var profile YAMLProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("parse YAML failed %s: %w", path, err)
	}

	return &profile, nil
}

// parseAllProfiles parses all YAML files in directory with security verification
func parseAllProfiles(dir string) ([]ProfileSpec, []string, error) {
	// security check: verify directory path
	allowedBaseDirs := []string{"profiles/specs", "cmd/profilegen/extract"}
	isAllowed := false
	
	cleanDir := filepath.Clean(dir)
	absDir, err := filepath.Abs(cleanDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get directory absolute path: %w", err)
	}
	
	for _, allowedDir := range allowedBaseDirs {
		absAllowedDir, err := filepath.Abs(allowedDir)
		if err != nil {
			continue
		}
		
		if strings.HasPrefix(absDir, absAllowedDir) || absDir == absAllowedDir {
			isAllowed = true
			break
		}
	}
	
	if !isAllowed {
		return nil, nil, fmt.Errorf("directory not in allowed range: %s (allowed directories: %v)", dir, allowedBaseDirs)
	}
	
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var profiles []ProfileSpec
	var files []string

	for _, entry := range entries {
		// security check: skip symbolic links
		if entry.Type()&fs.ModeSymlink != 0 {
			fmt.Printf("warning: skipping symbolic link %s\n", entry.Name())
			continue
		}
		
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		
		// double check: ensure joined path is still within allowed directory
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Printf("warning: skipping file %s (cannot get absolute path)\n", entry.Name())
			continue
		}
		
		validPath := false
		for _, allowedDir := range allowedBaseDirs {
			absAllowedDir, _ := filepath.Abs(allowedDir)
			if strings.HasPrefix(absPath, absAllowedDir) {
				validPath = true
				break
			}
		}
		
		if !validPath {
			fmt.Printf("warning: skipping file %s (path not in allowed range)\n", entry.Name())
			continue
		}
		
		yamlProfile, err := parseYAMLFile(path)
		if err != nil {
			return nil, nil, err
		}

		// convert to ProfileSpec
		spec := convertToSpec(yamlProfile)
		profiles = append(profiles, spec)
		files = append(files, entry.Name())
	}

	return profiles, files, nil
}

// convertToSpec converts YAMLProfile to ProfileSpec
func convertToSpec(yaml *YAMLProfile) ProfileSpec {
	return ProfileSpec{
		Name:                 yaml.Name,
		VarName:              yaml.VarName,
		DisplayName:          yaml.DisplayName,
		Client:               yaml.Client,
		Version:              yaml.Version,
		RandomExtensionOrder: yaml.RandomExtensionOrder,
		CipherSuites:         yaml.CipherSuites,
		CompressionMethods:   yaml.CompressionMethods,
		Extensions:           convertExtensions(yaml.Extensions),
		Settings:             yaml.Settings,
		SettingsOrder:        yaml.SettingsOrder,
		PseudoHeaderOrder:    yaml.PseudoHeaderOrder,
		ConnectionFlow:       yaml.ConnectionFlow,
		HeaderPriority:       convertHeaderPriority(yaml.HeaderPriority),
		Priorities:           convertPriorities(yaml.Priorities),
	}
}

// convertExtensions converts YAML extensions to Go code strings
func convertExtensions(exts []YAMLExtension) []string {
	var result []string
	for _, ext := range exts {
		code := generateExtensionCode(ext)
		result = append(result, code)
	}
	return result
}

// generateExtensionCode generates Go code for extensions
func generateExtensionCode(ext YAMLExtension) string {
	switch ext.Type {
	case "UtlsGREASEExtension":
		return "&tls.UtlsGREASEExtension{}"

	case "SessionTicketExtension":
		return "&tls.SessionTicketExtension{}"

	case "SignatureAlgorithmsExtension":
		return generateSignatureAlgorithmsExtension(ext.Params)

	case "ApplicationSettingsExtensionNew":
		return generateApplicationSettingsExtensionNew(ext.Params)

	case "KeyShareExtension":
		return generateKeyShareExtension(ext.Params)

	case "SCTExtension":
		return "&tls.SCTExtension{}"

	case "SupportedPointsExtension":
		return generateSupportedPointsExtension(ext.Params)

	case "SupportedVersionsExtension":
		return generateSupportedVersionsExtension(ext.Params)

	case "StatusRequestExtension":
		return "&tls.StatusRequestExtension{}"

	case "ALPNExtension":
		return generateALPNExtension(ext.Params)

	case "SNIExtension":
		return "&tls.SNIExtension{}"

	case "BoringGREASEECH":
		return "tls.BoringGREASEECH()"

	case "UtlsCompressCertExtension":
		return generateCompressCertExtension(ext.Params)

	case "SupportedCurvesExtension":
		return generateSupportedCurvesExtension(ext.Params)

	case "PSKKeyExchangeModesExtension":
		return generatePSKKeyExchangeModesExtension(ext.Params)

	case "ExtendedMasterSecretExtension":
		return "&tls.ExtendedMasterSecretExtension{}"

	case "RenegotiationInfoExtension":
		return generateRenegotiationInfoExtension(ext.Params)

	default:
		return fmt.Sprintf("// TODO: implement %s", ext.Type)
	}
}

// code generation functions for each extension type
func generateSignatureAlgorithmsExtension(params map[string]interface{}) string {
	algorithms, ok := params["supported_signature_algorithms"].([]interface{})
	if !ok {
		return "&tls.SignatureAlgorithmsExtension{}"
	}

	var parts []string
	for _, alg := range algorithms {
		parts = append(parts, fmt.Sprintf("tls.%v", alg))
	}

	return fmt.Sprintf("&tls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []tls.SignatureScheme{%s}}",
		strings.Join(parts, ", "))
}

func generateApplicationSettingsExtensionNew(params map[string]interface{}) string {
	protocols, ok := params["supported_protocols"].([]interface{})
	if !ok {
		return "&tls.ApplicationSettingsExtensionNew{}"
	}

	var parts []string
	for _, p := range protocols {
		parts = append(parts, fmt.Sprintf("\"%v\"", p))
	}

	return fmt.Sprintf("&tls.ApplicationSettingsExtensionNew{SupportedProtocols: []string{%s}}",
		strings.Join(parts, ", "))
}

func generateKeyShareExtension(params map[string]interface{}) string {
	shares, ok := params["key_shares"].([]interface{})
	if !ok {
		return "&tls.KeyShareExtension{}"
	}

	var parts []string
	for _, share := range shares {
		if m, ok := share.(map[string]interface{}); ok {
			group := m["group"]
			data, hasData := m["data"].([]interface{})

			if hasData && len(data) > 0 {
				parts = append(parts, fmt.Sprintf("{Group: tls.%v, Data: []byte{%v}}", group, data[0]))
			} else {
				parts = append(parts, fmt.Sprintf("{Group: tls.%v}", group))
			}
		}
	}

	return fmt.Sprintf("&tls.KeyShareExtension{KeyShares: []tls.KeyShare{%s}}",
		strings.Join(parts, ", "))
}

func generateSupportedPointsExtension(params map[string]interface{}) string {
	points, ok := params["supported_points"].([]interface{})
	if !ok {
		return "&tls.SupportedPointsExtension{}"
	}

	var parts []string
	for _, p := range points {
		parts = append(parts, fmt.Sprintf("tls.%v", p))
	}

	return fmt.Sprintf("&tls.SupportedPointsExtension{SupportedPoints: []byte{%s}}",
		strings.Join(parts, ", "))
}

func generateSupportedVersionsExtension(params map[string]interface{}) string {
	versions, ok := params["versions"].([]interface{})
	if !ok {
		return "&tls.SupportedVersionsExtension{}"
	}

	var parts []string
	for _, v := range versions {
		parts = append(parts, fmt.Sprintf("tls.%v", v))
	}

	return fmt.Sprintf("&tls.SupportedVersionsExtension{Versions: []uint16{%s}}",
		strings.Join(parts, ", "))
}

func generateALPNExtension(params map[string]interface{}) string {
	protocols, ok := params["alpn_protocols"].([]interface{})
	if !ok {
		return "&tls.ALPNExtension{}"
	}

	var parts []string
	for _, p := range protocols {
		parts = append(parts, fmt.Sprintf("\"%v\"", p))
	}

	return fmt.Sprintf("&tls.ALPNExtension{AlpnProtocols: []string{%s}}",
		strings.Join(parts, ", "))
}

func generateCompressCertExtension(params map[string]interface{}) string {
	algorithms, ok := params["algorithms"].([]interface{})
	if !ok {
		return "&tls.UtlsCompressCertExtension{}"
	}

	var parts []string
	for _, alg := range algorithms {
		parts = append(parts, fmt.Sprintf("tls.%v", alg))
	}

	return fmt.Sprintf("&tls.UtlsCompressCertExtension{Algorithms: []tls.CertCompressionAlgo{%s}}",
		strings.Join(parts, ", "))
}

func generateSupportedCurvesExtension(params map[string]interface{}) string {
	curves, ok := params["curves"].([]interface{})
	if !ok {
		return "&tls.SupportedCurvesExtension{}"
	}

	var parts []string
	for _, c := range curves {
		parts = append(parts, fmt.Sprintf("tls.%v", c))
	}

	return fmt.Sprintf("&tls.SupportedCurvesExtension{Curves: []tls.CurveID{%s}}",
		strings.Join(parts, ", "))
}

func generatePSKKeyExchangeModesExtension(params map[string]interface{}) string {
	modes, ok := params["modes"].([]interface{})
	if !ok {
		return "&tls.PSKKeyExchangeModesExtension{}"
	}

	var parts []string
	for _, m := range modes {
		parts = append(parts, fmt.Sprintf("tls.%v", m))
	}

	return fmt.Sprintf("&tls.PSKKeyExchangeModesExtension{Modes: []uint8{%s}}",
		strings.Join(parts, ", "))
}

func generateRenegotiationInfoExtension(params map[string]interface{}) string {
	reneg, ok := params["renegotiation"].(string)
	if !ok {
		return "&tls.RenegotiationInfoExtension{}"
	}

	return fmt.Sprintf("&tls.RenegotiationInfoExtension{Renegotiation: tls.%s}", reneg)
}

func convertHeaderPriority(yaml *YAMLPriorityParam) *PriorityParam {
	if yaml == nil {
		return nil
	}
	return &PriorityParam{
		StreamDep: yaml.StreamDep,
		Exclusive: yaml.Exclusive,
		Weight:    yaml.Weight,
	}
}

func convertPriorities(yamls []YAMLPriority) []Priority {
	var result []Priority
	for _, y := range yamls {
		result = append(result, Priority{
			StreamID: y.StreamID,
			PriorityParam: PriorityParam{
				StreamDep: y.StreamDep,
				Exclusive: y.Exclusive,
				Weight:    y.Weight,
			},
		})
	}
	return result
}
