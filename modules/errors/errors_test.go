package errors

import (
	"errors"
	"fmt"
	"testing"
)

// TestSentinelErrors tests sentinel errors
func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrProfileNotFound", ErrProfileNotFound},
		{"ErrInvalidFingerprint", ErrInvalidFingerprint},
		{"ErrClientHelloSpecNotImplemented", ErrClientHelloSpecNotImplemented},
		{"ErrInvalidUserAgent", ErrInvalidUserAgent},
		{"ErrNoProfilesAvailable", ErrNoProfilesAvailable},
		{"ErrUnsupportedBrowser", ErrUnsupportedBrowser},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s should not be nil", tt.name)
			}
		})
	}
}

// TestWrap tests error wrapping
func TestWrap(t *testing.T) {
	baseErr := errors.New("base error")
	wrapped := Wrap(baseErr, "context")

	if wrapped == nil {
		t.Fatal("Wrap should not return nil")
	}

	expectedMsg := "context: base error"
	if wrapped.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, wrapped.Error())
	}

	// test Unwrap
	if !errors.Is(wrapped, baseErr) {
		t.Error("Wrapped error should unwrap to base error")
	}
}

// TestWrapf tests formatted error wrapping
func TestWrapf(t *testing.T) {
	baseErr := errors.New("base error")
	wrapped := Wrapf(baseErr, "context %d", 42)

	expectedMsg := "context 42: base error"
	if wrapped.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, wrapped.Error())
	}
}

// TestWrapNil tests wrapping nil error
func TestWrapNil(t *testing.T) {
	result := Wrap(nil, "context")
	if result != nil {
		t.Error("Wrapping nil should return nil")
	}

	result = Wrapf(nil, "context %d", 42)
	if result != nil {
		t.Error("Wrapf nil should return nil")
	}
}

// TestCategorizedError tests categorized error
func TestCategorizedError(t *testing.T) {
	err := NewCategorizedError(CategoryConfig, ErrConfigNotLoaded, "test details")

	expectedMsg := "[config] config not loaded: test details"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}

	// test Unwrap
	if !errors.Is(err, ErrConfigNotLoaded) {
		t.Error("Should unwrap to ErrConfigNotLoaded")
	}
}

