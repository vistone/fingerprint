//go:build ignore

// Translated comment
// Translated comment
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/vistone/fingerprint/modules/gateway"
)

func main() {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("Gateway 反检测集成测试 - HTML 注入功能")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// Translated comment
	fmt.Println("【第一部分】测试反检测 API 端点")
	fmt.Println(strings.Repeat("-", 70))
	testAntiDetectAPIs()
	fmt.Println()

	// Translated comment
	fmt.Println("【第二部分】测试 HTML 注入中间件")
	fmt.Println(strings.Repeat("-", 70))
	testHTMLInjection()
	fmt.Println()

	// Translated comment
	fmt.Println("【第三部分】端到端集成测试")
	fmt.Println(strings.Repeat("-", 70))
	testEndToEnd()
	fmt.Println()

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("✅ 所有测试完成")
	fmt.Println(strings.Repeat("=", 70))
}

// Translated comment
func testAntiDetectAPIs() {
	// Translated comment
	config := *gateway.DefaultGatewayConfig
	config.Port = 8081
	config.AntiDetectEnabled = true
	config.AntiDetectProfileID = "chrome_134_default"
	config.AntiDetectConfigDir = "./profiles"
	config.AntiDetectInjectConsist = true

	// Translated comment
	gw := gateway.NewGateway(&config)

	// Translated comment
	fmt.Println("1️⃣ 测试 Profile 列表 API")
	req := httptest.NewRequest("GET", "/api/v1/antidetect/profiles", nil)
	w := httptest.NewRecorder()
	gw.ProfileListHandler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		fmt.Printf("   ✓ Profile 列表获取成功:\n")
		fmt.Printf("     %s\n", string(body))
	} else {
		fmt.Printf("   ✗ 失败: 状态码 %d\n", resp.StatusCode)
	}
	fmt.Println()

	// Translated comment
	fmt.Println("2️⃣ 测试 Profile 详情 API")
	req = httptest.NewRequest("GET", "/api/v1/antidetect/profile?id=chrome_134_default", nil)
	w = httptest.NewRecorder()
	gw.ProfileDetailHandler(w, req)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		fmt.Printf("   ✓ Profile 详情获取成功 (长度: %d 字节)\n", len(body))
		// Translated comment
		preview := string(body)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		fmt.Printf("     预览: %s\n", preview)
	} else {
		fmt.Printf("   ✗ 失败: 状态码 %d\n", resp.StatusCode)
	}
	fmt.Println()

	// Translated comment
	fmt.Println("3️⃣ 测试反检测代码生成 API")
	req = httptest.NewRequest("GET", "/api/v1/antidetect/antidetect.js", nil)
	w = httptest.NewRecorder()
	gw.AntiDetectCodeHandler(w, req)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		fmt.Printf("   ✓ 反检测代码生成成功\n")
		fmt.Printf("     代码长度: %d 字节\n", len(body))
		fmt.Printf("     Content-Type: %s\n", resp.Header.Get("Content-Type"))

		// Translated comment
		code := string(body)
		checks := []struct {
			keyword string
			desc    string
		}{
			{"<script>", "脚本标签"},
			{"WebGPU", "WebGPU 对抗点"},
			{"MediaDevices", "MediaDevices 对抗点"},
			{"Permissions", "Permissions 对抗点"},
			{"webdriver", "Automation 对抗点"},
		}

		allPassed := true
		for _, check := range checks {
			if strings.Contains(code, check.keyword) {
				fmt.Printf("     ✓ %s\n", check.desc)
			} else {
				fmt.Printf("     ✗ 缺少 %s\n", check.desc)
				allPassed = false
			}
		}

		if allPassed {
			fmt.Println("     ✓ 所有对抗点验证通过")
		}
	} else {
		fmt.Printf("   ✗ 失败: 状态码 %d\n", resp.StatusCode)
	}
	fmt.Println()

	// Translated comment
	fmt.Println("4️⃣ 测试切换 Profile")
	req = httptest.NewRequest("GET", "/api/v1/antidetect/antidetect.js?profile=firefox_132_default", nil)
	w = httptest.NewRecorder()
	gw.AntiDetectCodeHandler(w, req)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		fmt.Printf("   ✓ Firefox Profile 代码生成成功 (长度: %d 字节)\n", len(body))
	} else {
		fmt.Printf("   ✗ 失败: 状态码 %d\n", resp.StatusCode)
	}
}

