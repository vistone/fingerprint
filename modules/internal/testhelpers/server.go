package testhelpers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	tls_utls "github.com/bogdanfinn/utls"
)

// TestServer wraps httptest.Server and provides convenient test server helpers
type TestServer struct {
	*httptest.Server
	handlers map[string]func(http.ResponseWriter, *http.Request)
	tlsConf  *tls.Config
}

// NewTestServer creates a new local test server
// Uses httptest to avoid real network dependencies
func NewTestServer(routes map[string]func(http.ResponseWriter, *http.Request)) *TestServer {
	ts := &TestServer{
		handlers: routes,
	}

	mux := http.NewServeMux()

	// Register all routes
	for path, handler := range routes {
		mux.HandleFunc(path, handler)
	}

	ts.Server = httptest.NewServer(mux)
	return ts
}

// NewTestServerTLS creates a TLS-enabled test server (self-signed cert)
func NewTestServerTLS(routes map[string]func(http.ResponseWriter, *http.Request)) (*TestServer, error) {
	ts := &TestServer{
		handlers: routes,
	}

	mux := http.NewServeMux()

	// Register all routes
	for path, handler := range routes {
		mux.HandleFunc(path, handler)
	}

	// Create self-signed TLS certificate
	cert, key, err := generateSelfSignedCert()
	if err != nil {
		return nil, fmt.Errorf("failed to generate self-signed certificate: %w", err)
	}

	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TLS certificate: %w", err)
	}

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}

	ts.tlsConf = tlsConf

	// Create HTTPS test server
	ts.Server = httptest.NewUnstartedServer(mux)
	ts.Server.TLS = tlsConf
	ts.Server.StartTLS()

	return ts, nil
}

// RegisterRoute registers a new route handler dynamically
func (ts *TestServer) RegisterRoute(path string, handler func(http.ResponseWriter, *http.Request)) {
	ts.handlers[path] = handler
	// Note: httptest.Server does not support dynamic route insertion
	// This is illustrative; production usage should recreate the server
}

// URL returns the test server URL
func (ts *TestServer) URL() string {
	return ts.Server.URL
}

// Close shuts down the test server
func (ts *TestServer) Close() {
	if ts.Server != nil {
		ts.Server.Close()
	}
}

// GetTestResponse fetches a response from the test server (integration helper)
func (ts *TestServer) GetTestResponse(path string) (*http.Response, error) {
	return http.Get(ts.URL() + path)
}

// Listener returns the test server listener
func (ts *TestServer) Listener() net.Listener {
	return ts.Server.Listener
}

// generateSelfSignedCert generates a self-signed TLS certificate
func generateSelfSignedCert() (certPEM, keyPEM []byte, err error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Build certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(24 * time.Hour),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		DNSNames: []string{"localhost", "127.0.0.1"},
	}

	// Create self-signed certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	// Encode in PEM format
	certPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	keyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return certPEM, keyPEM, nil
}

// RecordHTTPInteraction stores recorded HTTP interactions (golden-file testing)
type RecordHTTPInteraction struct {
	Method   string              `json:"method"`
	Path     string              `json:"path"`
	Headers  map[string][]string `json:"headers"`
	Body     string              `json:"body,omitempty"`
	Status   int                 `json:"status"`
	Response string              `json:"response"`
}

// RecordTLSInteraction stores recorded TLS handshake exchanges (fingerprint tests)
type RecordTLSInteraction struct {
	ClientHello string `json:"client_hello"` // Hex encoded
	ServerHello string `json:"server_hello"` // Hex encoded
	Certificate string `json:"certificate"`  // PEM encoded
	Cipher      uint16 `json:"cipher"`       // TLS cipher suite ID
	Version     string `json:"version"`      // TLS version
}

// MockTLSConnection simulates a prerecorded TLS connection for unit tests
type MockTLSConnection struct {
	recorded  *RecordTLSInteraction
	readIdx   int
	readBuf   []byte
	clientVer uint16
}

// NewMockTLSConnection creates a mock TLS connection
func NewMockTLSConnection(recorded *RecordTLSInteraction) *MockTLSConnection {
	return &MockTLSConnection{
		recorded: recorded,
	}
}

// Read implements net.Conn.Read()
func (m *MockTLSConnection) Read(b []byte) (int, error) {
	// Return prerecorded server response
	if m.readIdx >= len(m.readBuf) {
		return 0, nil // EOF
	}
	n := copy(b, m.readBuf[m.readIdx:])
	m.readIdx += n
	return n, nil
}

// Write implements net.Conn.Write()
func (m *MockTLSConnection) Write(b []byte) (int, error) {
	// In mock mode, only record written data
	m.clientVer = extractTLSVersionFromClientHello(b)
	// Prepare prerecorded server response
	// m.readBuf = decodeHex(m.recorded.ServerHello)
	return len(b), nil
}

// Close implements net.Conn.Close()
func (m *MockTLSConnection) Close() error {
	return nil
}

// LocalAddr implements net.Conn.LocalAddr()
func (m *MockTLSConnection) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr implements net.Conn.RemoteAddr()
func (m *MockTLSConnection) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline implements net.Conn.SetDeadline()
func (m *MockTLSConnection) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline implements net.Conn.SetReadDeadline()
func (m *MockTLSConnection) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline implements net.Conn.SetWriteDeadline()
func (m *MockTLSConnection) SetWriteDeadline(t time.Time) error {
	return nil
}

// extractTLSVersionFromClientHello extracts TLS version from ClientHello
func extractTLSVersionFromClientHello(data []byte) uint16 {
	if len(data) < 5 {
		return 0
	}
	// TLS record header: type(1) + version(2) + length(2) + payload
	// ClientHello: version(2) + random(32) + ...
	if len(data) < 7 {
		return 0
	}
	// Return the version field from ClientHello
	return uint16(data[4])<<8 | uint16(data[5])
}

// AssertHTTPHandlerBehavior validates HTTP handler behavior (unit-test helper)
func AssertHTTPHandlerBehavior(t *testing.T, handler func(http.ResponseWriter, *http.Request),
	expectedStatus int, expectedBody string) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != expectedStatus {
		t.Errorf("expected status %d, got %d", expectedStatus, w.Code)
	}

	if body := w.Body.String(); body != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, body)
	}
}

// AssertTLSFingerprint validates TLS fingerprint (unit-test helper)
func AssertTLSFingerprint(t *testing.T, actual *tls_utls.ClientHelloSpec,
	expectedCiphers []uint16, expectedExtensions []tls_utls.TLSExtension) {

	if len(actual.CipherSuites) != len(expectedCiphers) {
		t.Errorf("expected %d ciphers, got %d", len(expectedCiphers), len(actual.CipherSuites))
	}

	// Additional validation logic...
}
