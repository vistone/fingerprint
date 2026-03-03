package logger

import (
	"log"
)

// Logger 简单的日志接口
type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// DefaultLogger 默认日志实现
type DefaultLogger struct{}

// NewDefaultLogger 创建默认日志器
func NewDefaultLogger() Logger {
	return &DefaultLogger{}
}

func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	if len(args) > 0 {
		log.Printf("[INFO] "+msg, args...)
	} else {
		log.Printf("[INFO] %s", msg)
	}
}

func (l *DefaultLogger) Warn(msg string, args ...interface{}) {
	if len(args) > 0 {
		log.Printf("[WARN] "+msg, args...)
	} else {
		log.Printf("[WARN] %s", msg)
	}
}

func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	if len(args) > 0 {
		log.Printf("[ERROR] "+msg, args...)
	} else {
		log.Printf("[ERROR] %s", msg)
	}
}

func (l *DefaultLogger) Debug(msg string, args ...interface{}) {
	if len(args) > 0 {
		log.Printf("[DEBUG] "+msg, args...)
	} else {
		log.Printf("[DEBUG] %s", msg)
	}
}

// Global 全局日志器
var Global Logger = NewDefaultLogger()
