//go:build ignore

// Gateway P3 集成测试 - HTML 注入功能验证
// 测试网关的 P3 反检测代码自动注入功能
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
	fmt.Println("Gateway P3 集成测试 - HTML 注入功能")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// 第一部分：测试 P3 API 端点
	fmt.Println("【第一部分】测试 P3 API 端点")
	fmt.Println(strings.Repeat("-", 70))
	testP3APIs()
	fmt.Println()

	// 第二部分：测试 HTML 注入中间件
	fmt.Println("【第二部分】测试 HTML 注入中间件")
	fmt.Println(strings.Repeat("-", 70))
	testHTMLInjection()
	fmt.Println()

	// 第三部分：端到端集成测试
	fmt.Println("【第三部分】端到端集成测试")
	fmt.Println(strings.Repeat("-", 70))
	testEndToEnd()
	fmt.Println()

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("✅ 所有测试完成")
	fmt.Println(strings.Repeat("=", 70))
}

// testP3APIs 测试 P3 相关的 API 端点
func testP3APIs() {
	// 创建网关配置
	config := *gateway.DefaultGatewayConfig
	config.Port = 8081
	config.P3Enabled = true
	config.P3ProfileID = "chrome_134_default"
	config.P3ConfigDir = "./profiles"
	config.P3InjectConsist = true

	// 创建网关
	gw := gateway.NewGateway(&config)

	// 测试 1: 获取 Profile 列表
	fmt.Println("1️⃣ 测试 Profile 列表 API")
	req := httptest.NewRequest("GET", "/api/v1/p3/profiles", nil)
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

	// 测试 2: 获取 Profile 详情
	fmt.Println("2️⃣ 测试 Profile 详情 API")
	req = httptest.NewRequest("GET", "/api/v1/p3/profile?id=chrome_134_default", nil)
	w = httptest.NewRecorder()
	gw.ProfileDetailHandler(w, req)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		fmt.Printf("   ✓ Profile 详情获取成功 (长度: %d 字节)\n", len(body))
		// 显示前200个字符
		preview := string(body)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		fmt.Printf("     预览: %s\n", preview)
	} else {
		fmt.Printf("   ✗ 失败: 状态码 %d\n", resp.StatusCode)
	}
	fmt.Println()

	// 测试 3: 获取反检测代码
	fmt.Println("3️⃣ 测试反检测代码生成 API")
	req = httptest.NewRequest("GET", "/api/v1/p3/antidetect.js", nil)
	w = httptest.NewRecorder()
	gw.AntiDetectCodeHandler(w, req)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		fmt.Printf("   ✓ 反检测代码生成成功\n")
		fmt.Printf("     代码长度: %d 字节\n", len(body))
		fmt.Printf("     Content-Type: %s\n", resp.Header.Get("Content-Type"))

		// 验证代码内容
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

	// 测试 4: 使用不同 Profile
	fmt.Println("4️⃣ 测试切换 Profile")
	req = httptest.NewRequest("GET", "/api/v1/p3/antidetect.js?profile=firefox_132_default", nil)
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

// testHTMLInjection 测试 HTML 注入功能
func testHTMLInjection() {
	// 创建网关配置
	config := *gateway.DefaultGatewayConfig
	config.P3Enabled = true
	gw := gateway.NewGateway(&config)

	// 创建一个简单的 HTML handler
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

	// 使用注入器中间件包装
	injectorMiddleware := gw.GetInjectorMiddleware()
	wrappedHandler := injectorMiddleware(htmlHandler)

	// 测试注入
	fmt.Println("1️⃣ 测试 HTML 注入")
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	fmt.Printf("   原始 HTML 长度: ~150 字节\n")
	fmt.Printf("   注入后 HTML 长度: %d 字节\n", len(html))

	// 验证注入
	if strings.Contains(html, "P3 Anti-Detection") {
		fmt.Println("   ✓ P3 代码已成功注入")
	} else {
		fmt.Println("   ✗ P3 代码未注入")
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

	// 显示注入片段
	fmt.Println("\n   注入代码预览:")
	lines := strings.Split(html, "\n")
	for i, line := range lines {
		if i > 5 && i < 10 { // 显示第6-10行
			fmt.Printf("     %s\n", line)
		}
	}
	fmt.Println()

	// 测试非 HTML 响应
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

// testEndToEnd 端到端集成测试
func testEndToEnd() {
	fmt.Println("1️⃣ 启动测试网关服务器")

	// 创建网关配置
	config := *gateway.DefaultGatewayConfig
	config.Port = 0 // 使用随机端口
	config.P3Enabled = true
	gw := gateway.NewGateway(&config)

	// 创建测试服务器
	mux := http.NewServeMux()

	// 注册路由
	mux.HandleFunc("/api/v1/p3/antidetect.js", gw.AntiDetectCodeHandler)
	mux.HandleFunc("/api/v1/p3/profiles", gw.ProfileListHandler)
	mux.HandleFunc("/api/v1/p3/profile", gw.ProfileDetailHandler)

	// 创建一个 HTML 测试页面
	injectorMiddleware := gw.GetInjectorMiddleware()
	htmlHandler := injectorMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>P3 Integration Test</title>
</head>
<body>
    <h1>P3 Anti-Detection Test Page</h1>
    <p>This page should have P3 anti-detection code injected.</p>
</body>
</html>`))
	}))
	mux.Handle("/", htmlHandler)

	// 启动服务器
	ts := httptest.NewServer(mux)
	defer ts.Close()

	fmt.Printf("   ✓ 测试服务器启动: %s\n", ts.URL)
	fmt.Println()

	// 测试 API 端点
	fmt.Println("2️⃣ 测试 API 端点")

	resp, err := http.Get(ts.URL + "/api/v1/p3/profiles")
	if err != nil {
		fmt.Printf("   ✗ 请求失败: %v\n", err)
	} else {
		fmt.Printf("   ✓ Profile 列表 API: 状态码 %d\n", resp.StatusCode)
		resp.Body.Close()
	}

	resp, err = http.Get(ts.URL + "/api/v1/p3/antidetect.js")
	if err != nil {
		fmt.Printf("   ✗ 请求失败: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   ✓ 反检测代码 API: 状态码 %d, 长度 %d 字节\n", resp.StatusCode, len(body))
		resp.Body.Close()
	}
	fmt.Println()

	// 测试 HTML 页面注入
	fmt.Println("3️⃣ 测试 HTML 页面注入")

	resp, err = http.Get(ts.URL + "/")
	if err != nil {
		fmt.Printf("   ✗ 请求失败: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		html := string(body)

		fmt.Printf("   ✓ 页面请求成功\n")
		fmt.Printf("   HTML 长度: %d 字节\n", len(html))

		// 验证注入
		checks := []string{
			"<script>",
			"P3 Anti-Detection",
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

	// 性能测试
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
