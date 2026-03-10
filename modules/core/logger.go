package core

// Logger standard logging interface
// design principle: minimal interface, adapts to most logging libraries
type Logger interface {
	// Debug logdebuginfo
	Debug(msg string, keysAndValues ...any)
	// Info logs general information
	Info(msg string, keysAndValues ...any)
	// Warn logwarninginfo
	Warn(msg string, keysAndValues ...any)
	// Error logerrorinfo
	Error(msg string, keysAndValues ...any)
}

// NoOpLogger empty log implementation, used for default values or testing
type NoOpLogger struct{}

func (l NoOpLogger) Debug(msg string, keysAndValues ...any) {}
func (l NoOpLogger) Info(msg string, keysAndValues ...any)  {}
func (l NoOpLogger) Warn(msg string, keysAndValues ...any)  {}
func (l NoOpLogger) Error(msg string, keysAndValues ...any) {}

// Ensure NoOpLogger implements Logger
var _ Logger = NoOpLogger{}
