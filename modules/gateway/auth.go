// Package gateway provides API authentication middleware.
// Supports API key validation via header or query parameter.
package gateway

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// APIKeyHeader is the default HTTP header for API key authentication.
const APIKeyHeader = "X-API-Key"

// apiKeyAuth implements API key based authentication middleware.
type apiKeyAuth struct {
	// keys holds the set of valid API keys for constant-time comparison.
	keys []string
	// skipPaths are path prefixes that bypass authentication (e.g. health check).
	skipPaths []string
}

// NewAPIKeyAuth creates a new API key authenticator.
// keys must not be empty; skipPaths defines unauthenticated path prefixes.
func NewAPIKeyAuth(keys []string, skipPaths []string) *apiKeyAuth {
	return &apiKeyAuth{
		keys:      keys,
		skipPaths: skipPaths,
	}
}

// Middleware returns an http.Handler that enforces API key authentication.
// Requests to skip paths are passed through without checks.
func (a *apiKeyAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for configured path prefixes
		for _, prefix := range a.skipPaths {
			if strings.HasPrefix(r.URL.Path, prefix) {
				next.ServeHTTP(w, r)
				return
			}
		}

		key := r.Header.Get(APIKeyHeader)
		if key == "" {
			key = r.URL.Query().Get("api_key")
		}

		if key == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"missing API key"}`))
			return
		}

		if !a.validateKey(key) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"invalid API key"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// validateKey checks the provided key against all configured keys using
// constant-time comparison to prevent timing attacks.
func (a *apiKeyAuth) validateKey(key string) bool {
	for _, valid := range a.keys {
		if subtle.ConstantTimeCompare([]byte(key), []byte(valid)) == 1 {
			return true
		}
	}
	return false
}
