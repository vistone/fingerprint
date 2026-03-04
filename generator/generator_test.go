package generator

import (
	"errors"
	"testing"
)

// TestErrorDefinitions 测试错误定义
func TestErrorDefinitions(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrNoProfilesAvailable",
			err:  ErrNoProfilesAvailable,
			want: "no profiles available: for generators",
		},
		{
			name: "ErrFailedToGenerateFingerprint",
			err:  ErrFailedToGenerateFingerprint,
			want: "invalid fingerprint format: fingerprint generation failed",
		},
		{
			name: "ErrInvalidRandomSource",
			err:  ErrInvalidRandomSource,
			want: "invalid fingerprint format: invalid random source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Error("error should not be nil")
			}
			if tt.err.Error() != tt.want {
				t.Errorf("error message = %v, want %v", tt.err.Error(), tt.want)
			}
		})
	}
}

// TestIsNoProfilesAvailable 测试无可用指纹错误检查
func TestIsNoProfilesAvailable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "matching error",
			err:  ErrNoProfilesAvailable,
			want: true,
		},
		{
			name: "wrapped matching error",
			err:  errors.Join(errors.New("context"), ErrNoProfilesAvailable),
			want: true,
		},
		{
			name: "different error",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "ErrFailedToGenerateFingerprint",
			err:  ErrFailedToGenerateFingerprint,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNoProfilesAvailable(tt.err)
			if got != tt.want {
				t.Errorf("IsNoProfilesAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsFailedToGenerateFingerprint 测试指纹生成失败错误检查
func TestIsFailedToGenerateFingerprint(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "matching error",
			err:  ErrFailedToGenerateFingerprint,
			want: true,
		},
		{
			name: "wrapped matching error",
			err:  errors.Join(errors.New("context"), ErrFailedToGenerateFingerprint),
			want: true,
		},
		{
			name: "different error",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "ErrNoProfilesAvailable",
			err:  ErrNoProfilesAvailable,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFailedToGenerateFingerprint(tt.err)
			if got != tt.want {
				t.Errorf("IsFailedToGenerateFingerprint() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestErrorWrapping 测试错误包装
func TestErrorWrapping(t *testing.T) {
	// 测试错误是否可以被 errors.Is 识别
	if !errors.Is(ErrNoProfilesAvailable, ErrNoProfilesAvailable) {
		t.Error("ErrNoProfilesAvailable should be self-identifiable")
	}

	if !errors.Is(ErrFailedToGenerateFingerprint, ErrFailedToGenerateFingerprint) {
		t.Error("ErrFailedToGenerateFingerprint should be self-identifiable")
	}

	// 测试错误不等于其他错误
	if errors.Is(ErrNoProfilesAvailable, ErrFailedToGenerateFingerprint) {
		t.Error("ErrNoProfilesAvailable should not match ErrFailedToGenerateFingerprint")
	}

	if errors.Is(ErrFailedToGenerateFingerprint, ErrNoProfilesAvailable) {
		t.Error("ErrFailedToGenerateFingerprint should not match ErrNoProfilesAvailable")
	}
}

// TestPackageExports 测试包导出
func TestPackageExports(t *testing.T) {
	// 确保所有导出变量不为 nil
	if ErrNoProfilesAvailable == nil {
		t.Error("ErrNoProfilesAvailable should not be nil")
	}

	if ErrFailedToGenerateFingerprint == nil {
		t.Error("ErrFailedToGenerateFingerprint should not be nil")
	}

	if ErrInvalidRandomSource == nil {
		t.Error("ErrInvalidRandomSource should not be nil")
	}
}
