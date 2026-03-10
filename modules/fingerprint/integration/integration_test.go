package fingerprint_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	tls "github.com/bogdanfinn/utls"

	"github.com/vistone/fingerprint/modules/generator/random"
	"github.com/vistone/fingerprint/modules/http/legacy/headers"
	"github.com/vistone/fingerprint/modules/http/legacy/useragent"
	"github.com/vistone/fingerprint/modules/errors"
	"github.com/vistone/fingerprint/modules/profiles/legacy"
	"github.com/vistone/fingerprint/modules/core/types"
)

// translated comment
func TestGetRandomFingerprintIntegration(t *testing.T) {
	result, err := random.GetRandomFingerprint()
	if err != nil {
		t.Fatalf("get随机指纹failed: %v", err)
	}

	// translated comment
	if result.Profile.GetClientHelloStr() == "" {
		t.Error("ClientHelloStr 不能为空")
	}
	if result.UserAgent == "" {
		t.Error("UserAgent 不能为空")
	}
	if result.HelloClientID == "" {
		t.Error("HelloClientID 不能为空")
	}
	if result.Headers == nil {
		t.Error("Headers 不能为 nil")
	}

	// translated comment
	headers := result.Headers.ToMap()
	if len(headers) == 0 {
		t.Error("Headers map 不能为空")
	}

	// translated comment
	if _, ok := headers["User-Agent"]; !ok {
		t.Error("Headers must包含 User-Agent")
	}
	if _, ok := headers["Accept"]; !ok {
		t.Error("Headers must包含 Accept")
	}
	if _, ok := headers["Accept-Language"]; !ok {
		t.Error("Headers must包含 Accept-Language")
	}

	t.Logf("successget随机指纹: %s", result.HelloClientID)
}

// translated comment
func TestGetRandomFingerprintByBrowserIntegration(t *testing.T) {
	browsers := []string{"chrome", "firefox", "safari", "opera"}

	for _, browser := range browsers {
		t.Run(browser, func(t *testing.T) {
			result, err := random.GetRandomFingerprintByBrowser(browser)
			if err != nil {
				t.Fatalf("get %s 指纹failed: %v", browser, err)
			}

			if result.Profile.GetClientHelloStr() == "" {
				t.Errorf("%s: ClientHelloStr 不能为空", browser)
			}
			if result.UserAgent == "" {
				t.Errorf("%s: UserAgent 不能为空", browser)
			}

			t.Logf("%s: %s", browser, result.HelloClientID)
		})
	}
}

// translated comment
func TestGetRandomFingerprintWithOSIntegration(t *testing.T) {
	oses := []types.OperatingSystem{
		types.OSWindows10,
		types.OSMacOS14,
		types.OSLinux,
	}

	for _, os := range oses {
		t.Run(string(os), func(t *testing.T) {
			result, err := random.GetRandomFingerprintWithOS(os)
			if err != nil {
				t.Fatalf("get指纹failed: %v", err)
			}

			if result.Profile.GetClientHelloStr() == "" {
				t.Error("ClientHelloStr 不能为空")
			}
			if result.UserAgent == "" {
				t.Error("UserAgent 不能为空")
			}

			t.Logf("OS: %s, Fingerprint: %s", os, result.HelloClientID)
		})
	}
}

