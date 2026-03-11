package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKeyAuth_ValidKey(t *testing.T) {
	auth := NewAPIKeyAuth([]string{"test-key-123"}, nil)
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analyze", nil)
	req.Header.Set(APIKeyHeader, "test-key-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	auth := NewAPIKeyAuth([]string{"test-key-123"}, nil)
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analyze", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	auth := NewAPIKeyAuth([]string{"test-key-123"}, nil)
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analyze", nil)
	req.Header.Set(APIKeyHeader, "wrong-key")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestAPIKeyAuth_SkipPaths(t *testing.T) {
	auth := NewAPIKeyAuth([]string{"test-key-123"}, []string{"/health", "/metrics"})
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, path := range []string{"/health", "/metrics"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("path %q: expected 200 (skip), got %d", path, rec.Code)
		}
	}
}

func TestAPIKeyAuth_QueryParam(t *testing.T) {
	auth := NewAPIKeyAuth([]string{"test-key-123"}, nil)
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analyze?api_key=test-key-123", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 via query param, got %d", rec.Code)
	}
}

func TestAPIKeyAuth_MultipleKeys(t *testing.T) {
	auth := NewAPIKeyAuth([]string{"key-a", "key-b"}, nil)
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, key := range []string{"key-a", "key-b"} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/analyze", nil)
		req.Header.Set(APIKeyHeader, key)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("key %q: expected 200, got %d", key, rec.Code)
		}
	}
}
