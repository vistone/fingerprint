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

// translated comment
type TestServer struct {
	*httptest.Server
	handlers map[string]func(http.ResponseWriter, *http.Request)
	tlsConf  *tls.Config
}

// translated comment
// translated comment
func NewTestServer(routes map[string]func(http.ResponseWriter, *http.Request)) *TestServer {
	ts := &TestServer{
		handlers: routes,
	}

	mux := http.NewServeMux()

	// translated comment
	for path, handler := range routes {
		mux.HandleFunc(path, handler)
	}

	ts.Server = httptest.NewServer(mux)
	return ts
}

// translated comment
func NewTestServerTLS(routes map[string]func(http.ResponseWriter, *http.Request)) (*TestServer, error) {
	ts := &TestServer{
		handlers: routes,
	}

	mux := http.NewServeMux()

	// translated comment
	for path, handler := range routes {
		mux.HandleFunc(path, handler)
	}

	// translated comment
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

	// translated comment
	ts.Server = httptest.NewUnstartedServer(mux)
	ts.Server.TLS = tlsConf
	ts.Server.StartTLS()

	return ts, nil
}

// translated comment
func (ts *TestServer) RegisterRoute(path string, handler func(http.ResponseWriter, *http.Request)) {
	ts.handlers[path] = handler
	// translated comment
	// translated comment
}

// translated comment
func (ts *TestServer) URL() string {
	return ts.Server.URL
}

// translated comment
func (ts *TestServer) Close() {
	if ts.Server != nil {
		ts.Server.Close()
	}
}

// translated comment
func (ts *TestServer) GetTestResponse(path string) (*http.Response, error) {
	return http.Get(ts.URL() + path)
}

// translated comment
func (ts *TestServer) Listener() net.Listener {
	return ts.Server.Listener
}

// translated comment
func generateSelfSignedCert() (certPEM, keyPEM []byte, err error) {
	// translated comment
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// translated comment
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

	// translated comment
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	// translated comment
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

// translated comment
type RecordHTTPInteraction struct {
	Method   string              `json:"method"`
	Path     string              `json:"path"`
	Headers  map[string][]string `json:"headers"`
	Body     string              `json:"body,omitempty"`
	Status   int                 `json:"status"`
	Response string              `json:"response"`
}

// translated comment
type RecordTLSInteraction struct {
	ClientHello string `json:"client_hello"` // translated comment
	ServerHello string `json:"server_hello"` // translated comment
	Certificate string `json:"certificate"`  // translated comment
	Cipher      uint16 `json:"cipher"`       // TLS cipher suite ID
	Version     string `json:"version"`      // TLS version
}

// translated comment
type MockTLSConnection struct {
	recorded  *RecordTLSInteraction
	readIdx   int
	readBuf   []byte
	clientVer uint16
}

// translated comment
func NewMockTLSConnection(recorded *RecordTLSInteraction) *MockTLSConnection {
	return &MockTLSConnection{
		recorded: recorded,
	}
}

// translated comment
func (m *MockTLSConnection) Read(b []byte) (int, error) {
	// translated comment
	if m.readIdx >= len(m.readBuf) {
		return 0, nil // EOF
	}
	n := copy(b, m.readBuf[m.readIdx:])
	m.readIdx += n
	return n, nil
}

// translated comment
func (m *MockTLSConnection) Write(b []byte) (int, error) {
	// translated comment
	m.clientVer = extractTLSVersionFromClientHello(b)
	// translated comment
	// m.readBuf = decodeHex(m.recorded.ServerHello)
	return len(b), nil
}

// translated comment
func (m *MockTLSConnection) Close() error {
	return nil
}

// translated comment
func (m *MockTLSConnection) LocalAddr() net.Addr {
	return nil
}

// translated comment
func (m *MockTLSConnection) RemoteAddr() net.Addr {
	return nil
}

// translated comment
func (m *MockTLSConnection) SetDeadline(t time.Time) error {
	return nil
}

// translated comment
func (m *MockTLSConnection) SetReadDeadline(t time.Time) error {
	return nil
}

// translated comment
func (m *MockTLSConnection) SetWriteDeadline(t time.Time) error {
	return nil
}

// translated comment
func extractTLSVersionFromClientHello(data []byte) uint16 {
	if len(data) < 5 {
		return 0
	}
	// TLS record header: type(1) + version(2) + length(2) + payload
	// ClientHello: version(2) + random(32) + ...
	if len(data) < 7 {
		return 0
	}
	// translated comment
	return uint16(data[4])<<8 | uint16(data[5])
}

// translated comment
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

// translated comment
func AssertTLSFingerprint(t *testing.T, actual *tls_utls.ClientHelloSpec,
	expectedCiphers []uint16, expectedExtensions []tls_utls.TLSExtension) {

	if len(actual.CipherSuites) != len(expectedCiphers) {
		t.Errorf("expected %d ciphers, got %d", len(expectedCiphers), len(actual.CipherSuites))
	}

	// translated comment
}
