package tcp

import (
	"crypto/md5"
	"fmt"
	"strings"
)

// TCPFlags TCP flag bits
type TCPFlags struct {
	SYN bool // Synchronize flag (connection establishment)
	ACK bool // Acknowledgment flag
	FIN bool // Finish flag
	RST bool // Reset flag
	PSH bool // Push flag
	URG bool // Urgent flag
}

// IPHeader IP packet header parameters
type IPHeader struct {
	Version    uint8  // IP version (4 or 6)
	TTL        uint8  // Time to Live
	TotalLen   uint16 // Total length
	Flags      uint8  // Flags (DF, MF, RF)
	FragOffset uint16 // Fragment offset
	ID         uint16 // Identifier
	Protocol   uint8  // Protocol number (6=TCP, 17=UDP)
	Checksum   uint16 // Checksum
	Src        string // Source IP
	Dst        string // Destination IP
}

// TCPOptions TCP options
type TCPOptions struct {
	MSS           *uint16 // Maximum Segment Size
	WindowScale   *uint8  // Window scale factor
	SACK          bool    // Selective Acknowledgment
	Timestamps    bool    // Timestamps
	SAckPermitted bool    // SACK Permitted
	NoOperation   int     // NOP count
	EndOfOptions  bool    // End of option list
	WindowSize    uint16  // Window size
	OptionsMD5    string  // Options hash fingerprint
}

// TCPPacket TCP packet parameters
type TCPPacket struct {
	// IP layer
	IPHeader IPHeader

	// TCP layer
	SrcPort    uint16   // Source port
	DstPort    uint16   // Destination port
	SeqNum     uint32   // Sequence number
	AckNum     uint32   // Acknowledgment number
	Flags      TCPFlags // Flags
	WindowSize uint16   // Window size
	Checksum   uint16   // TCP checksum
	UrgentPtr  uint16   // Urgent pointer
	DataLen    uint16   // Data length

	// TCP options
	Options TCPOptions

	// Timing and statistics
	Timestamp   int64 // Packet timestamp
	RoundTripMs int64 // Round-trip time (milliseconds)
}

// TCPIPSignature TCP/IP fingerprint signature
type TCPIPSignature struct {
	Hash              string            // MD5 hash
	RawSignature      string            // Raw signature string
	OS                string            // Identified operating system
	OSVersion         string            // Operating system version
	Confidence        float64           // Confidence level (0.0-1.0)
	MatchedProfiles   []string          // Matched fingerprint profiles
	TTLValue          int               // TTL value
	WindowSizeFamily  string            // Window size family (e.g., "Linux", "Windows")
	MSS               int               // Maximum Segment Size
	OptimizationLevel string            // Optimization level
	Features          map[string]string // Additional features
}

// OSFingerprint operating system fingerprint
type OSFingerprint struct {
	Name           string  // Operating system name (e.g., "Windows 11")
	Family         string  // OS family (Windows, Linux, macOS, etc.)
	Version        string  // Version number
	DefaultTTL     int     // Default TTL value
	WindowSizes    []int   // Common window sizes (ascending order)
	MSS            int     // Typical MSS value
	TCPOptions     string  // TCP option characteristics
	IPFlags        uint8   // IP flag characteristics
	Quirks         string  // OS quirks and peculiarities
	SYNACKSequence string  // SYN-ACK sequence characteristics
	Probability    float64 // Match probability
}

// TCPIPAnalyzer TCP/IP fingerprint analyzer
type TCPIPAnalyzer struct {
	packets    []TCPPacket              // Packet list
	signatures []TCPIPSignature         // Signature list
	osDatabase []OSFingerprint          // Operating system database
	matchCache map[string]OSFingerprint // Match cache
}

// TCPIPResult TCP/IP analysis result
type TCPIPResult struct {
	// Operating system identification
	OS            string   // Identified operating system
	OSFamily      string   // Operating system family
	Confidence    float64  // Identification confidence
	CandidateOSes []string // Candidate operating system list

	// Network characteristics
	InitialTTL        int   // Initial TTL value
	AverageWindowSize int   // Average window size
	MSS               int   // Maximum Segment Size
	NetworkLatency    int64 // Network latency (milliseconds)

	// TCP behavior
	SeqNumberBehavior string // Sequence number generation behavior
	AckBehavior       string // ACK number behavior
	ResetBehavior     string // Reset behavior

	// Security indicators
	RiskScore      float64  // Risk score
	AnomaliesFound []string // Detected anomalies
	IsVPN          bool     // Whether VPN is used
	IsProxy        bool     // Whether proxy is used
	IsNAT          bool     // Whether NAT is used

	// Detailed signature
	Signature TCPIPSignature
}

