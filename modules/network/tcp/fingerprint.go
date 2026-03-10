package tcp

import (
	"crypto/md5"
	"fmt"
	"strings"
)

// translated comment
type TCPFlags struct {
	SYN bool // translated comment
	ACK bool // translated comment
	FIN bool // translated comment
	RST bool // translated comment
	PSH bool // translated comment
	URG bool // translated comment
}

// translated comment
type IPHeader struct {
	Version    uint8  // translated comment
	TTL        uint8  // translated comment
	TotalLen   uint16 // translated comment
	Flags      uint8  // translated comment
	FragOffset uint16 // translated comment
	ID         uint16 // translated comment
	Protocol   uint8  // translated comment
	Checksum   uint16 // translated comment
	Src        string // translated comment
	Dst        string // translated comment
}

// translated comment
type TCPOptions struct {
	MSS           *uint16 // translated comment
	WindowScale   *uint8  // translated comment
	SACK          bool    // translated comment
	Timestamps    bool    // translated comment
	SAckPermitted bool    // translated comment
	NoOperation   int     // translated comment
	EndOfOptions  bool    // translated comment
	WindowSize    uint16  // translated comment
	OptionsMD5    string  // translated comment
}

// translated comment
type TCPPacket struct {
	// translated comment
	IPHeader IPHeader

	// translated comment
	SrcPort    uint16   // translated comment
	DstPort    uint16   // translated comment
	SeqNum     uint32   // translated comment
	AckNum     uint32   // translated comment
	Flags      TCPFlags // translated comment
	WindowSize uint16   // translated comment
	Checksum   uint16   // translated comment
	UrgentPtr  uint16   // translated comment
	DataLen    uint16   // translated comment

	// translated comment
	Options TCPOptions

	// translated comment
	Timestamp   int64 // translated comment
	RoundTripMs int64 // translated comment
}

// translated comment
type TCPIPSignature struct {
	Hash              string            // translated comment
	RawSignature      string            // translated comment
	OS                string            // translated comment
	OSVersion         string            // translated comment
	Confidence        float64           // translated comment
	MatchedProfiles   []string          // translated comment
	TTLValue          int               // translated comment
	WindowSizeFamily  string            // translated comment
	MSS               int               // translated comment
	OptimizationLevel string            // translated comment
	Features          map[string]string // translated comment
}

// translated comment
type OSFingerprint struct {
	Name           string  // translated comment
	Family         string  // translated comment
	Version        string  // translated comment
	DefaultTTL     int     // translated comment
	WindowSizes    []int   // translated comment
	MSS            int     // translated comment
	TCPOptions     string  // translated comment
	IPFlags        uint8   // translated comment
	Quirks         string  // translated comment
	SYNACKSequence string  // translated comment
	Probability    float64 // translated comment
}

// translated comment
type TCPIPAnalyzer struct {
	packets    []TCPPacket              // translated comment
	signatures []TCPIPSignature         // translated comment
	osDatabase []OSFingerprint          // translated comment
	matchCache map[string]OSFingerprint // translated comment
}

// translated comment
type TCPIPResult struct {
	// translated comment
	OS            string   // translated comment
	OSFamily      string   // translated comment
	Confidence    float64  // translated comment
	CandidateOSes []string // translated comment

	// translated comment
	InitialTTL        int   // translated comment
	AverageWindowSize int   // translated comment
	MSS               int   // translated comment
	NetworkLatency    int64 // translated comment

	// translated comment
	SeqNumberBehavior string // translated comment
	AckBehavior       string // translated comment
	ResetBehavior     string // translated comment

	// translated comment
	RiskScore      float64  // translated comment
	AnomaliesFound []string // translated comment
	IsVPN          bool     // translated comment
	IsProxy        bool     // translated comment
	IsNAT          bool     // translated comment

	// translated comment
	Signature TCPIPSignature
}

// translated comment
func NewTCPIPAnalyzer() *TCPIPAnalyzer {
	return &TCPIPAnalyzer{
		packets:    []TCPPacket{},
		signatures: []TCPIPSignature{},
		osDatabase: []OSFingerprint{},
		matchCache: make(map[string]OSFingerprint),
	}
}

// translated comment
func (t *TCPIPAnalyzer) AddPacket(packet TCPPacket) {
	t.packets = append(t.packets, packet)
}

// translated comment
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

	// translated comment
	sig.OS, sig.OSVersion, sig.Confidence = guessOSFromPacket(packet)
	sig.WindowSizeFamily = windowSizeFamily(packet.WindowSize)

	return sig, nil
}

// translated comment
func (t *TCPIPAnalyzer) AnalyzeStream() (TCPIPResult, error) {
	result := TCPIPResult{
		CandidateOSes:  []string{},
		AnomaliesFound: []string{},
	}

	if len(t.packets) == 0 {
		return result, nil
	}

	// translated comment
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

	// translated comment
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

	// translated comment
	if len(t.packets) > 0 {
		result.AverageWindowSize = int(totalWindowSize / int64(len(t.packets)))
	}

	// translated comment
	if latencyCount > 0 {
		result.NetworkLatency = totalLatency / int64(latencyCount)
	}

	// translated comment
	if len(t.packets) > 0 {
		first := t.packets[0]
		result.InitialTTL = nearestDefaultTTL(first.IPHeader.TTL)
		result.MSS = getMSSValue(first.Options)
	}

	// translated comment
	result.AnomaliesFound = t.DetectAnomalies()
	result.RiskScore = t.GetRiskScore()
	result.IsVPN = t.DetectVPN()
	result.IsProxy = t.DetectProxy()
	result.IsNAT = t.DetectNAT()

	// translated comment
	if len(t.signatures) > 0 {
		result.Signature = t.signatures[0]
	}

	return result, nil
}

