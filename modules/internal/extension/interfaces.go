package extension

import "context"

// translated comment
// translated comment
type Parser interface {
	// translated comment
	// translated comment
	// translated comment
	// translated comment
	Parse(data []byte, parentContext context.Context) (ExtensionData, error)

	// translated comment
	GetType() ExtensionType

	// translated comment
	GetVersion() string
}

// translated comment
// translated comment
type Analyzer interface {
	// translated comment
	// translated comment
	// translated comment
	// translated comment
	Analyze(data ExtensionData, config map[string]interface{}) (AnalysisResult, error)

	// translated comment
	GetType() ExtensionType

	// translated comment
	GetVersion() string

	// translated comment
	SupportsConfig() []string
}

// translated comment
// translated comment
type Handler interface {
	// translated comment
	// translated comment
	// translated comment
	Handle(event *ExtensionEvent) (*EventResult, error)

	// translated comment
	GetType() ExtensionType

	// translated comment
	// translated comment
	GetPriority() int

	// translated comment
	GetName() string
}

// translated comment
// translated comment
type Plugin interface {
	// translated comment
	GetInfo() *PluginInfo

	// translated comment
	// translated comment
	// translated comment
	Init(config map[string]interface{}) error

	// translated comment
	// translated comment
	Register() error

	// translated comment
	Unload() error

	// translated comment
	Validate() error

	// translated comment
	GetDependencies() []string

	// translated comment
	GetVersion() string
}

// translated comment
type ExtensionEvent struct {
	// translated comment
	Type string // "parse", "analyze", "transform"

	// translated comment
	ExtensionType ExtensionType

	// translated comment
	Data interface{}

	// translated comment
	Context context.Context

	// translated comment
	Metadata map[string]interface{}

	// translated comment
	Timestamp int64
}

// translated comment
type EventResult struct {
	// translated comment
	Success bool

	// translated comment
	Error string

	// translated comment
	Result interface{}

	// translated comment
	ContinueProcessing bool

	// translated comment
	HandlerName string
}

// translated comment
type PluginInfo struct {
	// translated comment
	ID string

	// translated comment
	Name string

	// translated comment
	Description string

	// translated comment
	Version string

	// translated comment
	Author string

	// translated comment
	Contact string

	// License
	License string

	// translated comment
	Homepage string

	// translated comment
	MinSDKVersion string

	// translated comment
	MaxSDKVersion string
}

// translated comment
// translated comment
type Transform interface {
	// translated comment
	// translated comment
	// translated comment
	Transform(input interface{}) (interface{}, error)

	// translated comment
	GetName() string

	// translated comment
	GetInputType() string

	// translated comment
	GetOutputType() string
}

// translated comment
// translated comment
type Validator interface {
	// translated comment
	// translated comment
	// translated comment
	Validate(value interface{}) (bool, error)

	// translated comment
	GetName() string
}

// ComparerComparable protocol for extension data comparison
// translated comment
type Comparer interface {
	// translated comment
	// translated comment
	// translated comment
	Compare(data1, data2 interface{}) (float64, []string, error)

	// translated comment
	GetName() string
}
