package logger

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// TestNew 测试创建日志记录器
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

// TestWithLevel 测试设置日志级别
func TestWithLevel(t *testing.T) {
	l := New(WithLevel(DebugLevel))
	if l.level != DebugLevel {
		t.Errorf("Expected level Debug, got %v", l.level)
	}
}

// TestWithFields 测试设置默认字段
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

// TestLogLevels 测试各日志级别
func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	l := New(
		WithLevel(DebugLevel),
		WithOutput(&buf),
	)

	// 测试 Debug
	l.Debug("debug message")
	if !strings.Contains(buf.String(), "debug message") {
		t.Error("Debug log not output correctly")
	}

	buf.Reset()

	// 测试 Info
	l.Info("info message")
	if !strings.Contains(buf.String(), "info message") {
		t.Error("Info log not output correctly")
	}

	buf.Reset()

	// 测试 Warn
	l.Warn("warn message")
	if !strings.Contains(buf.String(), "warn message") {
		t.Error("Warn log not output correctly")
	}

	buf.Reset()

	// 测试 Error
	l.Error("error message")
	if !strings.Contains(buf.String(), "error message") {
		t.Error("Error log not output correctly")
	}
}

// TestLogf 测试格式化日志
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

// TestLogw 测试带字段的日志
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

// TestLevelFiltering 测试日志级别过滤
func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := New(
		WithLevel(WarnLevel),
		WithOutput(&buf),
	)

	// Debug 和 Info 应该被过滤
	l.Debug("debug")
	l.Info("info")
	if buf.Len() > 0 {
		t.Error("Debug and Info should be filtered")
	}

	// Warn 和 Error 应该输出
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

// TestWith 测试创建子日志记录器
func TestWith(t *testing.T) {
	parent := New(WithFields(map[string]interface{}{"parent": "value"}))
	child := parent.With(map[string]interface{}{"child": "value"})

	// 子日志应该继承父日志的字段
	if child.fields["parent"] != "value" {
		t.Error("Child should inherit parent's fields")
	}
	if child.fields["child"] != "value" {
		t.Error("Child should have its own fields")
	}

	// 父日志的字段不应该被修改
	if _, exists := parent.fields["child"]; exists {
		t.Error("Parent's fields should not be modified")
	}
}

// TestSetLevel 测试动态设置日志级别
func TestSetLevel(t *testing.T) {
	var buf bytes.Buffer
	l := New(
		WithLevel(InfoLevel),
		WithOutput(&buf),
	)

	// Debug 应该被过滤
	l.Debug("debug1")
	if buf.Len() > 0 {
		t.Error("Debug should be filtered")
	}

	// 改变级别
	l.SetLevel(DebugLevel)

	// 现在 Debug 应该输出
	l.Debug("debug2")
	if !strings.Contains(buf.String(), "debug2") {
		t.Error("Debug should be output after level change")
	}
}

// TestDefaultFormatter 测试默认格式化器
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

// TestJSONFormatter 测试 JSON 格式化器
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

// TestLogLevelString 测试日志级别字符串
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

// TestGlobalFunctions 测试全局便捷函数
func TestGlobalFunctions(t *testing.T) {
	// 注意：全局函数使用 Default() 日志记录器
	// 由于 sync.Once 的存在，Default() 只会初始化一次
	// 这个测试只是验证函数调用不会 panic
	// 实际输出验证在 Logger 的方法测试中已完成

	// 这些调用不应 panic
	Info("global info")
	Warnf("formatted %s", "warn")
	Errorw("error with fields", "key", "value")
	// Debug 级别默认被过滤，不测试
}

// TestKvToMap 测试键值对转换
func TestKvToMap(t *testing.T) {
	// 正常情况
	m := kvToMap("key1", "value1", "key2", 42)
	if m["key1"] != "value1" {
		t.Error("key1 not converted correctly")
	}
	if m["key2"] != 42 {
		t.Error("key2 not converted correctly")
	}

	// 奇数个参数
	m2 := kvToMap("key1", "value1", "key2")
	if len(m2) != 1 {
		t.Error("Should handle odd number of arguments")
	}

	// 空参数
	m3 := kvToMap()
	if len(m3) != 0 {
		t.Error("Should handle empty arguments")
	}

	// 非字符串键
	m4 := kvToMap(123, "value")
	if len(m4) != 0 {
		t.Error("Should skip non-string keys")
	}
}

// BenchmarkInfo 基准测试 Info 日志
func BenchmarkInfo(b *testing.B) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("benchmark message")
	}
}

// BenchmarkInfow 基准测试带字段的日志
func BenchmarkInfow(b *testing.B) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Infow("benchmark message", "key1", "value1", "key2", 42)
	}
}

// BenchmarkInfof 基准测试格式化日志
func BenchmarkInfof(b *testing.B) {
	var buf bytes.Buffer
	l := New(WithOutput(&buf))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Infof("benchmark %s %d", "message", i)
	}
}
