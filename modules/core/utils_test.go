// Package core utility function tests
package core

import (
	"testing"
)

func TestStringSliceToString(t *testing.T) {
	input := []string{"a", "b", "c"}
	result := StringSliceToString(input)
	if result != "a,b,c" {
		t.Errorf("got %s, want a,b,c", result)
	}
}

func TestUint16SliceToString(t *testing.T) {
	input := []uint16{1, 2, 3}
	result := Uint16SliceToString(input)
	if result != "1,2,3" {
		t.Errorf("got %s, want 1,2,3", result)
	}
}

func TestCalculateSHA256(t *testing.T) {
	data := []byte("test")
	hash := CalculateSHA256(data)
	if len(hash) != 64 {
		t.Errorf("SHA256 should be 64 hex chars, got %d", len(hash))
	}
}

func TestRandomChoice(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}

	// test multiple selections
	chosen := make(map[int]bool)
	for i := 0; i < 100; i++ {
		c := RandomChoice(slice)
		chosen[c] = true
	}

	// verify at least some different values are selected
	if len(chosen) < 2 {
		t.Error("RandomChoice should return varied results")
	}
}

func TestRandomChoiceEmpty(t *testing.T) {
	var empty []int
	result := RandomChoice(empty)
	if result != 0 {
		t.Error("RandomChoice on empty slice should return zero value")
	}
}

func TestContains(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}

	if !Contains(slice, 3) {
		t.Error("Contains should return true for 3")
	}
	if Contains(slice, 10) {
		t.Error("Contains should return false for 10")
	}
}

func TestUnique(t *testing.T) {
	input := []int{1, 2, 2, 3, 3, 3}
	result := Unique(input)

	if len(result) != 3 {
		t.Errorf("Unique should return 3 elements, got %d", len(result))
	}

	expected := map[int]bool{1: true, 2: true, 3: true}
	for _, v := range result {
		if !expected[v] {
			t.Errorf("Unexpected value: %d", v)
		}
	}
}

func TestParseTLSVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected uint16
	}{
		{"1.0", 0x0301},
		{"1.1", 0x0302},
		{"1.2", 0x0303},
		{"1.3", 0x0304},
		{"2.0", 0x0303}, // default
	}

	for _, tt := range tests {
		result := ParseTLSVersion(tt.input)
		if result != tt.expected {
			t.Errorf("ParseTLSVersion(%s) = 0x%04x, want 0x%04x", tt.input, result, tt.expected)
		}
	}
}

func TestTLSVersionToString(t *testing.T) {
	tests := []struct {
		version  uint16
		expected string
	}{
		{0x0301, "1.0"},
		{0x0302, "1.1"},
		{0x0303, "1.2"},
		{0x0304, "1.3"},
		{0x9999, "unknown"},
	}

	for _, tt := range tests {
		result := TLSVersionToString(tt.version)
		if result != tt.expected {
			t.Errorf("TLSVersionToString(0x%04x) = %s, want %s", tt.version, result, tt.expected)
		}
	}
}

func TestNormalizeString(t *testing.T) {
	input := "  Hello World  "
	result := NormalizeString(input)
	if result != "hello world" {
		t.Errorf("got %s, want hello world", result)
	}
}

func TestTruncateString(t *testing.T) {
	input := "Hello World"
	result := TruncateString(input, 5)
	if result != "Hello" {
		t.Errorf("got %s, want Hello", result)
	}

	// no truncation
	result = TruncateString(input, 20)
	if result != input {
		t.Errorf("should not truncate when maxLen > len")
	}
}

func TestMergeMaps(t *testing.T) {
	m1 := map[string]string{"a": "1", "b": "2"}
	m2 := map[string]string{"b": "3", "c": "4"}

	result := MergeMaps(m1, m2)

	if len(result) != 3 {
		t.Errorf("merged map should have 3 keys, got %d", len(result))
	}
	if result["b"] != "3" {
		t.Error("second map should override first")
	}
}

func TestCopyMap(t *testing.T) {
	original := map[string]string{"a": "1", "b": "2"}
	copied := CopyMap(original)

	// modifying copy should not affect original
	copied["a"] = "modified"
	if original["a"] == "modified" {
		t.Error("CopyMap should create independent copy")
	}
}

func BenchmarkUint16SliceToString(b *testing.B) {
	slice := []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Uint16SliceToString(slice)
	}
}

func BenchmarkCalculateSHA256(b *testing.B) {
	data := []byte("test data for hashing")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateSHA256(data)
	}
}
