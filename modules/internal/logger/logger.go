package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// translated comment
type LogLevel int

const (
	// translated comment
	DebugLevel LogLevel = iota
	// translated comment
	InfoLevel
	// translated comment
	WarnLevel
	// translated comment
	ErrorLevel
	// translated comment
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

// translated comment
type Logger struct {
	mu         sync.RWMutex
	level      LogLevel
	output     io.Writer
	formatter  Formatter
	fields     map[string]interface{}
	callerSkip int
}

// translated comment
type Formatter interface {
	Format(level LogLevel, msg string, fields map[string]interface{}, timestamp time.Time) string
}

// translated comment
type DefaultFormatter struct {
	withTimestamp bool
	withLevel     bool
}

// translated comment
func (f *DefaultFormatter) Format(level LogLevel, msg string, fields map[string]interface{}, timestamp time.Time) string {
	var result string

	if f.withTimestamp {
		result += timestamp.Format("2006-01-02 15:04:05.000") + " "
	}

	if f.withLevel {
		result += fmt.Sprintf("[%s] ", level.String())
	}

	result += msg

	// translated comment
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

// translated comment
type JSONFormatter struct{}

// translated comment
func (f *JSONFormatter) Format(level LogLevel, msg string, fields map[string]interface{}, timestamp time.Time) string {
	// translated comment
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

// translated comment
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

// translated comment
type Option func(*Logger)

// translated comment
func WithLevel(level LogLevel) Option {
	return func(l *Logger) {
		l.level = level
	}
}

// translated comment
func WithOutput(output io.Writer) Option {
	return func(l *Logger) {
		l.output = output
	}
}

// translated comment
func WithFormatter(formatter Formatter) Option {
	return func(l *Logger) {
		l.formatter = formatter
	}
}

// translated comment
func WithFields(fields map[string]interface{}) Option {
	return func(l *Logger) {
		for k, v := range fields {
			l.fields[k] = v
		}
	}
}

// translated comment
func (l *Logger) log(level LogLevel, msg string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	// translated comment
	mergedFields := make(map[string]interface{}, len(l.fields)+len(fields))
	for k, v := range l.fields {
		mergedFields[k] = v
	}
	for k, v := range fields {
		mergedFields[k] = v
	}

	// translated comment
	formatted := l.formatter.Format(level, msg, mergedFields, time.Now())
	fmt.Fprintln(l.output, formatted)

	// translated comment
	if level == FatalLevel {
		os.Exit(1)
	}
}

// translated comment
func (l *Logger) Debug(msg string) {
	l.log(DebugLevel, msg, nil)
}

// translated comment
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DebugLevel, fmt.Sprintf(format, args...), nil)
}

// translated comment
func (l *Logger) Debugw(msg string, keysAndValues ...interface{}) {
	l.log(DebugLevel, msg, kvToMap(keysAndValues...))
}

// translated comment
func (l *Logger) Info(msg string) {
	l.log(InfoLevel, msg, nil)
}

// translated comment
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(InfoLevel, fmt.Sprintf(format, args...), nil)
}

// translated comment
func (l *Logger) Infow(msg string, keysAndValues ...interface{}) {
	l.log(InfoLevel, msg, kvToMap(keysAndValues...))
}

// translated comment
func (l *Logger) Warn(msg string) {
	l.log(WarnLevel, msg, nil)
}

// translated comment
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WarnLevel, fmt.Sprintf(format, args...), nil)
}

// translated comment
func (l *Logger) Warnw(msg string, keysAndValues ...interface{}) {
	l.log(WarnLevel, msg, kvToMap(keysAndValues...))
}

// translated comment
func (l *Logger) Error(msg string) {
	l.log(ErrorLevel, msg, nil)
}

// translated comment
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ErrorLevel, fmt.Sprintf(format, args...), nil)
}

// translated comment
func (l *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	l.log(ErrorLevel, msg, kvToMap(keysAndValues...))
}

// translated comment
func (l *Logger) Fatal(msg string) {
	l.log(FatalLevel, msg, nil)
}

// translated comment
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(FatalLevel, fmt.Sprintf(format, args...), nil)
}

// translated comment
func (l *Logger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.log(FatalLevel, msg, kvToMap(keysAndValues...))
}

// translated comment
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

// translated comment
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// translated comment
func kvToMap(keysAndValues ...interface{}) map[string]interface{} {
	if len(keysAndValues)%2 != 0 {
		// translated comment
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
// translated comment
// ============================================================================

var (
	defaultLogger     *Logger
	defaultLoggerOnce sync.Once
)

// translated comment
func Default() *Logger {
	defaultLoggerOnce.Do(func() {
		defaultLogger = New()
	})
	return defaultLogger
}

// translated comment
func SetDefault(l *Logger) {
	defaultLogger = l
}

// translated comment
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

// translated comment
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

	// translated comment
	log.SetOutput(defaultLogger.output)
}
