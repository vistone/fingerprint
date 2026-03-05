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

// TestServer 包装 httptest.Server，提供便捷的测试服务器功能
type TestServer struct {
	*httptest.Server
	handlers map[string]func(http.ResponseWriter, *http.Request)
	tlsConf  *tls.Config
}

// NewTestServer 创建一个新的本地测试服务器
// 使用 httptest 避免真实网络连接的依赖
func NewTestServer(routes map[string]func(http.ResponseWriter, *http.Request)) *TestServer {
	ts := &TestServer{
		handlers: routes,
	}

	mux := http.NewServeMux()

	// 注册所有路由
	for path, handler := range routes {
		mux.HandleFunc(path, handler)
	}

	ts.Server = httptest.NewServer(mux)
	return ts
}

// NewTestServerTLS 创建一个支持 TLS 的测试服务器（自签证书）
func NewTestServerTLS(routes map[string]func(http.ResponseWriter, *http.Request)) (*TestServer, error) {
	ts := &TestServer{
		handlers: routes,
	}

	mux := http.NewServeMux()

	// 注册所有路由
	for path, handler := range routes {
		mux.HandleFunc(path, handler)
	}

	// 创建自签TLS证书
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

	// 创建 HTTPS 测试服务器
	ts.Server = httptest.NewUnstartedServer(mux)
	ts.Server.TLS = tlsConf
	ts.Server.StartTLS()

	return ts, nil
}

// RegisterRoute 动态注册新的路由处理器
func (ts *TestServer) RegisterRoute(path string, handler func(http.ResponseWriter, *http.Request)) {
	ts.handlers[path] = handler
	// 注意：httptest.Server 不支持动态添加路由
	// 这里仅作为演示，实际应用需要重新创建服务器
}

// URL 返回测试服务器的 URL
func (ts *TestServer) URL() string {
	return ts.Server.URL
}

// Close 关闭测试服务器
func (ts *TestServer) Close() {
	if ts.Server != nil {
		ts.Server.Close()
	}
}

// GetTestResponse 从测试服务器获取响应（用于集成测试）
func (ts *TestServer) GetTestResponse(path string) (*http.Response, error) {
	return http.Get(ts.URL() + path)
}

// Listener 返回测试服务器的监听器
func (ts *TestServer) Listener() net.Listener {
	return ts.Server.Listener
}

// generateSelfSignedCert 生成自签 TLS 证书
func generateSelfSignedCert() (certPEM, keyPEM []byte, err error) {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// 创建证书模板
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

	// 自签证书
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	// 编码为 PEM 格式
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

// RecordHTTPInteraction 记录 HTTP 交互（用于 golden file 测试）
type RecordHTTPInteraction struct {
	Method   string              `json:"method"`
	Path     string              `json:"path"`
	Headers  map[string][]string `json:"headers"`
	Body     string              `json:"body,omitempty"`
	Status   int                 `json:"status"`
	Response string              `json:"response"`
}

// RecordTLSInteraction 记录 TLS 握手交互（用于指纹测试）
type RecordTLSInteraction struct {
	ClientHello string `json:"client_hello"` // 十六进制编码
	ServerHello string `json:"server_hello"` // 十六进制编码
	Certificate string `json:"certificate"`  // PEM 格式
	Cipher      uint16 `json:"cipher"`       // TLS cipher suite ID
	Version     string `json:"version"`      // TLS version
}

// MockTLSConnection 为单元测试模拟预录制的 TLS 连接
type MockTLSConnection struct {
	recorded  *RecordTLSInteraction
	readIdx   int
	readBuf   []byte
	clientVer uint16
}

// NewMockTLSConnection 创建一个模拟 TLS 连接
func NewMockTLSConnection(recorded *RecordTLSInteraction) *MockTLSConnection {
	return &MockTLSConnection{
		recorded: recorded,
	}
}

// Read 实现 net.Conn.Read() 接口
func (m *MockTLSConnection) Read(b []byte) (int, error) {
	// 返回预录制的服务器响应
	if m.readIdx >= len(m.readBuf) {
		return 0, nil // EOF
	}
	n := copy(b, m.readBuf[m.readIdx:])
	m.readIdx += n
	return n, nil
}

// Write 实现 net.Conn.Write() 接口
func (m *MockTLSConnection) Write(b []byte) (int, error) {
	// 在模拟模式下，只需记录写入数据
	m.clientVer = extractTLSVersionFromClientHello(b)
	// 准备返回预录制的服务器响应
	// m.readBuf = decodeHex(m.recorded.ServerHello)
	return len(b), nil
}

// Close 实现 net.Conn.Close() 接口
func (m *MockTLSConnection) Close() error {
	return nil
}

// LocalAddr 实现 net.Conn.LocalAddr() 接口
func (m *MockTLSConnection) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr 实现 net.Conn.RemoteAddr() 接口
func (m *MockTLSConnection) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline 实现 net.Conn.SetDeadline() 接口
func (m *MockTLSConnection) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline 实现 net.Conn.SetReadDeadline() 接口
func (m *MockTLSConnection) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline 实现 net.Conn.SetWriteDeadline() 接口
func (m *MockTLSConnection) SetWriteDeadline(t time.Time) error {
	return nil
}

// extractTLSVersionFromClientHello 从 ClientHello 消息中提取 TLS 版本
func extractTLSVersionFromClientHello(data []byte) uint16 {
	if len(data) < 5 {
		return 0
	}
	// TLS record header: type(1) + version(2) + length(2) + payload
	// ClientHello: version(2) + random(32) + ...
	if len(data) < 7 {
		return 0
	}
	// 返回 ClientHello 中的版本字段
	return uint16(data[4])<<8 | uint16(data[5])
}

// AssertHTTPHandlerBehavior 验证 HTTP 处理器的行为（单元测试辅助）
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

// AssertTLSFingerprint 验证 TLS 指纹（单元测试辅助）
func AssertTLSFingerprint(t *testing.T, actual *tls_utls.ClientHelloSpec,
	expectedCiphers []uint16, expectedExtensions []tls_utls.TLSExtension) {

	if len(actual.CipherSuites) != len(expectedCiphers) {
		t.Errorf("expected %d ciphers, got %d", len(expectedCiphers), len(actual.CipherSuites))
	}

	// 更多验证逻辑...
}
