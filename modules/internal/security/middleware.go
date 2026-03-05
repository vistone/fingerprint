package security

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Middleware provides HTTP security middleware
type Middleware struct {
	config    *SecurityConfig
	validator *Validator
	logger    *AuditLogger
	limiter   *RateLimiter
}

// RateLimiter provides rate limiting
type RateLimiter struct {
	requests map[string]*RateLimitEntry
	maxReq   int
	window   time.Duration
}

// RateLimitEntry tracks rate limit for a client
type RateLimitEntry struct {
	Count     int
	ResetTime time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxRequests int, windowSeconds int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*RateLimitEntry),
		maxReq:   maxRequests,
		window:   time.Duration(windowSeconds) * time.Second,
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(clientID string) bool {
	now := time.Now()
	
	entry, exists := rl.requests[clientID]
	if !exists || now.After(entry.ResetTime) {
		rl.requests[clientID] = &RateLimitEntry{
			Count:     1,
			ResetTime: now.Add(rl.window),
		}
		return true
	}

	if entry.Count >= rl.maxReq {
		return false
	}

	entry.Count++
	return true
}

// NewMiddleware creates new security middleware
func NewMiddleware(config *SecurityConfig) *Middleware {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	return &Middleware{
		config:    config,
		validator: NewValidator(),
		limiter:   NewRateLimiter(config.RateLimitRequests, config.RateLimitWindow),
	}
}

// SetAuditLogger sets the audit logger
func (m *Middleware) SetAuditLogger(logger *AuditLogger) {
	m.logger = logger
}

// SecureHeaders adds security headers to responses
func (m *Middleware) SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}

// RateLimit enforces rate limiting
func (m *Middleware) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.config.RateLimitEnabled {
			next.ServeHTTP(w, r)
			return
		}

		clientID := m.getClientIdentifier(r)

		if !m.limiter.Allow(clientID) {
			if m.logger != nil {
				m.logger.Warning("rate_limit", "check", "api", "blocked", map[string]interface{}{
					"client_id": clientID,
					"path":      r.URL.Path,
				})
			}

			w.Header().Set("Retry-After", fmt.Sprintf("%d", m.config.RateLimitWindow))
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ValidateRequest validates incoming requests
func (m *Middleware) ValidateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.config.ValidateInput {
			next.ServeHTTP(w, r)
			return
		}

		// Validate request size
		if r.ContentLength > m.config.MaxRequestSize {
			if m.logger != nil {
				m.logger.Warning("request", "validate", "api", "rejected", map[string]interface{}{
					"reason":          "content_too_large",
					"content_length":  r.ContentLength,
					"max_size":        m.config.MaxRequestSize,
				})
			}
			http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
			return
		}

		// Validate headers count
		if len(r.Header) > m.config.MaxHeaders {
			http.Error(w, "Too many headers", http.StatusBadRequest)
			return
		}

		// Validate User-Agent
		if ua := r.UserAgent(); ua != "" {
			if err := m.validator.ValidateUserAgent(ua); err != nil {
				if m.logger != nil {
					m.logger.Warning("request", "validate", "api", "rejected", map[string]interface{}{
						"reason":       "invalid_user_agent",
						"user_agent":   ua,
					})
				}
				http.Error(w, "Invalid User-Agent", http.StatusBadRequest)
				return
			}
		}

		// Validate header names and values
		for name, values := range r.Header {
			if err := m.validator.ValidateHeaderName(name); err != nil {
				http.Error(w, "Invalid header name", http.StatusBadRequest)
				return
			}

			for _, value := range values {
				if err := m.validator.ValidateHeaderValue(value); err != nil {
					http.Error(w, "Invalid header value", http.StatusBadRequest)
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// HostValidation validates the Host header
func (m *Middleware) HostValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(m.config.AllowedHosts) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		host := r.Host
		if !m.config.IsHostAllowed(host) {
			if m.logger != nil {
				m.logger.Warning("host", "validate", "api", "rejected", map[string]interface{}{
					"host": host,
				})
			}
			http.Error(w, "Invalid host", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// IPBlocklist checks against blocked IPs
func (m *Middleware) IPBlocklist(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := m.getClientIP(r)

		if m.config.IsIPBlocked(clientIP) {
			if m.logger != nil {
				m.logger.Warning("ip", "check", "api", "blocked", map[string]interface{}{
					"client_ip": clientIP,
				})
			}
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuditLogging logs requests to audit log
func (m *Middleware) AuditLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		if m.logger != nil && m.logger.IsEnabled() {
			level := Info
			if rw.statusCode >= 500 {
				level = Error
			} else if rw.statusCode >= 400 {
				level = Warning
			}

			result := "success"
			if rw.statusCode >= 400 {
				result = "failure"
			}

			m.logger.Log(level, "http", r.Method, r.URL.Path, result, map[string]interface{}{
				"client_ip":   m.getClientIP(r),
				"user_agent":  r.UserAgent(),
				"status_code": rw.statusCode,
				"duration_ms": duration.Milliseconds(),
				"request_size": r.ContentLength,
			})
		}
	})
}

// RequireTLS requires TLS connections
func (m *Middleware) RequireTLS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.config.RequireTLS && r.TLS == nil {
			if m.logger != nil {
				m.logger.Warning("tls", "require", "api", "rejected", map[string]interface{}{
					"client_ip": m.getClientIP(r),
				})
			}
			http.Error(w, "TLS required", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequestID adds a request ID to the context
func (m *Middleware) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add to response header
		w.Header().Set("X-Request-ID", requestID)

		// Add to context
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Chain chains multiple middleware together
func (m *Middleware) Chain(handlers ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(handlers) - 1; i >= 0; i-- {
			final = handlers[i](final)
		}
		return final
	}
}

// getClientIdentifier returns a unique identifier for the client
func (m *Middleware) getClientIdentifier(r *http.Request) string {
	// Try X-Forwarded-For first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	// Fall back to X-Real-Ip
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// getClientIP extracts the client IP
func (m *Middleware) getClientIP(r *http.Request) string {
	ip := m.getClientIdentifier(r)
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Write writes the response
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// ApplyAll applies all security middleware
func (m *Middleware) ApplyAll(handler http.Handler) http.Handler {
	chain := m.Chain(
		m.RequestID,
		m.SecureHeaders,
		m.IPBlocklist,
		m.HostValidation,
		m.RequireTLS,
		m.ValidateRequest,
		m.RateLimit,
		m.AuditLogging,
	)
	return chain(handler)
}
