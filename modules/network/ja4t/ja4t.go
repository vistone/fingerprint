package ja4t

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// JA4TResult JA4T fingerprint result (TCP client fingerprint)
type JA4TResult struct {
	// JA4T full fingerprint (raw string form)
	RawFingerprint string

	// JA4T hash (first 12 characters of SHA256)
	Hash string

	// Window size
	WindowSize uint16

	// TCP option type list (in original order)
	Options []uint8

	// Maximum Segment Size (MSS) value
	MSS uint16

	// Window scale factor
	WindowScale uint8

	// Anomaly flags
	AnomalyFlags []string

	// Risk score (0.0-1.0)
	RiskScore float64

	// Probable operating system
	ProbableOS string
}

// JA4TSResult JA4TS fingerprint result (TCP server fingerprint)
type JA4TSResult struct {
	// JA4TS full fingerprint (raw string form)
	RawFingerprint string

	// JA4TS hash (first 12 characters of SHA256)
	Hash string

	// Window size
	WindowSize uint16

	// TCP option type list (in original order)
	Options []uint8

	// Maximum Segment Size (MSS) value
	MSS uint16

	// Window scale factor
	WindowScale uint8
}

// TCPSYNData TCP SYN packet characteristics
type TCPSYNData struct {
	// Window size
	WindowSize uint16

	// TCP options (in original order, Option Kind values)
	// Common values: 2=MSS, 1=NOP, 3=Window Scale, 4=SACK Permitted, 8=Timestamps
	Options []uint8

	// Maximum Segment Size (MSS) value
	MSS uint16

	// Window scale factor
	WindowScale uint8

	// IP TTL (optional, used for OS detection)
	TTL uint8

	// IP DF flag (Don't Fragment)
	DF bool
}

// TCPOptionKind TCP option type constants
const (
	TCPOptionEndOfList   uint8 = 0 // End of option list
	TCPOptionNOP         uint8 = 1 // No operation (padding)
	TCPOptionMSS         uint8 = 2 // Maximum Segment Size
	TCPOptionWindowScale uint8 = 3 // Window scale
	TCPOptionSACKPermit  uint8 = 4 // Selective Acknowledgment Permitted
	TCPOptionSACK        uint8 = 5 // Selective Acknowledgment
	TCPOptionTimestamps  uint8 = 8 // Timestamps
)

// tcpOptionName TCP option type name
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

// ComputeJA4T computes the JA4T fingerprint from TCP SYN data
func ComputeJA4T(data TCPSYNData) *JA4TResult {
	result := &JA4TResult{
		WindowSize:   data.WindowSize,
		Options:      data.Options,
		MSS:          data.MSS,
		WindowScale:  data.WindowScale,
		AnomalyFlags: []string{},
	}

	// Build options string: option type values separated by "-"
	optionParts := make([]string, len(data.Options))
	for i, opt := range data.Options {
		optionParts[i] = fmt.Sprintf("%d", opt)
	}
	optionsStr := strings.Join(optionParts, "-")

	// JA4T format: {window_size}_{options}_{mss}_{window_scale}
	result.RawFingerprint = fmt.Sprintf("%d_%s_%d_%d",
		data.WindowSize,
		optionsStr,
		data.MSS,
		data.WindowScale,
	)

	// Compute SHA256 hash (first 12 characters)
	hash := sha256.Sum256([]byte(result.RawFingerprint))
	result.Hash = fmt.Sprintf("%x", hash)[:12]

	// Anomaly detection
	detectTCPAnomalies(data, result)

	// OS guessing
	result.ProbableOS = guessOS(data)

	return result
}

// ComputeJA4TS computes the JA4TS fingerprint from TCP SYN-ACK data (server side)
func ComputeJA4TS(data TCPSYNData) *JA4TSResult {
	result := &JA4TSResult{
		WindowSize:  data.WindowSize,
		Options:     data.Options,
		MSS:         data.MSS,
		WindowScale: data.WindowScale,
	}

	// Build options string
	optionParts := make([]string, len(data.Options))
	for i, opt := range data.Options {
		optionParts[i] = fmt.Sprintf("%d", opt)
	}
	optionsStr := strings.Join(optionParts, "-")

	// JA4TS format is the same as JA4T
	result.RawFingerprint = fmt.Sprintf("%d_%s_%d_%d",
		data.WindowSize,
		optionsStr,
		data.MSS,
		data.WindowScale,
	)

	// Compute hash
	hash := sha256.Sum256([]byte(result.RawFingerprint))
	result.Hash = fmt.Sprintf("%x", hash)[:12]

	return result
}

