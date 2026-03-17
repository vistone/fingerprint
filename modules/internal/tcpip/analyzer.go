// Package tcpip provides utility functions for TCP/IP fingerprint identification
package tcpip

import (
	"crypto/md5"
	"fmt"
	"strings"
)

var defaultOSDatabase = map[string]OSSignature{
	"Windows_11": {
		Name:       "Windows 11",
		Family:     "Windows",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,SACK,TS,NOP,WS",
		IPDFBit:    true,
		Quirks:     "Window scaling, Selective ACK",
	},
	"Windows_10": {
		Name:       "Windows 10",
		Family:     "Windows",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,SACK,TS,NOP,WS",
		IPDFBit:    true,
		Quirks:     "Window scaling, Selective ACK",
	},
	"Windows_Server_2019": {
		Name:       "Windows Server 2019",
		Family:     "Windows",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,SACK,TS,NOP,WS",
		IPDFBit:    true,
	},
	"Linux_Kernel_5.x": {
		Name:       "Linux Kernel 5.x",
		Family:     "Linux",
		DefaultTTL: 64,
		WindowBase: 29200,
		MSS:        1460,
		TCPOptions: "MSS,TS,TS,SACK,WS",
		IPDFBit:    true,
		Quirks:     "SYN flood protection",
	},
	"Linux_Kernel_4.x": {
		Name:       "Linux Kernel 4.x",
		Family:     "Linux",
		DefaultTTL: 64,
		WindowBase: 29200,
		MSS:        1460,
		TCPOptions: "MSS,TS,SACK,WS",
		IPDFBit:    true,
	},
	"Ubuntu_22.04": {
		Name:       "Ubuntu 22.04 LTS",
		Family:     "Linux",
		DefaultTTL: 64,
		WindowBase: 29200,
		MSS:        1460,
		TCPOptions: "MSS,TS,SACK,WS",
		IPDFBit:    true,
	},
	"macOS_13": {
		Name:       "macOS 13 (Ventura)",
		Family:     "macOS",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,NOP,WS,NOP,NOP,TS",
		IPDFBit:    true,
		Quirks:     "Special timestamp handling",
	},
	"macOS_12": {
		Name:       "macOS 12 (Monterey)",
		Family:     "macOS",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,NOP,WS,NOP,NOP,TS",
		IPDFBit:    true,
	},
	"iOS_16": {
		Name:       "iOS 16",
		Family:     "iOS",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,NOP,WS,NOP,NOP,TS",
		IPDFBit:    true,
	},
	"Android_13": {
		Name:       "Android 13",
		Family:     "Android",
		DefaultTTL: 64,
		WindowBase: 32768,
		MSS:        1460,
		TCPOptions: "MSS,SACK,TS,NOP,WS",
		IPDFBit:    true,
	},
}

// OSSignature represents an operating system signature definition
type OSSignature struct {
	Name          string         // OS name
	Family        string         // OS family
	DefaultTTL    int            // default TTL
	WindowBase    int            // base window size
	MSS           int            // Maximum Segment Size
	TCPOptions    string         // TCP options
	IPDFBit       bool           // IP DF bit
	Quirks        string         // quirks
	ProbeResponse map[int]string // responses to different probes
}

// BuildOSDatabase builds the operating system database
func BuildOSDatabase() map[string]OSSignature {
	return cloneOSDatabase(defaultOSDatabase)
}

func cloneOSDatabase(src map[string]OSSignature) map[string]OSSignature {
	db := make(map[string]OSSignature, len(src))
	for name, signature := range src {
		db[name] = signature
	}
	return db
}

