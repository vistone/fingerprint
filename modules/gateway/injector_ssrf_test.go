package gateway

import (
	"net/url"
	"testing"
)

func TestValidateProxyTarget_RejectsLoopback(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"localhost", "http://localhost:8080"},
		{"127.0.0.1", "http://127.0.0.1:80"},
		{"metadata", "http://metadata.google.internal"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.raw)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if err := validateProxyTarget(u, false); err == nil {
				t.Errorf("expected rejection for %q, got nil", tc.raw)
			}
		})
	}
}

func TestValidateProxyTarget_AllowsPublic(t *testing.T) {
	u, err := url.Parse("http://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateProxyTarget(u, false); err != nil {
		t.Errorf("expected allow for example.com, got: %v", err)
	}
}

func TestValidateProxyTarget_RejectsEmptyHost(t *testing.T) {
	u, err := url.Parse("/relative-path")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateProxyTarget(u, false); err == nil {
		t.Error("expected rejection for empty host")
	}
}

func TestValidateProxyTarget_AllowPrivateWhenEnabled(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"10.0.0.1", "http://10.0.0.1:8080"},
		{"172.17.0.2", "http://172.17.0.2:80"},
		{"192.168.1.1", "http://192.168.1.1:443"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.raw)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if err := validateProxyTarget(u, true); err != nil {
				t.Errorf("expected allow for %q with allowPrivate=true, got: %v", tc.raw, err)
			}
		})
	}
}
