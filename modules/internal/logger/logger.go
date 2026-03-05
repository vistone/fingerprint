package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	// DebugLevel 调试级别
	DebugLevel LogLevel = iota
	// InfoLevel 信息级别
	InfoLevel
	// WarnLevel 警告级别
	WarnLevel
	// ErrorLevel 错误级别
	ErrorLevel
	// FatalLevel 致命级别
	FatalLevel
)

func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger 日志记录器
type Logger struct {
	mu         sync.RWMutex
	level      LogLevel
	output     io.Writer
	formatter  Formatter
	fields     map[string]interface{}
	callerSkip int
}

// Formatter 日志格式化接口
type Formatter interface {
	Format(level LogLevel, msg string, fields map[string]interface{}, timestamp time.Time) string
}

// DefaultFormatter 默认格式化器
type DefaultFormatter struct {
	withTimestamp bool
	withLevel     bool
}

// Format 格式化日志
func (f *DefaultFormatter) Format(level LogLevel, msg string, fields map[string]interface{}, timestamp time.Time) string {
	var result string

	if f.withTimestamp {
		result += timestamp.Format("2006-01-02 15:04:05.000") + " "
	}

	if f.withLevel {
		result += fmt.Sprintf("[%s] ", level.String())
	}

	result += msg

	// 添加字段
	if len(fields) > 0 {
		result += " {"
		first := true
		for k, v := range fields {
			if !first {
				result += ", "
			}
			result += fmt.Sprintf("%s=%v", k, v)
			first = false
		}
		result += "}"
	}

	return result
}

// JSONFormatter JSON 格式化器
type JSONFormatter struct{}

// Format 格式化日志为 JSON
func (f *JSONFormatter) Format(level LogLevel, msg string, fields map[string]interface{}, timestamp time.Time) string {
	// 简化实现，实际应该使用 json.Marshal
	result := fmt.Sprintf(`{"timestamp":"%s","level":"%s","message":"%s"`,
		timestamp.Format(time.RFC3339Nano),
		level.String(),
		msg)

	for k, v := range fields {
		result += fmt.Sprintf(",\"%s\":%v", k, v)
	}

	result += "}"
	return result
}

// New 创建新的日志记录器
func New(opts ...Option) *Logger {
	l := &Logger{
		level:      InfoLevel,
		output:     os.Stdout,
		formatter:  &DefaultFormatter{withTimestamp: true, withLevel: true},
		fields:     make(map[string]interface{}),
		callerSkip: 2,
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// Option 日志选项
type Option func(*Logger)

// WithLevel 设置日志级别
func WithLevel(level LogLevel) Option {
	return func(l *Logger) {
		l.level = level
	}
}

// WithOutput 设置输出
func WithOutput(output io.Writer) Option {
	return func(l *Logger) {
		l.output = output
	}
}

// WithFormatter 设置格式化器
func WithFormatter(formatter Formatter) Option {
	return func(l *Logger) {
		l.formatter = formatter
	}
}

// WithFields 设置默认字段
func WithFields(fields map[string]interface{}) Option {
	return func(l *Logger) {
		for k, v := range fields {
			l.fields[k] = v
		}
	}
}

// log 内部日志方法
func (l *Logger) log(level LogLevel, msg string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	// 合并字段
	mergedFields := make(map[string]interface{}, len(l.fields)+len(fields))
	for k, v := range l.fields {
		mergedFields[k] = v
	}
	for k, v := range fields {
		mergedFields[k] = v
	}

	// 格式化并输出
	formatted := l.formatter.Format(level, msg, mergedFields, time.Now())
	fmt.Fprintln(l.output, formatted)

	// Fatal 级别退出程序
	if level == FatalLevel {
		os.Exit(1)
	}
}

// Debug 调试日志
func (l *Logger) Debug(msg string) {
	l.log(DebugLevel, msg, nil)
}

// Debugf 格式化调试日志
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DebugLevel, fmt.Sprintf(format, args...), nil)
}

// Debugw 带字段的调试日志
func (l *Logger) Debugw(msg string, keysAndValues ...interface{}) {
	l.log(DebugLevel, msg, kvToMap(keysAndValues...))
}

// Info 信息日志
func (l *Logger) Info(msg string) {
	l.log(InfoLevel, msg, nil)
}

// Infof 格式化信息日志
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(InfoLevel, fmt.Sprintf(format, args...), nil)
}

// Infow 带字段的信息日志
func (l *Logger) Infow(msg string, keysAndValues ...interface{}) {
	l.log(InfoLevel, msg, kvToMap(keysAndValues...))
}

