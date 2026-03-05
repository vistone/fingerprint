package ja4t

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// JA4TResult JA4T 指纹结果（TCP 客户端指纹）
type JA4TResult struct {
	// JA4T 完整指纹（原始字符串形式）
	RawFingerprint string

	// JA4T 哈希（SHA256 前 12 个字符）
	Hash string

	// 窗口大小
	WindowSize uint16

	// TCP 选项类型列表（按原始顺序）
	Options []uint8

	// 最大段大小（MSS）值
	MSS uint16

	// 窗口缩放因子
	WindowScale uint8

	// 异常标记
	AnomalyFlags []string

	// 风险评分 (0.0-1.0)
	RiskScore float64

	// 推测的操作系统
	ProbableOS string
}

// JA4TSResult JA4TS 指纹结果（TCP 服务端指纹）
type JA4TSResult struct {
	// JA4TS 完整指纹（原始字符串形式）
	RawFingerprint string

	// JA4TS 哈希（SHA256 前 12 个字符）
	Hash string

	// 窗口大小
	WindowSize uint16

	// TCP 选项类型列表（按原始顺序）
	Options []uint8

	// 最大段大小（MSS）值
	MSS uint16

	// 窗口缩放因子
	WindowScale uint8
}

// TCPSYNData TCP SYN 数据包特征
type TCPSYNData struct {
	// 窗口大小
	WindowSize uint16

	// TCP 选项（按原始顺序，Option Kind 值）
	// 常见值: 2=MSS, 1=NOP, 3=Window Scale, 4=SACK Permitted, 8=Timestamps
	Options []uint8

	// 最大段大小（MSS）值
	MSS uint16

	// 窗口缩放因子
	WindowScale uint8

	// IP TTL（可选，用于 OS 检测）
	TTL uint8

	// IP DF 标志（Don't Fragment）
	DF bool
}

// TCPOptionKind TCP 选项类型常量
const (
	TCPOptionEndOfList  uint8 = 0  // 选项列表结束
	TCPOptionNOP        uint8 = 1  // 无操作（填充）
	TCPOptionMSS        uint8 = 2  // 最大段大小
	TCPOptionWindowScale uint8 = 3  // 窗口缩放
	TCPOptionSACKPermit uint8 = 4  // 选择性确认许可
	TCPOptionSACK       uint8 = 5  // 选择性确认
	TCPOptionTimestamps uint8 = 8  // 时间戳
)

// tcpOptionName TCP 选项类型名称
func tcpOptionName(kind uint8) string {
	switch kind {
	case TCPOptionEndOfList:
		return "EOL"
	case TCPOptionNOP:
		return "NOP"
	case TCPOptionMSS:
		return "MSS"
	case TCPOptionWindowScale:
		return "WS"
	case TCPOptionSACKPermit:
		return "SACK"
	case TCPOptionSACK:
		return "SACK_DATA"
	case TCPOptionTimestamps:
		return "TS"
	default:
		return fmt.Sprintf("%d", kind)
	}
}

// ComputeJA4T 从 TCP SYN 数据计算 JA4T 指纹
func ComputeJA4T(data TCPSYNData) *JA4TResult {
	result := &JA4TResult{
		WindowSize:   data.WindowSize,
		Options:      data.Options,
		MSS:          data.MSS,
		WindowScale:  data.WindowScale,
		AnomalyFlags: []string{},
	}

	// 构建选项字符串：用 "-" 分隔的选项类型值
	optionParts := make([]string, len(data.Options))
	for i, opt := range data.Options {
		optionParts[i] = fmt.Sprintf("%d", opt)
	}
	optionsStr := strings.Join(optionParts, "-")

	// JA4T 格式: {window_size}_{options}_{mss}_{window_scale}
	result.RawFingerprint = fmt.Sprintf("%d_%s_%d_%d",
		data.WindowSize,
		optionsStr,
		data.MSS,
		data.WindowScale,
	)

	// 计算 SHA256 哈希（前 12 个字符）
	hash := sha256.Sum256([]byte(result.RawFingerprint))
	result.Hash = fmt.Sprintf("%x", hash)[:12]

	// 异常检测
	detectTCPAnomalies(data, result)

	// OS 推测
	result.ProbableOS = guessOS(data)

	return result
}

