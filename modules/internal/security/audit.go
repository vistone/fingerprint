package security

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// AuditLevel represents the severity of an audit event
type AuditLevel int

const (
	// Debug level for debugging
	Debug AuditLevel = iota
	// Info level for informational events
	Info
	// Warning level for warning events
	Warning
	// Error level for error events
	Error
	// Critical level for critical security events
	Critical
)

// String returns the string representation
func (l AuditLevel) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warning:
		return "WARNING"
	case Error:
		return "ERROR"
	case Critical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// AuditEvent represents a security audit event
type AuditEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Category  string                 `json:"category"`
	Action    string                 `json:"action"`
	UserID    string                 `json:"user_id,omitempty"`
	ClientIP  string                 `json:"client_ip,omitempty"`
	Resource  string                 `json:"resource"`
	Result    string                 `json:"result"`
	Details   map[string]interface{} `json:"details,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
}

// AuditLogger handles security audit logging
type AuditLogger struct {
	mu        sync.RWMutex
	writer    AuditWriter
	minLevel  AuditLevel
	enabled   bool
	sanitizer *Sanitizer
	handler   *SensitiveDataHandler
}

// AuditWriter defines the interface for audit log writers
type AuditWriter interface {
	Write(event *AuditEvent) error
	Close() error
}

// FileAuditWriter writes audit logs to a file
type FileAuditWriter struct {
	file    *os.File
	mu      sync.Mutex
	encoder *json.Encoder
}

// NewFileAuditWriter creates a new file audit writer
func NewFileAuditWriter(filepath string) (*FileAuditWriter, error) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}

	return &FileAuditWriter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

// Write writes an audit event
func (w *FileAuditWriter) Write(event *AuditEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.encoder.Encode(event)
}

// Close closes the writer
func (w *FileAuditWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.file.Close()
}

// ConsoleAuditWriter writes audit logs to console
type ConsoleAuditWriter struct {
	encoder *json.Encoder
}

// NewConsoleAuditWriter creates a new console audit writer
func NewConsoleAuditWriter() *ConsoleAuditWriter {
	return &ConsoleAuditWriter{
		encoder: json.NewEncoder(os.Stdout),
	}
}

// Write writes an audit event
func (w *ConsoleAuditWriter) Write(event *AuditEvent) error {
	return w.encoder.Encode(event)
}

// Close closes the writer (no-op for console)
func (w *ConsoleAuditWriter) Close() error {
	return nil
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(writer AuditWriter) *AuditLogger {
	return &AuditLogger{
		writer:    writer,
		minLevel:  Info,
		enabled:   true,
		sanitizer: NewSanitizer(),
		handler:   NewSensitiveDataHandler(),
	}
}

// SetMinLevel sets the minimum audit level
func (l *AuditLogger) SetMinLevel(level AuditLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

// Enable enables audit logging
func (l *AuditLogger) Enable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = true
}

// Disable disables audit logging
func (l *AuditLogger) Disable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = false
}

// IsEnabled returns whether audit logging is enabled
func (l *AuditLogger) IsEnabled() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.enabled
}

// AuditLogEntry groups the identifying fields of an audit event.
type AuditLogEntry struct {
	Category string
	Action   string
	Resource string
	Result   string
}

// Log logs an audit event
func (l *AuditLogger) Log(level AuditLevel, entry AuditLogEntry, details map[string]interface{}) {
	l.mu.RLock()
	if !l.enabled || level < l.minLevel {
		l.mu.RUnlock()
		return
	}
	l.mu.RUnlock()

	// Sanitize details
	sanitizedDetails := make(map[string]interface{})
	for k, v := range details {
		if str, ok := v.(string); ok {
			sanitizedDetails[k] = l.handler.SanitizeForLogging(str)
		} else {
			sanitizedDetails[k] = v
		}
	}

	event := &AuditEvent{
		Timestamp: time.Now().UTC(),
		Level:     level.String(),
		Category:  entry.Category,
		Action:    entry.Action,
		Resource:  entry.Resource,
		Result:    entry.Result,
		Details:   sanitizedDetails,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.writer != nil {
		l.writer.Write(event)
	}
}

// LogWithContext logs an audit event with context
func (l *AuditLogger) LogWithContext(ctx context.Context, level AuditLevel, entry AuditLogEntry, details map[string]interface{}) {
	// Extract context values if available
	if details == nil {
		details = make(map[string]interface{})
	}

	if userID, ok := ctx.Value("user_id").(string); ok {
		details["user_id"] = userID
	}

	if clientIP, ok := ctx.Value("client_ip").(string); ok {
		details["client_ip"] = clientIP
	}

	if requestID, ok := ctx.Value("request_id").(string); ok {
		details["request_id"] = requestID
	}

	l.Log(level, entry, details)
}

// Convenience methods for different levels

// Debug logs a debug event
func (l *AuditLogger) Debug(entry AuditLogEntry, details map[string]interface{}) {
	l.Log(Debug, entry, details)
}

// Info logs an info event
func (l *AuditLogger) Info(entry AuditLogEntry, details map[string]interface{}) {
	l.Log(Info, entry, details)
}

// Warning logs a warning event
func (l *AuditLogger) Warning(entry AuditLogEntry, details map[string]interface{}) {
	l.Log(Warning, entry, details)
}

// Error logs an error event
func (l *AuditLogger) Error(entry AuditLogEntry, details map[string]interface{}) {
	l.Log(Error, entry, details)
}

// Critical logs a critical event
func (l *AuditLogger) Critical(entry AuditLogEntry, details map[string]interface{}) {
	l.Log(Critical, entry, details)
}

// LogAuthentication logs authentication events
func (l *AuditLogger) LogAuthentication(userID, clientIP, method, result string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["user_id"] = userID
	details["client_ip"] = clientIP
	details["method"] = method

	level := Info
	if result != "success" {
		level = Warning
	}

	l.Log(level, AuditLogEntry{Category: "authentication", Action: "login", Resource: "auth_system", Result: result}, details)
}

// LogAuthorization logs authorization events
func (l *AuditLogger) LogAuthorization(userID, resource, action, result string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["user_id"] = userID

	level := Info
	if result != "allowed" {
		level = Warning
	}

	l.Log(level, AuditLogEntry{Category: "authorization", Action: action, Resource: resource, Result: result}, details)
}

// LogDataAccess logs data access events
func (l *AuditLogger) LogDataAccess(userID, resource, action string, sensitive bool, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["user_id"] = userID
	details["sensitive"] = sensitive

	level := Info
	if sensitive {
		level = Warning
	}

	l.Log(level, AuditLogEntry{Category: "data_access", Action: action, Resource: resource, Result: "completed"}, details)
}

// LogSecurityEvent logs general security events
func (l *AuditLogger) LogSecurityEvent(eventType, source, description string, severity AuditLevel, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["source"] = source
	details["description"] = description

	l.Log(severity, AuditLogEntry{Category: "security", Action: eventType, Resource: "security_system", Result: "detected"}, details)
}

// Close closes the audit logger
func (l *AuditLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.writer != nil {
		return l.writer.Close()
	}
	return nil
}

// AuditCategories defines common audit categories
var AuditCategories = struct {
	Authentication string
	Authorization  string
	DataAccess     string
	Configuration  string
	System         string
	Security       string
	API            string
}{
	Authentication: "authentication",
	Authorization:  "authorization",
	DataAccess:     "data_access",
	Configuration:  "configuration",
	System:         "system",
	Security:       "security",
	API:            "api",
}

// Common audit actions
var AuditActions = struct {
	Create   string
	Read     string
	Update   string
	Delete   string
	Login    string
	Logout   string
	Validate string
	Block    string
	Allow    string
}{
	Create:   "create",
	Read:     "read",
	Update:   "update",
	Delete:   "delete",
	Login:    "login",
	Logout:   "logout",
	Validate: "validate",
	Block:    "block",
	Allow:    "allow",
}
