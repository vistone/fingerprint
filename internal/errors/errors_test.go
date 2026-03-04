package errors

import (
	"errors"
	"fmt"
	"testing"
)

// TestSentinelErrors 测试哨兵错误
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

// TestWrap 测试错误包装
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

	// 测试 Unwrap
	if !errors.Is(wrapped, baseErr) {
		t.Error("Wrapped error should unwrap to base error")
	}
}

// TestWrapf 测试格式化错误包装
func TestWrapf(t *testing.T) {
	baseErr := errors.New("base error")
	wrapped := Wrapf(baseErr, "context %d", 42)

	expectedMsg := "context 42: base error"
	if wrapped.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, wrapped.Error())
	}
}

// TestWrapNil 测试包装 nil 错误
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

// TestCategorizedError 测试分类错误
func TestCategorizedError(t *testing.T) {
	err := NewCategorizedError(CategoryConfig, ErrConfigNotLoaded, "test details")

	expectedMsg := "[config] config not loaded: test details"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}

	// 测试 Unwrap
	if !errors.Is(err, ErrConfigNotLoaded) {
		t.Error("Should unwrap to ErrConfigNotLoaded")
	}
}

// TestCategorizedErrorNoDetails 测试无详细信息的分类错误
func TestCategorizedErrorNoDetails(t *testing.T) {
	err := NewCategorizedError(CategoryInput, ErrInvalidInput, "")

	expectedMsg := "[input] invalid input"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestIsClientHelloSpecNotImplemented 测试 ClientHelloSpec 未实现检查
func TestIsClientHelloSpecNotImplemented(t *testing.T) {
	// 测试哨兵错误
	if !IsClientHelloSpecNotImplemented(ErrClientHelloSpecNotImplemented) {
		t.Error("Should detect sentinel error")
	}

	// 测试包装后的错误
	wrapped := Wrap(ErrClientHelloSpecNotImplemented, "context")
	if !IsClientHelloSpecNotImplemented(wrapped) {
		t.Error("Should detect wrapped sentinel error")
	}

	// 测试消息匹配
	msgErr := errors.New("please implement this method")
	if !IsClientHelloSpecNotImplemented(msgErr) {
		t.Error("Should detect by message")
	}

	// 测试不匹配的错误
	if IsClientHelloSpecNotImplemented(errors.New("other error")) {
		t.Error("Should not match other errors")
	}

	// 测试 nil
	if IsClientHelloSpecNotImplemented(nil) {
		t.Error("Should not match nil")
	}
}

// TestIsNotFound 测试未找到错误检查
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

// TestIsInvalidInput 测试无效输入错误检查
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

// TestIsConfigError 测试配置错误检查
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

// TestNewConfigError 测试配置错误创建
func TestNewConfigError(t *testing.T) {
	baseErr := errors.New("connection failed")
	err := NewConfigError("load", baseErr)

	expectedMsg := "config operation 'load' failed: connection failed"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewValidationError 测试验证错误创建
func TestNewValidationError(t *testing.T) {
	err := NewValidationError("username", "cannot be empty")

	expectedMsg := "[input] invalid input: field 'username': cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewNotFoundError 测试未找到错误创建
func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("profile", "chrome_999")

	expectedMsg := "[not_found] profile not found: profile 'chrome_999' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewProtocolError 测试协议错误创建
func TestNewProtocolError(t *testing.T) {
	baseErr := errors.New("handshake failed")
	err := NewProtocolError("tls", baseErr)

	expectedMsg := "[protocol] handshake failed: protocol 'tls' error"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewNetworkError 测试网络错误创建
func TestNewNetworkError(t *testing.T) {
	baseErr := errors.New("timeout")
	err := NewNetworkError("connect", baseErr)

	expectedMsg := "[network] timeout: network operation 'connect' failed"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewf 测试格式化错误创建
func TestNewf(t *testing.T) {
	err := Newf("error code: %d", 500)

	expectedMsg := "error code: 500"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestCompatibilityWithStandardErrors 测试与标准错误库的兼容性
func TestCompatibilityWithStandardErrors(t *testing.T) {
	// 测试 Is
	if !Is(ErrProfileNotFound, ErrProfileNotFound) {
		t.Error("Is should work with sentinel errors")
	}

	// 测试 As
	var catErr *CategorizedError
	wrapped := NewCategorizedError(CategoryConfig, ErrConfigNotLoaded, "test")
	if !As(wrapped, &catErr) {
		t.Error("As should work with CategorizedError")
	}
	if catErr.Category != CategoryConfig {
		t.Errorf("Expected category %s, got %s", CategoryConfig, catErr.Category)
	}
}

// BenchmarkWrap 基准测试错误包装
func BenchmarkWrap(b *testing.B) {
	baseErr := errors.New("base error")
	for i := 0; i < b.N; i++ {
		_ = Wrap(baseErr, "context")
	}
}

// BenchmarkIs 基准测试错误检查
func BenchmarkIs(b *testing.B) {
	wrapped := Wrap(ErrProfileNotFound, "context")
	for i := 0; i < b.N; i++ {
		_ = Is(wrapped, ErrProfileNotFound)
	}
}

// BenchmarkCategorizedError 基准测试分类错误
func BenchmarkCategorizedError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		err := NewCategorizedError(CategoryInput, ErrInvalidInput, "benchmark")
		_ = err.Error()
	}
}

// ExampleWrap 示例：错误包装
func ExampleWrap() {
	baseErr := errors.New("file not found")
	wrapped := Wrap(baseErr, "loading config")
	fmt.Println(wrapped)
	// Output: loading config: file not found
}

// ExampleNewCategorizedError 示例：分类错误
func ExampleNewCategorizedError() {
	err := NewCategorizedError(CategoryConfig, ErrConfigNotLoaded, "missing file")
	fmt.Println(err)
	// Output: [config] config not loaded: missing file
}
