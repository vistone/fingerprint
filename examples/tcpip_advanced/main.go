package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/vistone/fingerprint/internal/tcpip"
)

func main() {
	fmt.Println("=== TCP/IP 高级指纹识别演示 ===\n")

	// 演示 1: 网络行为分析
	demonstrateNetworkBehaviorAnalysis()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// 演示 2: 设备指纹识别引擎
	demonstrateDeviceFingerprintingEngine()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// 演示 3: 数据包解析和签名
	demonstratePacketParsingAndSignatures()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// 演示 4: 风险评估
	demonstrateRiskAssessment()
}

// demonstrateNetworkBehaviorAnalysis 网络行为分析演示
func demonstrateNetworkBehaviorAnalysis() {
	fmt.Println("📊 演示 1: 网络行为分析")
	fmt.Println("-" + strings.Repeat("-", 49))

	// 创建行为分析器
	analyzer := tcpip.NewNetworkBehaviorAnalyzer()

	// 创建模拟数据包
	packets := createMockPackets(10)
	rttValues := []time.Duration{
		5 * time.Millisecond,
		6 * time.Millisecond,
		5 * time.Millisecond,
		7 * time.Millisecond,
		5 * time.Millisecond,
	}

	// 记录数据包和 RTT
	for i := 0; i < len(packets) && i < len(rttValues); i++ {
		analyzer.RecordPacket(packets[i], rttValues[i])
	}

	// 执行分析
	result := analyzer.AnalyzeBehavior()

	fmt.Printf("📦 总数据包: %d\n", result.TotalPackets)
	fmt.Printf("⏱️  平均 RTT: %v\n", result.RTTAnalysis.AverageRTT)
	fmt.Printf("   最小 RTT: %v\n", result.RTTAnalysis.MinRTT)
	fmt.Printf("   最大 RTT: %v\n", result.RTTAnalysis.MaxRTT)
	fmt.Printf("🌐 网络类型: %s\n", result.RTTAnalysis.NetworkType)
	fmt.Printf("🔢 序列号模式: %s\n", result.SequenceNumberPattern)
	fmt.Printf("⏳ 时间模式: %s\n", result.TimingPattern)
	fmt.Printf("🏷️  行为特征: %v\n", result.BehaviorCharacteristics)
	fmt.Printf("📊 协议分布: %v\n", result.ProtocolDistribution)
	fmt.Printf("\n✅ 网络行为分析完成!")
}

