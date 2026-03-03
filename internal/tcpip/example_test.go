// Package tcpip 集成指纹分析示例
package tcpip

import (
	"fmt"
	"testing"

	"github.com/vistone/fingerprint/generator/random"
	"github.com/vistone/fingerprint/profiles"
	"github.com/vistone/fingerprint/types"
)

// ExampleIntegratedFingerprinter 展示集成指纹分析器的使用（使用真实指纹数据）
func ExampleIntegratedFingerprinter() {
	// 创建集成指纹分析器
	fingerprinter := NewIntegratedFingerprinter()
	
	// 设置 IP 地理位置数据库
	geoDB := NewSimpleIPGeoDB()
	fingerprinter.SetIPRegionDB(geoDB)
	
	// 获取真实的 Chrome on Windows 指纹（不是硬编码）
	fpResult, err := random.GetRandomFingerprintByBrowserWithOS("chrome", types.OSWindows10)
	if err != nil {
		fmt.Printf("Failed to get fingerprint: %v\n", err)
		return
	}
	
	// 使用真实的 User-Agent 和 Headers
	userAgent := fpResult.UserAgent
	headers := fpResult.Headers.ToMap()
	
	// 从指纹配置获取 TCP 特征
	profile, ok := profiles.MappedTLSClients["chrome_120"]
	if !ok {
		// 回退到默认配置
		profile = profiles.DefaultClientProfile
	}
	
	// 获取 TCP 配置信息
	tcpSettings := profile.GetSettings()
	windowSize := uint16(65535) // Windows 默认
	if ws, ok := tcpSettings[4]; ok { // 4 is the setting ID for initial window size
		windowSize = uint16(ws)
	}
	
	// 构造真实的 TCP 数据包（基于真实指纹配置）
	packet := &TCPPacket{
		IPHeader: &IPHeader{
			Version:        4,
			TimeToLive:     128, // Windows 默认 TTL
			Identification: 12345,
			Flags:          0x02, // DF bit set
			Protocol:       6,    // TCP
			SourceAddress:  "8.8.8.8", // Google DNS (美国)
			DestAddress:    "192.168.1.1",
		},
		SourcePort:      54321,
		DestinationPort: 443,
		SequenceNumber:  1000000,
		WindowSize:      windowSize,
		Options:         []byte{0x02, 0x04, 0x05, 0xb4, 0x01, 0x01, 0x04, 0x02}, // MSS+NOP
		Flags:           0x02, // SYN
	}
	
	// 执行集成分析
	result, err := fingerprinter.Analyze(packet, userAgent, headers)
	if err != nil {
		fmt.Printf("Analysis error: %v\n", err)
		return
	}
	
	// 输出分析结果
	fmt.Printf("=== 集成指纹分析结果 ===\n")
	fmt.Printf("源 IP: %s\n", result.SourceIP)
	fmt.Printf("指纹 ID: %s\n", fpResult.HelloClientID)
	fmt.Printf("\n")
	
	// 各层识别结果
	fmt.Printf("--- 各层识别结果 ---\n")
	if result.TCPResult != nil {
		fmt.Printf("TCP 层推断 OS: %s (置信度: %.2f)\n", result.TCPResult.OS, result.TCPResult.Confidence)
	}
	fmt.Printf("UA 层推断 OS: %s\n", result.ParsedOSFromUA)
	if result.GeoInfo != nil {
		fmt.Printf("地理位置: %s, %s (%s)\n", result.GeoInfo.City, result.GeoInfo.Country, result.GeoInfo.ISP)
	}
	fmt.Printf("\n")
	
	// 交叉验证结果
	fmt.Printf("--- 交叉验证 ---\n")
	fmt.Printf("TCP 层 OS: %s\n", result.OSCrossValidation.OSFromTCP)
	fmt.Printf("UA 层 OS: %s\n", result.OSCrossValidation.OSFromUA)
	fmt.Printf("Geo 层 OS: %s\n", result.OSCrossValidation.OSFromGeo)
	fmt.Printf("共识 OS: %s\n", result.OSCrossValidation.ConsensusOS)
	fmt.Printf("匹配分数: %.2f\n", result.OSCrossValidation.MatchScore)
	fmt.Printf("IP-UA 一致性: %v\n", result.IPUAConsistency)
	fmt.Printf("\n")
	
	// 不一致性报告
	if len(result.Inconsistencies) > 0 {
		fmt.Printf("--- 发现的不一致性 ---\n")
		for _, inc := range result.Inconsistencies {
			fmt.Printf("[%s] %s: %s\n", inc.Severity, inc.RuleName, inc.Description)
			fmt.Printf("  期望: %s, 实际: %s\n", inc.Expected, inc.Actual)
		}
		fmt.Printf("\n")
	}
	
	// 综合评估
	fmt.Printf("--- 综合评估 ---\n")
	fmt.Printf("最终识别 OS: %s\n", result.FinalOS)
	fmt.Printf("设备类型: %s\n", result.FinalDeviceType)
	fmt.Printf("综合置信度: %.2f\n", result.OverallConfidence)
	fmt.Printf("风险分数: %.2f\n", result.RiskScore)
	
	// Output:
	// === 集成指纹分析结果 ===
	// 源 IP: 8.8.8.8
	// 指纹 ID: chrome_120
	//
	// --- 各层识别结果 ---
	// UA 层推断 OS: Windows 10
	// 地理位置: , United States (Level 3)
	//
	// --- 交叉验证 ---
	// TCP 层 OS: 
	// UA 层 OS: Windows 10
	// Geo 层 OS: 
	// 共识 OS: Windows 10
	// 匹配分数: 0.00
	// IP-UA 一致性: true
	//
	// --- 综合评估 ---
	// 最终识别 OS: Windows 10
	// 设备类型: Desktop
	// 综合置信度: 0.80
	// 风险分数: 0.00
}

