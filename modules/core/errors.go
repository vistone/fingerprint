// Package core re-exports error types from the canonical errors package
// for backward compatibility. New code should import modules/errors directly.
package core

import (
	errs "github.com/vistone/fingerprint/modules/errors"
)

// Type aliases — identical to modules/errors types
type ErrorCode = errs.ErrorCode
type CoreError = errs.CoreError

// Validation error codes (VAL001-VAL010)
const (
	ErrCodeInvalidInput   = errs.ErrCodeInvalidInput
	ErrCodeRequiredField  = errs.ErrCodeRequiredField
	ErrCodeInvalidFormat  = errs.ErrCodeInvalidFormat
	ErrCodeOutOfRange     = errs.ErrCodeOutOfRange
	ErrCodeNilPointer     = errs.ErrCodeNilPointer
	ErrCodeInvalidType    = errs.ErrCodeInvalidType
	ErrCodeEmptyValue     = errs.ErrCodeEmptyValue
	ErrCodeInvalidLength  = errs.ErrCodeInvalidLength
	ErrCodeInvalidCharset = errs.ErrCodeInvalidCharset
	ErrCodeInvalidPattern = errs.ErrCodeInvalidPattern
)

// Not Found error codes (NTF001-NTF005)
const (
	ErrCodeProfileNotFound  = errs.ErrCodeProfileNTF
	ErrCodeResourceNotFound = errs.ErrCodeResourceNTF
	ErrCodeKeyNotFound      = errs.ErrCodeKeyNTF
	ErrCodeFileNotFound     = errs.ErrCodeFileNTF
	ErrCodeRouteNotFound    = errs.ErrCodeRouteNTF
)

// Security error codes (SEC001-SEC010)
const (
	ErrCodeSecurityError       = errs.ErrCodeSecurityError
	ErrCodeAuthFailed          = errs.ErrCodeAuthFailed
	ErrCodeUnauthorized        = errs.ErrCodeUnauthorized
	ErrCodeForbidden           = errs.ErrCodeForbidden
	ErrCodeInvalidToken        = errs.ErrCodeInvalidToken
	ErrCodeReplayAttack        = errs.ErrCodeReplayAttack
	ErrCodeFingerprintMismatch = errs.ErrCodeFingerprintMismatch
	ErrCodeInvalidTLSVersion   = errs.ErrCodeInvalidTLSVersion
	ErrCodeInvalidJA3Hash      = errs.ErrCodeInvalidJA3Hash
	ErrCodeInvalidJA4Hash      = errs.ErrCodeInvalidJA4Hash
)

// Sentinel errors re-exported from errors package
var (
	ErrInvalidProfile          = errs.ErrInvalidProfile
	ErrProfileNotFound         = errs.ErrProfileNotFound
	ErrUnsupportedBrowser      = errs.ErrUnsupportedBrowser
	ErrInvalidTLSVersion       = errs.ErrInvalidTLSVersion
	ErrInvalidJA3Hash          = errs.ErrInvalidJA3Hash
	ErrInvalidJA4Hash          = errs.ErrInvalidJA4Hash
	ErrFeatureExtractionFailed = errs.ErrFeatureExtractionFailed
	ErrClassificationFailed    = errs.ErrClassificationFailed
	ErrRiskAssessmentFailed    = errs.ErrRiskAssessmentFailed
	ErrNilPointer              = errs.ErrNilPointer
	ErrInvalidInput            = errs.ErrInvalidInput
	ErrOutOfRange              = errs.ErrOutOfRange
	ErrEmptyValue              = errs.ErrEmptyValue
	ErrAlreadyExists           = errs.ErrAlreadyExists
)

// Error creation functions re-exported from errors package
func NewCodedError(code ErrorCode, op string, err error) *CoreError {
	return errs.NewCodedError(code, op, err)
}

func NewCodedErrorf(code ErrorCode, op, format string, args ...interface{}) *CoreError {
	return errs.NewCodedErrorf(code, op, format, args...)
}

func WrapError(code ErrorCode, op string, err error, context string) *CoreError {
	return errs.WrapError(code, op, err, context)
}

func WrapErrorf(code ErrorCode, op string, err error, format string, args ...interface{}) *CoreError {
	return errs.WrapErrorf(code, op, err, format, args...)
}

func IsCoreError(err error) bool {
	return errs.IsCoreError(err)
}

func GetErrorCode(err error) ErrorCode {
	return errs.GetErrorCode(err)
}

func IsErrorCode(err error, code ErrorCode) bool {
	return errs.IsErrorCode(err, code)
}