// Warn 警告日志
func (l *Logger) Warn(msg string) {
	l.log(WarnLevel, msg, nil)
}

// Warnf 格式化警告日志
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WarnLevel, fmt.Sprintf(format, args...), nil)
}

// Warnw 带字段的警告日志
func (l *Logger) Warnw(msg string, keysAndValues ...interface{}) {
	l.log(WarnLevel, msg, kvToMap(keysAndValues...))
}

// Error 错误日志
func (l *Logger) Error(msg string) {
	l.log(ErrorLevel, msg, nil)
}

// Errorf 格式化错误日志
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ErrorLevel, fmt.Sprintf(format, args...), nil)
}

// Errorw 带字段的错误日志
func (l *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	l.log(ErrorLevel, msg, kvToMap(keysAndValues...))
}

// Fatal 致命日志
func (l *Logger) Fatal(msg string) {
	l.log(FatalLevel, msg, nil)
}

// Fatalf 格式化致命日志
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(FatalLevel, fmt.Sprintf(format, args...), nil)
}

// Fatalw 带字段的致命日志
func (l *Logger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.log(FatalLevel, msg, kvToMap(keysAndValues...))
}

// With 创建带字段的子日志记录器
func (l *Logger) With(fields map[string]interface{}) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	newFields := make(map[string]interface{}, len(l.fields)+len(fields))
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &Logger{
		level:      l.level,
		output:     l.output,
		formatter:  l.formatter,
		fields:     newFields,
		callerSkip: l.callerSkip,
	}
}

// SetLevel 动态设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// kvToMap 将键值对转换为 map
func kvToMap(keysAndValues ...interface{}) map[string]interface{} {
	if len(keysAndValues)%2 != 0 {
		// 奇数个参数，忽略最后一个
		keysAndValues = keysAndValues[:len(keysAndValues)-1]
	}

	m := make(map[string]interface{}, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}
		m[key] = keysAndValues[i+1]
	}
	return m
}

// ============================================================================
// 全局日志记录器
// ============================================================================

var (
	defaultLogger     *Logger
	defaultLoggerOnce sync.Once
)

// Default 获取默认日志记录器
func Default() *Logger {
	defaultLoggerOnce.Do(func() {
		defaultLogger = New()
	})
	return defaultLogger
}

// SetDefault 设置默认日志记录器
func SetDefault(l *Logger) {
	defaultLogger = l
}

// 全局便捷函数
func Debug(msg string)                         { Default().Debug(msg) }
func Debugf(format string, args ...interface{}) { Default().Debugf(format, args...) }
func Debugw(msg string, keysAndValues ...interface{}) {
	Default().Debugw(msg, keysAndValues...)
}
func Info(msg string)                          { Default().Info(msg) }
func Infof(format string, args ...interface{})  { Default().Infof(format, args...) }
func Infow(msg string, keysAndValues ...interface{}) {
	Default().Infow(msg, keysAndValues...)
}
func Warn(msg string)                          { Default().Warn(msg) }
func Warnf(format string, args ...interface{})  { Default().Warnf(format, args...) }
func Warnw(msg string, keysAndValues ...interface{}) {
	Default().Warnw(msg, keysAndValues...)
}
func Error(msg string)                         { Default().Error(msg) }
func Errorf(format string, args ...interface{}) { Default().Errorf(format, args...) }
func Errorw(msg string, keysAndValues ...interface{}) {
	Default().Errorw(msg, keysAndValues...)
}
func Fatal(msg string)                         { Default().Fatal(msg) }
func Fatalf(format string, args ...interface{}) { Default().Fatalf(format, args...) }
func Fatalw(msg string, keysAndValues ...interface{}) {
	Default().Fatalw(msg, keysAndValues...)
}

// InitFromEnv 从环境变量初始化日志
func InitFromEnv() {
	level := InfoLevel
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		switch lvl {
		case "debug", "DEBUG":
			level = DebugLevel
		case "info", "INFO":
			level = InfoLevel
		case "warn", "WARN", "warning", "WARNING":
			level = WarnLevel
		case "error", "ERROR":
			level = ErrorLevel
		case "fatal", "FATAL":
			level = FatalLevel
		}
	}

	var formatter Formatter = &DefaultFormatter{withTimestamp: true, withLevel: true}
	if fmt := os.Getenv("LOG_FORMAT"); fmt == "json" {
		formatter = &JSONFormatter{}
	}

	defaultLogger = New(
		WithLevel(level),
		WithFormatter(formatter),
	)

	// 同时设置标准库日志
	log.SetOutput(defaultLogger.output)
}
