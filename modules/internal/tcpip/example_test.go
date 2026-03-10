// translated comment
package tcpip

import (
	"fmt"
	"testing"

	"github.com/vistone/fingerprint/modules/generator/random"
	"github.com/vistone/fingerprint/modules/http/legacy/useragent"
	"github.com/vistone/fingerprint/modules/profiles/legacy"
	"github.com/vistone/fingerprint/modules/core/types"
)

// translated comment
func ExampleIntegratedFingerprinter() {
	// translated comment
	fingerprinter := NewIntegratedFingerprinter()
	
	// translated comment
	geoDB := NewSimpleIPGeoDB()
	fingerprinter.SetIPRegionDB(geoDB)
	
	// translated comment
	fpResult, err := random.GetRandomFingerprintByBrowserWithOS("chrome", types.OSWindows10)
	if err != nil {
		fmt.Printf("Failed to get fingerprint: %v\n", err)
		return
	}
	
	// translated comment
	userAgent := fpResult.UserAgent
	headers := fpResult.Headers.ToMap()
	
	// translated comment
	profile, ok := profiles.MappedTLSClients["chrome_120"]
	if !ok {
		// translated comment
		profile = profiles.DefaultClientProfile
	}
	
	// translated comment
	tcpSettings := profile.GetSettings()
	windowSize := uint16(65535) // translated comment
	if ws, ok := tcpSettings[4]; ok { // 4 is the setting ID for initial window size
		windowSize = uint16(ws)
	}
	
	// translated comment
	packet := &TCPPacket{
		IPHeader: &IPHeader{
			Version:        4,
			TimeToLive:     128, // translated comment
			Identification: 12345,
			Flags:          0x02, // DF bit set
			Protocol:       6,    // TCP
			SourceAddress:  "8.8.8.8", // translated comment
			DestAddress:    "192.168.1.1",
		},
		SourcePort:      54321,
		DestinationPort: 443,
		SequenceNumber:  1000000,
		WindowSize:      windowSize,
		Options:         []byte{0x02, 0x04, 0x05, 0xb4, 0x01, 0x01, 0x04, 0x02}, // MSS+NOP
		Flags:           0x02, // SYN
	}
	
	// translated comment
	result, err := fingerprinter.Analyze(packet, userAgent, headers)
	if err != nil {
		fmt.Printf("Analysis error: %v\n", err)
		return
	}
	
	// translated comment
	fmt.Printf("=== 集成指纹分析结果 ===\n")
	fmt.Printf("源 IP: %s\n", result.SourceIP)
	fmt.Printf("指纹 ID: %s\n", fpResult.HelloClientID)
	fmt.Printf("\n")
	
	// translated comment
	fmt.Printf("--- 各层识别结果 ---\n")
	if result.TCPResult != nil {
		fmt.Printf("TCP 层推断 OS: %s (置信度: %.2f)\n", result.TCPResult.OS, result.TCPResult.Confidence)
	}
	fmt.Printf("UA 层推断 OS: %s\n", result.ParsedOSFromUA)
	if result.GeoInfo != nil {
		fmt.Printf("地理位置: %s, %s (%s)\n", result.GeoInfo.City, result.GeoInfo.Country, result.GeoInfo.ISP)
	}
	fmt.Printf("\n")
	
	// translated comment
	fmt.Printf("--- 交叉验证 ---\n")
	fmt.Printf("TCP 层 OS: %s\n", result.OSCrossValidation.OSFromTCP)
	fmt.Printf("UA 层 OS: %s\n", result.OSCrossValidation.OSFromUA)
	fmt.Printf("Geo 层 OS: %s\n", result.OSCrossValidation.OSFromGeo)
	fmt.Printf("共识 OS: %s\n", result.OSCrossValidation.ConsensusOS)
	fmt.Printf("匹配分数: %.2f\n", result.OSCrossValidation.MatchScore)
	fmt.Printf("IP-UA 一致性: %v\n", result.IPUAConsistency)
	fmt.Printf("\n")
	
	// translated comment
	if len(result.Inconsistencies) > 0 {
		fmt.Printf("--- 发现的不一致性 ---\n")
		for _, inc := range result.Inconsistencies {
			fmt.Printf("[%s] %s: %s\n", inc.Severity, inc.RuleName, inc.Description)
			fmt.Printf("  期望: %s, 实际: %s\n", inc.Expected, inc.Actual)
		}
		fmt.Printf("\n")
	}
	
	// translated comment
	fmt.Printf("--- 综合评估 ---\n")
	fmt.Printf("最终识别 OS: %s\n", result.FinalOS)
	fmt.Printf("设备类型: %s\n", result.FinalDeviceType)
	fmt.Printf("综合置信度: %.2f\n", result.OverallConfidence)
	fmt.Printf("风险分数: %.2f\n", result.RiskScore)
}