// translated comment
func TestHeadersCustomizationIntegration(t *testing.T) {
	result, err := random.GetRandomFingerprint()
	if err != nil {
		t.Fatalf("get随机指纹failed: %v", err)
	}

	// translated comment
	result.Headers.Set("Cookie", "session_id=test123")
	result.Headers.Set("Authorization", "Bearer token456")
	result.Headers.Set("X-Custom-Header", "custom-value")

	headers := result.Headers.ToMap()

	// translated comment
	if cookie, ok := headers["Cookie"]; !ok || cookie != "session_id=test123" {
		t.Error("Cookie header settingfailed")
	}
	if auth, ok := headers["Authorization"]; !ok || auth != "Bearer token456" {
		t.Error("Authorization header settingfailed")
	}
	if custom, ok := headers["X-Custom-Header"]; !ok || custom != "custom-value" {
		t.Error("X-Custom-Header settingfailed")
	}

	// translated comment
	result.Headers.SetHeaders(map[string]string{
		"Cookie":       "session_id=updated",
		"X-API-Key":    "api-key-123",
		"X-Request-ID": "req-123",
	})

	headers = result.Headers.ToMap()
	if cookie, ok := headers["Cookie"]; !ok || cookie != "session_id=updated" {
		t.Error("批量setting Cookie failed")
	}
	if apiKey, ok := headers["X-API-Key"]; !ok || apiKey != "api-key-123" {
		t.Error("批量setting X-API-Key failed")
	}
}

// translated comment
func TestHeadersCloneIntegration(t *testing.T) {
	result, err := random.GetRandomFingerprint()
	if err != nil {
		t.Fatalf("get随机指纹failed: %v", err)
	}

	// translated comment
	result.Headers.Set("Cookie", "original")
	result.Headers.Set("X-Test", "value")

	// translated comment
	cloned := result.Headers.Clone()
	if cloned == nil {
		t.Fatal("克隆failed，return nil")
	}

	// translated comment
	cloned.Set("Cookie", "modified")
	cloned.Set("X-New", "new-value")

	// translated comment
	originalHeaders := result.Headers.ToMap()
	if cookie, ok := originalHeaders["Cookie"]; !ok || cookie != "original" {
		t.Error("原对象的 Cookie 被意外modify")
	}
	if _, ok := originalHeaders["X-New"]; ok {
		t.Error("原对象不应该包含 X-New")
	}

	// translated comment
	clonedHeaders := cloned.ToMap()
	if cookie, ok := clonedHeaders["Cookie"]; !ok || cookie != "modified" {
		t.Error("克隆对象的 Cookie 未正确modify")
	}
	if newVal, ok := clonedHeaders["X-New"]; !ok || newVal != "new-value" {
		t.Error("克隆对象未正确添加 X-New")
	}
}

// translated comment
func TestTLSClientHelloIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip集成测试（使用 -short 标志）")
	}

	testProfiles := []string{"chrome_133", "firefox_135", "safari_ios_18_0"}

	for _, profileName := range testProfiles {
		t.Run(profileName, func(t *testing.T) {
			profile, ok := profiles.MappedTLSClients[profileName]
			if !ok {
				t.Fatalf("Profile %s 不存在", profileName)
			}

			spec, err := profile.GetClientHelloSpec()
			if err != nil {
				if errors.IsClientHelloSpecNotImplemented(err) {
					t.Skipf("Profile %s 使用预定义 ID，skip测试", profileName)
					return
				}
				t.Fatalf("get ClientHelloSpec failed: %v", err)
			}

			// translated comment
			if len(spec.CipherSuites) == 0 {
				t.Error("CipherSuites 不能为空")
			}

			t.Logf("Profile %s: %d cipher suites, %d extensions",
				profileName, len(spec.CipherSuites), len(spec.Extensions))
		})
	}
}

// translated comment
func TestConcurrentAccess(t *testing.T) {
	const goroutines = 100
	const iterations = 10

	done := make(chan bool, goroutines)
	errors := make(chan error, goroutines*iterations)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer func() { done <- true }()

			for j := 0; j < iterations; j++ {
				// translated comment
				result, err := random.GetRandomFingerprint()
				if err != nil {
					errors <- fmt.Errorf("GetRandomFingerprint failed: %v", err)
					continue
				}

				// translated comment
				result.Headers.Set("Cookie", fmt.Sprintf("test_%d_%d", i, j))
				_ = result.Headers.ToMap()

				// translated comment
				_ = result.Headers.Clone()

				// translated comment
				_ = headers.RandomLanguage()

				// translated comment
				_ = useragent.RandomOS()
			}
		}()
	}

	// translated comment
	for i := 0; i < goroutines; i++ {
		<-done
	}

	close(errors)

	// translated comment
	var errCount int
	for err := range errors {
		t.Error(err)
		errCount++
		if errCount >= 10 {
			t.Fatal("concurrent测试error过多，stop测试")
		}
	}

	if errCount > 0 {
		t.Fatalf("concurrent测试发现 %d 个error", errCount)
	}

	t.Logf("concurrent测试通过：%d goroutines × %d iterations = %d 次操作",
		goroutines, iterations, goroutines*iterations)
}

