package main

import (
	"fmt"

	"github.com/vistone/fingerprint"
	"github.com/vistone/fingerprint/internal/tcpip"
)

func main() {
	fmt.Println("=== TCP/IP 指纹识别示例 ===\n")

	// 例子 1: 分析单个 TCP/IP 数据包
	example1SinglePacket()

	// 例子 2: 匹配操作系统
	example2MatchOS()

	// 例子 3: 检测网络异常
	example3DetectAnomalies()

	// 例子 4: 分析网络行为
	example4AnalyzeNetworkBehavior()
}

// 例子 1: 分析单个 TCP/IP 数据包
func example1SinglePacket() {
	fmt.Println("1️⃣  分析 TCP/IP 数据包:\n")

	analyzer := fingerprint.NewTCPIPAnalyzer()

	// 创建一个模拟的 TCP/IP 数据包
	packet := fingerprint.TCPPacket{
		IPHeader: fingerprint.IPHeader{
			Version:  4,
			TTL:      64,
			ID:       0x1234,
			Protocol: 6, // TCP
		},
		SrcPort: 1234,
		DstPort: 80,
		SeqNum:  0x12345678,
		AckNum:  0x87654321,
		Flags: fingerprint.TCPFlags{
			SYN: true,
			ACK: false,
		},
		WindowSize: 65535,
		Options: fingerprint.TCPOptions{
			MSS:         ptrUint16(1460),
			WindowScale: ptrUint8(7),
			SACK:        true,
			Timestamps:  true,
		},
	}

	// 添加数据包
	analyzer.AddPacket(packet)

	// 分析数据包
	sig, err := analyzer.AnalyzePacket(packet)
	if err != nil {
		fmt.Printf("❌ 分析失败: %v\n", err)
		return
	}

	fmt.Printf("✓ 数据包哈希: %s\n", sig.Hash)
	fmt.Printf("✓ 操作系统: %s\n", sig.OS)
	fmt.Printf("✓ TTL 值: %d\n", sig.TTLValue)
	fmt.Printf("✓ 窗口大小: %v\n", sig.Features["WindowSize"])
}

// 例子 2: 操作系统匹配
func example2MatchOS() {
	fmt.Println("\n2️⃣  操作系统匹配:\n")

	// 构建 OS 数据库
	osDB := tcpip.BuildOSDatabase()

	// 测试数据包特征
	testCases := []struct {
		name    string
		ttl     int
		mss     int
		options string
	}{
		{"Windows 11", 64, 1460, "MSS,SACK,TS,NOP,WS"},
		{"Linux", 64, 1460, "MSS,TS,SACK,WS"},
		{"macOS", 64, 1460, "MSS,NOP,WS,NOP,NOP,TS"},
	}

	for _, tc := range testCases {
		matched := tcpip.MatchOSSignature(osDB, tc.ttl, tc.mss, tc.options)
		fmt.Printf("✓ 特征 (%s): 匹配到 %s\n", tc.name, matched)
	}
}

// 例子 3: 检测网络异常
func example3DetectAnomalies() {
	fmt.Println("\n3️⃣  检测网络异常:\n")

	// 正常特征
	fmt.Println("正常特征:")
	normal := tcpip.DetectAnomalies(64, 1460, 65535)
	if len(normal) == 0 {
		fmt.Println("✓ 无异常检测到")
	} else {
		for _, a := range normal {
			fmt.Printf("⚠️  %s\n", a)
		}
	}

	// 异常特征
	fmt.Println("\n异常特征:")
	abnormal := tcpip.DetectAnomalies(100, 500, 512)
	for _, a := range abnormal {
		fmt.Printf("⚠️  %s\n", a)
	}
}

// 例子 4: 分析网络行为
func example4AnalyzeNetworkBehavior() {
	fmt.Println("\n4️⃣  分析网络行为:\n")

	// 模拟 RTT 数据
	rtts := []int64{5, 6, 4, 5, 7, 5, 6, 4, 5, 6}

	analysis := tcpip.AnalyzeNetworkBehavior(rtts)

	fmt.Printf("✓ 平均 RTT: %d ms\n", analysis["average_rtt_ms"])
	fmt.Printf("✓ 最小 RTT: %d ms\n", analysis["min_rtt_ms"])
	fmt.Printf("✓ 最大 RTT: %d ms\n", analysis["max_rtt_ms"])
	fmt.Printf("✓ 网络类型: %s\n", analysis["network_type"])
	fmt.Printf("✓ 抖动: %d ms\n", analysis["variance"])
}

// 辅助函数
func ptr[T any](val T) *T {
	return &val
}

func ptrUint16(val uint16) *uint16 {
	return &val
}

func ptrUint8(val uint8) *uint8 {
	return &val
}
