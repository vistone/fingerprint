package extension

// translated comment
//
// translated comment
// translated comment
//
// translated comment
//
//	var comp Component = myParser
//	info := comp.GetInfo()
type Component interface {
	// translated comment
	GetInfo() ComponentInfo
}

// translated comment
type ComponentInfo struct {
	// translated comment
	Name string

	// translated comment
	Version string

	// translated comment
	Description string

	// translated comment
	Author string
}

// translated comment
//
// translated comment
type Auditable interface {
	// translated comment
	// translated comment
	RecordEvent(eventType, severity, message string, details map[string]interface{}) error
}

// translated comment
//
// translated comment
// translated comment
type Closeable interface {
	// translated comment
	Close() error
}

// translated comment
//
// translated comment
type Initializable interface {
	// translated comment
	// translated comment
	// translated comment
	Initialize(config map[string]interface{}) error

	// translated comment
	IsInitialized() bool
}

// translated comment
//
// translated comment
type Identifiable interface {
	// translated comment
	GetID() string
}