// translated comment
func TestRealTLSConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skip网络测试（使用 -short 标志）")
	}

	result, err := random.GetRandomFingerprint()
	if err != nil {
		t.Fatalf("get随机指纹failed: %v", err)
	}

	spec, err := result.Profile.GetClientHelloSpec()
	if err != nil {
		if errors.IsClientHelloSpecNotImplemented(err) {
			t.Skip("skip预定义 ID 的 TLS connect测试")
			return
		}
		t.Fatalf("get ClientHelloSpec failed: %v", err)
	}

	// create TLS configuration
	tlsConfig := &tls.Config{
		ServerName: "www.google.com",
	}

	// translated comment
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}

	conn, err := dialer.DialContext(ctx, "tcp", "www.google.com:443")
	if err != nil {
		t.Skipf("无法connect到测试服务器: %v", err)
		return
	}
	defer conn.Close()

	// translated comment
	tlsConn := tls.UClient(conn, tlsConfig, result.Profile.GetClientHelloId(), false, false, false)
	if err := tlsConn.ApplyPreset(&spec); err != nil {
		t.Fatalf("应用 TLS preset failed: %v", err)
	}

	// translated comment
	if err := tlsConn.Handshake(); err != nil {
		if strings.Contains(err.Error(), "empty psk detected") {
			t.Skipf("skip当前随机指纹的 TLS 握手测试（uTLS PSK limit）: %v", err)
			return
		}
		t.Fatalf("TLS 握手failed: %v", err)
	}

	// translated comment
	request := fmt.Sprintf("GET / HTTP/1.1\r\nHost: www.google.com\r\nUser-Agent: %s\r\nConnection: close\r\n\r\n", result.UserAgent)
	if _, err := tlsConn.Write([]byte(request)); err != nil {
		t.Fatalf("sendrequestfailed: %v", err)
	}

	// translated comment
	response := make([]byte, 1024)
	n, err := tlsConn.Read(response)
	if err != nil && err != io.EOF {
		t.Fatalf("读取responsefailed: %v", err)
	}

	if n == 0 {
		t.Fatal("未收到response")
	}

	t.Logf("success建立 TLS connect并收到response: %d bytes", n)
	t.Logf("response预览: %s", string(response[:min(200, n)]))
}

// translated comment
func TestAllProfilesWithUserAgent(t *testing.T) {
	failCount := 0
	successCount := 0

	for name := range profiles.MappedTLSClients {
		t.Run(name, func(t *testing.T) {
			ua, err := useragent.GetUserAgentByProfileName(name)
			if err != nil {
				t.Errorf("Profile %s: get User-Agent failed: %v", name, err)
				failCount++
				return
			}

			if ua == "" {
				t.Errorf("Profile %s: User-Agent 不能为空", name)
				failCount++
				return
			}

			// verify User-Agent format
			if len(ua) < 20 {
				t.Errorf("Profile %s: User-Agent format可能不正确: %s", name, ua)
				failCount++
				return
			}

			successCount++
			t.Logf("Profile %s: %s", name, ua)
		})
	}

	t.Logf("User-Agent 测试complete: success %d, failed %d", successCount, failCount)

	if failCount > 0 {
		t.Errorf("有 %d 个 profile 的 User-Agent generatefailed", failCount)
	}
}

// translated comment
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
