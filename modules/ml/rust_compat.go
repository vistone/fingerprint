// Package ml provides data compatibility with Rust fingerprint library
// Supports importing training data and models from Rust versions
package ml

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vistone/fingerprint/modules/core"
)

// RustFingerprint Rust format fingerprint data
type RustFingerprint struct {
	ID       string             `json:"id"`
	Browser  string             `json:"browser"`
	Version  string             `json:"version"`
	OS       string             `json:"os"`
	TLS      RustTLSData        `json:"tls"`
	HTTP2    RustHTTP2Data      `json:"http2"`
	Headers  map[string]string  `json:"headers"`
	Features map[string]float64 `json:"features"`
}

// RustTLSData Rust TLS data
type RustTLSData struct {
	Version         uint16   `json:"version"`
	CipherSuites    []uint16 `json:"cipher_suites"`
	Extensions      []uint16 `json:"extensions"`
	SupportedCurves []uint16 `json:"supported_curves"`
	PointFormats    []uint8  `json:"point_formats"`
}

// RustHTTP2Data Rust HTTP/2 data
type RustHTTP2Data struct {
	Settings          map[string]uint32 `json:"settings"`
	Priorities        []RustPriority    `json:"priorities"`
	PseudoHeaderOrder []string          `json:"pseudo_header_order"`
}

// RustPriority Rust HTTP/2 priority
type RustPriority struct {
	StreamID  uint32 `json:"stream_id"`
	Weight    uint8  `json:"weight"`
	DependsOn uint32 `json:"depends_on"`
	Exclusive bool   `json:"exclusive"`
}

// RustDataset Rust format dataset
type RustDataset struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Fingerprints []RustFingerprint `json:"fingerprints"`
}

// RustModel Rust format model
type RustModel struct {
	Name          string                         `json:"name"`
	Version       string                         `json:"version"`
	ProtocolLayer RustClassifierLayer            `json:"protocol_layer"`
	FamilyLayers  map[string]RustClassifierLayer `json:"family_layers"`  // protocol -> layer
	VersionLayers map[string]RustClassifierLayer `json:"version_layers"` // family -> layer
}

// RustClassifierLayer Rust classifier layer
type RustClassifierLayer struct {
	Centroids    map[string][]float64 `json:"centroids"` // label -> center
	Weights      []float64            `json:"weights"`
	FeatureNames []string             `json:"feature_names"`
}

// RustImporter Rust data importer
type RustImporter struct {
	dataPath string
}

// NewRustImporter create new Rust data importer
func NewRustImporter(dataPath string) *RustImporter {
	return &RustImporter{dataPath: dataPath}
}

// ImportDataset import Rust format dataset
func (ri *RustImporter) ImportDataset(filename string) (*Dataset, error) {
	path := filepath.Join(ri.dataPath, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read rust dataset: %w", err)
	}

	var rustDataset RustDataset
	if err := json.Unmarshal(data, &rustDataset); err != nil {
		return nil, fmt.Errorf("unmarshal rust dataset: %w", err)
	}

	return ri.convertDataset(&rustDataset), nil
}

// convertDataset convert Rust dataset to Go format
func (ri *RustImporter) convertDataset(rust *RustDataset) *Dataset {
	dataset := &Dataset{
		Name:        rust.Name,
		Version:     rust.Version,
		Description: rust.Description,
		Samples:     make([]TrainingSample, 0, len(rust.Fingerprints)),
	}

	for _, fp := range rust.Fingerprints {
		sample := ri.convertFingerprint(&fp)
		dataset.Samples = append(dataset.Samples, sample)
	}

	dataset.updateStatistics()
	return dataset
}

// convertFingerprint convert Rust fingerprint to Go training sample
func (ri *RustImporter) convertFingerprint(rust *RustFingerprint) TrainingSample {
	// Create feature vector
	fv := convertRustFeatures(rust.Features)

	// Convert label
	label := TrainingLabel{
		Protocol: inferProtocolFromFeatures(rust.Features),
		Family:   inferBrowserFamily(rust.Browser),
		Version:  rust.Version,
		OS:       rust.OS,
	}

	return TrainingSample{
		ID:       rust.ID,
		Features: fv,
		Label:    label,
		Metadata: map[string]interface{}{
			"source":      "rust",
			"browser":     rust.Browser,
			"os":          rust.OS,
			"tls_version": rust.TLS.Version,
		},
	}
}

