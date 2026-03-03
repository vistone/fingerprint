// Package tcpip 提供 TCP/IP 指纹识别的工具函数
package tcpip

import (
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
)

// OSSignature 操作系统签名定义
type OSSignature struct {
	Name          string         // OS 名称
	Family        string         // OS 家族
	DefaultTTL    int            // 默认 TTL
	WindowBase    int            // 窗口大小基数
	MSS           int            // 最大段大小
	TCPOptions    string         // TCP 选项
	IPDFBit       bool           // IP DF 位
	Quirks        string         // 特殊行为
	ProbeResponse map[int]string // 根据不同 probe 的响应
}

// BuildOSDatabase 构建操作系统数据库
func BuildOSDatabase() map[string]OSSignature {
	db := make(map[string]OSSignature)

	// Windows 系列
	db["Windows_11"] = OSSignature{
		Name:       "Windows 11",
		Family:     "Windows",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,SACK,TS,NOP,WS",
		IPDFBit:    true,
		Quirks:     "Window scaling, Selective ACK",
	}

	db["Windows_10"] = OSSignature{
		Name:       "Windows 10",
		Family:     "Windows",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,SACK,TS,NOP,WS",
		IPDFBit:    true,
		Quirks:     "Window scaling, Selective ACK",
	}

	db["Windows_Server_2019"] = OSSignature{
		Name:       "Windows Server 2019",
		Family:     "Windows",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,SACK,TS,NOP,WS",
		IPDFBit:    true,
	}

	// Linux 系列
	db["Linux_Kernel_5.x"] = OSSignature{
		Name:       "Linux Kernel 5.x",
		Family:     "Linux",
		DefaultTTL: 64,
		WindowBase: 29200,
		MSS:        1460,
		TCPOptions: "MSS,TS,TS,SACK,WS",
		IPDFBit:    true,
		Quirks:     "SYN flood protection",
	}

	db["Linux_Kernel_4.x"] = OSSignature{
		Name:       "Linux Kernel 4.x",
		Family:     "Linux",
		DefaultTTL: 64,
		WindowBase: 29200,
		MSS:        1460,
		TCPOptions: "MSS,TS,SACK,WS",
		IPDFBit:    true,
	}

	db["Ubuntu_22.04"] = OSSignature{
		Name:       "Ubuntu 22.04 LTS",
		Family:     "Linux",
		DefaultTTL: 64,
		WindowBase: 29200,
		MSS:        1460,
		TCPOptions: "MSS,TS,SACK,WS",
		IPDFBit:    true,
	}

	// macOS 系列
	db["macOS_13"] = OSSignature{
		Name:       "macOS 13 (Ventura)",
		Family:     "macOS",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,NOP,WS,NOP,NOP,TS",
		IPDFBit:    true,
		Quirks:     "Special timestamp handling",
	}

	db["macOS_12"] = OSSignature{
		Name:       "macOS 12 (Monterey)",
		Family:     "macOS",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,NOP,WS,NOP,NOP,TS",
		IPDFBit:    true,
	}

	// iOS
	db["iOS_16"] = OSSignature{
		Name:       "iOS 16",
		Family:     "iOS",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,NOP,WS,NOP,NOP,TS",
		IPDFBit:    true,
	}

	// Android
	db["Android_13"] = OSSignature{
		Name:       "Android 13",
		Family:     "Android",
		DefaultTTL: 64,
		WindowBase: 32768,
		MSS:        1460,
		TCPOptions: "MSS,SACK,TS,NOP,WS",
		IPDFBit:    true,
	}

	return db
}

