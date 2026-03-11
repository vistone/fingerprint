package extension

import "context"

// Parser is the extension parsing interface
// Responsible for parsing raw byte data into structured ExtensionData
type Parser interface {
	// Parse parses extension data
	// data: raw extension data (excluding extension header)
	// parentContext: parent context (e.g. complete ClientHello)
	// Returns: parsed ExtensionData, error
	Parse(data []byte, parentContext context.Context) (ExtensionData, error)

	// GetType returns the extension type this parser handles
	GetType() ExtensionType

	// GetVersion returns the parser version
	GetVersion() string
}

// Analyzer is the extension analysis interface
// Responsible for analyzing ExtensionData and producing AnalysisResult
type Analyzer interface {
	// Analyze analyzes extension data
	// data: parsed ExtensionData
	// config: analysis configuration (optional)
	// Returns: analysis result, error
	Analyze(data ExtensionData, config map[string]interface{}) (AnalysisResult, error)

	// GetType returns the extension type this analyzer handles
	GetType() ExtensionType

	// GetVersion returns the analyzer version
	GetVersion() string

	// SupportsConfig returns the list of supported configuration keys
	SupportsConfig() []string
}

// Handler is the extension handling interface
// Event-driven extension handling for streaming and middleware
type Handler interface {
	// Handle handles an extension event
	// event: extension event
	// Returns: handling result, error
	Handle(event *ExtensionEvent) (*EventResult, error)

	// GetType returns the extension type this handler handles
	GetType() ExtensionType

	// GetPriority returns handler priority (higher priority executes first)
	// 0-100, default 50
	GetPriority() int

	// GetName returns the handler name
	GetName() string
}

// Plugin is the third-party plugin interface
// Allows external developers to extend fingerprint library functionality
type Plugin interface {
	// GetInfo returns plugin information
	GetInfo() *PluginInfo

	// Init initializes the plugin
	// config: plugin configuration
	// Returns: error
	Init(config map[string]interface{}) error

	// Register registers the plugin's extensions
	// This method should call RegisterExtension, RegisterParser, etc.
	Register() error

	// Unload unloads the plugin
	Unload() error

	// Validate validates plugin validity
	Validate() error

	// GetDependencies returns the plugin dependency list
	GetDependencies() []string

	// GetVersion returns the plugin version
	GetVersion() string
}

// ExtensionEvent represents an extension event
type ExtensionEvent struct {
	// Event type
	Type string // "parse", "analyze", "transform"

	// Extension type
	ExtensionType ExtensionType

	// Event data
	Data interface{}

	// Context at the time of the event
	Context context.Context

	// Event metadata
	Metadata map[string]interface{}

	// Timestamp
	Timestamp int64
}

// EventResult holds event handling results
type EventResult struct {
	// Whether handling was successful
	Success bool

	// Error message
	Error string

	// Handling result data
	Result interface{}

	// Whether to continue passing to the next handler
	ContinueProcessing bool

	// Handler name
	HandlerName string
}

// PluginInfo holds plugin information
type PluginInfo struct {
	// Plugin ID (unique identifier)
	ID string

	// Plugin name
	Name string

	// Plugin description
	Description string

	// Plugin version
	Version string

	// Author
	Author string

	// Contact information
	Contact string

	// License
	License string

	// Homepage
	Homepage string

	// Minimum SDK version
	MinSDKVersion string

	// Maximum SDK version
	MaxSDKVersion string
}

// Transform is the data transformation interface
// Used for chained transformations on extensions
type Transform interface {
	// Transform performs the transformation
	// input: input data
	// Returns: transformed data, error
	Transform(input interface{}) (interface{}, error)

	// GetName returns the transform name
	GetName() string

	// GetInputType returns the input type name
	GetInputType() string

	// GetOutputType returns the output type name
	GetOutputType() string
}

// Validator is the data validation interface
// Used for extension data validation
type Validator interface {
	// Validate validates data
	// value: value to validate
	// Returns: whether valid, error
	Validate(value interface{}) (bool, error)

	// GetName returns the validator name
	GetName() string
}

// ComparerComparable protocol for extension data comparison
// Used for comparison and difference detection between extensions
type Comparer interface {
	// Compare compares two data sets
	// data1, data2: data to compare
	// Returns: similarity (0.0-1.0), list of differences
	Compare(data1, data2 interface{}) (float64, []string, error)

	// GetName returns the comparer name
	GetName() string
}
