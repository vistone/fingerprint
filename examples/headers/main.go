package main

import (
	"fmt"
	"log"

	"github.com/vistone/fingerprint"
)

func main() {
	fmt.Println("=== 完整指纹、User-Agent 和 Headers 示例 ===")

	// 获取随机指纹（包含 Headers）
	result, err := fingerprint.GetRandomFingerprint()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("HelloClientID: %s\n", result.HelloClientID)
	fmt.Printf("\n标准 HTTP Headers（包含 User-Agent 和 Accept-Language）:\n")

	// 将 Headers 转换为 map
	headers := result.Headers.ToMap()
	for key, value := range headers {
		fmt.Printf("  %s: %s\n", key, value)
	}

	fmt.Println("\n=== 使用 Headers 进行 HTTP 请求 ===")
	fmt.Println("示例代码：")
	fmt.Println("  req, _ := http.NewRequest(\"GET\", url, nil)")
	fmt.Println("  for key, value := range result.Headers.ToMap() {")
	fmt.Println("      req.Header.Set(key, value)")
	fmt.Println("  }")
	fmt.Println("  client.Do(req)")

	fmt.Println("\n=== 不同浏览器的 Headers ===")

	// Chrome
	chromeResult, _ := fingerprint.GetRandomFingerprintByBrowser("chrome")
	fmt.Printf("\nChrome (%s) Headers:\n", chromeResult.HelloClientID)
	chromeHeaders := chromeResult.Headers.ToMap()
	for key, value := range chromeHeaders {
		fmt.Printf("  %s: %s\n", key, value)
	}

	// Firefox
	firefoxResult, _ := fingerprint.GetRandomFingerprintByBrowser("firefox")
	fmt.Printf("\nFirefox (%s) Headers:\n", firefoxResult.HelloClientID)
	firefoxHeaders := firefoxResult.Headers.ToMap()
	for key, value := range firefoxHeaders {
		fmt.Printf("  %s: %s\n", key, value)
	}

	// Safari
	safariResult, _ := fingerprint.GetRandomFingerprintByBrowser("safari")
	fmt.Printf("\nSafari (%s) Headers:\n", safariResult.HelloClientID)
	safariHeaders := safariResult.Headers.ToMap()
	for key, value := range safariHeaders {
		fmt.Printf("  %s: %s\n", key, value)
	}
}
