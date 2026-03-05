// Package tcpip 提供 TCP/IP 指纹识别的工具函数
package tcpip

import (
	"crypto/md5"
	"fmt"
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
// 完整实现 TCP 选项解析，返回选项名称的逗号分隔字符串
// 参考: RFC 793, RFC 1323, RFC 2018
// TCP 选项格式: Kind(1B) | Length(1B) | Data(variable)
func ExtractTCPOptions(packet []byte) string {
	// TCP 头部最小长度为 20 字节
	if len(packet) < 20 {
		return ""
	}

	// 从第 12 字节提取 Data Offset (高 4 位)
	// Data Offset 表示 TCP 头部长度，单位是 32 位字（4 字节）
	dataOffset := (packet[12] >> 4) * 4

	// 验证头部长度
	if dataOffset < 20 || int(dataOffset) > len(packet) {
		return ""
	}

	// 选项从第 20 字节开始，到 Data Offset 结束
	if dataOffset == 20 {
		return "" // 没有选项
	}

	optionsData := packet[20:dataOffset]
	options := make([]string, 0, len(optionsData)/2)

	i := 0
	for i < len(optionsData) {
		kind := optionsData[i]

		// Kind 0 (EOL) - End of Option List
		if kind == 0 {
			break
		}

		// Kind 1 (NOP) - No-Operation (单字节选项)
		if kind == 1 {
			options = append(options, "NOP")
			i++
			continue
		}

		// 其他选项至少需要 2 字节 (Kind + Length)
		if i+1 >= len(optionsData) {
			break
		}

		length := int(optionsData[i+1])

		// 验证长度
		if length < 2 || i+length > len(optionsData) {
			break
		}

		// 解析具体选项
		switch kind {
		case 2:
			options = append(options, "MSS")
		case 3:
			options = append(options, "WS")
		case 4:
			options = append(options, "SACK_PERMITTED")
		case 5:
			options = append(options, "SACK")
		case 8:
			options = append(options, "TS")
		case 14:
			options = append(options, "TCP_ALTCHK")
		case 15:
			options = append(options, "TCP_ALTCHK_DATA")
		case 28:
			options = append(options, "UTO")
		case 29:
			options = append(options, "TCP_AO")
		case 30:
			options = append(options, "MP_CAPABLE")
		case 34:
			options = append(options, "TFO")
		default:
			// 未知选项，记录其 Kind 值
			options = append(options, fmt.Sprintf("OPT_%d", kind))
		}

		i += length
	}

	if len(options) == 0 {
		return ""
	}

	return strings.Join(options, ",")
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
	anomalies := make([]string, 0, 3)

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
