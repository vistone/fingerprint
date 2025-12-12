package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/vistone/fingerprint"
)

func main() {
	fmt.Println("=== Headers 自定义示例（系统自动合并）===")

	// 1. 获取标准 headers
	result, err := fingerprint.GetRandomFingerprint()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n1. 标准 Headers（系统生成）:")
	standardHeaders := result.Headers.ToMap()
	fmt.Printf("  Cookie: %s\n", standardHeaders["Cookie"])
	fmt.Printf("  X-API-Key: %s\n", standardHeaders["X-API-Key"])
	fmt.Printf("  User-Agent: %s\n", standardHeaders["User-Agent"][:50]+"...")

	// 2. 设置自定义 headers（系统会自动合并，无需手动调用 Merge）
	fmt.Println("\n2. 设置自定义 Headers（系统自动合并）:")
	
	// 方式 1: 使用 Set 方法（推荐）
	result.Headers.Set("Cookie", "session_id=abc123; csrf_token=xyz789")
	result.Headers.Set("X-API-Key", "your-api-key-here")
	result.Headers.Set("Authorization", "Bearer token123456")
	
	// 方式 2: 使用 SetHeaders 批量设置
	result.Headers.SetHeaders(map[string]string{
		"X-Custom-Header": "custom-value",
		"Accept-Language": "zh-CN,zh;q=0.9", // 可以覆盖系统生成的 headers
	})

	// 3. 调用 ToMap()，系统自动合并（无需任何额外操作）
	fmt.Println("\n3. 调用 ToMap()，系统自动合并:")
	mergedHeaders := result.Headers.ToMap()
	fmt.Printf("  Cookie: %s\n", mergedHeaders["Cookie"])
	fmt.Printf("  X-API-Key: %s\n", mergedHeaders["X-API-Key"])
	fmt.Printf("  Authorization: %s\n", mergedHeaders["Authorization"])
	fmt.Printf("  Accept-Language: %s\n", mergedHeaders["Accept-Language"])
	fmt.Printf("  User-Agent: %s\n", mergedHeaders["User-Agent"][:50]+"...")

	// 4. 实际使用示例：创建 HTTP 请求
	fmt.Println("\n4. 实际使用示例（创建 HTTP 请求）:")
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	
	// 直接使用 ToMap()，系统自动包含所有自定义 headers
	headers := result.Headers.ToMap()
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	fmt.Println("  请求头已自动设置:")
	fmt.Printf("    Cookie: %s\n", req.Header.Get("Cookie"))
	fmt.Printf("    X-API-Key: %s\n", req.Header.Get("X-Api-Key"))
	fmt.Printf("    Authorization: %s\n", req.Header.Get("Authorization"))

	// 5. 动态更新 session（系统自动处理）
	fmt.Println("\n5. 动态更新 Session（系统自动处理）:")
	result.Headers.Set("Cookie", "session_id=updated123; csrf_token=new789")
	updatedHeaders := result.Headers.ToMap()
	fmt.Printf("  更新后 Cookie: %s\n", updatedHeaders["Cookie"])
	fmt.Printf("  更新后 X-API-Key: %s\n", updatedHeaders["X-API-Key"]) // 保持不变

	// 6. 使用场景：多次更新（系统自动处理）
	fmt.Println("\n6. 使用场景：多次更新（系统自动处理）")
	result2, _ := fingerprint.GetRandomFingerprint()
	
	// 第一次设置
	result2.Headers.Set("Cookie", "session_id=initial")
	fmt.Printf("  第一次设置后: %s\n", result2.Headers.ToMap()["Cookie"])
	
	// 第二次更新
	result2.Headers.Set("Cookie", "session_id=updated")
	fmt.Printf("  第二次更新后: %s\n", result2.Headers.ToMap()["Cookie"])
	
	// 添加新的 header
	result2.Headers.Set("X-API-Key", "new-api-key")
	fmt.Printf("  添加 API Key 后 Cookie: %s\n", result2.Headers.ToMap()["Cookie"])
	fmt.Printf("  添加 API Key 后 X-API-Key: %s\n", result2.Headers.ToMap()["X-API-Key"])
	
	fmt.Println("\n✓ 所有操作都是系统自动完成的，无需手动调用 Merge 或 ToMapWithCustom！")
}