// MatchJA4T checks whether two JA4T hashes match
func MatchJA4T(hash1, hash2 string) bool {
	return len(hash1) == 12 && len(hash2) == 12 && hash1 == hash2
}

// detectTCPAnomalies detects TCP anomaly characteristics
func detectTCPAnomalies(data TCPSYNData, result *JA4TResult) {
	baseScore := 0.0

	// Anomaly 1: Window size is 0 or abnormally small
	if data.WindowSize == 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "ZERO_WINDOW")
		baseScore += 0.3
	} else if data.WindowSize < 512 {
		result.AnomalyFlags = append(result.AnomalyFlags, "SMALL_WINDOW")
		baseScore += 0.15
	}

	// Anomaly 2: Missing MSS option (most normal implementations include it)
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

	// Anomaly 3: Abnormal MSS value
	if data.MSS > 0 && data.MSS < 536 {
		result.AnomalyFlags = append(result.AnomalyFlags, "LOW_MSS")
		baseScore += 0.15
	}

	// Anomaly 4: No TCP options (very rare)
	if len(data.Options) == 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "NO_OPTIONS")
		baseScore += 0.25
	}

	// Anomaly 5: Excessive window scale factor (>14 is unreasonable)
	if data.WindowScale > 14 {
		result.AnomalyFlags = append(result.AnomalyFlags, "EXCESSIVE_WINDOW_SCALE")
		baseScore += 0.2
	}

	// Anomaly 6: Abnormal TTL value
	if data.TTL > 0 && data.TTL < 32 {
		result.AnomalyFlags = append(result.AnomalyFlags, "LOW_TTL")
		baseScore += 0.15
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}
	result.RiskScore = baseScore
}

// guessOS guesses the operating system based on TCP SYN characteristics
func guessOS(data TCPSYNData) string {
	// OS guessing based on common TCP/IP stack characteristics
	// Primary references: default TTL, window size, MSS, option order

	optionsStr := formatOptions(data.Options)

	// Windows characteristics: TTL=128, window size=65535 or multiple of 8192
	if data.TTL > 96 && data.TTL <= 128 {
		if data.WindowSize == 65535 || data.WindowSize%8192 == 0 {
			return "Windows"
		}
	}

	// macOS/iOS characteristics: TTL=64, window size=65535, options contain 1-1-8-4-2-3 or 2-1-1-4-8-3
	if data.TTL > 32 && data.TTL <= 64 {
		if data.WindowSize == 65535 && strings.Contains(optionsStr, "1-1-4-8") {
			return "macOS"
		}
	}

	// Linux characteristics: TTL=64, common option order 2-1-3-1-1-8-4 or 2-4-8-1-3
	if data.TTL > 32 && data.TTL <= 64 {
		if strings.HasPrefix(optionsStr, "2-1-3") || strings.HasPrefix(optionsStr, "2-4-8-1-3") {
			return "Linux"
		}
	}

	// FreeBSD characteristics: TTL=64, window size=65535
	if data.TTL > 32 && data.TTL <= 64 {
		if data.WindowSize == 65535 {
			return "FreeBSD"
		}
	}

	// Solaris/AIX: TTL=254 or 255
	if data.TTL >= 254 {
		return "Solaris/AIX"
	}

	return "Unknown"
}

// formatOptions formats the option list as a string
func formatOptions(options []uint8) string {
	parts := make([]string, len(options))
	for i, opt := range options {
		parts[i] = fmt.Sprintf("%d", opt)
	}
	return strings.Join(parts, "-")
}

// GetOptionNames returns the human-readable name list of options
func GetOptionNames(options []uint8) []string {
	names := make([]string, len(options))
	for i, opt := range options {
		names[i] = tcpOptionName(opt)
	}
	return names
}

// KnownOSProfiles returns TCP characteristics of known operating systems
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

// OSProfile operating system TCP characteristics
type OSProfile struct {
	// Operating system name
	Name string

	// Default TTL
	TTL uint8

	// Default window size
	WindowSize uint16

	// Default MSS
	MSS uint16

	// Default window scale
	WindowScale uint8

	// TCP option order
	Options []uint8
}

// ToSYNData converts to TCPSYNData
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
