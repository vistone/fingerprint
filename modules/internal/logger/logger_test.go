package logger

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// translated comment
func TestNew(t *testing.T) {
	l := New()
	if l == nil {
		t.Fatal("New() returned nil")
	}

	if l.level != InfoLevel {
		t.Errorf("Expected default level Info, got %v", l.level)
	}

	if l.fields == nil {
		t.Error("fields should not be nil")
	}
}

// translated comment
func TestWithLevel(t *testing.T) {
	l := New(WithLevel(DebugLevel))
	if l.level != DebugLevel {
		t.Errorf("Expected level Debug, got %v", l.level)
	}
}

// translated comment
func TestWithFields(t *testing.T) {
	fields := map[string]interface{}{
		"service": "test",
		"version": "1.0",
	}
	l := New(WithFields(fields))

	if l.fields["service"] != "test" {
		t.Error("field 'service' not set correctly")
	}
	if l.fields["version"] != "1.0" {
		t.Error("field 'version' not set correctly")
	}
}

// translated comment
func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	l := New(
		WithLevel(DebugLevel),
		WithOutput(&buf),
	)

	// translated comment
	l.Debug("debug message")
	if !strings.Contains(buf.String(), "debug message") {
		t.Error("Debug log not output correctly")
	}

	buf.Reset()

	// translated comment
	l.Info("info message")
	if !strings.Contains(buf.String(), "info message") {
		t.Error("Info log not output correctly")
	}

	buf.Reset()

	// translated comment
	l.Warn("warn message")
	if !strings.Contains(buf.String(), "warn message") {
		t.Error("Warn log not output correctly")
	}

	buf.Reset()

	// translated comment
	l.Error("error message")
	if !strings.Contains(buf.String(), "error message") {
		t.Error("Error log not output correctly")
	}
}

// translated comment
func TestLogf(t *testing.T) {
	var buf bytes.Buffer
	l := New(
		WithLevel(InfoLevel),
		WithOutput(&buf),
	)

	l.Infof("test %s %d", "message", 42)
	output := buf.String()
	if !strings.Contains(output, "test message 42") {
		t.Errorf("Formatted log not output correctly: %s", output)
	}
}

// translated comment
func TestLogw(t *testing.T) {
	var buf bytes.Buffer
	l := New(
		WithLevel(InfoLevel),
		WithOutput(&buf),
	)

	l.Infow("test message", "key1", "value1", "key2", 42)
	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Error("Message not in output")
	}
	if !strings.Contains(output, "key1=value1") {
		t.Error("Field key1 not in output")
	}
	if !strings.Contains(output, "key2=42") {
		t.Error("Field key2 not in output")
	}
}

// translated comment
func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := New(
		WithLevel(WarnLevel),
		WithOutput(&buf),
	)

	// translated comment
	l.Debug("debug")
	l.Info("info")
	if buf.Len() > 0 {
		t.Error("Debug and Info should be filtered")
	}

	// translated comment
	l.Warn("warn")
	if !strings.Contains(buf.String(), "warn") {
		t.Error("Warn should be output")
	}

	buf.Reset()
	l.Error("error")
	if !strings.Contains(buf.String(), "error") {
		t.Error("Error should be output")
	}
}

// translated comment
func TestWith(t *testing.T) {
	parent := New(WithFields(map[string]interface{}{"parent": "value"}))
	child := parent.With(map[string]interface{}{"child": "value"})

	// translated comment
	if child.fields["parent"] != "value" {
		t.Error("Child should inherit parent's fields")
	}
	if child.fields["child"] != "value" {
		t.Error("Child should have its own fields")
	}

	// translated comment
	if _, exists := parent.fields["child"]; exists {
		t.Error("Parent's fields should not be modified")
	}
}

// translated comment
func TestSetLevel(t *testing.T) {
	var buf bytes.Buffer
	l := New(
		WithLevel(InfoLevel),
		WithOutput(&buf),
	)

	// translated comment
	l.Debug("debug1")
	if buf.Len() > 0 {
		t.Error("Debug should be filtered")
	}

	// translated comment
	l.SetLevel(DebugLevel)

	// translated comment
	l.Debug("debug2")
	if !strings.Contains(buf.String(), "debug2") {
		t.Error("Debug should be output after level change")
	}
}

// translated comment
func TestDefaultFormatter(t *testing.T) {
	f := &DefaultFormatter{withTimestamp: true, withLevel: true}
	output := f.Format(InfoLevel, "test", map[string]interface{}{"key": "value"}, time.Now())

	if !strings.Contains(output, "test") {
		t.Error("Message not in output")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("Level not in output")
	}
	if !strings.Contains(output, "key=value") {
		t.Error("Fields not in output")
	}
}

// translated comment
func TestJSONFormatter(t *testing.T) {
	f := &JSONFormatter{}
	output := f.Format(InfoLevel, "test", map[string]interface{}{"key": "value"}, time.Now())

	if !strings.Contains(output, `"level":"INFO"`) {
		t.Error("Level not in JSON output")
	}
	if !strings.Contains(output, `"message":"test"`) {
		t.Error("Message not in JSON output")
	}
}

// translated comment
func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{FatalLevel, "FATAL"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		if tt.level.String() != tt.expected {
			t.Errorf("Level %v: expected %s, got %s", tt.level, tt.expected, tt.level.String())
		}
	}
}

// translated comment
func TestGlobalFunctions(t *testing.T) {
	// translated comment
	// translated comment
	// translated comment
	// translated comment

	// translated comment
	Info("global info")
	Warnf("formatted %s", "warn")
	Errorw("error with fields", "key", "value")
	// translated comment
}

// translated comment
func TestKvToMap(t *testing.T) {
	// translated comment
	m := kvToMap("key1", "value1", "key2", 42)
	if m["key1"] != "value1" {
		t.Error("key1 not converted correctly")
	}
	if m["key2"] != 42 {
		t.Error("key2 not converted correctly")
	}

	// translated comment
	m2 := kvToMap("key1", "value1", "key2")
	if len(m2) != 1 {
		t.Error("Should handle odd number of arguments")
	}

	// translated comment
	m3 := kvToMap()
	if len(m3) != 0 {
		t.Error("Should handle empty arguments")
	}

	// translated comment
	m4 := kvToMap(123, "value")
	if len(m4) != 0 {
		t.Error("Should skip non-string keys")
	}
}

// translated comment
func BenchmarkInfo(b *testing.B) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("benchmark message")
	}
}

// translated comment
func BenchmarkInfow(b *testing.B) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Infow("benchmark message", "key1", "value1", "key2", 42)
	}
}

// translated comment
func BenchmarkInfof(b *testing.B) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Infof("benchmark %s %d", "message", i)
	}
}