// NewTCPIPAnalyzer creates a new analyzer
func NewTCPIPAnalyzer() *TCPIPAnalyzer {
	return &TCPIPAnalyzer{
		packets:    []TCPPacket{},
		signatures: []TCPIPSignature{},
		osDatabase: []OSFingerprint{},
		matchCache: make(map[string]OSFingerprint),
	}
}

// AddPacket adds a TCP/IP packet
func (t *TCPIPAnalyzer) AddPacket(packet TCPPacket) {
	t.packets = append(t.packets, packet)
}

// AnalyzePacket analyzes a single packet
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

	// OS guessing
	sig.OS, sig.OSVersion, sig.Confidence = guessOSFromPacket(packet)
	sig.WindowSizeFamily = windowSizeFamily(packet.WindowSize)

	return sig, nil
}

// AnalyzeStream analyzes a packet stream
func (t *TCPIPAnalyzer) AnalyzeStream() (TCPIPResult, error) {
	result := TCPIPResult{
		CandidateOSes:  []string{},
		AnomaliesFound: []string{},
	}

	if len(t.packets) == 0 {
		return result, nil
	}

	// Analyze each packet
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

	// Determine the most probable operating system
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

	// Average window size
	if len(t.packets) > 0 {
		result.AverageWindowSize = int(totalWindowSize / int64(len(t.packets)))
	}

	// Network latency
	if latencyCount > 0 {
		result.NetworkLatency = totalLatency / int64(latencyCount)
	}

	// Get first packet characteristics
	if len(t.packets) > 0 {
		first := t.packets[0]
		result.InitialTTL = nearestDefaultTTL(first.IPHeader.TTL)
		result.MSS = getMSSValue(first.Options)
	}

	// Anomaly detection
	result.AnomaliesFound = t.DetectAnomalies()
	result.RiskScore = t.GetRiskScore()
	result.IsVPN = t.DetectVPN()
	result.IsProxy = t.DetectProxy()
	result.IsNAT = t.DetectNAT()

	// Generate comprehensive signature
	if len(t.signatures) > 0 {
		result.Signature = t.signatures[0]
	}

	return result, nil
}

// ComputeSignature computes the TCP/IP signature
func (t *TCPIPAnalyzer) ComputeSignature(packet TCPPacket) (string, error) {
	// Signature format: TTL:WindowSize:DF:Options:MSS:WindowScale
	var b strings.Builder
	b.Grow(64)

	// TTL (estimate initial TTL)
	fmt.Fprintf(&b, "%d:", nearestDefaultTTL(packet.IPHeader.TTL))

	// Window size
	fmt.Fprintf(&b, "%d:", packet.WindowSize)

	// DF flag
	if packet.IPHeader.Flags&0x02 != 0 {
		b.WriteByte('1')
	} else {
		b.WriteByte('0')
	}
	b.WriteByte(':')

	// TCP options fingerprint
	b.WriteString(formatTCPOptions(packet.Options))
	b.WriteByte(':')

	// MSS
	fmt.Fprintf(&b, "%d:", getMSSValue(packet.Options))

	// Window scale
	fmt.Fprintf(&b, "%d", getWindowScale(packet.Options))

	return b.String(), nil
}

// GetOSFingerprints returns the operating system fingerprint database
func (t *TCPIPAnalyzer) GetOSFingerprints() []OSFingerprint {
	return t.osDatabase
}

// SetOSDatabase sets the operating system database
func (t *TCPIPAnalyzer) SetOSDatabase(db []OSFingerprint) {
	t.osDatabase = db
}

