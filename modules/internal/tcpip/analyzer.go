// translated comment
package tcpip

import (
	"crypto/md5"
	"fmt"
	"strings"
)

// translated comment
type OSSignature struct {
	Name          string         // translated comment
	Family        string         // translated comment
	DefaultTTL    int            // translated comment
	WindowBase    int            // translated comment
	MSS           int            // translated comment
	TCPOptions    string         // translated comment
	IPDFBit       bool           // translated comment
	Quirks        string         // translated comment
	ProbeResponse map[int]string // translated comment
}

// translated comment
func BuildOSDatabase() map[string]OSSignature {
	db := make(map[string]OSSignature)

	// translated comment
	db["Windows_11"] = OSSignature{
		Name:       "Windows 11",
		Family:     "Windows",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,SACK,TS,NOP,WS",
		IPDFBit:    true,
		Quirks:     "Window scaling, Selective ACK",
	}

	db["Windows_10"] = OSSignature{
		Name:       "Windows 10",
		Family:     "Windows",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,SACK,TS,NOP,WS",
		IPDFBit:    true,
		Quirks:     "Window scaling, Selective ACK",
	}

	db["Windows_Server_2019"] = OSSignature{
		Name:       "Windows Server 2019",
		Family:     "Windows",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,SACK,TS,NOP,WS",
		IPDFBit:    true,
	}

	// translated comment
	db["Linux_Kernel_5.x"] = OSSignature{
		Name:       "Linux Kernel 5.x",
		Family:     "Linux",
		DefaultTTL: 64,
		WindowBase: 29200,
		MSS:        1460,
		TCPOptions: "MSS,TS,TS,SACK,WS",
		IPDFBit:    true,
		Quirks:     "SYN flood protection",
	}

	db["Linux_Kernel_4.x"] = OSSignature{
		Name:       "Linux Kernel 4.x",
		Family:     "Linux",
		DefaultTTL: 64,
		WindowBase: 29200,
		MSS:        1460,
		TCPOptions: "MSS,TS,SACK,WS",
		IPDFBit:    true,
	}

	db["Ubuntu_22.04"] = OSSignature{
		Name:       "Ubuntu 22.04 LTS",
		Family:     "Linux",
		DefaultTTL: 64,
		WindowBase: 29200,
		MSS:        1460,
		TCPOptions: "MSS,TS,SACK,WS",
		IPDFBit:    true,
	}

	// translated comment
	db["macOS_13"] = OSSignature{
		Name:       "macOS 13 (Ventura)",
		Family:     "macOS",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,NOP,WS,NOP,NOP,TS",
		IPDFBit:    true,
		Quirks:     "Special timestamp handling",
	}

	db["macOS_12"] = OSSignature{
		Name:       "macOS 12 (Monterey)",
		Family:     "macOS",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,NOP,WS,NOP,NOP,TS",
		IPDFBit:    true,
	}

	// iOS
	db["iOS_16"] = OSSignature{
		Name:       "iOS 16",
		Family:     "iOS",
		DefaultTTL: 64,
		WindowBase: 65535,
		MSS:        1460,
		TCPOptions: "MSS,NOP,WS,NOP,NOP,TS",
		IPDFBit:    true,
	}

	// Android
	db["Android_13"] = OSSignature{
		Name:       "Android 13",
		Family:     "Android",
		DefaultTTL: 64,
		WindowBase: 32768,
		MSS:        1460,
		TCPOptions: "MSS,SACK,TS,NOP,WS",
		IPDFBit:    true,
	}

	return db
}