// TestCategorizedErrorNoDetails tests categorized error without details
func TestCategorizedErrorNoDetails(t *testing.T) {
	err := NewCategorizedError(CategoryInput, ErrInvalidInput, "")

	expectedMsg := "[input] invalid input"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestIsClientHelloSpecNotImplemented tests ClientHelloSpec not implemented check
func TestIsClientHelloSpecNotImplemented(t *testing.T) {
	// test sentinel error
	if !IsClientHelloSpecNotImplemented(ErrClientHelloSpecNotImplemented) {
		t.Error("Should detect sentinel error")
	}

	// test wrapped error
	wrapped := Wrap(ErrClientHelloSpecNotImplemented, "context")
	if !IsClientHelloSpecNotImplemented(wrapped) {
		t.Error("Should detect wrapped sentinel error")
	}

	// test message match
	msgErr := errors.New("please implement this method")
	if !IsClientHelloSpecNotImplemented(msgErr) {
		t.Error("Should detect by message")
	}

	// test non-matching error
	if IsClientHelloSpecNotImplemented(errors.New("other error")) {
		t.Error("Should not match other errors")
	}

	// test nil
	if IsClientHelloSpecNotImplemented(nil) {
		t.Error("Should not match nil")
	}
}

// TestIsNotFound tests not found error check
func TestIsNotFound(t *testing.T) {
	if !IsNotFound(ErrProfileNotFound) {
		t.Error("Should detect ErrProfileNotFound")
	}

	if !IsNotFound(ErrConfigPathNotFound) {
		t.Error("Should detect ErrConfigPathNotFound")
	}

	if !IsNotFound(ErrVersionNotFound) {
		t.Error("Should detect ErrVersionNotFound")
	}

	if IsNotFound(ErrInvalidInput) {
		t.Error("Should not detect ErrInvalidInput")
	}

	if IsNotFound(nil) {
		t.Error("Should not detect nil")
	}
}

// TestIsInvalidInput tests invalid input error check
func TestIsInvalidInput(t *testing.T) {
	if !IsInvalidInput(ErrInvalidInput) {
		t.Error("Should detect ErrInvalidInput")
	}

	if !IsInvalidInput(ErrInvalidFingerprint) {
		t.Error("Should detect ErrInvalidFingerprint")
	}

	if !IsInvalidInput(ErrInvalidUserAgent) {
		t.Error("Should detect ErrInvalidUserAgent")
	}

	if IsInvalidInput(ErrProfileNotFound) {
		t.Error("Should not detect ErrProfileNotFound")
	}

	if IsInvalidInput(nil) {
		t.Error("Should not detect nil")
	}
}

// TestIsConfigError tests configuration error check
func TestIsConfigError(t *testing.T) {
	if !IsConfigError(ErrConfigNotLoaded) {
		t.Error("Should detect ErrConfigNotLoaded")
	}

	if !IsConfigError(ErrConfigValidation) {
		t.Error("Should detect ErrConfigValidation")
	}

	if !IsConfigError(ErrConfigPathNotFound) {
		t.Error("Should detect ErrConfigPathNotFound")
	}

	if IsConfigError(ErrProfileNotFound) {
		t.Error("Should not detect ErrProfileNotFound")
	}

	if IsConfigError(nil) {
		t.Error("Should not detect nil")
	}
}

// TestNewConfigError tests configuration error creation
func TestNewConfigError(t *testing.T) {
	baseErr := errors.New("connection failed")
	err := NewConfigError("load", baseErr)

	expectedMsg := "config operation 'load' failed: connection failed"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewValidationError tests validation error creation
func TestNewValidationError(t *testing.T) {
	err := NewValidationError("username", "cannot be empty")

	expectedMsg := "[input] invalid input: field 'username': cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewNotFoundError tests not found error creation
func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("profile", "chrome_999")

	expectedMsg := "[not_found] profile not found: profile 'chrome_999' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewProtocolError tests protocol error creation
func TestNewProtocolError(t *testing.T) {
	baseErr := errors.New("handshake failed")
	err := NewProtocolError("tls", baseErr)

	expectedMsg := "[protocol] handshake failed: protocol 'tls' error"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewNetworkError tests network error creation
func TestNewNetworkError(t *testing.T) {
	baseErr := errors.New("timeout")
	err := NewNetworkError("connect", baseErr)

	expectedMsg := "[network] timeout: network operation 'connect' failed"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewf tests formatted error creation
func TestNewf(t *testing.T) {
	err := Newf("error code: %d", 500)

	expectedMsg := "error code: 500"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestCompatibilityWithStandardErrors tests compatibility with standard error library
func TestCompatibilityWithStandardErrors(t *testing.T) {
	// test Is
	if !Is(ErrProfileNotFound, ErrProfileNotFound) {
		t.Error("Is should work with sentinel errors")
	}

	// test As
	var catErr *CategorizedError
	wrapped := NewCategorizedError(CategoryConfig, ErrConfigNotLoaded, "test")
	if !As(wrapped, &catErr) {
		t.Error("As should work with CategorizedError")
	}
	if catErr.Category != CategoryConfig {
		t.Errorf("Expected category %s, got %s", CategoryConfig, catErr.Category)
	}
}

// BenchmarkWrap benchmarks error wrapping
func BenchmarkWrap(b *testing.B) {
	baseErr := errors.New("base error")
	for i := 0; i < b.N; i++ {
		_ = Wrap(baseErr, "context")
	}
}

// BenchmarkIs benchmarks error check
func BenchmarkIs(b *testing.B) {
	wrapped := Wrap(ErrProfileNotFound, "context")
	for i := 0; i < b.N; i++ {
		_ = Is(wrapped, ErrProfileNotFound)
	}
}

// BenchmarkCategorizedError benchmarks categorized error
func BenchmarkCategorizedError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		err := NewCategorizedError(CategoryInput, ErrInvalidInput, "benchmark")
		_ = err.Error()
	}
}

// ExampleWrap example: error wrapping
func ExampleWrap() {
	baseErr := errors.New("file not found")
	wrapped := Wrap(baseErr, "loading config")
	fmt.Println(wrapped)
	// Output: loading config: file not found
}

// ExampleNewCategorizedError example:classifyerror
func ExampleNewCategorizedError() {
	err := NewCategorizedError(CategoryConfig, ErrConfigNotLoaded, "missing file")
	fmt.Println(err)
	// Output: [config] config not loaded: missing file
}