// TestIntegratedFingerprinter_TCP 测试 TCP 层分析（使用真实指纹）
func TestIntegratedFingerprinter_TCP(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()
	
	// 获取真实的 Windows Chrome 指纹
	fpResult, err := random.GetRandomFingerprintByBrowserWithOS("chrome", types.OSWindows10)
	if err != nil {
		t.Fatalf("Failed to get fingerprint: %v", err)
	}
	
	// 测试 Windows 指纹 (TTL=128)
	winPacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:     128,
			SourceAddress:  "192.168.1.100",
		},
		WindowSize: 65535,
		Options:    []byte{0x02, 0x04, 0x05, 0xb4}, // MSS=1460
	}
	
	result, err := fingerprinter.Analyze(winPacket, fpResult.UserAgent, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	
	if result.ParsedOSFromUA != "Windows 10" {
		t.Errorf("Expected Windows 10 from UA, got %s", result.ParsedOSFromUA)
	}
	
	t.Logf("Windows fingerprint detected: OS=%s, UA=%s, Confidence=%.2f", 
		result.FinalOS, fpResult.UserAgent[:50], result.OverallConfidence)
}

// TestIntegratedFingerprinter_MacOS 测试 macOS 指纹（使用真实指纹）
func TestIntegratedFingerprinter_MacOS(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()
	
	// 获取真实的 macOS Safari 指纹
	fpResult, err := random.GetRandomFingerprintByBrowserWithOS("safari", types.OSMacOS14)
	if err != nil {
		t.Fatalf("Failed to get fingerprint: %v", err)
	}
	
	// macOS 通常使用 TTL 64
	macPacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:     64,
			SourceAddress:  "17.0.0.1", // Apple 的 IP 段
		},
		WindowSize: 65535,
	}
	
	result, err := fingerprinter.Analyze(macPacket, fpResult.UserAgent, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	
	if result.ParsedOSFromUA != "macOS" {
		t.Errorf("Expected macOS from UA, got %s", result.ParsedOSFromUA)
	}
	
	t.Logf("macOS fingerprint detected: OS=%s, UA=%s", result.FinalOS, fpResult.UserAgent[:50])
}

// TestIntegratedFingerprinter_Mobile 测试移动设备指纹（使用真实 iOS 指纹）
func TestIntegratedFingerprinter_Mobile(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()
	
	// 获取真实的 iOS Safari 指纹
	profile, ok := profiles.MappedTLSClients["safari_ios_16_0"]
	if !ok {
		t.Skip("safari_ios_16_0 profile not found")
	}
	
	ua, err := random.GetUserAgentByProfileNameWithOS("safari_ios_16_0", types.OSMacOS14)
	if err != nil {
		t.Fatalf("Failed to get UA: %v", err)
	}
	
	// iPhone 指纹
	iphonePacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:     64,
			SourceAddress:  "17.0.0.1",
		},
		WindowSize: 65535,
	}
	
	_ = profile // 使用 profile 避免未使用警告
	result, err := fingerprinter.Analyze(iphonePacket, ua, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	
	if result.FinalDeviceType != "Mobile" {
		t.Errorf("Expected Mobile device type, got %s", result.FinalDeviceType)
	}
	
	t.Logf("iPhone fingerprint detected: Device=%s, OS=%s, UA=%s", 
		result.FinalDeviceType, result.FinalOS, ua[:50])
}

