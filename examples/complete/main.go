package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/vistone/fingerprint"
)

func main() {
	fmt.Println("=== 完整使用示例：指纹 + User-Agent + Headers ===")

	// 1. 获取随机指纹（包含所有信息）
	result, err := fingerprint.GetRandomFingerprint()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("指纹名称: %s\n", result.Name)
	fmt.Printf("\n完整 HTTP Headers（包含 User-Agent 和 Accept-Language）:\n")

	// 2. 获取完整的 HTTP Headers
	headers := result.Headers.ToMap()
	for key, value := range headers {
		fmt.Printf("  %s: %s\n", key, value)
	}

	// 3. 使用示例：创建 HTTP 请求
	fmt.Println("\n=== 使用示例 ===")
	fmt.Println("创建 HTTP 请求：")
	fmt.Println("  req, _ := http.NewRequest(\"GET\", \"https://example.com\", nil)")
	fmt.Println("  for key, value := range result.Headers.ToMap() {")
	fmt.Println("      req.Header.Set(key, value)")
	fmt.Println("  }")
	fmt.Println("  client.Do(req)")

	// 4. 实际使用示例
	fmt.Println("\n=== 实际 HTTP 请求示例 ===")
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	fmt.Println("请求头已设置:")
	for key, values := range req.Header {
		fmt.Printf("  %s: %v\n", key, values)
	}

	// 5. 注意：User-Agent 和 Accept-Language 已经包含在 Headers 中，无需单独获取
}