// ComputeJA4TS 从 TCP SYN-ACK 数据计算 JA4TS 指纹（服务端）
func ComputeJA4TS(data TCPSYNData) *JA4TSResult {
	result := &JA4TSResult{
		WindowSize:  data.WindowSize,
		Options:     data.Options,
		MSS:         data.MSS,
		WindowScale: data.WindowScale,
	}

	// 构建选项字符串
	optionParts := make([]string, len(data.Options))
	for i, opt := range data.Options {
		optionParts[i] = fmt.Sprintf("%d", opt)
	}
	optionsStr := strings.Join(optionParts, "-")

	// JA4TS 格式与 JA4T 相同
	result.RawFingerprint = fmt.Sprintf("%d_%s_%d_%d",
		data.WindowSize,
		optionsStr,
		data.MSS,
		data.WindowScale,
	)

	// 计算哈希
	hash := sha256.Sum256([]byte(result.RawFingerprint))
	result.Hash = fmt.Sprintf("%x", hash)[:12]

	return result
}

// MatchJA4T 比较两个 JA4T 哈希是否匹配
func MatchJA4T(hash1, hash2 string) bool {
	return len(hash1) == 12 && len(hash2) == 12 && hash1 == hash2
}

// detectTCPAnomalies 检测 TCP 异常特征
func detectTCPAnomalies(data TCPSYNData, result *JA4TResult) {
	baseScore := 0.0

	// 异常 1: 窗口大小为 0 或异常小
	if data.WindowSize == 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "ZERO_WINDOW")
		baseScore += 0.3
	} else if data.WindowSize < 512 {
		result.AnomalyFlags = append(result.AnomalyFlags, "SMALL_WINDOW")
		baseScore += 0.15
	}

	// 异常 2: 缺少 MSS 选项（大多数正常实现都包含）
	hasMSS := false
	for _, opt := range data.Options {
		if opt == TCPOptionMSS {
			hasMSS = true
			break
		}
	}
	if !hasMSS {
		result.AnomalyFlags = append(result.AnomalyFlags, "NO_MSS")
		baseScore += 0.2
	}

	// 异常 3: MSS 值异常
	if data.MSS > 0 && data.MSS < 536 {
		result.AnomalyFlags = append(result.AnomalyFlags, "LOW_MSS")
		baseScore += 0.15
	}

	// 异常 4: 无 TCP 选项（非常罕见）
	if len(data.Options) == 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "NO_OPTIONS")
		baseScore += 0.25
	}

	// 异常 5: 窗口缩放因子过大（>14 不合理）
	if data.WindowScale > 14 {
		result.AnomalyFlags = append(result.AnomalyFlags, "EXCESSIVE_WINDOW_SCALE")
		baseScore += 0.2
	}

	// 异常 6: TTL 异常值
	if data.TTL > 0 && data.TTL < 32 {
		result.AnomalyFlags = append(result.AnomalyFlags, "LOW_TTL")
		baseScore += 0.15
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}
	result.RiskScore = baseScore
}