// Translated comment
func testHTMLInjection() {
	// Translated comment
	config := *gateway.DefaultGatewayConfig
	config.AntiDetectEnabled = true
	gw := gateway.NewGateway(&config)

	// Translated comment
	htmlHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>Test Page</title>
</head>
<body>
    <h1>Hello World</h1>
</body>
</html>`))
	})

	// Translated comment
	injectorMiddleware := gw.GetInjectorMiddleware()
	wrappedHandler := injectorMiddleware(htmlHandler)

	// Translated comment
	fmt.Println("1️⃣ 测试 HTML 注入")
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	fmt.Printf("   原始 HTML 长度: ~150 字节\n")
	fmt.Printf("   注入后 HTML 长度: %d 字节\n", len(html))

	// Translated comment
	if strings.Contains(html, "Anti-Detection") {
		fmt.Println("   ✓ 反检测代码已成功注入")
	} else {
		fmt.Println("   ✗ 反检测代码未注入")
	}

	if strings.Contains(html, "<head>") {
		headPos := strings.Index(html, "<head>")
		scriptPos := strings.Index(html, "<script>")
		if scriptPos > headPos && scriptPos < strings.Index(html, "</head>") {
			fmt.Println("   ✓ 代码注入位置正确（在 <head> 标签内）")
		} else {
			fmt.Println("   ✗ 代码注入位置不正确")
		}
	}

	// Translated comment
	fmt.Println("\n   注入代码预览:")
	lines := strings.Split(html, "\n")
	for i, line := range lines {
		if i > 5 && i < 10 { // Translated comment
			fmt.Printf("     %s\n", line)
		}
	}
	fmt.Println()

	// Translated comment
	fmt.Println("2️⃣ 测试非 HTML 响应（不应注入）")
	jsonHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	wrappedJSON := injectorMiddleware(jsonHandler)
	req = httptest.NewRequest("GET", "/api/test", nil)
	w = httptest.NewRecorder()
	wrappedJSON.ServeHTTP(w, req)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)

	if !strings.Contains(string(body), "<script>") {
		fmt.Println("   ✓ JSON 响应未被注入（正确）")
	} else {
		fmt.Println("   ✗ JSON 响应被错误注入")
	}
}

// Translated comment
func testEndToEnd() {
	fmt.Println("1️⃣ 启动测试网关服务器")

	// Translated comment
	config := *gateway.DefaultGatewayConfig
	config.Port = 0 // Translated comment
	config.AntiDetectEnabled = true
	gw := gateway.NewGateway(&config)

	// Translated comment
	mux := http.NewServeMux()

	// Translated comment
	mux.HandleFunc("/api/v1/antidetect/antidetect.js", gw.AntiDetectCodeHandler)
	mux.HandleFunc("/api/v1/antidetect/profiles", gw.ProfileListHandler)
	mux.HandleFunc("/api/v1/antidetect/profile", gw.ProfileDetailHandler)

	// Translated comment
	injectorMiddleware := gw.GetInjectorMiddleware()
	htmlHandler := injectorMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>Anti-Detection Integration Test</title>
</head>
<body>
    <h1>Anti-Detection Test Page</h1>
    <p>This page should have anti-detection code injected.</p>
</body>
</html>`))
	}))
	mux.Handle("/", htmlHandler)

	// Translated comment
	ts := httptest.NewServer(mux)
	defer ts.Close()

	fmt.Printf("   ✓ 测试服务器启动: %s\n", ts.URL)
	fmt.Println()

	// Translated comment
	fmt.Println("2️⃣ 测试 API 端点")

	resp, err := http.Get(ts.URL + "/api/v1/antidetect/profiles")
	if err != nil {
		fmt.Printf("   ✗ 请求失败: %v\n", err)
	} else {
		fmt.Printf("   ✓ Profile 列表 API: 状态码 %d\n", resp.StatusCode)
		resp.Body.Close()
	}

	resp, err = http.Get(ts.URL + "/api/v1/antidetect/antidetect.js")
	if err != nil {
		fmt.Printf("   ✗ 请求失败: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   ✓ 反检测代码 API: 状态码 %d, 长度 %d 字节\n", resp.StatusCode, len(body))
		resp.Body.Close()
	}
	fmt.Println()

	// Translated comment
	fmt.Println("3️⃣ 测试 HTML 页面注入")

	resp, err = http.Get(ts.URL + "/")
	if err != nil {
		fmt.Printf("   ✗ 请求失败: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		html := string(body)

		fmt.Printf("   ✓ 页面请求成功\n")
		fmt.Printf("   HTML 长度: %d 字节\n", len(html))

		// Translated comment
		checks := []string{
			"<script>",
			"Anti-Detection",
			"WebGPU",
			"navigator.mediaDevices",
			"navigator.permissions",
			"webdriver",
		}

		passed := 0
		for _, check := range checks {
			if strings.Contains(html, check) {
				passed++
			}
		}

		fmt.Printf("   验证通过: %d/%d\n", passed, len(checks))

		if passed == len(checks) {
			fmt.Println("   ✓ 所有注入验证通过")
		} else {
			fmt.Println("   ⚠️  部分验证未通过")
		}

		resp.Body.Close()
	}
	fmt.Println()

	// Translated comment
	fmt.Println("4️⃣ 性能测试")

	iterations := 100
	start := time.Now()

	for i := 0; i < iterations; i++ {
		resp, err := http.Get(ts.URL + "/")
		if err != nil {
			fmt.Printf("   ✗ 请求 %d 失败: %v\n", i, err)
			continue
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}

	elapsed := time.Since(start)
	avgTime := elapsed / time.Duration(iterations)

	fmt.Printf("   ✓ %d 次请求完成\n", iterations)
	fmt.Printf("   总耗时: %v\n", elapsed)
	fmt.Printf("   平均耗时: %v\n", avgTime)
	fmt.Printf("   吞吐量: %.2f req/s\n", float64(iterations)/elapsed.Seconds())
}
