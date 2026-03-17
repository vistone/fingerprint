package gateway

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"
)

func TestGRPCService_TLS(t *testing.T) {
	caCert, caKey := generateTestCA(t)
	serverCert := generateSignedCert(t, caCert, caKey, true, false)

	rootPool := x509.NewCertPool()
	rootPool.AddCert(caCert)

	gw := NewGateway(nil)
	defer gw.Close()

	srv := NewGRPCServerWithTLS(gw, &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{serverCert},
	})

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer lis.Close()

	go func() {
		_ = srv.server.Serve(lis)
	}()
	defer srv.Stop()

	client, err := NewGRPCClientWithTLS(lis.Addr().String(), &tls.Config{
		RootCAs:    rootPool,
		MinVersion: tls.VersionTLS12,
		ServerName: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("tls client connect failed: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	health, err := client.Health(ctx)
	if err != nil {
		t.Fatalf("health failed: %v", err)
	}
	if health.GetStatus() != "SERVING" {
		t.Fatalf("unexpected health status: %q", health.GetStatus())
	}
}

func TestGRPCService_MTLS(t *testing.T) {
	caCert, caKey := generateTestCA(t)
	serverCert := generateSignedCert(t, caCert, caKey, true, false)
	clientCert := generateSignedCert(t, caCert, caKey, false, true)

	rootPool := x509.NewCertPool()
	rootPool.AddCert(caCert)

	gw := NewGateway(nil)
	defer gw.Close()

	srv := NewGRPCServerWithTLS(gw, &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    rootPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	})

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	defer lis.Close()

	go func() {
		_ = srv.server.Serve(lis)
	}()
	defer srv.Stop()

	goodClient, err := NewGRPCClientWithTLS(lis.Addr().String(), &tls.Config{
		RootCAs:      rootPool,
		Certificates: []tls.Certificate{clientCert},
		MinVersion:   tls.VersionTLS12,
		ServerName:   "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("mtls client connect failed: %v", err)
	}
	defer goodClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := goodClient.Health(ctx); err != nil {
		t.Fatalf("mtls health failed: %v", err)
	}

	badClient, err := NewGRPCClientWithTLS(lis.Addr().String(), &tls.Config{
		RootCAs:    rootPool,
		MinVersion: tls.VersionTLS12,
		ServerName: "127.0.0.1",
	})
	if err == nil {
		defer badClient.Close()
		_, callErr := badClient.Health(ctx)
		if callErr == nil {
			t.Fatal("expected mTLS health call to fail without client certificate")
		}
		return
	}
}

func generateTestCA(t *testing.T) (*x509.Certificate, *rsa.PrivateKey) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate CA key failed: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("generate CA cert failed: %v", err)
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse CA cert failed: %v", err)
	}
	return cert, key
}

func generateSignedCert(t *testing.T, caCert *x509.Certificate, caKey *rsa.PrivateKey, isServer bool, isClient bool) tls.Certificate {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key failed: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: "127.0.0.1",
		},
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{},
	}
	if isServer {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
		template.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	}
	if isClient {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	}

	der, err := x509.CreateCertificate(rand.Reader, template, caCert, &key.PublicKey, caKey)
	if err != nil {
		t.Fatalf("generate cert failed: %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("load key pair failed: %v", err)
	}
	return cert
}