// ComputeTCPSignature computes a TCP signature
func ComputeTCPSignature(mss int, windowSize int, options string, flags string) string {
	data := fmt.Sprintf("%d,%d,%s,%s", mss, windowSize, options, flags)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// ComputeIPSignature computes an IP signature
func ComputeIPSignature(ttl int, flags uint8, id uint16) string {
	data := fmt.Sprintf("%d,%d,%d", ttl, flags, id)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// MatchOSSignature matches an operating system signature
func MatchOSSignature(db map[string]OSSignature, ttl int, mss int, options string) string {
	bestMatch := ""
	bestScore := 0.0

	for osName, sig := range db {
		score := 0.0

		// TTL match (weight 40%)
		if sig.DefaultTTL == ttl {
			score += 0.4
		} else if sig.DefaultTTL-ttl <= 10 && sig.DefaultTTL-ttl >= 0 {
			score += 0.2
		}

		// MSS match (weight 30%)
		if sig.MSS == mss {
			score += 0.3
		} else if mss > sig.MSS-100 && mss < sig.MSS+100 {
			score += 0.15
		}

		// TCP options match (weight 30%)
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

// ExtractTCPOptions extracts a TCP options string
//
// Full implementation of TCP options parsing, returns a comma-separated string of option names
// Reference: RFC 793, RFC 1323, RFC 2018
// TCP option format: Kind(1B) | Length(1B) | Data(variable)
func ExtractTCPOptions(packet []byte) string {
	optionsData, ok := extractTCPOptionsData(packet)
	if !ok || len(optionsData) == 0 {
		return ""
	}

	options := make([]string, 0, len(optionsData)/2)
	parseTCPOptions(optionsData, &options)
	if len(options) == 0 {
		return ""
	}
	return strings.Join(options, ",")
}

func extractTCPOptionsData(packet []byte) ([]byte, bool) {
	// Minimum TCP header length is 20 bytes
	if len(packet) < 20 {
		return nil, false
	}
	// Data Offset represents the TCP header length in 32-bit words.
	dataOffset := (packet[12] >> 4) * 4
	if dataOffset < 20 || int(dataOffset) > len(packet) {
		return nil, false
	}
	// Data offset 20 means no TCP options are present.
	if dataOffset == 20 {
		return nil, true
	}
	return packet[20:dataOffset], true
}

func parseTCPOptions(optionsData []byte, options *[]string) {
	i := 0
	for i < len(optionsData) {
		kind := optionsData[i]

		if kind == 0 {
			break
		}

		if kind == 1 {
			*options = append(*options, "NOP")
			i++
			continue
		}

		if i+1 >= len(optionsData) {
			break
		}

		length := int(optionsData[i+1])
		if length < 2 || i+length > len(optionsData) {
			break
		}

		*options = append(*options, tcpOptionName(kind))
		i += length
	}
}

func tcpOptionName(kind byte) string {
	switch kind {
	case 2:
		return "MSS"
	case 3:
		return "WS"
	case 4:
		return "SACK_PERMITTED"
	case 5:
		return "SACK"
	case 8:
		return "TS"
	case 14:
		return "TCP_ALTCHK"
	case 15:
		return "TCP_ALTCHK_DATA"
	case 28:
		return "UTO"
	case 29:
		return "TCP_AO"
	case 30:
		return "MP_CAPABLE"
	case 34:
		return "TFO"
	default:
		return fmt.Sprintf("OPT_%d", kind)
	}
}

// AnalyzeTTL analyzes the TTL value to infer the initial TTL
func AnalyzeTTL(currentTTL int) int {
	// Common initial TTLs: 64/128 (Linux/Windows)
	// Infer the initial value based on the current TTL
	if currentTTL > 64 {
		return 128
	} else if currentTTL > 32 {
		return 64
	}
	return 32
}

// AnalyzeWindowSize analyzes the window size
func AnalyzeWindowSize(size int) string {
	if size > 60000 {
		return "Large (Windows/macOS style)"
	} else if size > 20000 {
		return "Medium (Linux style)"
	}
	return "Small"
}

// DetectSequenceNumberPattern detects sequence number patterns
func DetectSequenceNumberPattern(seqNumbers []uint32) string {
	if len(seqNumbers) < 2 {
		return "Insufficient data"
	}

	// Calculate differences
	diffs := make([]int64, len(seqNumbers)-1)
	for i := 0; i < len(diffs); i++ {
		diffs[i] = int64(seqNumbers[i+1]) - int64(seqNumbers[i])
	}

	// Check if random or sequential
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

// AnalyzeNetworkBehavior analyzes network behavior
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

	// Classify
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

// DetectAnomalies detects network anomalies
func DetectAnomalies(ttl int, mss int, windowSize int) []string {
	anomalies := make([]string, 0, 3)

	// TTL is not set to a standard value
	if ttl != 64 && ttl != 128 && ttl != 32 {
		anomalies = append(anomalies, fmt.Sprintf("Non-standard TTL: %d", ttl))
	}

	// MSS anomaly
	if mss < 536 {
		anomalies = append(anomalies, "MSS too small (potential DoS)")
	}

	// Window size anomaly
	if windowSize < 1024 {
		anomalies = append(anomalies, "Unusually small window size")
	}

	return anomalies
}

// CalculateConfidence calculates matching confidence
func CalculateConfidence(matches int, total int) float64 {
	if total == 0 {
		return 0.0
	}
	return float64(matches) / float64(total)
}