// GetRiskScore computes the risk score
func (t *TCPIPAnalyzer) GetRiskScore() float64 {
	if len(t.packets) == 0 {
		return 0.0
	}

	score := 0.0

	for _, packet := range t.packets {
		// TTL anomaly
		if packet.IPHeader.TTL > 0 && packet.IPHeader.TTL < 32 {
			score += 0.15
		}

		// Window size anomaly
		if packet.WindowSize == 0 {
			score += 0.2
		}

		// RST flood
		if packet.Flags.RST {
			score += 0.1
		}
	}

	// Normalize
	score = score / float64(len(t.packets))
	if score > 1.0 {
		score = 1.0
	}
	return score
}

// DetectAnomalies detects network anomalies
func (t *TCPIPAnalyzer) DetectAnomalies() []string {
	var anomalies []string

	if len(t.packets) == 0 {
		return anomalies
	}

	// Detect TTL inconsistency (may indicate route changes or man-in-the-middle)
	ttlSet := make(map[uint8]bool)
	for _, p := range t.packets {
		if p.IPHeader.TTL > 0 {
			ttlSet[p.IPHeader.TTL] = true
		}
	}
	if len(ttlSet) > 3 {
		anomalies = append(anomalies, "TTL_INCONSISTENCY")
	}

	// Detect abnormal RST count
	rstCount := 0
	for _, p := range t.packets {
		if p.Flags.RST {
			rstCount++
		}
	}
	if rstCount > len(t.packets)/3 {
		anomalies = append(anomalies, "EXCESSIVE_RST")
	}

	// Detect packets with zero window size (window probes or anomaly)
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

// DetectVPN detects VPN usage
func (t *TCPIPAnalyzer) DetectVPN() bool {
	if len(t.packets) == 0 {
		return false
	}

	for _, p := range t.packets {
		// Common VPN characteristic: lower MSS value (due to encapsulation overhead)
		mss := getMSSValue(p.Options)
		if mss > 0 && mss < 1400 && mss != 1460 {
			return true
		}

		// Abnormal TTL fluctuation
		ttl := p.IPHeader.TTL
		if ttl > 0 && ttl != 64 && ttl != 128 && ttl != 255 {
			initialTTL := nearestDefaultTTL(ttl)
			hops := initialTTL - int(ttl)
			// VPN typically adds 1-2 extra hops
			if hops > 20 {
				return true
			}
		}
	}
	return false
}

// DetectProxy detects proxy usage
func (t *TCPIPAnalyzer) DetectProxy() bool {
	if len(t.packets) == 0 {
		return false
	}

	// Proxy characteristics: multiple different IP ID sequence patterns
	// or sudden window size changes
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

	// Frequent sudden window size changes may indicate a proxy
	return windowChanges > len(t.packets)/2
}

// DetectNAT detects NAT usage
func (t *TCPIPAnalyzer) DetectNAT() bool {
	if len(t.packets) < 2 {
		return false
	}

	// NAT characteristics: non-sequential or overlapping IP ID field
	// When multiple hosts share the same IP, ID values will have gaps
	ids := make([]uint16, 0, len(t.packets))
	for _, p := range t.packets {
		if p.IPHeader.ID > 0 {
			ids = append(ids, p.IPHeader.ID)
		}
	}

	if len(ids) < 2 {
		return false
	}

	// Detect ID gaps (jumps greater than 1000 may indicate multi-host NAT)
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

// NewAnalyzer creates a TCP/IP analyzer (unified module naming).
func NewAnalyzer() *TCPIPAnalyzer {
	return NewTCPIPAnalyzer()
}

// Helper functions: create packets from byte stream
// ParseTCPPacket(data []byte) (TCPPacket, error) { ... }
// ParseIPHeader(data []byte) (IPHeader, error) { ... }

// nearestDefaultTTL estimates the initial TTL
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

// getMSSValue retrieves the MSS value from TCP options
func getMSSValue(options TCPOptions) int {
	if options.MSS != nil {
		return int(*options.MSS)
	}
	return 0
}

// getWindowScale retrieves the window scale value from TCP options
func getWindowScale(options TCPOptions) int {
	if options.WindowScale != nil {
		return int(*options.WindowScale)
	}
	return 0
}

// formatTCPOptions formats TCP options into a fingerprint string
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

// guessOSFromPacket guesses the operating system based on packet characteristics
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

// windowSizeFamily determines the OS family based on window size
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
