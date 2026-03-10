package ja4t

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// translated comment
type JA4TResult struct {
	// translated comment
	RawFingerprint string

	// translated comment
	Hash string

	// translated comment
	WindowSize uint16

	// translated comment
	Options []uint8

	// translated comment
	MSS uint16

	// translated comment
	WindowScale uint8

	// translated comment
	AnomalyFlags []string

	// translated comment
	RiskScore float64

	// translated comment
	ProbableOS string
}

// translated comment
type JA4TSResult struct {
	// translated comment
	RawFingerprint string

	// translated comment
	Hash string

	// translated comment
	WindowSize uint16

	// translated comment
	Options []uint8

	// translated comment
	MSS uint16

	// translated comment
	WindowScale uint8
}

// translated comment
type TCPSYNData struct {
	// translated comment
	WindowSize uint16

	// translated comment
	// translated comment
	Options []uint8

	// translated comment
	MSS uint16

	// translated comment
	WindowScale uint8

	// translated comment
	TTL uint8

	// translated comment
	DF bool
}

// translated comment
const (
	TCPOptionEndOfList  uint8 = 0  // translated comment
	TCPOptionNOP        uint8 = 1  // translated comment
	TCPOptionMSS        uint8 = 2  // translated comment
	TCPOptionWindowScale uint8 = 3  // translated comment
	TCPOptionSACKPermit uint8 = 4  // translated comment
	TCPOptionSACK       uint8 = 5  // translated comment
	TCPOptionTimestamps uint8 = 8  // translated comment
)

// translated comment
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

// translated comment
func ComputeJA4T(data TCPSYNData) *JA4TResult {
	result := &JA4TResult{
		WindowSize:   data.WindowSize,
		Options:      data.Options,
		MSS:          data.MSS,
		WindowScale:  data.WindowScale,
		AnomalyFlags: []string{},
	}

	// translated comment
	optionParts := make([]string, len(data.Options))
	for i, opt := range data.Options {
		optionParts[i] = fmt.Sprintf("%d", opt)
	}
	optionsStr := strings.Join(optionParts, "-")

	// translated comment
	result.RawFingerprint = fmt.Sprintf("%d_%s_%d_%d",
		data.WindowSize,
		optionsStr,
		data.MSS,
		data.WindowScale,
	)

	// translated comment
	hash := sha256.Sum256([]byte(result.RawFingerprint))
	result.Hash = fmt.Sprintf("%x", hash)[:12]

	// translated comment
	detectTCPAnomalies(data, result)

	// translated comment
	result.ProbableOS = guessOS(data)

	return result
}

// translated comment
func ComputeJA4TS(data TCPSYNData) *JA4TSResult {
	result := &JA4TSResult{
		WindowSize:  data.WindowSize,
		Options:     data.Options,
		MSS:         data.MSS,
		WindowScale: data.WindowScale,
	}

	// translated comment
	optionParts := make([]string, len(data.Options))
	for i, opt := range data.Options {
		optionParts[i] = fmt.Sprintf("%d", opt)
	}
	optionsStr := strings.Join(optionParts, "-")

	// translated comment
	result.RawFingerprint = fmt.Sprintf("%d_%s_%d_%d",
		data.WindowSize,
		optionsStr,
		data.MSS,
		data.WindowScale,
	)

	// translated comment
	hash := sha256.Sum256([]byte(result.RawFingerprint))
	result.Hash = fmt.Sprintf("%x", hash)[:12]

	return result
}

// translated comment
func MatchJA4T(hash1, hash2 string) bool {
	return len(hash1) == 12 && len(hash2) == 12 && hash1 == hash2
}

// translated comment
func detectTCPAnomalies(data TCPSYNData, result *JA4TResult) {
	baseScore := 0.0

	// translated comment
	if data.WindowSize == 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "ZERO_WINDOW")
		baseScore += 0.3
	} else if data.WindowSize < 512 {
		result.AnomalyFlags = append(result.AnomalyFlags, "SMALL_WINDOW")
		baseScore += 0.15
	}

	// translated comment
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

	// translated comment
	if data.MSS > 0 && data.MSS < 536 {
		result.AnomalyFlags = append(result.AnomalyFlags, "LOW_MSS")
		baseScore += 0.15
	}

	// translated comment
	if len(data.Options) == 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "NO_OPTIONS")
		baseScore += 0.25
	}

	// translated comment
	if data.WindowScale > 14 {
		result.AnomalyFlags = append(result.AnomalyFlags, "EXCESSIVE_WINDOW_SCALE")
		baseScore += 0.2
	}

	// translated comment
	if data.TTL > 0 && data.TTL < 32 {
		result.AnomalyFlags = append(result.AnomalyFlags, "LOW_TTL")
		baseScore += 0.15
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}
	result.RiskScore = baseScore
}

// translated comment
func guessOS(data TCPSYNData) string {
	// translated comment
	// translated comment

	optionsStr := formatOptions(data.Options)

	// translated comment
	if data.TTL > 96 && data.TTL <= 128 {
		if data.WindowSize == 65535 || data.WindowSize%8192 == 0 {
			return "Windows"
		}
	}

	// translated comment
	if data.TTL > 32 && data.TTL <= 64 {
		if data.WindowSize == 65535 && strings.Contains(optionsStr, "1-1-4-8") {
			return "macOS"
		}
	}

	// translated comment
	if data.TTL > 32 && data.TTL <= 64 {
		if strings.HasPrefix(optionsStr, "2-1-3") || strings.HasPrefix(optionsStr, "2-4-8-1-3") {
			return "Linux"
		}
	}

	// translated comment
	if data.TTL > 32 && data.TTL <= 64 {
		if data.WindowSize == 65535 {
			return "FreeBSD"
		}
	}

	// translated comment
	if data.TTL >= 254 {
		return "Solaris/AIX"
	}

	return "Unknown"
}

// translated comment
func formatOptions(options []uint8) string {
	parts := make([]string, len(options))
	for i, opt := range options {
		parts[i] = fmt.Sprintf("%d", opt)
	}
	return strings.Join(parts, "-")
}

// translated comment
func GetOptionNames(options []uint8) []string {
	names := make([]string, len(options))
	for i, opt := range options {
		names[i] = tcpOptionName(opt)
	}
	return names
}

// translated comment
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

// translated comment
type OSProfile struct {
	// translated comment
	Name string

	// translated comment
	TTL uint8

	// translated comment
	WindowSize uint16

	// translated comment
	MSS uint16

	// translated comment
	WindowScale uint8

	// translated comment
	Options []uint8
}

// translated comment
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
