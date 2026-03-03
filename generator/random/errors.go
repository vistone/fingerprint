package random

import (
	"errors"
	"fmt"

	"github.com/vistone/fingerprint/generator"
	errs "github.com/vistone/fingerprint/internal/errors"
)

// ============================================================================
// Random 子包错误定义
// ============================================================================

var (
	// ErrBrowserTypeNotSupported 表示浏览器类型不支持
	ErrBrowserTypeNotSupported = fmt.Errorf("%w: random generator", errs.ErrUnsupportedBrowser)

	// ErrNoRandomProfileFound 表示没有找到随机指纹
	ErrNoRandomProfileFound = fmt.Errorf("%w: random profiles", errs.ErrNoProfilesAvailable)

	// ErrRandomProfileInvalid 表示随机指纹无效
	ErrRandomProfileInvalid = fmt.Errorf("%w: random profile is invalid", generator.ErrFailedToGenerateFingerprint)
)

// IsBrowserTypeNotSupported 检查错误是否为浏览器类型不支持
func IsBrowserTypeNotSupported(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrBrowserTypeNotSupported)
}

// IsNoRandomProfileFound 检查错误是否为没有随机指纹
func IsNoRandomProfileFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrNoRandomProfileFound)
}

// IsRandomProfileInvalid 检查错误是否为随机指纹无效
func IsRandomProfileInvalid(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrRandomProfileInvalid)
}
