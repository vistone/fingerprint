package tcpip

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// PacketParser 数据包解析器
type PacketParser struct {
	rawData []byte
}

// NewPacketParser 创建新的解析器
func NewPacketParser(data []byte) *PacketParser {
	return &PacketParser{rawData: data}
}

// ParseIPHeader 解析 IP 头
func (p *PacketParser) ParseIPHeader() error {
	if len(p.rawData) < 20 {
		return fmt.Errorf("IP header too short")
	}

	version := p.rawData[0] >> 4
	if version != 4 && version != 6 {
		return fmt.Errorf("unsupported IP version: %d", version)
	}

	return nil
}

// ParseTCPHeader 解析 TCP 头
func (p *PacketParser) ParseTCPHeader() error {
	if len(p.rawData) < 20 {
		return fmt.Errorf("TCP header too short")
	}

	return nil
}

// ExtractFlags 提取 TCP 标志
func (p *PacketParser) ExtractFlags(flagByte byte) []string {
	var flags []string

	if flagByte&0x01 != 0 {
		flags = append(flags, "FIN")
	}
	if flagByte&0x02 != 0 {
		flags = append(flags, "SYN")
	}
	if flagByte&0x04 != 0 {
		flags = append(flags, "RST")
	}
	if flagByte&0x08 != 0 {
		flags = append(flags, "PSH")
	}
	if flagByte&0x10 != 0 {
		flags = append(flags, "ACK")
	}
	if flagByte&0x20 != 0 {
		flags = append(flags, "URG")
	}

	return flags
}

// FormatSignature 格式化签名字符串
func FormatSignature(ttl int, mss int, ws int, opts string) string {
	return fmt.Sprintf("%d,%d,%d,%s", ttl, mss, ws, opts)
}

// ParseSignature 解析签名字符串
func ParseSignature(sigStr string) (int, int, int, string, error) {
	parts := strings.Split(sigStr, ",")
	if len(parts) < 4 {
		return 0, 0, 0, "", fmt.Errorf("invalid signature format")
	}

	ttl, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, "", err
	}

	mss, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, "", err
	}

	ws, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, "", err
	}

	opts := parts[3]

	return ttl, mss, ws, opts, nil
}

// CalculateChecksum 计算校验和
func CalculateChecksum(data []byte) uint16 {
	sum := uint32(0)

	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}

	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}

	for sum>>16 > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}

	return ^uint16(sum)
}

// IsPrivateIP 检查是否为私有 IP
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	return parsedIP.IsPrivate()
}

// IsReservedIP 检查是否为保留 IP
func IsReservedIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	return parsedIP.IsLoopback() ||
		parsedIP.IsLinkLocalUnicast() ||
		parsedIP.IsMulticast()
}

// DetectNATUsage 检测 NAT 使用（基于 IP ID 和序列号）
func DetectNATUsage(ipIDs []uint16) bool {
	if len(ipIDs) < 3 {
		return false
	}

	// 计算 IP ID 的差值
	diffs := make([]int, len(ipIDs)-1)
	for i := 0; i < len(diffs); i++ {
		diffs[i] = int(ipIDs[i+1]) - int(ipIDs[i])
	}

	// 如果 IP ID 线性增长，可能使用了 NAT
	consecutive := 0
	for i := 0; i < len(diffs)-1; i++ {
		if diffs[i] == diffs[i+1] && diffs[i] > 0 {
			consecutive++
		}
	}

	// 如果有明显的线性模式，可能是 NAT
	return float64(consecutive) > float64(len(diffs))*0.5
}

// InferInitialTTL 推断初始 TTL 值
func InferInitialTTL(currentTTL int, hopCount int) int {
	ttl := currentTTL + hopCount

	// 使用最接近的标准 TTL 值
	standards := []int{32, 64, 128, 255}
	minDiff := 255
	bestMatch := 64

	for _, std := range standards {
		diff := std - ttl
		if diff < 0 {
			diff = -diff
		}
		if diff < minDiff {
			minDiff = diff
			bestMatch = std
		}
	}

	return bestMatch
}

// GenerateSignatureHash 生成签名哈希
func GenerateSignatureHash(parts ...interface{}) string {
	str := fmt.Sprintf("%v", parts)
	hash := 0
	for _, c := range str {
		hash = ((hash << 5) - hash) + int(c)
	}
	return fmt.Sprintf("%x", hash)
}

// ProfileMatch 轮廓匹配结果
type ProfileMatch struct {
	OS         string
	Confidence float64
	Matches    int
	Total      int
}

// MatchProfile 匹配操作系统轮廓
func MatchProfile(ttl int, mss int, ws int, opts string, profiles map[string]ProfileTemplate) ProfileMatch {
	maxScore := 0.0
	bestOS := ""
	totalMatches := 0

	for osName, profile := range profiles {
		matches := 0
		if profile.TTL == ttl {
			matches++
		}
		if profile.MSS == mss {
			matches++
		}
		if profile.WindowScale > 0 && (ws&profile.WindowScale) > 0 {
			matches++
		}
		if strings.Contains(opts, profile.Options) {
			matches++
		}

		score := float64(matches) / 4.0
		if score > maxScore {
			maxScore = score
			bestOS = osName
			totalMatches = matches
		}
	}

	return ProfileMatch{
		OS:         bestOS,
		Confidence: maxScore,
		Matches:    totalMatches,
		Total:      4,
	}
}

// ProfileTemplate 操作系统轮廓模板
type ProfileTemplate struct {
	Name        string
	TTL         int
	MSS         int
	WindowScale int
	Options     string
	DF          bool
	EcnCapable  bool
}
