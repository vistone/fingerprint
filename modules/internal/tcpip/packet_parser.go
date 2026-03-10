package tcpip

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// PacketParser is a packet parser
type PacketParser struct {
	rawData []byte
}

// NewPacketParser creates a new parser
func NewPacketParser(data []byte) *PacketParser {
	return &PacketParser{rawData: data}
}

// ParseIPHeader parses an IP header
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

// ParseTCPHeader parses a TCP header
func (p *PacketParser) ParseTCPHeader() error {
	if len(p.rawData) < 20 {
		return fmt.Errorf("TCP header too short")
	}

	return nil
}

// ExtractFlags extracts TCP flags
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

// FormatSignature formats a signature string
func FormatSignature(ttl int, mss int, ws int, opts string) string {
	return fmt.Sprintf("%d,%d,%d,%s", ttl, mss, ws, opts)
}

// ParseSignature parses a signature string
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

// CalculateChecksum calculates a checksum
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

// IsPrivateIP checks whether the IP is a private IP
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	return parsedIP.IsPrivate()
}

// IsReservedIP checks whether the IP is a reserved IP
func IsReservedIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	return parsedIP.IsLoopback() ||
		parsedIP.IsLinkLocalUnicast() ||
		parsedIP.IsMulticast()
}

// DetectNATUsage detects NAT usage (based on IP ID and sequence numbers)
func DetectNATUsage(ipIDs []uint16) bool {
	if len(ipIDs) < 3 {
		return false
	}

	// Calculate IP ID differences
	diffs := make([]int, len(ipIDs)-1)
	for i := 0; i < len(diffs); i++ {
		diffs[i] = int(ipIDs[i+1]) - int(ipIDs[i])
	}

	// If IP IDs grow linearly, NAT may be in use
	consecutive := 0
	for i := 0; i < len(diffs)-1; i++ {
		if diffs[i] == diffs[i+1] && diffs[i] > 0 {
			consecutive++
		}
	}

	// If there is an obvious linear pattern, it may be NAT
	return float64(consecutive) > float64(len(diffs))*0.5
}

// InferInitialTTL infers the initial TTL value
func InferInitialTTL(currentTTL int, hopCount int) int {
	ttl := currentTTL + hopCount

	// Use the closest standard TTL value
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

// GenerateSignatureHash generates a signature hash
func GenerateSignatureHash(parts ...interface{}) string {
	str := fmt.Sprintf("%v", parts)
	hash := 0
	for _, c := range str {
		hash = ((hash << 5) - hash) + int(c)
	}
	return fmt.Sprintf("%x", hash)
}

// ProfileMatch represents a profile match result
type ProfileMatch struct {
	OS         string
	Confidence float64
	Matches    int
	Total      int
}

// MatchProfile matches an operating system profile
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

// ProfileTemplate represents an operating system profile template
type ProfileTemplate struct {
	Name        string
	TTL         int
	MSS         int
	WindowScale int
	Options     string
	DF          bool
	EcnCapable  bool
}
