package main

import (
	"fmt"
	"github.com/vistone/fingerprint"
)

func main() {
	// 示例1: 使用默认指纹
	fmt.Println("=== 默认指纹 ===")
	defaultProfile := fingerprint.DefaultClientProfile
	fmt.Printf("指纹标识: %s\n", defaultProfile.GetClientHelloStr())
	
	// 示例2: 通过名称获取指纹
	fmt.Println("\n=== 通过名称获取指纹 ===")
	profile, ok := fingerprint.MappedTLSClients["chrome_133"]
	if ok {
		fmt.Printf("找到指纹: %s\n", profile.GetClientHelloStr())
	}
	
	// 示例3: 列出所有可用指纹
	fmt.Println("\n=== 所有可用指纹 ===")
	count := 0
	for name := range fingerprint.MappedTLSClients {
		fmt.Printf("  - %s\n", name)
		count++
	}
	fmt.Printf("\n总计: %d 个指纹\n", count)
	
	// 示例4: 获取指纹的详细信息
	fmt.Println("\n=== 指纹详细信息 ===")
	spec, err := profile.GetClientHelloSpec()
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	
	fmt.Printf("密码套件数量: %d\n", len(spec.CipherSuites))
	fmt.Printf("扩展数量: %d\n", len(spec.Extensions))
	fmt.Printf("HTTP/2 设置: %v\n", profile.GetSettings())
}

