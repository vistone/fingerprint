package extension

// Component is the base interface for all governable components
//
// Follows the single responsibility principle, each component defines only minimal necessary behavior
// Concrete components can extend functionality through composition
//
// Usage example:
//
//	var comp Component = myParser
//	info := comp.GetInfo()
type Component interface {
	// GetInfo returns component information
	GetInfo() ComponentInfo
}

// ComponentInfo holds component information
type ComponentInfo struct {
	// Component name
	Name string

	// Component version
	Version string

	// Component description
	Description string

	// Component author
	Author string
}

// Auditable is the interface for auditable components
//
// Follows Go's implicit interface design, components need not explicitly declare implementation
type Auditable interface {
	// RecordEvent records an event
	// Returns: error
	RecordEvent(eventType, severity, message string, details map[string]interface{}) error
}

// Closeable is the interface for components that support resource cleanup
//
// Any component that needs to release resources can implement this interface
// Consistent with io.Closer for ease of use
type Closeable interface {
	// Close closes the component and releases resources
	Close() error
}

// Initializable is the interface for components that support initialization
//
// Used for lazy initialization and complex setup processes
type Initializable interface {
	// Initialize initializes the component
	// config: initialization configuration
	// Returns: error
	Initialize(config map[string]interface{}) error

	// IsInitialized checks whether the component is initialized
	IsInitialized() bool
}

// Identifiable is the interface for identifiable components
//
// Any object with a unique identifier should implement this interface
type Identifiable interface {
	// GetID returns the unique identifier
	GetID() string
}
