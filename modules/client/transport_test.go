package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

func TestRoundTripHTTP1Compat_RejectsInvalidCertificateByDefault(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport, err := NewSmartTransport(profiles.ClientProfile{
		BrowserType: core.BrowserChrome,
		TLSVersion:  0x0304,
	})
	if err != nil {
		t.Fatalf("NewSmartTransport returned error: %v", err)
	}

	req, err := fhttp.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	resp, err := transport.roundTripHTTP1Compat(context.Background(), req)
	if err == nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		t.Fatal("expected TLS verification failure for self-signed certificate, got success")
	}
}