// guessOS 根据 TCP SYN 特征推测操作系统
func guessOS(data TCPSYNData) string {
	// 基于常见的 TCP/IP 栈特征进行 OS 推测
	// 主要参考: TTL 默认值、窗口大小、MSS、选项顺序

	optionsStr := formatOptions(data.Options)

	// Windows 特征: TTL=128, 窗口大小=65535 或 8192 的倍数
	if data.TTL > 96 && data.TTL <= 128 {
		if data.WindowSize == 65535 || data.WindowSize%8192 == 0 {
			return "Windows"
		}
	}

	// macOS/iOS 特征: TTL=64, 窗口大小=65535, 选项包含 1-1-8-4-2-3 或 2-1-1-4-8-3
	if data.TTL > 32 && data.TTL <= 64 {
		if data.WindowSize == 65535 && strings.Contains(optionsStr, "1-1-4-8") {
			return "macOS"
		}
	}

	// Linux 特征: TTL=64, 常见选项顺序 2-1-3-1-1-8-4 或 2-4-8-1-3
	if data.TTL > 32 && data.TTL <= 64 {
		if strings.HasPrefix(optionsStr, "2-1-3") || strings.HasPrefix(optionsStr, "2-4-8-1-3") {
			return "Linux"
		}
	}

	// FreeBSD 特征: TTL=64, 窗口大小=65535
	if data.TTL > 32 && data.TTL <= 64 {
		if data.WindowSize == 65535 {
			return "FreeBSD"
		}
	}

	// Solaris/AIX: TTL=254 或 255
	if data.TTL >= 254 {
		return "Solaris/AIX"
	}

	return "Unknown"
}

// formatOptions 格式化选项列表为字符串
func formatOptions(options []uint8) string {
	parts := make([]string, len(options))
	for i, opt := range options {
		parts[i] = fmt.Sprintf("%d", opt)
	}
	return strings.Join(parts, "-")
}

// GetOptionNames 获取选项的可读名称列表
func GetOptionNames(options []uint8) []string {
	names := make([]string, len(options))
	for i, opt := range options {
		names[i] = tcpOptionName(opt)
	}
	return names
}

// KnownOSProfiles 返回已知操作系统的 TCP 特征
func KnownOSProfiles() []OSProfile {
	return []OSProfile{
		{
			Name:        "Windows 10/11",
			TTL:         128,
			WindowSize:  65535,
			MSS:         1460,
			WindowScale: 8,
			Options:     []uint8{2, 1, 3, 1, 1, 4},
		},
		{
			Name:        "Linux 5.x/6.x",
			TTL:         64,
			WindowSize:  65535,
			MSS:         1460,
			WindowScale: 7,
			Options:     []uint8{2, 4, 8, 1, 3},
		},
		{
			Name:        "macOS 14.x (Sonoma)",
			TTL:         64,
			WindowSize:  65535,
			MSS:         1460,
			WindowScale: 6,
			Options:     []uint8{2, 1, 1, 4, 8, 3},
		},
		{
			Name:        "iOS 17.x",
			TTL:         64,
			WindowSize:  65535,
			MSS:         1460,
			WindowScale: 6,
			Options:     []uint8{2, 1, 1, 4, 8, 3},
		},
		{
			Name:        "Android 14",
			TTL:         64,
			WindowSize:  65535,
			MSS:         1460,
			WindowScale: 7,
			Options:     []uint8{2, 4, 8, 1, 3},
		},
		{
			Name:        "FreeBSD 14",
			TTL:         64,
			WindowSize:  65535,
			MSS:         1460,
			WindowScale: 7,
			Options:     []uint8{2, 1, 3, 4, 8, 1, 1},
		},
	}
}

// OSProfile 操作系统 TCP 特征
type OSProfile struct {
	// 操作系统名称
	Name string

	// 默认 TTL
	TTL uint8

	// 默认窗口大小
	WindowSize uint16

	// 默认 MSS
	MSS uint16

	// 默认窗口缩放
	WindowScale uint8

	// TCP 选项顺序
	Options []uint8
}

// ToSYNData 转换为 TCPSYNData
func (p *OSProfile) ToSYNData() TCPSYNData {
	return TCPSYNData{
		WindowSize:  p.WindowSize,
		Options:     p.Options,
		MSS:         p.MSS,
		WindowScale: p.WindowScale,
		TTL:         p.TTL,
		DF:          true,
	}
}
