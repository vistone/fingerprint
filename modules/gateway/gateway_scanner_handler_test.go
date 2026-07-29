package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestV8ScannerHandler_ErrorResponseIsValidJSON(t *testing.T) {
	g := &Gateway{config: DefaultGatewayConfig}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scan", strings.NewReader(`{"url":"http://[::1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	g.V8ScannerHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, w.Code, w.Body.String())
	}

	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response is not valid JSON: %v; body=%s", err, w.Body.String())
	}

	if _, ok := payload["error"]; !ok {
		t.Fatalf("expected error field in response, got: %v", payload)
	}
}

func TestV8ScannerHandler_RejectsBlockedSSRFURL(t *testing.T) {
	g := &Gateway{config: DefaultGatewayConfig}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scan", strings.NewReader(`{"url":"http://127.0.0.1:8080"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	g.V8ScannerHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, w.Code, w.Body.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response is not valid JSON: %v; body=%s", err, w.Body.String())
	}

	if _, ok := payload["error"]; !ok {
		t.Fatalf("expected error field in response, got: %v", payload)
	}
}
