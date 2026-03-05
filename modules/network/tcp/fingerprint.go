package tcp

import (
	"crypto/md5"
	"fmt"
	"strings"
)

// TCPFlags TCP 标志位
type TCPFlags struct {
	SYN bool // 同步标志（建立连接）
	ACK bool // 确认标志
	FIN bool // 结束标志
	RST bool // 重置标志
	PSH bool // 推送标志
	URG bool // 紧急标志
}

// IPHeader IP 包头参数
type IPHeader struct {
	Version    uint8  // IP 版本（4 或 6）
	TTL        uint8  // 生存时间
	TotalLen   uint16 // 总长度
	Flags      uint8  // 标志位（DF、MF、RF）
	FragOffset uint16 // 分片偏移
	ID         uint16 // 标识符
	Protocol   uint8  // 协议号（6=TCP, 17=UDP）
	Checksum   uint16 // 校验和
	Src        string // 源 IP
	Dst        string // 目标 IP
}

// TCPOptions TCP 选项
type TCPOptions struct {
	MSS           *uint16 // 最大段大小
	WindowScale   *uint8  // 窗口缩放因子
	SACK          bool    // 选择性确认
	Timestamps    bool    // 时间戳
	SAckPermitted bool    // SACK 许可
	NoOperation   int     // NOP 数量
	EndOfOptions  bool    // 选项列表结束
	WindowSize    uint16  // 窗口大小
	OptionsMD5    string  // 选项的哈希指纹
}

// TCPPacket TCP 数据包参数
type TCPPacket struct {
	// IP 层
	IPHeader IPHeader

	// TCP 层
	SrcPort    uint16   // 源端口
	DstPort    uint16   // 目标端口
	SeqNum     uint32   // 序列号
	AckNum     uint32   // 确认号
	Flags      TCPFlags // 标志位
	WindowSize uint16   // 窗口大小
	Checksum   uint16   // TCP 校验和
	UrgentPtr  uint16   // 紧急指针
	DataLen    uint16   // 数据长度

	// TCP 选项
	Options TCPOptions

	// 时间和统计
	Timestamp   int64 // 数据包时间戳
	RoundTripMs int64 // 往返时间（毫秒）
}

// TCPIPSignature TCP/IP 指纹签名
type TCPIPSignature struct {
	Hash              string            // MD5 哈希
	RawSignature      string            // 原始签名字符串
	OS                string            // 识别的操作系统
	OSVersion         string            // 操作系统版本
	Confidence        float64           // 置信度（0.0-1.0）
	MatchedProfiles   []string          // 匹配的指纹配置文件
	TTLValue          int               // TTL 值
	WindowSizeFamily  string            // 窗口大小系列（如 "Linux", "Windows"）
	MSS               int               // 最大段大小
	OptimizationLevel string            // 优化级别
	Features          map[string]string // 其他特征
}

// OSFingerprint 操作系统指纹
type OSFingerprint struct {
	Name           string  // 操作系统名称（如 "Windows 11"）
	Family         string  // OS 系列（Windows、Linux、macOS 等）
	Version        string  // 版本号
	DefaultTTL     int     // 默认 TTL 值
	WindowSizes    []int   // 常见窗口大小（从小到大）
	MSS            int     // 典型的 MSS 值
	TCPOptions     string  // TCP 选项特征
	IPFlags        uint8   // IP 标志特征
	Quirks         string  // 操作系统特性和古怪行为
	SYNACKSequence string  // SYN-ACK 序列特征
	Probability    float64 // 匹配概率
}

// TCPIPAnalyzer TCP/IP 指纹分析器
type TCPIPAnalyzer struct {
	packets    []TCPPacket              // 数据包列表
	signatures []TCPIPSignature         // 签名列表
	osDatabase []OSFingerprint          // 操作系统数据库
	matchCache map[string]OSFingerprint // 匹配缓存
}

// TCPIPResult TCP/IP 分析结果
type TCPIPResult struct {
	// 操作系统识别
	OS            string   // 识别的操作系统
	OSFamily      string   // 操作系统家族
	Confidence    float64  // 识别置信度
	CandidateOSes []string // 候选操作系统列表

	// 网络特征
	InitialTTL        int   // 初始 TTL 值
	AverageWindowSize int   // 平均窗口大小
	MSS               int   // 最大段大小
	NetworkLatency    int64 // 网络延迟（毫秒）

	// TCP 行为
	SeqNumberBehavior string // 序列号生成行为
	AckBehavior       string // ACK 号行为
	ResetBehavior     string // 重置行为

	// 安全指标
	RiskScore      float64  // 风险评分
	AnomaliesFound []string // 发现的异常
	IsVPN          bool     // 是否使用 VPN
	IsProxy        bool     // 是否使用代理
	IsNAT          bool     // 是否使用 NAT

	// 详细签名
	Signature TCPIPSignature
}