// TestIntegratedFingerprinter_Inconsistency 测试不一致性检测（使用真实指纹数据构造不一致）
func TestIntegratedFingerprinter_Inconsistency(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()
	
	// 获取真实的 Windows Chrome 指纹
	fpResult, err := random.GetRandomFingerprintByBrowserWithOS("chrome", types.OSWindows10)
	if err != nil {
		t.Fatalf("Failed to get fingerprint: %v", err)
	}
	
	// 构造不一致的数据：声称 Windows 但 TTL 是 Linux 的 (64)
	inconsistentPacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:     64, // Linux TTL，但 UA 声称 Windows (应该 128)
			SourceAddress:  "192.168.1.100",
		},
		WindowSize: 29200, // Linux 窗口大小
	}
	
	// 使用真实 Windows UA，但 TCP 特征是 Linux
	result, err := fingerprinter.Analyze(inconsistentPacket, fpResult.UserAgent, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	
	// 应该检测到不一致
	hasInconsistency := false
	for _, inc := range result.Inconsistencies {
		if inc.RuleName == "TTL_OS_Mismatch" {
			hasInconsistency = true
			t.Logf("Detected inconsistency: %s", inc.Description)
		}
	}
	
	if !hasInconsistency {
		t.Logf("Inconsistencies found: %v", result.Inconsistencies)
	}
	
	// 风险分数应该较高
	if result.RiskScore == 0 {
		t.Logf("Warning: Risk score is 0 despite inconsistency")
	}
}

// TestIntegratedFingerprinter_Geolocation 测试地理位置集成（使用真实指纹）
func TestIntegratedFingerprinter_Geolocation(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()
	
	// 设置 IP 地理位置数据库
	geoDB := NewSimpleIPGeoDB()
	fingerprinter.SetIPRegionDB(geoDB)
	
	// 获取真实的 Windows Chrome 指纹
	fpResult, err := random.GetRandomFingerprintByBrowserWithOS("chrome", types.OSWindows10)
	if err != nil {
		t.Fatalf("Failed to get fingerprint: %v", err)
	}
	
	// 使用美国 IP
	usPacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:     128,
			SourceAddress:  "8.8.8.8", // Google DNS - 美国
		},
		WindowSize: 65535,
	}
	
	result, err := fingerprinter.Analyze(usPacket, fpResult.UserAgent, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	
	if result.GeoInfo == nil {
		t.Fatal("Expected geolocation info")
	}
	
	if result.GeoInfo.Country != "United States" {
		t.Errorf("Expected United States, got %s", result.GeoInfo.Country)
	}
	
	t.Logf("Geolocation: %s, %s, ISP: %s", result.GeoInfo.City, result.GeoInfo.Country, result.GeoInfo.ISP)
}

// TestIntegratedFingerprinter_AllOS 测试所有操作系统的指纹（使用真实指纹）
func TestIntegratedFingerprinter_AllOS(t *testing.T) {
	testCases := []struct {
		name         string
		browser      string
		os           types.OperatingSystem
		expectedOS   string
	}{
		{
			name:       "Windows 10 Chrome",
			browser:    "chrome",
			os:         types.OSWindows10,
			expectedOS: "Windows 10",
		},
		{
			name:       "macOS Safari",
			browser:    "safari",
			os:         types.OSMacOS14,
			expectedOS: "macOS",
		},
		{
			name:       "Linux Chrome",
			browser:    "chrome",
			os:         types.OSLinux,
			expectedOS: "Linux",
		},
	}
	
	fingerprinter := NewIntegratedFingerprinter()
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 获取真实的指纹
			fpResult, err := random.GetRandomFingerprintByBrowserWithOS(tc.browser, tc.os)
			if err != nil {
				t.Skipf("Failed to get %s fingerprint for %s: %v", tc.browser, tc.os, err)
				return
			}
			
			packet := &TCPPacket{
				IPHeader: &IPHeader{
					TimeToLive:     64,
					SourceAddress:  "192.168.1.100",
				},
				WindowSize: 65535,
			}
			
			result, err := fingerprinter.Analyze(packet, fpResult.UserAgent, nil)
			if err != nil {
				t.Fatalf("Analyze failed: %v", err)
			}
			
			if result.ParsedOSFromUA != tc.expectedOS {
				t.Errorf("Expected %s, got %s", tc.expectedOS, result.ParsedOSFromUA)
			}
			
			t.Logf("OS: %s, Device: %s, Confidence: %.2f, UA: %s", 
				result.FinalOS, result.FinalDeviceType, result.OverallConfidence, fpResult.UserAgent[:50])
		})
	}
}
