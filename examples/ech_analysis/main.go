package main

import (
	"fmt"

	fp "github.com/vistone/fingerprint"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("ECH (Encrypted Client Hello) 分析示例")
	fmt.Println("========================================")
	fmt.Println()

	// 示例 1: 标准 Chrome 请求（无 ECH）
	fmt.Println("=== 示例 1: 标准 Chrome 请求（无 ECH） ===")
	standardChrome := fp.ClientHelloData{
		TLSVersion: 0x0304, // TLS 1.3
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0x0000, Data: []byte{0x00, 0x0b, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d}}, // SNI: example.com
			{Type: 0x000d, Data: []byte{}}, // signature_algorithms
			{Type: 0x0033, Data: []byte{}}, // key_share
			{Type: 0x002b, Data: []byte{}}, // supported_versions
			{Type: 0x000a, Data: []byte{}}, // supported_groups
		},
		HasSNI: true,
		SNI:    "example.com",
	}

	result1, _ := fp.AnalyzeECH(standardChrome)
	printECHResult("标准 Chrome", result1)

	// 示例 2: Chrome with GREASE ECH（兼容性测试）
	fmt.Println("\n=== 示例 2: Chrome with GREASE ECH ===")
	greaseChrome := fp.ClientHelloData{
		TLSVersion: 0x0304,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0x0000, Data: []byte{0x00, 0x0b, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d}}, // SNI
			{Type: 0xfe0d, Data: []byte{0x00, 0x00, 0x00}},                                                             // GREASE ECH
			{Type: 0x000d, Data: []byte{}},
			{Type: 0x0033, Data: []byte{}},
		},
		HasSNI: true,
		SNI:    "example.com",
	}

	result2, _ := fp.AnalyzeECH(greaseChrome)
	printECHResult("Chrome (GREASE ECH)", result2)

	// 示例 3: Firefox with ECH enabled（完整 ECH）
	fmt.Println("\n=== 示例 3: Firefox with ECH enabled ===")
	firefoxECH := fp.ClientHelloData{
		TLSVersion: 0x0304,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc030,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0xfd00, Data: []byte{0xfe, 0x0d, 0x01, 0x02}}, // ECH outer extensions
			{Type: 0x000d, Data: []byte{}},                       // signature_algorithms
			{Type: 0x0033, Data: []byte{}},                       // key_share
			{Type: 0x002b, Data: []byte{}},                       // supported_versions
			{Type: 0x000a, Data: []byte{}},                       // supported_groups
			{Type: 0x0010, Data: []byte{}},                       // alpn
		},
		HasSNI: false, // ECH 加密了 SNI
		SNI:    "",
	}

	result3, _ := fp.AnalyzeECH(firefoxECH)
	printECHResult("Firefox (ECH)", result3)
	printAlternativeStrategies(result3)

	// 示例 4: 异常场景 - ECH 但 SNI 仍可见（配置错误）
	fmt.Println("\n=== 示例 4: 异常场景 - ECH 但 SNI 仍可见 ===")
	misconfigured := fp.ClientHelloData{
		TLSVersion: 0x0304,
		CipherSuites: []uint16{
			0x1301, 0x1302,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0x0000, Data: []byte{0x00, 0x0b, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d}}, // SNI
			{Type: 0xfe0d, Data: []byte{0xfe, 0x0d, 0x01, 0xff}},                                                       // ECH (非 GREASE)
		},
		HasSNI: true,
		SNI:    "example.com",
	}

	result4, _ := fp.AnalyzeECH(misconfigured)
	printECHResult("错误配置", result4)

	// 示例 5: 异常场景 - 旧 TLS 版本使用 ECH
	fmt.Println("\n=== 示例 5: 异常场景 - 旧 TLS 版本使用 ECH ===")
	oldTLS := fp.ClientHelloData{
		TLSVersion: 0x0303, // TLS 1.2
		CipherSuites: []uint16{
			0xc02b, 0xc02f,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0xfe0d, Data: []byte{0xfe, 0x0d, 0x01, 0x02}}, // ECH
		},
		HasSNI: false,
	}

	result5, _ := fp.AnalyzeECH(oldTLS)
	printECHResult("TLS 1.2 with ECH", result5)

	// 示例 6: Inner ClientHello
	fmt.Println("\n=== 示例 6: Inner ClientHello ===")
	innerHello := fp.ClientHelloData{
		TLSVersion: 0x0304,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0xfe0d, Data: []byte{0xfe, 0x0d, 0x01, 0xff}}, // ECH inner
			{Type: 0x000d, Data: []byte{}},
			{Type: 0x0033, Data: []byte{}},
			{Type: 0x002b, Data: []byte{}},
		},
		HasSNI: false,
	}

	result6, _ := fp.AnalyzeECH(innerHello)
	printECHResult("Inner ClientHello", result6)

	fmt.Println("\n========================================")
	fmt.Println("ECH 分析总结")
	fmt.Println("========================================")
	printSummary()
}