// NewTCPIPAnalyzer 创建新的分析器
func NewTCPIPAnalyzer() *TCPIPAnalyzer {
	return &TCPIPAnalyzer{
		packets:    []TCPPacket{},
		signatures: []TCPIPSignature{},
		osDatabase: []OSFingerprint{},
		matchCache: make(map[string]OSFingerprint),
	}
}

// AddPacket 添加 TCP/IP 数据包
func (t *TCPIPAnalyzer) AddPacket(packet TCPPacket) {
	t.packets = append(t.packets, packet)
}

// AnalyzePacket 分析单个数据包
func (t *TCPIPAnalyzer) AnalyzePacket(packet TCPPacket) (TCPIPSignature, error) {
	sigStr, err := t.ComputeSignature(packet)
	if err != nil {
		return TCPIPSignature{}, err
	}

	hash := md5.Sum([]byte(sigStr))
	sig := TCPIPSignature{
		Hash:         fmt.Sprintf("%x", hash),
		RawSignature: sigStr,
		TTLValue:     int(packet.IPHeader.TTL),
		MSS:          getMSSValue(packet.Options),
		Features:     make(map[string]string),
	}

	// OS 推测
	sig.OS, sig.OSVersion, sig.Confidence = guessOSFromPacket(packet)
	sig.WindowSizeFamily = windowSizeFamily(packet.WindowSize)

	return sig, nil
}

// AnalyzeStream 分析数据包流
func (t *TCPIPAnalyzer) AnalyzeStream() (TCPIPResult, error) {
	result := TCPIPResult{
		CandidateOSes:  []string{},
		AnomaliesFound: []string{},
	}

	if len(t.packets) == 0 {
		return result, nil
	}

	// 分析每个数据包
	osCounts := make(map[string]int)
	var totalWindowSize int64
	var totalLatency int64
	latencyCount := 0

	for _, packet := range t.packets {
		sig, err := t.AnalyzePacket(packet)
		if err != nil {
			continue
		}
		t.signatures = append(t.signatures, sig)

		if sig.OS != "" && sig.OS != "Unknown" {
			osCounts[sig.OS]++
		}
		totalWindowSize += int64(packet.WindowSize)

		if packet.RoundTripMs > 0 {
			totalLatency += packet.RoundTripMs
			latencyCount++
		}
	}

	// 确定最可能的操作系统
	maxCount := 0
	for os, count := range osCounts {
		if count > maxCount {
			maxCount = count
			result.OS = os
		}
		result.CandidateOSes = append(result.CandidateOSes, os)
	}

	if maxCount > 0 {
		result.Confidence = float64(maxCount) / float64(len(t.packets))
	}

	// 平均窗口大小
	if len(t.packets) > 0 {
		result.AverageWindowSize = int(totalWindowSize / int64(len(t.packets)))
	}

	// 网络延迟
	if latencyCount > 0 {
		result.NetworkLatency = totalLatency / int64(latencyCount)
	}

	// 获取首包特征
	if len(t.packets) > 0 {
		first := t.packets[0]
		result.InitialTTL = nearestDefaultTTL(first.IPHeader.TTL)
		result.MSS = getMSSValue(first.Options)
	}

	// 异常检测
	result.AnomaliesFound = t.DetectAnomalies()
	result.RiskScore = t.GetRiskScore()
	result.IsVPN = t.DetectVPN()
	result.IsProxy = t.DetectProxy()
	result.IsNAT = t.DetectNAT()

	// 生成综合签名
	if len(t.signatures) > 0 {
		result.Signature = t.signatures[0]
	}

	return result, nil
}

