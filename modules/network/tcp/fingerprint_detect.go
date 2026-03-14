package tcp

import "strings"

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
