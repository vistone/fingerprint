package core

import (
	"log"
	"log/slog"
	"os"
)

// ============================================================================
// StdLoggerAdapter - 标准库 log 适配器
// ============================================================================

// StdLoggerAdapter 适配标准库 log.Logger
type StdLoggerAdapter struct {
	logger *log.Logger
}

// NewStdLoggerAdapter 创建标准库日志适配器
func NewStdLoggerAdapter(logger *log.Logger) Logger {
	if logger == nil {
		logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	return &StdLoggerAdapter{logger: logger}
}

func (a *StdLoggerAdapter) Debug(msg string, keysAndValues ...any) {
	a.logger.Printf("[DEBUG] %s %v", msg, keysAndValues)
}

func (a *StdLoggerAdapter) Info(msg string, keysAndValues ...any) {
	a.logger.Printf("[INFO] %s %v", msg, keysAndValues)
}

func (a *StdLoggerAdapter) Warn(msg string, keysAndValues ...any) {
	a.logger.Printf("[WARN] %s %v", msg, keysAndValues)
}

func (a *StdLoggerAdapter) Error(msg string, keysAndValues ...any) {
	a.logger.Printf("[ERROR] %s %v", msg, keysAndValues)
}

// ============================================================================
// SlogAdapter - Go 1.21+ slog 适配器
// ============================================================================

// SlogAdapter 适配 slog.Logger
type SlogAdapter struct {
	logger *slog.Logger
}

// NewSlogAdapter 创建 slog 适配器
func NewSlogAdapter(logger *slog.Logger) Logger {
	if logger == nil {
		logger = slog.Default()
	}
	return &SlogAdapter{logger: logger}
}

func (a *SlogAdapter) Debug(msg string, keysAndValues ...any) {
	a.logger.Debug(msg, keysAndValues...)
}

func (a *SlogAdapter) Info(msg string, keysAndValues ...any) {
	a.logger.Info(msg, keysAndValues...)
}

func (a *SlogAdapter) Warn(msg string, keysAndValues ...any) {
	a.logger.Warn(msg, keysAndValues...)
}

func (a *SlogAdapter) Error(msg string, keysAndValues ...any) {
	a.logger.Error(msg, keysAndValues...)
}

// ============================================================================
// ZapLoggerAdapter - Uber Zap 适配器（条件编译）
// ============================================================================

// ZapLogger 接口定义（避免直接依赖 zap）
type ZapLogger interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
}

// ZapAdapter 适配 zap.Logger
type ZapAdapter struct {
	logger ZapLogger
}

// NewZapAdapter 创建 zap 适配器
func NewZapAdapter(logger ZapLogger) Logger {
	return &ZapAdapter{logger: logger}
}

func (a *ZapAdapter) Debug(msg string, keysAndValues ...any) {
	if a.logger != nil {
		a.logger.Debug(msg, keysAndValues...)
	}
}

func (a *ZapAdapter) Info(msg string, keysAndValues ...any) {
	if a.logger != nil {
		a.logger.Info(msg, keysAndValues...)
	}
}

func (a *ZapAdapter) Warn(msg string, keysAndValues ...any) {
	if a.logger != nil {
		a.logger.Warn(msg, keysAndValues...)
	}
}

func (a *ZapAdapter) Error(msg string, keysAndValues ...any) {
	if a.logger != nil {
		a.logger.Error(msg, keysAndValues...)
	}
}

// ============================================================================
// LogrusLoggerAdapter - Logrus 适配器（条件编译）
// ============================================================================

// LogrusLogger 接口定义（避免直接依赖 logrus）
type LogrusLogger interface {
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	WithFields(fields map[string]any) LogrusLogger
}

// LogrusAdapter 适配 logrus.Logger
type LogrusAdapter struct {
	logger LogrusLogger
}

// NewLogrusAdapter 创建 logrus 适配器
func NewLogrusAdapter(logger LogrusLogger) Logger {
	return &LogrusAdapter{logger: logger}
}

func (a *LogrusAdapter) Debug(msg string, keysAndValues ...any) {
	if a.logger != nil {
		fields := keyValuesToMap(keysAndValues)
		a.logger.WithFields(fields).Debug(msg)
	}
}

func (a *LogrusAdapter) Info(msg string, keysAndValues ...any) {
	if a.logger != nil {
		fields := keyValuesToMap(keysAndValues)
		a.logger.WithFields(fields).Info(msg)
	}
}

func (a *LogrusAdapter) Warn(msg string, keysAndValues ...any) {
	if a.logger != nil {
		fields := keyValuesToMap(keysAndValues)
		a.logger.WithFields(fields).Warn(msg)
	}
}

func (a *LogrusAdapter) Error(msg string, keysAndValues ...any) {
	if a.logger != nil {
		fields := keyValuesToMap(keysAndValues)
		a.logger.WithFields(fields).Error(msg)
	}
}

// keyValuesToMap 将键值对转换为 map
func keyValuesToMap(keysAndValues []any) map[string]any {
	fields := make(map[string]any)
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		if key, ok := keysAndValues[i].(string); ok {
			fields[key] = keysAndValues[i+1]
		}
	}
	return fields
}

// ============================================================================
// 工具函数
// ============================================================================

// NewDefaultLogger 创建默认的 slog 日志器
func NewDefaultLogger(level string) Logger {
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	return NewSlogAdapter(logger)
}