// translated comment
func ComputeTCPSignature(mss int, windowSize int, options string, flags string) string {
	data := fmt.Sprintf("%d,%d,%s,%s", mss, windowSize, options, flags)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// translated comment
func ComputeIPSignature(ttl int, flags uint8, id uint16) string {
	data := fmt.Sprintf("%d,%d,%d", ttl, flags, id)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// translated comment
func MatchOSSignature(db map[string]OSSignature, ttl int, mss int, options string) string {
	bestMatch := ""
	bestScore := 0.0

	for osName, sig := range db {
		score := 0.0

		// translated comment
		if sig.DefaultTTL == ttl {
			score += 0.4
		} else if sig.DefaultTTL-ttl <= 10 && sig.DefaultTTL-ttl >= 0 {
			score += 0.2
		}

		// translated comment
		if sig.MSS == mss {
			score += 0.3
		} else if mss > sig.MSS-100 && mss < sig.MSS+100 {
			score += 0.15
		}

		// translated comment
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

// translated comment
//
// translated comment
// translated comment
// translated comment
func ExtractTCPOptions(packet []byte) string {
	// translated comment
	if len(packet) < 20 {
		return ""
	}

	// translated comment
	// translated comment
	dataOffset := (packet[12] >> 4) * 4

	// translated comment
	if dataOffset < 20 || int(dataOffset) > len(packet) {
		return ""
	}

	// translated comment
	if dataOffset == 20 {
		return "" // translated comment
	}

	optionsData := packet[20:dataOffset]
	options := make([]string, 0, len(optionsData)/2)

	i := 0
	for i < len(optionsData) {
		kind := optionsData[i]

		// Kind 0 (EOL) - End of Option List
		if kind == 0 {
			break
		}

		// translated comment
		if kind == 1 {
			options = append(options, "NOP")
			i++
			continue
		}

		// translated comment
		if i+1 >= len(optionsData) {
			break
		}

		length := int(optionsData[i+1])

		// translated comment
		if length < 2 || i+length > len(optionsData) {
			break
		}

		// translated comment
		switch kind {
		case 2:
			options = append(options, "MSS")
		case 3:
			options = append(options, "WS")
		case 4:
			options = append(options, "SACK_PERMITTED")
		case 5:
			options = append(options, "SACK")
		case 8:
			options = append(options, "TS")
		case 14:
			options = append(options, "TCP_ALTCHK")
		case 15:
			options = append(options, "TCP_ALTCHK_DATA")
		case 28:
			options = append(options, "UTO")
		case 29:
			options = append(options, "TCP_AO")
		case 30:
			options = append(options, "MP_CAPABLE")
		case 34:
			options = append(options, "TFO")
		default:
			// translated comment
			options = append(options, fmt.Sprintf("OPT_%d", kind))
		}

		i += length
	}

	if len(options) == 0 {
		return ""
	}

	return strings.Join(options, ",")
}

// translated comment
func AnalyzeTTL(currentTTL int) int {
	// translated comment
	// translated comment
	if currentTTL > 64 {
		return 128
	} else if currentTTL > 32 {
		return 64
	}
	return 32
}

// translated comment
func AnalyzeWindowSize(size int) string {
	if size > 60000 {
		return "Large (Windows/macOS style)"
	} else if size > 20000 {
		return "Medium (Linux style)"
	}
	return "Small"
}

// translated comment
func DetectSequenceNumberPattern(seqNumbers []uint32) string {
	if len(seqNumbers) < 2 {
		return "Insufficient data"
	}

	// translated comment
	diffs := make([]int64, len(seqNumbers)-1)
	for i := 0; i < len(diffs); i++ {
		diffs[i] = int64(seqNumbers[i+1]) - int64(seqNumbers[i])
	}

	// translated comment
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

// translated comment
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

	// translated comment
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

// translated comment
func DetectAnomalies(ttl int, mss int, windowSize int) []string {
	anomalies := make([]string, 0, 3)

	// translated comment
	if ttl != 64 && ttl != 128 && ttl != 32 {
		anomalies = append(anomalies, fmt.Sprintf("Non-standard TTL: %d", ttl))
	}

	// translated comment
	if mss < 536 {
		anomalies = append(anomalies, "MSS too small (potential DoS)")
	}

	// translated comment
	if windowSize < 1024 {
		anomalies = append(anomalies, "Unusually small window size")
	}

	return anomalies
}

// translated comment
func CalculateConfidence(matches int, total int) float64 {
	if total == 0 {
		return 0.0
	}
	return float64(matches) / float64(total)
}