// demonstrateDeviceFingerprintingEngine 设备指纹识别引擎演示
func demonstrateDeviceFingerprintingEngine() {
	fmt.Println("🔍 演示 2: 设备指纹识别引擎")
	fmt.Println("-" + strings.Repeat("-", 49))

	// 创建引擎
	engine := tcpip.NewDeviceFingerprintingEngine()

	// 注册设备轮廓
	windows11Profile := &tcpip.DeviceProfile{
		Name:         "Windows11_Desktop",
		DeviceType:   "desktop",
		Manufacturer: "Microsoft",
		OS:           "Windows",
		OSVersion:    "11",
		BrowserSignatures: map[string]string{
			"Chrome": "Chrome/120",
		},
		NetworkSignatures: map[string]string{
			"tcp_window": "65535",
			"mss":        "1460",
		},
		Applications: []string{"Chrome", "Firefox", "Edge"},
		Confidence:   0.85,
	}

	linuxProfile := &tcpip.DeviceProfile{
		Name:         "Linux_Server",
		DeviceType:   "server",
		Manufacturer: "Various",
		OS:           "Linux",
		OSVersion:    "5.x",
		NetworkSignatures: map[string]string{
			"tcp_window": "29200",
			"mss":        "1460",
		},
		Applications: []string{"nginx", "apache"},
		Confidence:   0.90,
	}

	engine.RegisterDeviceProfile(windows11Profile)
	engine.RegisterDeviceProfile(linuxProfile)

	// 创建测试数据包
	packets := createMockPackets(5)

	// 创建行为分析结果
	behaviorAnalyzer := tcpip.NewNetworkBehaviorAnalyzer()
	for _, pkt := range packets {
		behaviorAnalyzer.RecordPacket(pkt, 5*time.Millisecond)
	}
	behaviorResult := behaviorAnalyzer.AnalyzeBehavior()

	// 分析设备
	result := engine.AnalyzeDevice(packets, behaviorResult)

	fmt.Printf("📊 分析数据包数: %d\n", result.AnalyzedPackets)
	fmt.Printf("🎯 总体信心度: %.2f%%\n", result.Confidence*100)
	fmt.Printf("🌐 网络类型: %s\n", result.NetworkType)

	if len(result.DeviceMatches) > 0 {
		fmt.Println("\n🖥️  设备匹配:")
		for i, match := range result.DeviceMatches {
			fmt.Printf("   %d. %s (%s) - 匹配度: %.2f%%\n",
				i+1, match.DeviceName, match.DeviceType, match.MatchScore*100)
		}
	}

	if len(result.OSMatches) > 0 {
		fmt.Println("\n💻 操作系统匹配:")
		for i, match := range result.OSMatches {
			fmt.Printf("   %d. %s (%s) - 匹配度: %.2f%%\n",
				i+1, match.OSName, match.OSFamily, match.MatchScore*100)
		}
	}

	if len(result.RiskIndicators) > 0 {
		fmt.Println("\n⚠️  风险指标:")
		for _, risk := range result.RiskIndicators {
			fmt.Printf("   • %s\n", risk)
		}
	} else {
		fmt.Println("\n✅ 未检测到风险指标")
	}
}

// demonstratePacketParsingAndSignatures 数据包解析和签名演示
func demonstratePacketParsingAndSignatures() {
	fmt.Println("🔐 演示 3: 数据包解析和签名")
	fmt.Println("-" + strings.Repeat("-", 49))

	// 创建模拟数据包
	packet := &tcpip.TCPPacket{
		IPHeader: &tcpip.IPHeader{
			Version:        4,
			TimeToLive:     64,
			Identification: 12345,
			Flags:          0x40, // DF 标志
			Protocol:       6,    // TCP
			SourceAddress:  "192.168.1.100",
			DestAddress:    "8.8.8.8",
		},
		SourcePort:      56789,
		DestinationPort: 443,
		SequenceNumber:  1000000,
		AckNumber:       0,
		Flags:           0x02, // SYN
		WindowSize:      65535,
		Options:         []byte{2, 4, 5, 0xb4, 3, 3, 8, 1, 1, 4, 2}, // MSS, WS, SACK, TS
		Timestamp:       time.Now(),
	}

	parser := tcpip.NewPacketParser([]byte{})

	// 生成签名
	signature := tcpip.FormatSignature(
		int(packet.IPHeader.TimeToLive),
		1460,
		65535,
		"MSS,WS,SACK,TS",
	)

	fmt.Printf("📋 数据包详情:\n")
	fmt.Printf("   源地址: %s:%d\n", packet.IPHeader.SourceAddress, packet.SourcePort)
	fmt.Printf("   目标地址: %s:%d\n", packet.IPHeader.DestAddress, packet.DestinationPort)
	fmt.Printf("   TTL: %d\n", packet.IPHeader.TimeToLive)
	fmt.Printf("   IP ID: %d\n", packet.IPHeader.Identification)
	fmt.Printf("   窗口大小: %d\n", packet.WindowSize)

	// 提取标志
	flags := parser.ExtractFlags(packet.Flags)
	fmt.Printf("   TCP 标志: %v\n", flags)

	// 检查 DF 标志
	if (packet.IPHeader.Flags & 0x40) != 0 {
		fmt.Printf("   DF 标志: 已设置\n")
	}

	fmt.Printf("\n🔑 生成的签名:\n")
	fmt.Printf("   %s\n", signature)

	// 检验 IP
	fmt.Printf("\n🔒 IP 验证:\n")
	if tcpip.IsPrivateIP(packet.IPHeader.SourceAddress) {
		fmt.Printf("   源地址为私有 IP\n")
	}
	if !tcpip.IsPrivateIP(packet.IPHeader.DestAddress) {
		fmt.Printf("   目标地址为公网 IP\n")
	}

	// 推断初始 TTL
	initialTTL := tcpip.InferInitialTTL(int(packet.IPHeader.TimeToLive), 3)
	fmt.Printf("\n📊 推断初始 TTL (跳转数=3): %d\n", initialTTL)

	fmt.Printf("\n✅ 数据包解析完成!")
}