// translated comment
func (t *TCPIPAnalyzer) ComputeSignature(packet TCPPacket) (string, error) {
	// translated comment
	var b strings.Builder
	b.Grow(64)

	// translated comment
	fmt.Fprintf(&b, "%d:", nearestDefaultTTL(packet.IPHeader.TTL))

	// translated comment
	fmt.Fprintf(&b, "%d:", packet.WindowSize)

	// translated comment
	if packet.IPHeader.Flags&0x02 != 0 {
		b.WriteByte('1')
	} else {
		b.WriteByte('0')
	}
	b.WriteByte(':')

	// translated comment
	b.WriteString(formatTCPOptions(packet.Options))
	b.WriteByte(':')

	// MSS
	fmt.Fprintf(&b, "%d:", getMSSValue(packet.Options))

	// translated comment
	fmt.Fprintf(&b, "%d", getWindowScale(packet.Options))

	return b.String(), nil
}

// translated comment
func (t *TCPIPAnalyzer) GetOSFingerprints() []OSFingerprint {
	return t.osDatabase
}

// translated comment
func (t *TCPIPAnalyzer) SetOSDatabase(db []OSFingerprint) {
	t.osDatabase = db
}

// translated comment
func (t *TCPIPAnalyzer) GetRiskScore() float64 {
	if len(t.packets) == 0 {
		return 0.0
	}

	score := 0.0

	for _, packet := range t.packets {
		// translated comment
		if packet.IPHeader.TTL > 0 && packet.IPHeader.TTL < 32 {
			score += 0.15
		}

		// translated comment
		if packet.WindowSize == 0 {
			score += 0.2
		}

		// translated comment
		if packet.Flags.RST {
			score += 0.1
		}
	}

	// translated comment
	score = score / float64(len(t.packets))
	if score > 1.0 {
		score = 1.0
	}
	return score
}

// translated comment
func (t *TCPIPAnalyzer) DetectAnomalies() []string {
	var anomalies []string

	if len(t.packets) == 0 {
		return anomalies
	}

	// translated comment
	ttlSet := make(map[uint8]bool)
	for _, p := range t.packets {
		if p.IPHeader.TTL > 0 {
			ttlSet[p.IPHeader.TTL] = true
		}
	}
	if len(ttlSet) > 3 {
		anomalies = append(anomalies, "TTL_INCONSISTENCY")
	}

	// translated comment
	rstCount := 0
	for _, p := range t.packets {
		if p.Flags.RST {
			rstCount++
		}
	}
	if rstCount > len(t.packets)/3 {
		anomalies = append(anomalies, "EXCESSIVE_RST")
	}

	// translated comment
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

// translated comment
func (t *TCPIPAnalyzer) DetectVPN() bool {
	if len(t.packets) == 0 {
		return false
	}

	for _, p := range t.packets {
		// translated comment
		mss := getMSSValue(p.Options)
		if mss > 0 && mss < 1400 && mss != 1460 {
			return true
		}

		// translated comment
		ttl := p.IPHeader.TTL
		if ttl > 0 && ttl != 64 && ttl != 128 && ttl != 255 {
			initialTTL := nearestDefaultTTL(ttl)
			hops := initialTTL - int(ttl)
			// translated comment
			if hops > 20 {
				return true
			}
		}
	}
	return false
}

// translated comment
func (t *TCPIPAnalyzer) DetectProxy() bool {
	if len(t.packets) == 0 {
		return false
	}

	// translated comment
	// translated comment
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

	// translated comment
	return windowChanges > len(t.packets)/2
}

// translated comment
func (t *TCPIPAnalyzer) DetectNAT() bool {
	if len(t.packets) < 2 {
		return false
	}

	// translated comment
	// translated comment
	ids := make([]uint16, 0, len(t.packets))
	for _, p := range t.packets {
		if p.IPHeader.ID > 0 {
			ids = append(ids, p.IPHeader.ID)
		}
	}

	if len(ids) < 2 {
		return false
	}

	// translated comment
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

// translated comment
func NewAnalyzer() *TCPIPAnalyzer {
	return NewTCPIPAnalyzer()
}

// translated comment
// ParseTCPPacket(data []byte) (TCPPacket, error) { ... }
// ParseIPHeader(data []byte) (IPHeader, error) { ... }

// translated comment
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

// translated comment
func getMSSValue(options TCPOptions) int {
	if options.MSS != nil {
		return int(*options.MSS)
	}
	return 0
}

// translated comment
func getWindowScale(options TCPOptions) int {
	if options.WindowScale != nil {
		return int(*options.WindowScale)
	}
	return 0
}

// translated comment
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

// translated comment
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

// translated comment
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