func printECHResult(name string, result *fp.ECHAnalysisResult) {
	fmt.Printf("客户端: %s\n", name)
	fmt.Printf("ECH 检测: %v\n", result.ECHPresent)

	if result.ECHPresent {
		fmt.Printf("ECH 类型: %s\n", result.ECHType)
		fmt.Printf("ECH 版本: 0x%04x\n", result.ECHVersion)
		if result.ClientHelloType != "" {
			fmt.Printf("ClientHello 类型: %s\n", result.ClientHelloType)
		}
		fmt.Printf("影响等级: %s\n", result.Impact.ImpactLevel)
		fmt.Printf("SNI 可见: %v\n", result.Impact.SNIVisible)
		fmt.Printf("可见字段签名: %s\n", result.VisibleFieldsSignature)
	}

	fmt.Printf("风险分数: %.2f\n", result.RiskScore)

	if len(result.AnomalyFlags) > 0 {
		fmt.Printf("⚠️ 异常标记: %v\n", result.AnomalyFlags)
	} else {
		fmt.Println("✓ 无异常")
	}

	if len(result.Hash) > 0 {
		if len(result.Hash) >= 32 {
			fmt.Printf("哈希: %s\n", result.Hash[:32])
		} else {
			fmt.Printf("哈希: %s\n", result.Hash)
		}
	}
	fmt.Printf("影响摘要: %s\n", result.GetImpactSummary())
}

func printAlternativeStrategies(result *fp.ECHAnalysisResult) {
	if len(result.AlternativeStrategies) > 0 {
		fmt.Println("\n建议的替代策略:")
		for i, strategy := range result.AlternativeStrategies {
			fmt.Printf("  %d. %s\n", i+1, strategy)
		}
	}
}

func printSummary() {
	fmt.Println()
	fmt.Println(`ECH (Encrypted Client Hello) 是 TLS 1.3 的扩展，它加密 ClientHello 中的敏感信息。

关键点:
1. ECH 类型:
   - GREASE: 兼容性测试，实际不加密（影响低）
   - Outer: 真实 ECH，SNI 被加密（影响高）
   - Inner: 内部 ClientHello（影响中）

2. 影响:
   - SNI 不可见: 基于 SNI 的路由和策略受影响
   - 其他字段仍可见: Cipher Suites, 扩展顺序等
   - 需要使用替代指纹方法

3. 异常检测:
   - GREASE_ECH: 测试模式
   - ECH_WITH_VISIBLE_SNI: 配置错误
   - ECH_WITH_OLD_TLS: 协议异常
   - INCOMPLETE_OUTER_HELLO: 结构异常

4. 应对策略:
   - 使用 JA3/JA4 基于可见字段的指纹
   - 结合 HTTP/2, QUIC 等传输层特征
   - 应用层行为分析
   - 跨请求关联分析
   - 多层防御，不依赖单一方法`)
}
