// 示例：使用 fingerprint v3.0 API
// 展示 Go Workspace 重构后的新 API 用法
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/vistone/fingerprint"
)

func main() {
	fmt.Println("=== Fingerprint v3.0 API Demo ===\n")

	// 1. 获取随机指纹
	fmt.Println("1. Get Random Fingerprint:")
	profile := fingerprint.GetRandom()
	fmt.Printf("   Browser: %s %s\n", profile.BrowserType, profile.BrowserVersion)
	fmt.Printf("   OS: %s\n", profile.OS)
	if profile.Headers != nil {
		fmt.Printf("   User-Agent: %s\n", profile.Headers.UserAgent)
	}

	// 2. 获取指定浏览器的指纹
	fmt.Println("\n2. Get Chrome Fingerprint:")
	chrome := fingerprint.GetByBrowser(fingerprint.BrowserChrome)
	if chrome != nil {
		fmt.Printf("   Chrome Version: %s\n", chrome.BrowserVersion)
	}

	// 3. 提取特征向量
	fmt.Println("\n3. Extract Features:")
	analyzer := fingerprint.NewAnalyzer()
	features := analyzer.ExtractFeatures(profile)
	fmt.Printf("   Extracted %d features\n", len(features.Features))
	for ft, val := range features.Features {
		fmt.Printf("     %s: %.2f\n", ft, val)
	}

	// 4. 分类
	fmt.Println("\n4. Classify:")
	classification := analyzer.Classify(features)
	fmt.Printf("   Protocol: %s (confidence: %.2f)\n", classification.Protocol, classification.ProtocolConfidence)
	fmt.Printf("   Family: %s (confidence: %.2f)\n", classification.Family, classification.FamilyConfidence)
	fmt.Printf("   Version: %s (confidence: %.2f)\n", classification.Version, classification.VersionConfidence)
	fmt.Printf("   Overall Confidence: %.2f\n", classification.Confidence)

	// 5. 风险评估
	fmt.Println("\n5. Risk Assessment:")
	risk := analyzer.EvaluateRisk(features, classification)
	fmt.Printf("   Risk Score: %.2f\n", risk.Score)
	fmt.Printf("   Risk Level: %s\n", risk.Level.String())
	if len(risk.Suggestions) > 0 {
		fmt.Println("   Suggestions:")
		for _, s := range risk.Suggestions {
			fmt.Printf("     - %s\n", s)
		}
	}

	// 6. 快速分析 HTTP 头
	fmt.Println("\n6. Quick Analyze HTTP Headers:")
	headers := &fingerprint.HTTPHeaders{
		Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		AcceptLanguage:  "en-US,en;q=0.5",
		AcceptEncoding:  "gzip, deflate, br",
		UserAgent:       profile.Headers.UserAgent,
		SecFetchSite:    "none",
		SecFetchMode:    "navigate",
		SecFetchDest:    "document",
		SecCHUA:         `"Google Chrome";v="133"`,
		SecCHUAMobile:   "?0",
		SecCHUAPlatform: `"Windows"`,
	}

	result := fingerprint.QuickAnalyze(headers, "GET")
	if result.JA4H != nil {
		fmt.Printf("   JA4H: %s\n", result.JA4H.Fingerprint)
	}
	fmt.Printf("   Risk Level: %s\n", result.RiskLevel.String())

	// 7. 打印 JSON 格式输出
	fmt.Println("\n7. Classification Result (JSON):")
	jsonData, _ := json.MarshalIndent(classification, "   ", "  ")
	fmt.Println("   " + string(jsonData))

	fmt.Println("\n=== Demo Complete ===")
}

// 示例：启动网关服务
func runGateway() {
	config := fingerprint.DefaultGatewayConfig
	config.Port = 8080
	config.CacheEnabled = true
	config.RateLimitRequests = 1000

	log.Printf("Starting gateway on port %d...\n", config.Port)
	if err := fingerprint.StartGateway(config); err != nil {
		log.Fatal(err)
	}
}
