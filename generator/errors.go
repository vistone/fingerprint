package generator

import (
	"errors"
	"fmt"

	errs "github.com/vistone/fingerprint/internal/errors"
)

// ============================================================================
// 生成器子包错误定义
// ============================================================================

var (
	// ErrNoProfilesAvailable 表示没有可用的指纹配置
	ErrNoProfilesAvailable = fmt.Errorf("%w: for generators", errs.ErrNoProfilesAvailable)

	// ErrFailedToGenerateFingerprint 表示指纹生成失败
	ErrFailedToGenerateFingerprint = fmt.Errorf("%w: fingerprint generation failed", errs.ErrInvalidFingerprint)

	// ErrInvalidRandomSource 表示随机源无效
	ErrInvalidRandomSource = fmt.Errorf("%w: invalid random source", errs.ErrInvalidFingerprint)
)

// IsNoProfilesAvailable 检查错误是否为无可用指纹
func IsNoProfilesAvailable(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrNoProfilesAvailable)
}

// IsFailedToGenerateFingerprint 检查错误是否为指纹生成失败
func IsFailedToGenerateFingerprint(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrFailedToGenerateFingerprint)
}