// translated comment
func TestIntegratedFingerprinter_TCP(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()
	
	// translated comment
	fpResult, err := random.GetRandomFingerprintByBrowserWithOS("chrome", types.OSWindows10)
	if err != nil {
		t.Fatalf("Failed to get fingerprint: %v", err)
	}
	
	// translated comment
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

// translated comment
func TestIntegratedFingerprinter_MacOS(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()
	
	// translated comment
	fpResult, err := random.GetRandomFingerprintByBrowserWithOS("safari", types.OSMacOS14)
	if err != nil {
		t.Fatalf("Failed to get fingerprint: %v", err)
	}
	
	// translated comment
	macPacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:     64,
			SourceAddress:  "17.0.0.1", // translated comment
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

// translated comment
func TestIntegratedFingerprinter_Mobile(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()
	
	// translated comment
	profile, ok := profiles.MappedTLSClients["safari_ios_16_0"]
	if !ok {
		t.Skip("safari_ios_16_0 profile not found")
	}
	
	ua, err := useragent.GetUserAgentByProfileNameWithOS("safari_ios_16_0", types.OSMacOS14)
	if err != nil {
		t.Fatalf("Failed to get UA: %v", err)
	}
	
	// translated comment
	iphonePacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:     64,
			SourceAddress:  "17.0.0.1",
		},
		WindowSize: 65535,
	}
	
	_ = profile // translated comment
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

// translated comment
func TestIntegratedFingerprinter_Inconsistency(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()
	
	// translated comment
	fpResult, err := random.GetRandomFingerprintByBrowserWithOS("chrome", types.OSWindows10)
	if err != nil {
		t.Fatalf("Failed to get fingerprint: %v", err)
	}
	
	// translated comment
	inconsistentPacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:     64, // translated comment
			SourceAddress:  "192.168.1.100",
		},
		WindowSize: 29200, // translated comment
	}
	
	// translated comment
	result, err := fingerprinter.Analyze(inconsistentPacket, fpResult.UserAgent, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	
	// translated comment
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
	
	// translated comment
	if result.RiskScore == 0 {
		t.Logf("Warning: Risk score is 0 despite inconsistency")
	}
}

// translated comment
func TestIntegratedFingerprinter_Geolocation(t *testing.T) {
	fingerprinter := NewIntegratedFingerprinter()
	
	// translated comment
	geoDB := NewSimpleIPGeoDB()
	fingerprinter.SetIPRegionDB(geoDB)
	
	// translated comment
	fpResult, err := random.GetRandomFingerprintByBrowserWithOS("chrome", types.OSWindows10)
	if err != nil {
		t.Fatalf("Failed to get fingerprint: %v", err)
	}
	
	// translated comment
	usPacket := &TCPPacket{
		IPHeader: &IPHeader{
			TimeToLive:     128,
			SourceAddress:  "8.8.8.8", // translated comment
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

// translated comment
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
			// translated comment
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
