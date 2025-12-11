package main

import (
	"fmt"
	"log"

	"github.com/vistone/fingerprint"
)

func main() {
	fmt.Println("=== 随机指纹和 User-Agent 示例 ===")

	// 示例 1: 完全随机（所有指纹）
	fmt.Println("1. 完全随机指纹（所有浏览器）:")
	for i := 0; i < 5; i++ {
		result, err := fingerprint.GetRandomFingerprint()
		if err != nil {
			log.Printf("错误: %v", err)
			continue
		}
		fmt.Printf("   %d. %s\n", i+1, result.Name)
		fmt.Printf("      ClientHello: %s\n\n", result.Profile.GetClientHelloStr())
	}

	// 示例 2: 随机 Chrome 指纹
	fmt.Println("2. 随机 Chrome 指纹:")
	for i := 0; i < 3; i++ {
		result, err := fingerprint.GetRandomFingerprintByBrowser("chrome")
		if err != nil {
			log.Printf("错误: %v", err)
			continue
		}
		fmt.Printf("   %d. %s\n\n", i+1, result.Name)
	}

	// 示例 3: 随机 Firefox 指纹
	fmt.Println("3. 随机 Firefox 指纹:")
	for i := 0; i < 3; i++ {
		result, err := fingerprint.GetRandomFingerprintByBrowser("firefox")
		if err != nil {
			log.Printf("错误: %v", err)
			continue
		}
		fmt.Printf("   %d. %s\n\n", i+1, result.Name)
	}

	// 示例 4: 指定操作系统的随机指纹
	fmt.Println("4. 指定操作系统的随机指纹（Windows）:")
	for i := 0; i < 3; i++ {
		result, err := fingerprint.GetRandomFingerprintWithOS(fingerprint.OSWindows10)
		if err != nil {
			log.Printf("错误: %v", err)
			continue
		}
		fmt.Printf("   %d. %s\n\n", i+1, result.Name)
	}

	// 示例 5: 指定浏览器和操作系统
	fmt.Println("5. 指定浏览器和操作系统（Chrome + macOS）:")
	for i := 0; i < 3; i++ {
		result, err := fingerprint.GetRandomFingerprintByBrowserWithOS("chrome", fingerprint.OSMacOS15)
		if err != nil {
			log.Printf("错误: %v", err)
			continue
		}
		fmt.Printf("   %d. %s\n\n", i+1, result.Name)
	}

	fmt.Println("=== 使用示例 ===")
	fmt.Println("最简单的使用方式：")
	fmt.Println("  result, err := fingerprint.GetRandomFingerprint()")
	fmt.Println("  if err != nil {")
	fmt.Println("      log.Fatal(err)")
	fmt.Println("  }")
	fmt.Println("  // result.Profile 是 TLS 指纹配置")
	fmt.Println("  // result.Headers 包含完整的 HTTP Headers（包括 User-Agent 和 Accept-Language）")
	fmt.Println("  // result.Name 是指纹名称")
}
