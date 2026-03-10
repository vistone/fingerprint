// Package web provides the log buffer system for real-time log streaming
package web

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// translated comment
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
}

// translated comment
type LogBuffer struct {
	mu      sync.RWMutex
	entries []LogEntry
	maxSize int
	// translated comment
	subscribers map[chan LogEntry]struct{}
	subMu       sync.RWMutex
}

// translated comment
var globalLogBuffer = NewLogBuffer(500)

// translated comment
func NewLogBuffer(size int) *LogBuffer {
	return &LogBuffer{
		entries:     make([]LogEntry, 0, size),
		maxSize:     size,
		subscribers: make(map[chan LogEntry]struct{}),
	}
}

// translated comment
func InitLogCapture() {
	pr, pw, err := os.Pipe()
	if err != nil {
		log.Printf("Warning: failed to capture log output: %v", err)
		return
	}

	// translated comment
	log.SetOutput(pw)
	log.SetFlags(0) // translated comment

	// translated comment
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := pr.Read(buf)
			if n > 0 {
				msg := string(buf[:n])
				// translated comment
				os.Stderr.WriteString(msg)
				// translated comment
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

// translated comment
func (lb *LogBuffer) Append(entry LogEntry) {
	lb.mu.Lock()
	lb.entries = append(lb.entries, entry)
	if len(lb.entries) > lb.maxSize {
		lb.entries = lb.entries[len(lb.entries)-lb.maxSize:]
	}
	lb.mu.Unlock()

	// translated comment
	lb.subMu.RLock()
	for ch := range lb.subscribers {
		select {
		case ch <- entry:
		default:
			// translated comment
		}
	}
	lb.subMu.RUnlock()
}

// translated comment
func (lb *LogBuffer) GetAll() []LogEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	result := make([]LogEntry, len(lb.entries))
	copy(result, lb.entries)
	return result
}

// translated comment
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

// translated comment
func (lb *LogBuffer) Subscribe() chan LogEntry {
	ch := make(chan LogEntry, 64)
	lb.subMu.Lock()
	lb.subscribers[ch] = struct{}{}
	lb.subMu.Unlock()
	return ch
}

// translated comment
func (lb *LogBuffer) Unsubscribe(ch chan LogEntry) {
	lb.subMu.Lock()
	delete(lb.subscribers, ch)
	lb.subMu.Unlock()
	close(ch)
}

// translated comment
func WriteLog(level, source, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	globalLogBuffer.Append(LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Source:    source,
	})
	// translated comment
	fmt.Fprintf(os.Stderr, "[%s] %s: %s\n", level, source, msg)
}

// translated comment
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
