package core

// Logger 标准日志接口
// 设计原则：最小接口，满足大多数日志库适配
type Logger interface {
	// Debug 记录调试信息
	Debug(msg string, keysAndValues ...any)
	// Info 记录一般信息
	Info(msg string, keysAndValues ...any)
	// Warn 记录警告信息
	Warn(msg string, keysAndValues ...any)
	// Error 记录错误信息
	Error(msg string, keysAndValues ...any)
}

// NoOpLogger 空日志实现，用于默认值或测试
type NoOpLogger struct{}

func (l NoOpLogger) Debug(msg string, keysAndValues ...any) {}
func (l NoOpLogger) Info(msg string, keysAndValues ...any)  {}
func (l NoOpLogger) Warn(msg string, keysAndValues ...any)  {}
func (l NoOpLogger) Error(msg string, keysAndValues ...any) {}

// Ensure NoOpLogger implements Logger
var _ Logger = NoOpLogger{}
