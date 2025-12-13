package fingerprint_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	tls "github.com/bogdanfinn/utls"
	"github.com/vistone/fingerprint"
)

// TestGetRandomFingerprintIntegration 集成测试：随机指纹完整流程
func TestGetRandomFingerprintIntegration(t *testing.T) {
	result, err := fingerprint.GetRandomFingerprint()
	if err != nil {
		t.Fatalf("获取随机指纹失败: %v", err)
	}

	// 验证结果完整性
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

	// 验证 Headers 完整性
	headers := result.Headers.ToMap()
	if len(headers) == 0 {
		t.Error("Headers map 不能为空")
	}

	// 验证必需的 header 字段
	if _, ok := headers["User-Agent"]; !ok {
		t.Error("Headers 必须包含 User-Agent")
	}
	if _, ok := headers["Accept"]; !ok {
		t.Error("Headers 必须包含 Accept")
	}
	if _, ok := headers["Accept-Language"]; !ok {
		t.Error("Headers 必须包含 Accept-Language")
	}

	t.Logf("成功获取随机指纹: %s", result.HelloClientID)
}

// TestGetRandomFingerprintByBrowserIntegration 集成测试：按浏览器类型获取指纹
func TestGetRandomFingerprintByBrowserIntegration(t *testing.T) {
	browsers := []string{"chrome", "firefox", "safari", "opera"}

	for _, browser := range browsers {
		t.Run(browser, func(t *testing.T) {
			result, err := fingerprint.GetRandomFingerprintByBrowser(browser)
			if err != nil {
				t.Fatalf("获取 %s 指纹失败: %v", browser, err)
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

// TestGetRandomFingerprintWithOSIntegration 集成测试：指定操作系统获取指纹
func TestGetRandomFingerprintWithOSIntegration(t *testing.T) {
	oses := []fingerprint.OperatingSystem{
		fingerprint.OSWindows10,
		fingerprint.OSMacOS14,
		fingerprint.OSLinux,
	}

	for _, os := range oses {
		t.Run(string(os), func(t *testing.T) {
			result, err := fingerprint.GetRandomFingerprintWithOS(os)
			if err != nil {
				t.Fatalf("获取指纹失败: %v", err)
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

// TestHeadersCustomizationIntegration 集成测试：自定义 Headers
func TestHeadersCustomizationIntegration(t *testing.T) {
	result, err := fingerprint.GetRandomFingerprint()
	if err != nil {
		t.Fatalf("获取随机指纹失败: %v", err)
	}

	// 测试 Set 方法
	result.Headers.Set("Cookie", "session_id=test123")
	result.Headers.Set("Authorization", "Bearer token456")
	result.Headers.Set("X-Custom-Header", "custom-value")

	headers := result.Headers.ToMap()

	// 验证自定义 header
	if cookie, ok := headers["Cookie"]; !ok || cookie != "session_id=test123" {
		t.Error("Cookie header 设置失败")
	}
	if auth, ok := headers["Authorization"]; !ok || auth != "Bearer token456" {
		t.Error("Authorization header 设置失败")
	}
	if custom, ok := headers["X-Custom-Header"]; !ok || custom != "custom-value" {
		t.Error("X-Custom-Header 设置失败")
	}

	// 测试 SetHeaders 方法
	result.Headers.SetHeaders(map[string]string{
		"Cookie":      "session_id=updated",
		"X-API-Key":   "api-key-123",
		"X-Request-ID": "req-123",
	})

	headers = result.Headers.ToMap()
	if cookie, ok := headers["Cookie"]; !ok || cookie != "session_id=updated" {
		t.Error("批量设置 Cookie 失败")
	}
	if apiKey, ok := headers["X-API-Key"]; !ok || apiKey != "api-key-123" {
		t.Error("批量设置 X-API-Key 失败")
	}
}

// TestHeadersCloneIntegration 集成测试：Headers 克隆
func TestHeadersCloneIntegration(t *testing.T) {
	result, err := fingerprint.GetRandomFingerprint()
	if err != nil {
		t.Fatalf("获取随机指纹失败: %v", err)
	}

	// 设置自定义 header
	result.Headers.Set("Cookie", "original")
	result.Headers.Set("X-Test", "value")

	// 克隆
	cloned := result.Headers.Clone()
	if cloned == nil {
		t.Fatal("克隆失败，返回 nil")
	}

	// 修改克隆对象
	cloned.Set("Cookie", "modified")
	cloned.Set("X-New", "new-value")

	// 验证原对象未被修改
	originalHeaders := result.Headers.ToMap()
	if cookie, ok := originalHeaders["Cookie"]; !ok || cookie != "original" {
		t.Error("原对象的 Cookie 被意外修改")
	}
	if _, ok := originalHeaders["X-New"]; ok {
		t.Error("原对象不应该包含 X-New")
	}

	// 验证克隆对象已被修改
	clonedHeaders := cloned.ToMap()
	if cookie, ok := clonedHeaders["Cookie"]; !ok || cookie != "modified" {
		t.Error("克隆对象的 Cookie 未正确修改")
	}
	if newVal, ok := clonedHeaders["X-New"]; !ok || newVal != "new-value" {
		t.Error("克隆对象未正确添加 X-New")
	}
}

// TestTLSClientHelloIntegration 集成测试：TLS Client Hello
func TestTLSClientHelloIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试（使用 -short 标志）")
	}

	testProfiles := []string{"chrome_133", "firefox_135", "safari_ios_18_0"}

	for _, profileName := range testProfiles {
		t.Run(profileName, func(t *testing.T) {
			profile, ok := fingerprint.MappedTLSClients[profileName]
			if !ok {
				t.Fatalf("Profile %s 不存在", profileName)
			}

			spec, err := profile.GetClientHelloSpec()
			if err != nil {
				if err.Error() == "please implement this method" {
					t.Skipf("Profile %s 使用预定义 ID，跳过测试", profileName)
					return
				}
				t.Fatalf("获取 ClientHelloSpec 失败: %v", err)
			}

			// 验证 spec 的完整性
			if len(spec.CipherSuites) == 0 {
				t.Error("CipherSuites 不能为空")
			}

			t.Logf("Profile %s: %d cipher suites, %d extensions",
				profileName, len(spec.CipherSuites), len(spec.Extensions))
		})
	}
}

// TestConcurrentAccess 并发访问测试
func TestConcurrentAccess(t *testing.T) {
	const goroutines = 100
	const iterations = 10

	done := make(chan bool, goroutines)
	errors := make(chan error, goroutines*iterations)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer func() { done <- true }()

			for j := 0; j < iterations; j++ {
				// 测试随机指纹获取
				result, err := fingerprint.GetRandomFingerprint()
				if err != nil {
					errors <- fmt.Errorf("GetRandomFingerprint 失败: %v", err)
					continue
				}

				// 测试 Headers 操作
				result.Headers.Set("Cookie", fmt.Sprintf("test_%d_%d", i, j))
				_ = result.Headers.ToMap()

				// 测试克隆
				_ = result.Headers.Clone()

				// 测试随机语言
				_ = fingerprint.RandomLanguage()

				// 测试随机 OS
				_ = fingerprint.RandomOS()
			}
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < goroutines; i++ {
		<-done
	}

	close(errors)

	// 检查是否有错误
	var errCount int
	for err := range errors {
		t.Error(err)
		errCount++
		if errCount >= 10 {
			t.Fatal("并发测试错误过多，停止测试")
		}
	}

	if errCount > 0 {
		t.Fatalf("并发测试发现 %d 个错误", errCount)
	}

	t.Logf("并发测试通过：%d goroutines × %d iterations = %d 次操作",
		goroutines, iterations, goroutines*iterations)
}

// TestRealTLSConnection 真实 TLS 连接测试（可选，需要网络）
func TestRealTLSConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过网络测试（使用 -short 标志）")
	}

	result, err := fingerprint.GetRandomFingerprint()
	if err != nil {
		t.Fatalf("获取随机指纹失败: %v", err)
	}

	spec, err := result.Profile.GetClientHelloSpec()
	if err != nil {
		if err.Error() == "please implement this method" {
			t.Skip("跳过预定义 ID 的 TLS 连接测试")
			return
		}
		t.Fatalf("获取 ClientHelloSpec 失败: %v", err)
	}

	// 创建 TLS 配置
	tlsConfig := &tls.Config{
		ServerName: "www.google.com",
	}

	// 尝试建立 TLS 连接（超时 5 秒）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}

	conn, err := dialer.DialContext(ctx, "tcp", "www.google.com:443")
	if err != nil {
		t.Skipf("无法连接到测试服务器: %v", err)
		return
	}
	defer conn.Close()

	// 使用 utls 创建客户端
	tlsConn := tls.UClient(conn, tlsConfig, result.Profile.GetClientHelloId(), false, false, false)
	if err := tlsConn.ApplyPreset(&spec); err != nil {
		t.Fatalf("应用 TLS preset 失败: %v", err)
	}

	// 执行 TLS 握手
	if err := tlsConn.Handshake(); err != nil {
		t.Fatalf("TLS 握手失败: %v", err)
	}

	// 发送简单的 HTTP 请求
	request := fmt.Sprintf("GET / HTTP/1.1\r\nHost: www.google.com\r\nUser-Agent: %s\r\nConnection: close\r\n\r\n", result.UserAgent)
	if _, err := tlsConn.Write([]byte(request)); err != nil {
		t.Fatalf("发送请求失败: %v", err)
	}

	// 读取响应
	response := make([]byte, 1024)
	n, err := tlsConn.Read(response)
	if err != nil && err != io.EOF {
		t.Fatalf("读取响应失败: %v", err)
	}

	if n == 0 {
		t.Fatal("未收到响应")
	}

	t.Logf("成功建立 TLS 连接并收到响应: %d bytes", n)
	t.Logf("响应预览: %s", string(response[:min(200, n)]))
}

// TestAllProfilesWithUserAgent 测试所有 profile 的 User-Agent 生成
func TestAllProfilesWithUserAgent(t *testing.T) {
	failCount := 0
	successCount := 0

	for name := range fingerprint.MappedTLSClients {
		t.Run(name, func(t *testing.T) {
			ua, err := fingerprint.GetUserAgentByProfileName(name)
			if err != nil {
				t.Errorf("Profile %s: 获取 User-Agent 失败: %v", name, err)
				failCount++
				return
			}

			if ua == "" {
				t.Errorf("Profile %s: User-Agent 不能为空", name)
				failCount++
				return
			}

			// 验证 User-Agent 格式
			if len(ua) < 20 {
				t.Errorf("Profile %s: User-Agent 格式可能不正确: %s", name, ua)
				failCount++
				return
			}

			successCount++
			t.Logf("Profile %s: %s", name, ua)
		})
	}

	t.Logf("User-Agent 测试完成: 成功 %d, 失败 %d", successCount, failCount)
	
	if failCount > 0 {
		t.Errorf("有 %d 个 profile 的 User-Agent 生成失败", failCount)
	}
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
