// Package testhelpers provides testing utilities and helpers
package testhelpers

import (
	"reflect"
	"testing"
)

// AssertEqual asserts that two values are equal
func AssertEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// AssertNotNil asserts that the value is not nil
func AssertNotNil(t *testing.T, got interface{}) {
	t.Helper()
	if got == nil {
		t.Errorf("expected non-nil value, got nil")
	}
}

// AssertNil asserts that the value is nil
func AssertNil(t *testing.T, got interface{}) {
	t.Helper()
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

// AssertNoError asserts that there is no error
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// AssertError asserts that there is an error
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

// AssertTrue asserts that the condition is true
func AssertTrue(t *testing.T, condition bool) {
	t.Helper()
	if !condition {
		t.Errorf("expected true, got false")
	}
}

// AssertFalse asserts that the condition is false
func AssertFalse(t *testing.T, condition bool) {
	t.Helper()
	if condition {
		t.Errorf("expected false, got true")
	}
}

// AssertContains asserts that the string contains the substring
func AssertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}

// Contains checks if string contains substring
func Contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || ContainsHelper(s, substr))
}

// ContainsHelper is the helper for substring search
func ContainsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
