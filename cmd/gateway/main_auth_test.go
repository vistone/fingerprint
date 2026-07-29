package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	gatewaypkg "github.com/vistone/fingerprint/modules/gateway"
)

func TestRegisterRoutes_AdminRequiresAPIKey(t *testing.T) {
	t.Setenv("FP_API_KEYS", "test-key-123")

	cfg := buildGatewayConfig(8080)
	gw := gatewaypkg.NewGateway(&cfg)
	mux := http.NewServeMux()
	registerRoutes(mux, gw, &runtimeModules{}, &cfg)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated admin request to be rejected with 401, got %d", rec.Code)
	}
}

func TestRegisterRoutes_HealthSkipsAPIKey(t *testing.T) {
	t.Setenv("FP_API_KEYS", "test-key-123")

	cfg := buildGatewayConfig(8080)
	gw := gatewaypkg.NewGateway(&cfg)
	mux := http.NewServeMux()
	registerRoutes(mux, gw, &runtimeModules{}, &cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected health endpoint to skip auth, got %d", rec.Code)
	}
}

func TestBuildGatewayConfig_ReadsAPIKeysFromEnv(t *testing.T) {
	t.Setenv("FP_API_KEYS", " key-a ,key-b ,, ")

	cfg := buildGatewayConfig(8080)

	if len(cfg.APIKeys) != 2 {
		t.Fatalf("expected 2 API keys, got %d", len(cfg.APIKeys))
	}

	if cfg.APIKeys[0] != "key-a" || cfg.APIKeys[1] != "key-b" {
		t.Fatalf("unexpected API keys: %#v", cfg.APIKeys)
	}
}

func TestMain(m *testing.M) {
	_ = os.Unsetenv("FP_API_KEYS")
	os.Exit(m.Run())
}