// convertRustFeatures convert Rust features to Go feature vector
func convertRustFeatures(rustFeatures map[string]float64) *core.FeatureVector {
	fv := core.NewFeatureVector()

	// Feature name mapping
	featureMapping := map[string]core.FeatureType{
		"tls_version":      core.FeatureTLSVersion,
		"cipher_suites":    core.FeatureCipherSuites,
		"extensions":       core.FeatureExtensions,
		"http2_settings":   core.FeatureHTTP2Settings,
		"http_headers":     core.FeatureHTTPHeaders,
		"user_agent":       core.FeatureUserAgent,
		"canvas":           core.FeatureCanvas,
		"webgl":            core.FeatureWebGL,
		"audio":            core.FeatureAudio,
		"fonts":            core.FeatureFonts,
		"storage":          core.FeatureStorage,
		"webrtc":           core.FeatureWebRTC,
		"hardware":         core.FeatureHardware,
		"timing":           core.FeatureTiming,
		"headless_browser": core.FeatureHeadlessBrowser,
		"entropy":          core.FeatureEntropy,
		"tool_marker":      core.FeatureToolMarker,
		"behavior_pattern": core.FeatureBehaviorPattern,
	}

	for rustName, value := range rustFeatures {
		if goFeature, ok := featureMapping[rustName]; ok {
			fv.Set(goFeature, value)
		} else {
			// Store unknown features in metadata
			fv.Metadata[rustName] = value
		}
	}

	return fv
}

// inferProtocolFromFeatures infer protocol type from features
func inferProtocolFromFeatures(features map[string]float64) core.ProtocolType {
	if _, ok := features["quic_version"]; ok {
		return core.ProtocolQUIC
	}
	if tlsVer, ok := features["tls_version"]; ok {
		if tlsVer >= 0x0304 {
			return core.ProtocolHTTP3
		}
	}
	if _, ok := features["http2_settings"]; ok {
		return core.ProtocolHTTP2
	}
	return core.ProtocolTLS
}

// inferBrowserFamily infer browser family
func inferBrowserFamily(browser string) core.BrowserType {
	switch browser {
	case "chrome", "Chrome":
		return core.BrowserChrome
	case "firefox", "Firefox":
		return core.BrowserFirefox
	case "safari", "Safari":
		return core.BrowserSafari
	case "edge", "Edge":
		return core.BrowserEdge
	case "opera", "Opera":
		return core.BrowserOpera
	default:
		return core.BrowserChrome
	}
}

// ImportModel import Rust format model
func (ri *RustImporter) ImportModel(filename string) (*PretrainedModel, error) {
	path := filepath.Join(ri.dataPath, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read rust model: %w", err)
	}

	var rustModel RustModel
	if err := json.Unmarshal(data, &rustModel); err != nil {
		return nil, fmt.Errorf("unmarshal rust model: %w", err)
	}

	return ri.convertModel(&rustModel), nil
}

// convertModel convert Rust model to Go format
func (ri *RustImporter) convertModel(rust *RustModel) *PretrainedModel {
	model := &PretrainedModel{
		Name:            rust.Name,
		Version:         rust.Version,
		ProtocolCenters: rust.ProtocolLayer.Centroids,
		FamilyCenters:   make(map[string]map[string][]float64),
		VersionCenters:  make(map[string]map[string][]float64),
		FeatureWeights:  rust.ProtocolLayer.Weights,
	}

	// Convert family layers
	for proto, layer := range rust.FamilyLayers {
		model.FamilyCenters[proto] = layer.Centroids
	}

	// Convert version layers
	for family, layer := range rust.VersionLayers {
		model.VersionCenters[family] = layer.Centroids
	}

	return model
}

// ExportToRust export to Rust format (bidirectional compatibility)
func (d *Dataset) ExportToRust() *RustDataset {
	rust := &RustDataset{
		Name:         d.Name,
		Version:      d.Version,
		Description:  d.Description,
		Fingerprints: make([]RustFingerprint, 0, len(d.Samples)),
	}

	for _, sample := range d.Samples {
		rustFp := RustFingerprint{
			ID:       sample.ID,
			Browser:  string(sample.Label.Family),
			Version:  sample.Label.Version,
			OS:       sample.Label.OS,
			Features: make(map[string]float64),
		}

		// Convert features
		for ft, value := range sample.Features.Features {
			rustName := convertGoFeatureToRust(ft)
			rustFp.Features[rustName] = value
		}

		rust.Fingerprints = append(rust.Fingerprints, rustFp)
	}

	return rust
}

