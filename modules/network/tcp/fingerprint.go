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