// ComputeSignature 计算 TCP/IP 签名
func (t *TCPIPAnalyzer) ComputeSignature(packet TCPPacket) (string, error) {
	// 签名格式: TTL:WindowSize:DF:Options:MSS:WindowScale
	var parts []string

	// TTL (推测初始 TTL)
	parts = append(parts, fmt.Sprintf("%d", nearestDefaultTTL(packet.IPHeader.TTL)))

	// 窗口大小
	parts = append(parts, fmt.Sprintf("%d", packet.WindowSize))

	// DF 标志
	df := "0"
	if packet.IPHeader.Flags&0x02 != 0 { // DF bit
		df = "1"
	}
	parts = append(parts, df)

	// TCP 选项指纹
	parts = append(parts, formatTCPOptions(packet.Options))

	// MSS
	mss := getMSSValue(packet.Options)
	parts = append(parts, fmt.Sprintf("%d", mss))

	// 窗口缩放
	ws := getWindowScale(packet.Options)
	parts = append(parts, fmt.Sprintf("%d", ws))

	return strings.Join(parts, ":"), nil
}

// GetOSFingerprints 获取操作系统指纹库
func (t *TCPIPAnalyzer) GetOSFingerprints() []OSFingerprint {
	return t.osDatabase
}

// SetOSDatabase 设置操作系统数据库
func (t *TCPIPAnalyzer) SetOSDatabase(db []OSFingerprint) {
	t.osDatabase = db
}

// GetRiskScore 计算风险评分
func (t *TCPIPAnalyzer) GetRiskScore() float64 {
	if len(t.packets) == 0 {
		return 0.0
	}

	score := 0.0

	for _, packet := range t.packets {
		// TTL 异常
		if packet.IPHeader.TTL > 0 && packet.IPHeader.TTL < 32 {
			score += 0.15
		}

		// 窗口大小异常
		if packet.WindowSize == 0 {
			score += 0.2
		}

		// RST 泛滥
		if packet.Flags.RST {
			score += 0.1
		}
	}

	// 归一化
	score = score / float64(len(t.packets))
	if score > 1.0 {
		score = 1.0
	}
	return score
}

// DetectAnomalies 检测网络异常
func (t *TCPIPAnalyzer) DetectAnomalies() []string {
	var anomalies []string

	if len(t.packets) == 0 {
		return anomalies
	}

	// 检测 TTL 不一致（可能表示路由变更或中间人）
	ttlSet := make(map[uint8]bool)
	for _, p := range t.packets {
		if p.IPHeader.TTL > 0 {
			ttlSet[p.IPHeader.TTL] = true
		}
	}
	if len(ttlSet) > 3 {
		anomalies = append(anomalies, "TTL_INCONSISTENCY")
	}

	// 检测异常的 RST 数量
	rstCount := 0
	for _, p := range t.packets {
		if p.Flags.RST {
			rstCount++
		}
	}
	if rstCount > len(t.packets)/3 {
		anomalies = append(anomalies, "EXCESSIVE_RST")
	}

	// 检测窗口大小为 0 的包（窗口探测或异常）
	zeroWindowCount := 0
	for _, p := range t.packets {
		if p.WindowSize == 0 && !p.Flags.RST {
			zeroWindowCount++
		}
	}
	if zeroWindowCount > 2 {
		anomalies = append(anomalies, "ZERO_WINDOW_PROBES")
	}

	return anomalies
}

// DetectVPN 检测 VPN 使用
func (t *TCPIPAnalyzer) DetectVPN() bool {
	if len(t.packets) == 0 {
		return false
	}

	for _, p := range t.packets {
		// VPN 常见特征: MSS 值较低（因为封装开销）
		mss := getMSSValue(p.Options)
		if mss > 0 && mss < 1400 && mss != 1460 {
			return true
		}

		// TTL 异常跳动
		ttl := p.IPHeader.TTL
		if ttl > 0 && ttl != 64 && ttl != 128 && ttl != 255 {
			initialTTL := nearestDefaultTTL(ttl)
			hops := initialTTL - int(ttl)
			// VPN 通常增加 1-2 跳
			if hops > 20 {
				return true
			}
		}
	}
	return false
}

// DetectProxy 检测代理使用
func (t *TCPIPAnalyzer) DetectProxy() bool {
	if len(t.packets) == 0 {
		return false
	}

	// 代理特征: 多个不同的 IP ID 序列模式
	// 或者窗口大小突变
	if len(t.packets) < 2 {
		return false
	}

	windowChanges := 0
	for i := 1; i < len(t.packets); i++ {
		prev := t.packets[i-1].WindowSize
		curr := t.packets[i].WindowSize
		if prev > 0 && curr > 0 {
			ratio := float64(curr) / float64(prev)
			if ratio > 2.0 || ratio < 0.5 {
				windowChanges++
			}
		}
	}

	// 频繁的窗口大小突变可能表示代理
	return windowChanges > len(t.packets)/2
}