// convertGoFeatureToRust convert Go feature name to Rust feature name
func convertGoFeatureToRust(ft core.FeatureType) string {
	featureMapping := map[core.FeatureType]string{
		core.FeatureTLSVersion:      "tls_version",
		core.FeatureCipherSuites:    "cipher_suites",
		core.FeatureExtensions:      "extensions",
		core.FeatureHTTP2Settings:   "http2_settings",
		core.FeatureHTTPHeaders:     "http_headers",
		core.FeatureUserAgent:       "user_agent",
		core.FeatureCanvas:          "canvas",
		core.FeatureWebGL:           "webgl",
		core.FeatureAudio:           "audio",
		core.FeatureFonts:           "fonts",
		core.FeatureStorage:         "storage",
		core.FeatureWebRTC:          "webrtc",
		core.FeatureHardware:        "hardware",
		core.FeatureTiming:          "timing",
		core.FeatureHeadlessBrowser: "headless_browser",
		core.FeatureEntropy:         "entropy",
		core.FeatureToolMarker:      "tool_marker",
		core.FeatureBehaviorPattern: "behavior_pattern",
	}

	if rustName, ok := featureMapping[ft]; ok {
		return rustName
	}
	return string(ft)
}

// SaveToRustFormat save to Rust format
func (d *Dataset) SaveToRustFormat(path string) error {
	rust := d.ExportToRust()

	data, err := json.MarshalIndent(rust, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal rust dataset: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write rust dataset: %w", err)
	}

	return nil
}

// CompatibilityChecker compatibility checker
type CompatibilityChecker struct {
	goFeatures   map[string]bool
	rustFeatures map[string]bool
}

// NewCompatibilityChecker create new compatibility checker
func NewCompatibilityChecker() *CompatibilityChecker {
	return &CompatibilityChecker{
		goFeatures: map[string]bool{
			"tls_version":      true,
			"cipher_suites":    true,
			"extensions":       true,
			"http2_settings":   true,
			"http_headers":     true,
			"user_agent":       true,
			"canvas":           true,
			"webgl":            true,
			"audio":            true,
			"fonts":            true,
			"storage":          true,
			"webrtc":           true,
			"hardware":         true,
			"timing":           true,
			"headless_browser": true,
			"entropy":          true,
			"tool_marker":      true,
			"behavior_pattern": true,
		},
		rustFeatures: map[string]bool{
			"tls_version":      true,
			"cipher_suites":    true,
			"extensions":       true,
			"http2_settings":   true,
			"http_headers":     true,
			"user_agent":       true,
			"canvas":           true,
			"webgl":            true,
			"audio":            true,
			"fonts":            true,
			"storage":          true,
			"webrtc":           true,
			"hardware":         true,
			"timing":           true,
			"headless_browser": true,
			"entropy":          true,
			"tool_marker":      true,
			"behavior_pattern": true,
		},
	}
}

// CheckFeatureCompatibility check feature compatibility
func (cc *CompatibilityChecker) CheckFeatureCompatibility() map[string]string {
	result := make(map[string]string)

	for feature := range cc.goFeatures {
		if cc.rustFeatures[feature] {
			result[feature] = "compatible"
		} else {
			result[feature] = "go_only"
		}
	}

	for feature := range cc.rustFeatures {
		if !cc.goFeatures[feature] {
			result[feature] = "rust_only"
		}
	}

	return result
}

// ImportWithValidation import and validate
func (ri *RustImporter) ImportWithValidation(filename string) (*Dataset, error) {
	dataset, err := ri.ImportDataset(filename)
	if err != nil {
		return nil, err
	}

	// Validate data integrity
	issues := ri.validateDataset(dataset)
	if len(issues) > 0 {
		fmt.Printf("Import validation issues:\n")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
	}

	return dataset, nil
}

// validateDataset validate dataset
func (ri *RustImporter) validateDataset(dataset *Dataset) []string {
	var issues []string

	if len(dataset.Samples) == 0 {
		issues = append(issues, "dataset has no samples")
	}

	// Check sample integrity
	for i, sample := range dataset.Samples {
		if sample.ID == "" {
			issues = append(issues, fmt.Sprintf("sample %d has no ID", i))
		}
		if sample.Label.Family == "" && sample.Label.Protocol == "" {
			issues = append(issues, fmt.Sprintf("sample %s has no label", sample.ID))
		}
		if len(sample.Features.Features) == 0 {
			issues = append(issues, fmt.Sprintf("sample %s has no features", sample.ID))
		}
	}

	return issues
}
