// Package core 提供输入验证工具
package core

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Validator 输入验证器
type Validator struct {
	errors []error
}

// NewValidator 创建新的验证器
func NewValidator() *Validator {
	return &Validator{
		errors: make([]error, 0),
	}
}

// HasErrors 是否有错误
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors 获取所有错误
func (v *Validator) Errors() []error {
	return v.errors
}

// Error 返回组合错误
func (v *Validator) Error() error {
	if !v.HasErrors() {
		return nil
	}
	msgs := make([]string, len(v.errors))
	for i, err := range v.errors {
		msgs[i] = err.Error()
	}
	return NewCodedError(ErrCodeInvalidInput, "Validator", errors.New(strings.Join(msgs, "; ")))
}

// AddError 添加错误
func (v *Validator) AddError(err error) {
	if err != nil {
		v.errors = append(v.errors, err)
	}
}

// AddErrorf 添加格式化错误
func (v *Validator) AddErrorf(format string, args ...interface{}) {
	v.errors = append(v.errors, fmt.Errorf(format, args...))
}

// NotNil 验证非 nil
func (v *Validator) NotNil(val interface{}, name string) *Validator {
	if val == nil {
		v.AddErrorf("%s cannot be nil", name)
	}
	return v
}

// NotEmpty 验证字符串非空
func (v *Validator) NotEmpty(s string, name string) *Validator {
	if strings.TrimSpace(s) == "" {
		v.AddErrorf("%s cannot be empty", name)
	}
	return v
}

// NotZero 验证整数非零
func (v *Validator) NotZero(n int, name string) *Validator {
	if n == 0 {
		v.AddErrorf("%s cannot be zero", name)
	}
	return v
}

// Positive 验证正数
func (v *Validator) Positive(n int, name string) *Validator {
	if n <= 0 {
		v.AddErrorf("%s must be positive, got %d", name, n)
	}
	return v
}

// InRange 验证范围
func (v *Validator) InRange(n, min, max int, name string) *Validator {
	if n < min || n > max {
		v.AddErrorf("%s must be in range [%d, %d], got %d", name, min, max, n)
	}
	return v
}

// MinLength 验证最小长度
func (v *Validator) MinLength(s string, min int, name string) *Validator {
	if len(s) < min {
		v.AddErrorf("%s must be at least %d characters, got %d", name, min, len(s))
	}
	return v
}

// MaxLength 验证最大长度
func (v *Validator) MaxLength(s string, max int, name string) *Validator {
	if len(s) > max {
		v.AddErrorf("%s must be at most %d characters, got %d", name, max, len(s))
	}
	return v
}

// Matches 验证正则匹配
func (v *Validator) Matches(s string, pattern *regexp.Regexp, name string) *Validator {
	if !pattern.MatchString(s) {
		v.AddErrorf("%s format is invalid", name)
	}
	return v
}

// ValidBrowserType 验证浏览器类型
func (v *Validator) ValidBrowserType(bt BrowserType, name string) *Validator {
	valid := []BrowserType{
		BrowserChrome, BrowserFirefox, BrowserSafari,
		BrowserEdge, BrowserOpera, BrowserBrave,
	}
	for _, b := range valid {
		if bt == b {
			return v
		}
	}
	v.AddErrorf("%s is not a valid browser type: %s", name, bt)
	return v
}

// ValidOS 验证操作系统
func (v *Validator) ValidOS(os OperatingSystem, name string) *Validator {
	valid := []OperatingSystem{
		OSWindows10, OSWindows11, OSMacOS13, OSMacOS14, OSMacOS15,
		OSLinux, OSLinuxUbuntu, OSLinuxDebian, OSLinuxFedora,
		OSiOS, OSiPadOS, OSAndroid,
	}
	for _, o := range valid {
		if os == o {
			return v
		}
	}
	v.AddErrorf("%s is not a valid OS: %s", name, os)
	return v
}

// ValidateTLSVersion 验证 TLS 版本
func ValidateTLSVersion(version uint16) error {
	switch version {
	case 0x0301, 0x0302, 0x0303, 0x0304:
		return nil
	default:
		return NewCodedErrorf(ErrCodeInvalidTLSVersion, "ValidateTLSVersion",
			"invalid TLS version: 0x%04x", version)
	}
}

// ValidateJA3Hash 验证 JA3 哈希
func ValidateJA3Hash(hash string) error {
	if len(hash) != 32 {
		return NewCodedErrorf(ErrCodeInvalidJA3Hash, "ValidateJA3Hash",
			"invalid JA3 hash length: %d, expected 32", len(hash))
	}
	// JA3 是 MD5，应该是 32 位十六进制
	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return NewCodedError(ErrCodeInvalidJA3Hash, "ValidateJA3Hash",
				fmt.Errorf("invalid character in JA3 hash: %c", c))
		}
	}
	return nil
}

// SanitizeString 清理字符串输入
func SanitizeString(s string, maxLen int) string {
	// 移除控制字符
	s = strings.Map(func(r rune) rune {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return -1
		}
		return r
	}, s)
	
	// 截断过长的字符串
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	
	return strings.TrimSpace(s)
}

// SafeDereference 安全解引用指针
func SafeDereference(ptr *int, defaultVal int) int {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// SafeSliceAccess 安全访问切片
func SafeSliceAccess(slice []interface{}, index int) (interface{}, bool) {
	if index < 0 || index >= len(slice) {
		return nil, false
	}
	return slice[index], true
}
