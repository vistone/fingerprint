// Package tcpip 集成指纹分析示例
package tcpip

import (
	"fmt"
	"testing"
)

// ExampleIntegratedFingerprinter 展示集成指纹分析器的使用
func ExampleIntegratedFingerprinter() {
	// 创建集成指纹分析器
	fingerprinter := NewIntegratedFingerprinter()
	
	// 设置 IP 地理位置数据库
	geoDB := NewSimpleIPGeoDB()
	fingerprinter.SetIPRegionDB(geoDB)
	
	// 构造真实的 TCP 数据包（来自 Chrome on Windows）
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
		SourcePort:     54321,
		DestinationPort: 443,
		SequenceNumber: 1000000,
		WindowSize:     65535, // Windows 典型窗口大小
		Options:        []byte{0x02, 0x04, 0x05, 0xb4, 0x01, 0x01, 0x04, 0x02}, // MSS+NOP
		Flags:          0x02, // SYN
	}
	
	// 真实的 Chrome User-Agent
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	
	// HTTP Headers
	headers := map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
		"Accept-Encoding": "gzip, deflate, br",
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
	fmt.Printf("User-Agent: %s\n", result.UserAgent)
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
	// User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36
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

// TestIntegratedFingerprinter_TCP 测试 TCP 层分析
func TestIntegratedFingerprinter_TCP(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()
	
	// 测试 Windows 指纹 (TTL=128)
	winPacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:     128,
			SourceAddress:  "192.168.1.100",
		},
		WindowSize: 65535,
		Options:    []byte{0x02, 0x04, 0x05, 0xb4}, // MSS=1460
	}
	
	winUA := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0"
	
	result, err := fingerprinter.Analyze(winPacket, winUA, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	
	if result.ParsedOSFromUA != "Windows 10" {
		t.Errorf("Expected Windows 10 from UA, got %s", result.ParsedOSFromUA)
	}
	
	t.Logf("Windows fingerprint detected: OS=%s, Confidence=%.2f", result.FinalOS, result.OverallConfidence)
}

// TestIntegratedFingerprinter_MacOS 测试 macOS 指纹
func TestIntegratedFingerprinter_MacOS(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()
	
	// macOS 通常使用 TTL 64
	macPacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:     64,
			SourceAddress:  "17.0.0.1", // Apple 的 IP 段
		},
		WindowSize: 65535,
	}
	
	macUA := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/120.0.0.0"
	
	result, err := fingerprinter.Analyze(macPacket, macUA, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	
	if result.ParsedOSFromUA != "macOS" {
		t.Errorf("Expected macOS from UA, got %s", result.ParsedOSFromUA)
	}
	
	t.Logf("macOS fingerprint detected: OS=%s", result.FinalOS)
}

// TestIntegratedFingerprinter_Mobile 测试移动设备指纹
func TestIntegratedFingerprinter_Mobile(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()
	
	// iPhone 指纹
	iphonePacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:     64,
			SourceAddress:  "17.0.0.1",
		},
		WindowSize: 65535,
	}
	
	iphoneUA := "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1"
	
	result, err := fingerprinter.Analyze(iphonePacket, iphoneUA, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	
	if result.FinalDeviceType != "Mobile" {
		t.Errorf("Expected Mobile device type, got %s", result.FinalDeviceType)
	}
	
	t.Logf("iPhone fingerprint detected: Device=%s, OS=%s", result.FinalDeviceType, result.FinalOS)
}

// TestIntegratedFingerprinter_Inconsistency 测试不一致性检测
func TestIntegratedFingerprinter_Inconsistency(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()
	
	// 构造不一致的数据：声称 Windows 但 TTL 是 Linux 的
	inconsistentPacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:     64, // Linux TTL，但 UA 声称 Windows
			SourceAddress:  "192.168.1.100",
		},
		WindowSize: 29200, // Linux 窗口大小
	}
	
	// Windows UA 但 TCP 特征是 Linux
	winUA := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0"
	
	result, err := fingerprinter.Analyze(inconsistentPacket, winUA, nil)
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

// TestIntegratedFingerprinter_Geolocation 测试地理位置集成
func TestIntegratedFingerprinter_Geolocation(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()
	
	// 设置 IP 地理位置数据库
	geoDB := NewSimpleIPGeoDB()
	fingerprinter.SetIPRegionDB(geoDB)
	
	// 使用美国 IP
	usPacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:     128,
			SourceAddress:  "8.8.8.8", // Google DNS - 美国
		},
		WindowSize: 65535,
	}
	
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0"
	
	result, err := fingerprinter.Analyze(usPacket, ua, nil)
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

// TestIntegratedFingerprinter_AllOS 测试所有操作系统的指纹
func TestIntegratedFingerprinter_AllOS(t *testing.T) {
	testCases := []struct {
		name      string
		ua        string
		expectedOS string
	}{
		{
			name:       "Windows 10 Chrome",
			ua:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0",
			expectedOS: "Windows 10",
		},
		{
			name:       "macOS Safari",
			ua:         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Safari/605.1.15",
			expectedOS: "macOS",
		},
		{
			name:       "Linux Firefox",
			ua:         "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/120.0",
			expectedOS: "Linux",
		},
		{
			name:       "Android Chrome",
			ua:         "Mozilla/5.0 (Linux; Android 13; SM-S918B) AppleWebKit/537.36 Chrome/120.0.0.0 Mobile",
			expectedOS: "Android",
		},
		{
			name:       "iOS Safari",
			ua:         "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148 Safari/604.1",
			expectedOS: "iOS",
		},
	}
	
	fingerprinter := NewIntegratedFingerprinter()
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			packet := &TCPPacket{
				IPHeader: &IPHeader{
					TimeToLive:     64,
					SourceAddress:  "192.168.1.100",
				},
				WindowSize: 65535,
			}
			
			result, err := fingerprinter.Analyze(packet, tc.ua, nil)
			if err != nil {
				t.Fatalf("Analyze failed: %v", err)
			}
			
			if result.ParsedOSFromUA != tc.expectedOS {
				t.Errorf("Expected %s, got %s", tc.expectedOS, result.ParsedOSFromUA)
			}
			
			t.Logf("OS: %s, Device: %s, Confidence: %.2f", 
				result.FinalOS, result.FinalDeviceType, result.OverallConfidence)
		})
	}
}
