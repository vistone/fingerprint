package main

import (
	"fmt"
	"log"

	"github.com/vistone/fingerprint"
)

func main() {
	// 最简单的使用方式：随机获取指纹和完整的 HTTP Headers
	result, err := fingerprint.GetRandomFingerprint()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("指纹名称: %s\n", result.Name)
	fmt.Printf("ClientHello: %s\n", result.Profile.GetClientHelloStr())
	fmt.Printf("\n完整 HTTP Headers（包含 User-Agent 和 Accept-Language）:\n")
	headers := result.Headers.ToMap()
	for key, value := range headers {
		fmt.Printf("  %s: %s\n", key, value)
	}

	// 使用指纹进行 TLS 握手
	spec, err := result.Profile.GetClientHelloSpec()
	if err != nil {
		log.Printf("注意: 某些指纹可能无法获取完整 Spec: %v", err)
	} else {
		fmt.Printf("密码套件数量: %d\n", len(spec.CipherSuites))
	}

	// 也可以指定浏览器类型
	fmt.Println("\n随机 Chrome 指纹:")
	chromeResult, err := fingerprint.GetRandomFingerprintByBrowser("chrome")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  名称: %s\n", chromeResult.Name)
	fmt.Printf("  Headers（包含 User-Agent）:\n")
	chromeHeaders := chromeResult.Headers.ToMap()
	for key, value := range chromeHeaders {
		fmt.Printf("    %s: %s\n", key, value)
	}
}

