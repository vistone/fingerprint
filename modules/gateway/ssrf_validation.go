package gateway

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type blockedTargetError struct {
	cause error
}

func (e *blockedTargetError) Error() string {
	if e == nil || e.cause == nil {
		return "blocked target"
	}
	return e.cause.Error()
}

func (e *blockedTargetError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// ValidateOutboundTarget rejects malformed or internal-only targets before any outbound request is attempted.
func ValidateOutboundTarget(rawURL string, allowPrivate bool) error {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return &blockedTargetError{cause: fmt.Errorf("target URL is required")}
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return &blockedTargetError{cause: fmt.Errorf("invalid target URL: %w", err)}
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return &blockedTargetError{cause: fmt.Errorf("target URL scheme %q is not allowed", parsed.Scheme)}
	}

	if err := validateProxyTarget(parsed, allowPrivate); err != nil {
		return &blockedTargetError{cause: err}
	}

	return nil
}

func isBlockedTargetError(err error) bool {
	var targetErr *blockedTargetError
	return errors.As(err, &targetErr)
}