// demonstrateRiskAssessment 风险评估演示
func demonstrateRiskAssessment() {
	fmt.Println("⚠️  演示 4: 风险评估")
	fmt.Println("-" + strings.Repeat("-", 49))

	// 创建各种风险指标
	riskIndicators := tcpip.RiskIndicators{
		IsBot:      false,
		IsScanner:  false,
		IsVPN:      false,
		IsProxy:    false,
		IsNAT:      false,
		IsSpoofed:  false,
		Suspicious: []string{"unusual_tcp_options", "non_standard_mss"},
	}

	// 计算风险评分
	riskScore := tcpip.CalculateRiskScore(riskIndicators)

	fmt.Printf("🎯 风险评估结果:\n")
	fmt.Printf("   机器人检测: %v\n", riskIndicators.IsBot)
	fmt.Printf("   扫描器检测: %v\n", riskIndicators.IsScanner)
	fmt.Printf("   VPN 检测: %v\n", riskIndicators.IsVPN)
	fmt.Printf("   代理检测: %v\n", riskIndicators.IsProxy)
	fmt.Printf("   NAT 检测: %v\n", riskIndicators.IsNAT)
	fmt.Printf("   IP 欺骗检测: %v\n", riskIndicators.IsSpoofed)

	if len(riskIndicators.Suspicious) > 0 {
		fmt.Println("\n   🚩 可疑特征:")
		for _, susp := range riskIndicators.Suspicious {
			fmt.Printf("      • %s\n", susp)
		}
	}

	fmt.Printf("\n📊 风险评分: %.2f (满分为 1.0)\n", riskScore)

	// 风险等级
	riskLevel := "低"
	if riskScore > 0.7 {
		riskLevel = "严重"
	} else if riskScore > 0.5 {
		riskLevel = "中等"
	} else if riskScore > 0.3 {
		riskLevel = "较低"
	}

	fmt.Printf("⚡ 风险等级: %s\n", riskLevel)
	fmt.Printf("\n✅ 风险评估完成!")
}

// createMockPackets 创建模拟数据包
func createMockPackets(count int) []*tcpip.TCPPacket {
	packets := make([]*tcpip.TCPPacket, count)

	for i := 0; i < count; i++ {
		packets[i] = &tcpip.TCPPacket{
			IPHeader: &tcpip.IPHeader{
				Version:        4,
				TimeToLive:     64,
				Identification: uint16(1000 + i),
				Flags:          0x40,
				Protocol:       6,
				SourceAddress:  "192.168.1.100",
				DestAddress:    "8.8.8.8",
			},
			SourcePort:      56789,
			DestinationPort: 443,
			SequenceNumber:  uint32(1000000 + i*100),
			AckNumber:       0,
			Flags:           0x02,
			WindowSize:      65535,
			Options:         []byte{2, 4, 5, 0xb4, 3, 3, 8, 1, 1, 4, 2},
			Timestamp:       time.Now().Add(time.Duration(i*100) * time.Millisecond),
			Payload:         make([]byte, 0),
		}
	}

	return packets
}
