// Plugin example demonstrating how to create and use custom plugins
package main

import (
	"context"
	"fmt"

	"github.com/vistone/fingerprint/modules/plugin"
)

// CustomAnalyzer is a custom analyzer plugin
type CustomAnalyzer struct {
	plugin.BasePlugin
	threshold float64
}

// NewCustomAnalyzer creates a new custom analyzer
func NewCustomAnalyzer() *CustomAnalyzer {
	return &CustomAnalyzer{
		BasePlugin: plugin.BasePlugin{},
		threshold:  0.8,
	}
}

// Name returns the plugin name
func (a *CustomAnalyzer) Name() string {
	return "custom-analyzer"
}

// Type returns the plugin type
func (a *CustomAnalyzer) Type() plugin.PluginType {
	return plugin.TypeAnalyzer
}

// Version returns the plugin version
func (a *CustomAnalyzer) Version() string {
	return "1.0.0"
}

// Init initializes the plugin
func (a *CustomAnalyzer) Init(config map[string]interface{}) error {
	if threshold, ok := config["threshold"].(float64); ok {
		a.threshold = threshold
	}
	return nil
}

// Analyze performs custom analysis
func (a *CustomAnalyzer) Analyze(ctx context.Context, data interface{}) (*plugin.AnalysisResult, error) {
	// Custom analysis logic
	return &plugin.AnalysisResult{
		Score:      0.9,
		Confidence: a.threshold,
		Labels: map[string]string{
			"custom": "true",
			"source": "custom-analyzer",
		},
		Annotations: map[string]interface{}{
			"threshold": a.threshold,
		},
	}, nil
}

// CustomTransformer is a custom transformer plugin
type CustomTransformer struct {
	plugin.BasePlugin
	prefix string
}

// NewCustomTransformer creates a new custom transformer
func NewCustomTransformer() *CustomTransformer {
	return &CustomTransformer{
		BasePlugin: plugin.BasePlugin{},
		prefix:     "transformed_",
	}
}

// Name returns the plugin name
func (t *CustomTransformer) Name() string {
	return "custom-transformer"
}

// Type returns the plugin type
func (t *CustomTransformer) Type() plugin.PluginType {
	return plugin.TypeTransformer
}

// Version returns the plugin version
func (t *CustomTransformer) Version() string {
	return "1.0.0"
}

// Init initializes the plugin
func (t *CustomTransformer) Init(config map[string]interface{}) error {
	if prefix, ok := config["prefix"].(string); ok {
		t.prefix = prefix
	}
	return nil
}

// Transform transforms the data
func (t *CustomTransformer) Transform(ctx context.Context, data interface{}) (interface{}, error) {
	// Custom transformation logic
	if str, ok := data.(string); ok {
		return t.prefix + str, nil
	}
	return data, nil
}

// CustomValidator is a custom validator plugin
type CustomValidator struct {
	plugin.BasePlugin
	requiredFields []string
}

// NewCustomValidator creates a new custom validator
func NewCustomValidator() *CustomValidator {
	return &CustomValidator{
		BasePlugin:     plugin.BasePlugin{},
		requiredFields: []string{"tls_version", "cipher_suites"},
	}
}

// Name returns the plugin name
func (v *CustomValidator) Name() string {
	return "custom-validator"
}

// Type returns the plugin type
func (v *CustomValidator) Type() plugin.PluginType {
	return plugin.TypeValidator
}

// Version returns the plugin version
func (v *CustomValidator) Version() string {
	return "1.0.0"
}

// Init initializes the plugin
func (v *CustomValidator) Init(config map[string]interface{}) error {
	if fields, ok := config["required_fields"].([]string); ok {
		v.requiredFields = fields
	}
	return nil
}

// Validate performs custom validation
func (v *CustomValidator) Validate(ctx context.Context, data interface{}) (*plugin.ValidationResult, error) {
	result := &plugin.ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check required fields
	if dataMap, ok := data.(map[string]interface{}); ok {
		for _, field := range v.requiredFields {
			if _, exists := dataMap[field]; !exists {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("missing required field: %s", field))
			}
		}
	}

	return result, nil
}

func main() {
	fmt.Println("=== Plugin Example ===")

	// Create plugin manager
	manager := plugin.NewManager()

	// Register plugins
	analyzer := NewCustomAnalyzer()
	transformer := NewCustomTransformer()
	validator := NewCustomValidator()

	// Initialize plugins
	analyzer.Init(map[string]interface{}{"threshold": 0.85})
	transformer.Init(map[string]interface{}{"prefix": "my_"})
	validator.Init(map[string]interface{}{
		"required_fields": []string{"tls_version", "extensions"},
	})

	// Register with manager
	manager.Registry().Register(analyzer)
	manager.Registry().Register(transformer)
	manager.Registry().Register(validator)

	fmt.Println("Registered Plugins:")
	for _, info := range manager.Registry().List() {
		fmt.Printf("  - %s (type: %s, version: %s, enabled: %v)\n",
			info.Name, info.Type, info.Version, info.Enabled)
	}

	// Execute analyzers
	fmt.Println("\nExecuting Analyzers...")
	ctx := context.Background()
	results, err := manager.ExecuteAnalyzers(ctx, "test-data")
	if err != nil {
		fmt.Printf("Analyzer error: %v\n", err)
	} else {
		for _, result := range results {
			fmt.Printf("  Score: %.2f, Confidence: %.2f\n", result.Score, result.Confidence)
			fmt.Printf("  Labels: %v\n", result.Labels)
		}
	}

	// Execute transformers
	fmt.Println("\nExecuting Transformers...")
	transformed, err := manager.ExecuteTransformers(ctx, "hello")
	if err != nil {
		fmt.Printf("Transformer error: %v\n", err)
	} else {
		fmt.Printf("  Result: %v\n", transformed)
	}

	// Execute validators
	fmt.Println("\nExecuting Validators...")
	validationData := map[string]interface{}{
		"tls_version": 0x0303,
		"extensions":  []int{0, 5, 10},
	}
	valResult, err := manager.ExecuteValidators(ctx, validationData)
	if err != nil {
		fmt.Printf("Validator error: %v\n", err)
	} else {
		fmt.Printf("  Valid: %v\n", valResult.Valid)
		if len(valResult.Errors) > 0 {
			fmt.Printf("  Errors: %v\n", valResult.Errors)
		}
		if len(valResult.Warnings) > 0 {
			fmt.Printf("  Warnings: %v\n", valResult.Warnings)
		}
	}

	// Test with invalid data
	fmt.Println("\nTesting Validation with Invalid Data...")
	invalidData := map[string]interface{}{
		"tls_version": 0x0303,
		// Missing "extensions" field
	}
	valResult, _ = manager.ExecuteValidators(ctx, invalidData)
	fmt.Printf("  Valid: %v\n", valResult.Valid)
	fmt.Printf("  Errors: %v\n", valResult.Errors)

	// Disable a plugin
	fmt.Println("\nDisabling transformer plugin...")
	manager.Registry().Disable("custom-transformer")
	fmt.Printf("  custom-transformer enabled: %v\n",
		manager.Registry().IsEnabled("custom-transformer"))

	fmt.Println("\nPlugin example completed!")
}
