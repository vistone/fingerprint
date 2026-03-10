// Package web provides the log buffer system for real-time log streaming
package web

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// LogEntry 一条日志记录
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
}

// LogBuffer 环形日志缓冲区，捕获 Go log 输出并支持 SSE 推流
type LogBuffer struct {
	mu      sync.RWMutex
	entries []LogEntry
	maxSize int
	// SSE 订阅者
	subscribers map[chan LogEntry]struct{}
	subMu       sync.RWMutex
}

// globalLogBuffer 全局日志缓冲区
var globalLogBuffer = NewLogBuffer(500)

// NewLogBuffer 创建日志缓冲区
func NewLogBuffer(size int) *LogBuffer {
	return &LogBuffer{
		entries:     make([]LogEntry, 0, size),
		maxSize:     size,
		subscribers: make(map[chan LogEntry]struct{}),
	}
}

// InitLogCapture 拦截标准 log 输出，转存到缓冲区
func InitLogCapture() {
	pr, pw, err := os.Pipe()
	if err != nil {
		log.Printf("Warning: failed to capture log output: %v", err)
		return
	}

	// 重定向标准 log 输出
	log.SetOutput(pw)
	log.SetFlags(0) // 我们自己管理时间戳

	// 后台读取 pipe 数据
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := pr.Read(buf)
			if n > 0 {
				msg := string(buf[:n])
				// 同时写到 stderr 保留原有行为
				os.Stderr.WriteString(msg)
				// 解析级别
				level := parseLogLevel(msg)
				globalLogBuffer.Append(LogEntry{
					Timestamp: time.Now(),
					Level:     level,
					Message:   msg,
					Source:    "system",
				})
			}
			if err != nil {
				break
			}
		}
	}()
}

// Append 添加一条日志到缓冲区
func (lb *LogBuffer) Append(entry LogEntry) {
	lb.mu.Lock()
	lb.entries = append(lb.entries, entry)
	if len(lb.entries) > lb.maxSize {
		lb.entries = lb.entries[len(lb.entries)-lb.maxSize:]
	}
	lb.mu.Unlock()

	// 通知所有 SSE 订阅者
	lb.subMu.RLock()
	for ch := range lb.subscribers {
		select {
		case ch <- entry:
		default:
			// 订阅者跟不上，丢弃
		}
	}
	lb.subMu.RUnlock()
}

// GetAll 获取缓冲区中所有日志
func (lb *LogBuffer) GetAll() []LogEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	result := make([]LogEntry, len(lb.entries))
	copy(result, lb.entries)
	return result
}

// GetFiltered 按级别过滤日志
func (lb *LogBuffer) GetFiltered(level string) []LogEntry {
	if level == "" || level == "all" {
		return lb.GetAll()
	}
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	var result []LogEntry
	for _, e := range lb.entries {
		if e.Level == level {
			result = append(result, e)
		}
	}
	return result
}

// Subscribe 订阅实时日志流，返回 channel。调用 Unsubscribe 释放。
func (lb *LogBuffer) Subscribe() chan LogEntry {
	ch := make(chan LogEntry, 64)
	lb.subMu.Lock()
	lb.subscribers[ch] = struct{}{}
	lb.subMu.Unlock()
	return ch
}

// Unsubscribe 取消订阅
func (lb *LogBuffer) Unsubscribe(ch chan LogEntry) {
	lb.subMu.Lock()
	delete(lb.subscribers, ch)
	lb.subMu.Unlock()
	close(ch)
}

// WriteLog 对外公开的日志写入接口（其他模块可直接调用）
func WriteLog(level, source, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	globalLogBuffer.Append(LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Source:    source,
	})
	// 同时输出到 stderr
	fmt.Fprintf(os.Stderr, "[%s] %s: %s\n", level, source, msg)
}

// parseLogLevel 从日志消息中解析级别
func parseLogLevel(msg string) string {
	switch {
	case len(msg) > 6 && msg[:6] == "ERROR " || len(msg) > 7 && msg[:7] == "[ERROR]":
		return "ERROR"
	case len(msg) > 5 && msg[:5] == "WARN " || len(msg) > 6 && msg[:6] == "[WARN]":
		return "WARN"
	case len(msg) > 6 && msg[:6] == "DEBUG " || len(msg) > 7 && msg[:7] == "[DEBUG]":
		return "DEBUG"
	default:
		return "INFO"
	}
}