// ComputeTCPSignature 计算 TCP 签名
func ComputeTCPSignature(mss int, windowSize int, options string, flags string) string {
	data := fmt.Sprintf("%d,%d,%s,%s", mss, windowSize, options, flags)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// ComputeIPSignature 计算 IP 签名
func ComputeIPSignature(ttl int, flags uint8, id uint16) string {
	data := fmt.Sprintf("%d,%d,%d", ttl, flags, id)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// MatchOSSignature 匹配操作系统签名
func MatchOSSignature(db map[string]OSSignature, ttl int, mss int, options string) string {
	bestMatch := ""
	bestScore := 0.0

	for osName, sig := range db {
		score := 0.0

		// TTL 匹配（权重 40%）
		if sig.DefaultTTL == ttl {
			score += 0.4
		} else if sig.DefaultTTL-ttl <= 10 && sig.DefaultTTL-ttl >= 0 {
			score += 0.2
		}

		// MSS 匹配（权重 30%）
		if sig.MSS == mss {
			score += 0.3
		} else if mss > sig.MSS-100 && mss < sig.MSS+100 {
			score += 0.15
		}

		// TCP 选项匹配（权重 30%）
		if strings.Contains(options, sig.TCPOptions) {
			score += 0.3
		}

		if score > bestScore {
			bestScore = score
			bestMatch = osName
		}
	}

	return bestMatch
}

// ExtractTCPOptions 提取 TCP 选项字符串
//
// TODO: 当前为简化实现，需要完整实现 TCP 选项解析
// 参考: github.com/google/gopacket 的实现方式
// TCP 选项格式: Kind(1B) | Length(1B) | Data(variable)
func ExtractTCPOptions(packet []byte) string {
	// 当前返回简化结果，实际实现需要解析 TCP 头部后的选项字段
	// 选项偏移量 = TCP 头部长度 (Data Offset * 4)
	if len(packet) < 20 {
		return ""
	}

	// TODO: 实现完整的 TCP 选项解析
	// 1. 从第 12 字节提取 Data Offset (高 4 位)
	// 2. 计算选项起始偏移量
	// 3. 解析选项直到达到选项长度
	_ = packet // 使用参数避免编译警告

	// 临时返回固定值，表示未实现
	return "MSS,SACK,TS,NOP,WS"
}

// AnalyzeTTL 分析 TTL 值推断初始 TTL
func AnalyzeTTL(currentTTL int) int {
	// 常见的初始 TTL：64/128 (Linux/Windows)
	// 根据当前 TTL 推断初始值
	if currentTTL > 64 {
		return 128
	} else if currentTTL > 32 {
		return 64
	}
	return 32
}

// AnalyzeWindowSize 分析窗口大小
func AnalyzeWindowSize(size int) string {
	if size > 60000 {
		return "Large (Windows/macOS style)"
	} else if size > 20000 {
		return "Medium (Linux style)"
	}
	return "Small"
}

// DetectSequenceNumberPattern 检测序列号模式
func DetectSequenceNumberPattern(seqNumbers []uint32) string {
	if len(seqNumbers) < 2 {
		return "Insufficient data"
	}

	// 计算差值
	diffs := make([]int64, len(seqNumbers)-1)
	for i := 0; i < len(diffs); i++ {
		diffs[i] = int64(seqNumbers[i+1]) - int64(seqNumbers[i])
	}

	// 检查是否为随机或序列
	var sum int64
	for _, d := range diffs {
		sum += d
	}
	avg := sum / int64(len(diffs))

	if avg > 1000 && avg < 100000 {
		return "Random (cryptographically secure)"
	} else if avg > 100000 {
		return "Time-based"
	}
	return "Sequential or low-entropy"
}

// AnalyzeNetworkBehavior 分析网络行为
func AnalyzeNetworkBehavior(rttValues []int64) map[string]interface{} {
	result := make(map[string]interface{})

	if len(rttValues) == 0 {
		return result
	}

	var sum, min, max int64
	min = rttValues[0]
	max = rttValues[0]

	for _, rtt := range rttValues {
		sum += rtt
		if rtt < min {
			min = rtt
		}
		if rtt > max {
			max = rtt
		}
	}

	avg := sum / int64(len(rttValues))

	result["average_rtt_ms"] = avg
	result["min_rtt_ms"] = min
	result["max_rtt_ms"] = max
	result["variance"] = max - min

	// 分类
	if avg < 10 {
		result["network_type"] = "Local LAN"
	} else if avg < 50 {
		result["network_type"] = "Domestic network"
	} else if avg < 150 {
		result["network_type"] = "Regional network"
	} else {
		result["network_type"] = "International/Satellite"
	}

	return result
}

// DetectAnomalies 检测网络异常
func DetectAnomalies(ttl int, mss int, windowSize int) []string {
	var anomalies []string

	// TTL 未设置为标准值
	if ttl != 64 && ttl != 128 && ttl != 32 {
		anomalies = append(anomalies, fmt.Sprintf("Non-standard TTL: %d", ttl))
	}

	// MSS 异常
	if mss < 536 {
		anomalies = append(anomalies, "MSS too small (potential DoS)")
	}

	// 窗口大小异常
	if windowSize < 1024 {
		anomalies = append(anomalies, "Unusually small window size")
	}

	return anomalies
}

// CalculateConfidence 计算匹配置信度
func CalculateConfidence(matches int, total int) float64 {
	if total == 0 {
		return 0.0
	}
	return float64(matches) / float64(total)
}
