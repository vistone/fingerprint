// Package web provides the log buffer system for real-time log streaming
package web

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// LogEntry is one captured log record.
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
}

// LogBuffer stores recent logs and supports SSE streaming to subscribers.
type LogBuffer struct {
	mu      sync.RWMutex
	entries []LogEntry
	maxSize int
	// SSE subscribers.
	subscribers map[chan LogEntry]struct{}
	subMu       sync.RWMutex
}

// globalLogBuffer is the process-wide log buffer.
var globalLogBuffer = NewLogBuffer(500)

// NewLogBuffer creates a bounded in-memory log buffer.
func NewLogBuffer(size int) *LogBuffer {
	return &LogBuffer{
		entries:     make([]LogEntry, 0, size),
		maxSize:     size,
		subscribers: make(map[chan LogEntry]struct{}),
	}
}

// InitLogCapture redirects standard log output into the in-memory buffer.
func InitLogCapture() {
	pr, pw, err := os.Pipe()
	if err != nil {
		log.Printf("Warning: failed to capture log output: %v", err)
		return
	}

	// Redirect standard log output.
	log.SetOutput(pw)
	log.SetFlags(0) // Timestamp is attached when buffering log entries.

	// Read redirected logs in background.
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := pr.Read(buf)
			if n > 0 {
				msg := string(buf[:n])
				// Mirror logs to stderr to preserve default behavior.
				os.Stderr.WriteString(msg)
				// Parse log level from message text.
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

// Append adds one entry to the log buffer.
func (lb *LogBuffer) Append(entry LogEntry) {
	lb.mu.Lock()
	lb.entries = append(lb.entries, entry)
	if len(lb.entries) > lb.maxSize {
		lb.entries = lb.entries[len(lb.entries)-lb.maxSize:]
	}
	lb.mu.Unlock()

	// Broadcast to SSE subscribers.
	lb.subMu.RLock()
	for ch := range lb.subscribers {
		select {
		case ch <- entry:
		default:
			// Drop when subscriber is slow to consume.
		}
	}
	lb.subMu.RUnlock()
}

// GetAll returns all buffered logs.
func (lb *LogBuffer) GetAll() []LogEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	result := make([]LogEntry, len(lb.entries))
	copy(result, lb.entries)
	return result
}

// GetFiltered returns logs filtered by level.
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

// Subscribe registers a real-time log stream subscriber.
func (lb *LogBuffer) Subscribe() chan LogEntry {
	ch := make(chan LogEntry, 64)
	lb.subMu.Lock()
	lb.subscribers[ch] = struct{}{}
	lb.subMu.Unlock()
	return ch
}

// Unsubscribe removes and closes a subscriber channel.
func (lb *LogBuffer) Unsubscribe(ch chan LogEntry) {
	lb.subMu.Lock()
	delete(lb.subscribers, ch)
	lb.subMu.Unlock()
	close(ch)
}

// WriteLog appends a log entry and mirrors it to stderr.
func WriteLog(level, source, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	globalLogBuffer.Append(LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Source:    source,
	})
	// Mirror to stderr.
	fmt.Fprintf(os.Stderr, "[%s] %s: %s\n", level, source, msg)
}

// parseLogLevel extracts a coarse log level from message text.
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
