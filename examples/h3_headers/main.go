package main

import (
	"fmt"
	"log"

	"github.com/vistone/fingerprint"
)

func main() {
	fmt.Println("=== HTTP/3 头部使用说明 ===")

	// 1. 获取指纹和标准 headers
	result, err := fingerprint.GetRandomFingerprint()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n1. HTTP/3 头部字段说明：")
	fmt.Println("   - HTTP/3 的头部字段定义与 HTTP/1.1 和 HTTP/2 相同")
	fmt.Println("   - 没有引入新的专有头部字段")
	fmt.Println("   - 使用 QPACK 进行头部压缩（而不是 HTTP/2 的 HPACK）")
	fmt.Println("   - 伪头部（:method, :scheme, :path, :authority）由传输层自动处理")

	fmt.Println("\n2. 标准 HTTP Headers（适用于 HTTP/1.1、HTTP/2、HTTP/3）：")
	headers := result.Headers.ToMap()
	for key, value := range headers {
		if len(value) > 60 {
			fmt.Printf("   %s: %s...\n", key, value[:60])
		} else {
			fmt.Printf("   %s: %s\n", key, value)
		}
	}

	fmt.Println("\n3. HTTP/3 伪头部说明：")
	fmt.Println("   伪头部由 HTTP/3 客户端库自动添加，不需要手动设置：")
	fmt.Println("   - :method  - HTTP 方法（GET, POST 等）")
	fmt.Println("   - :scheme - 协议方案（http, https）")
	fmt.Println("   - :path   - 请求路径")
	fmt.Println("   - :authority - 主机名（替代 Host 头）")

	fmt.Println("\n4. 使用自定义 headers（适用于所有 HTTP 版本）：")
	result.Headers.Set("Cookie", "session_id=abc123")
	result.Headers.Set("X-API-Key", "api-key-123")
	
	allHeaders := result.Headers.ToMap()
	fmt.Printf("   Cookie: %s\n", allHeaders["Cookie"])
	fmt.Printf("   X-API-Key: %s\n", allHeaders["X-API-Key"])

	fmt.Println("\n5. 重要提示：")
	fmt.Println("   - 当前的 HTTPHeaders 结构体可以同时用于 HTTP/1.1、HTTP/2 和 HTTP/3")
	fmt.Println("   - 伪头部由传输层（HTTP/2 或 HTTP/3 客户端库）自动处理")
	fmt.Println("   - 应用层只需要提供常规的 HTTP 头部字段")
	fmt.Println("   - HTTP/3 使用 QPACK 压缩，但头部字段定义不变")

	fmt.Println("\n✓ HTTP/3 头部使用说明完成！")
}