// DetectNAT 检测 NAT 使用
func (t *TCPIPAnalyzer) DetectNAT() bool {
	if len(t.packets) < 2 {
		return false
	}

	// NAT 特征: IP ID 字段不连续或重叠
	// 多个主机共享同一 IP 时 ID 会有间隙
	ids := make([]uint16, 0, len(t.packets))
	for _, p := range t.packets {
		if p.IPHeader.ID > 0 {
			ids = append(ids, p.IPHeader.ID)
		}
	}

	if len(ids) < 2 {
		return false
	}

	// 检测 ID 间隙（大于 1000 的跳跃可能表示多主机 NAT）
	gapCount := 0
	for i := 1; i < len(ids); i++ {
		var gap uint16
		if ids[i] > ids[i-1] {
			gap = ids[i] - ids[i-1]
		} else {
			gap = ids[i-1] - ids[i]
		}
		if gap > 1000 {
			gapCount++
		}
	}

	return gapCount > len(ids)/3
}

// NewAnalyzer 创建 TCP/IP 分析器（模块统一命名）。
func NewAnalyzer() *TCPIPAnalyzer {
	return NewTCPIPAnalyzer()
}

// Helper 函数：从字节流创建数据包
// ParseTCPPacket(data []byte) (TCPPacket, error) { ... }
// ParseIPHeader(data []byte) (IPHeader, error) { ... }

// nearestDefaultTTL 推测初始 TTL
func nearestDefaultTTL(ttl uint8) int {
	switch {
	case ttl <= 32:
		return 32
	case ttl <= 64:
		return 64
	case ttl <= 128:
		return 128
	default:
		return 255
	}
}

// getMSSValue 从 TCP 选项中获取 MSS 值
func getMSSValue(options TCPOptions) int {
	if options.MSS != nil {
		return int(*options.MSS)
	}
	return 0
}

// getWindowScale 从 TCP 选项中获取窗口缩放值
func getWindowScale(options TCPOptions) int {
	if options.WindowScale != nil {
		return int(*options.WindowScale)
	}
	return 0
}

// formatTCPOptions 格式化 TCP 选项为指纹字符串
func formatTCPOptions(options TCPOptions) string {
	var parts []string

	if options.MSS != nil {
		parts = append(parts, "M")
	}
	if options.WindowScale != nil {
		parts = append(parts, "W")
	}
	if options.SAckPermitted {
		parts = append(parts, "S")
	}
	if options.Timestamps {
		parts = append(parts, "T")
	}
	for i := 0; i < options.NoOperation; i++ {
		parts = append(parts, "N")
	}
	if options.EndOfOptions {
		parts = append(parts, "E")
	}

	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ",")
}

// guessOSFromPacket 根据数据包特征推测操作系统
func guessOSFromPacket(packet TCPPacket) (os, version string, confidence float64) {
	ttl := packet.IPHeader.TTL
	ws := packet.WindowSize
	mss := getMSSValue(packet.Options)

	switch {
	case ttl > 96 && ttl <= 128:
		if ws == 65535 || ws%8192 == 0 {
			if mss == 1460 {
				return "Windows", "10/11", 0.75
			}
			return "Windows", "", 0.6
		}
		return "Windows", "", 0.5

	case ttl > 32 && ttl <= 64:
		if packet.Options.Timestamps && packet.Options.SAckPermitted {
			if ws == 65535 {
				if packet.Options.WindowScale != nil && *packet.Options.WindowScale == 6 {
					return "macOS", "", 0.65
				}
				return "Linux", "5.x/6.x", 0.7
			}
			return "Linux", "", 0.6
		}
		if ws == 65535 {
			return "macOS/Linux", "", 0.5
		}
		return "Linux", "", 0.5

	case ttl >= 254:
		return "Solaris/AIX", "", 0.6

	default:
		return "Unknown", "", 0.0
	}
}

// windowSizeFamily 根据窗口大小判断操作系统家族
func windowSizeFamily(ws uint16) string {
	switch {
	case ws == 65535:
		return "Linux/macOS/Windows"
	case ws%8192 == 0:
		return "Windows"
	case ws == 5840 || ws == 14600 || ws == 29200:
		return "Linux"
	case ws == 32768:
		return "FreeBSD"
	default:
		return "Unknown"
	}
}
