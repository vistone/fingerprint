package agent

import (
	"sort"

	"github.com/vistone/fingerprint/modules/core"
)

func (kb *KnowledgeBase) loadBuiltinKnowledge() {
	// Chrome knowledge
	kb.browsers[core.BrowserChrome] = &BrowserKnowledge{
		Family:             core.BrowserChrome,
		CommonCipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
		CommonExtensions:   []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0016, 0x000d, 0x002b, 0x002d, 0x0033},
		CommonCurves:       []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		MarketShare:        0.65,
		Versions: []VersionKnowledge{
			{Version: "115", VersionMajor: 115, ReleasedYear: 2023, SupportedOS: desktopOSList(),
				TLSVersion: 0x0303, CipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8, 0xc013, 0xc014, 0x002f, 0x0035, 0x000a},
				Extensions:          []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0016, 0x000d, 0x002b, 0x002d, 0x0033, 0x001c},
				Curves:              []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536,
				PseudoHeaderOrder: []string{":method", ":authority", ":scheme", ":path"}, ConnectionFlow: 15663105,
				SecCHUAPattern: `"Google Chrome";v="%d"`, AcceptPattern: "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp"},
			{Version: "119", VersionMajor: 119, ReleasedYear: 2023, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536,
				ConnectionFlow: 15663105, SecCHUAPattern: `"Google Chrome";v="%d"`, AcceptPattern: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp"},
			{Version: "121", VersionMajor: 121, ReleasedYear: 2024, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536,
				ConnectionFlow: 15663105, Curves: []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384, core.CurveP256Kyber}},
			{Version: "131", VersionMajor: 131, ReleasedYear: 2024, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536, ConnectionFlow: 15663105},
			{Version: "133", VersionMajor: 133, ReleasedYear: 2025, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536, ConnectionFlow: 15663105},
			{Version: "134", VersionMajor: 134, ReleasedYear: 2025, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536, ConnectionFlow: 15663105},
		},
	}

	// Firefox knowledge
	kb.browsers[core.BrowserFirefox] = &BrowserKnowledge{
		Family:             core.BrowserFirefox,
		CommonCipherSuites: []uint16{0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030},
		CommonExtensions:   []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
		CommonCurves:       []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		MarketShare:        0.03,
		Versions: []VersionKnowledge{
			{Version: "115", VersionMajor: 115, ReleasedYear: 2023, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc00a, 0xc009, 0xc013, 0xc014, 0x0033, 0x0039, 0x002f, 0x0035},
				H2InitialWindowSize: 131072, H2MaxConcurrentStreams: 100, H2HeaderTableSize: 65536, ConnectionFlow: 12517377,
				PseudoHeaderOrder: []string{":method", ":path", ":authority", ":scheme"}},
			{Version: "120", VersionMajor: 120, ReleasedYear: 2023, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030, 0xc00a, 0xc009, 0xc013, 0xc014},
				H2InitialWindowSize: 131072, H2MaxConcurrentStreams: 100, H2HeaderTableSize: 65536, ConnectionFlow: 12517377},
			{Version: "124", VersionMajor: 124, ReleasedYear: 2024, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030},
				H2InitialWindowSize: 131072, H2MaxConcurrentStreams: 100, H2HeaderTableSize: 65536, ConnectionFlow: 12517377},
			{Version: "135", VersionMajor: 135, ReleasedYear: 2025, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1303, 0x1302, 0xc02b, 0xc02f, 0xcca9, 0xcca8, 0xc02c, 0xc030},
				H2InitialWindowSize: 131072, H2MaxConcurrentStreams: 100, H2HeaderTableSize: 65536, ConnectionFlow: 12517377},
		},
	}

	// Safari knowledge
	kb.browsers[core.BrowserSafari] = &BrowserKnowledge{
		Family:             core.BrowserSafari,
		CommonCipherSuites: []uint16{0xc02c, 0xc02b, 0xc030, 0xc02f, 0xcca9, 0xcca8, 0xc00a, 0xc009, 0xc014, 0xc013},
		CommonExtensions:   []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x000d, 0x002b, 0x002d, 0x0033},
		CommonCurves:       []core.CurveID{core.CurveP256, core.CurveP384, core.CurveP521, core.CurveX25519},
		MarketShare:        0.19,
		Versions: []VersionKnowledge{
			{Version: "16", VersionMajor: 16, ReleasedYear: 2022, SupportedOS: appleOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0xc02c, 0xc02b, 0xc030, 0xc02f, 0xcca9, 0xcca8, 0xc00a, 0xc009, 0xc014, 0xc013, 0x002f, 0x0035, 0x000a},
				H2InitialWindowSize: 4194304, H2MaxConcurrentStreams: 100, H2HeaderTableSize: 4096, ConnectionFlow: 10485760,
				PseudoHeaderOrder: []string{":method", ":scheme", ":path", ":authority"}},
			{Version: "17", VersionMajor: 17, ReleasedYear: 2023, SupportedOS: appleOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0xc02c, 0xc02b, 0xc030, 0xc02f, 0xcca9, 0xcca8, 0xc00a, 0xc009, 0xc014, 0xc013},
				H2InitialWindowSize: 4194304, H2MaxConcurrentStreams: 100, H2HeaderTableSize: 4096, ConnectionFlow: 10485760},
			{Version: "18", VersionMajor: 18, ReleasedYear: 2024, SupportedOS: appleOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0xc02c, 0xc02b, 0xc030, 0xc02f, 0xcca9, 0xcca8, 0xc00a, 0xc009},
				H2InitialWindowSize: 4194304, H2MaxConcurrentStreams: 100, H2HeaderTableSize: 4096, ConnectionFlow: 10485760},
		},
	}

	// Edge knowledge (Chromium kernel)
	kb.browsers[core.BrowserEdge] = &BrowserKnowledge{
		Family:             core.BrowserEdge,
		CommonCipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
		CommonExtensions:   []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0016, 0x000d, 0x002b, 0x002d, 0x0033},
		CommonCurves:       []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		MarketShare:        0.05,
		Versions: []VersionKnowledge{
			{Version: "118", VersionMajor: 118, ReleasedYear: 2023, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536, ConnectionFlow: 15663105},
			{Version: "131", VersionMajor: 131, ReleasedYear: 2024, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536, ConnectionFlow: 15663105},
		},
	}

	// Opera knowledge (Chromium kernel)
	kb.browsers[core.BrowserOpera] = &BrowserKnowledge{
		Family:             core.BrowserOpera,
		CommonCipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
		CommonExtensions:   []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0016, 0x000d, 0x002b, 0x002d, 0x0033},
		CommonCurves:       []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		MarketShare:        0.03,
		Versions: []VersionKnowledge{
			{Version: "104", VersionMajor: 104, ReleasedYear: 2023, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536, ConnectionFlow: 15663105},
		},
	}

	// Brave knowledge (Chromium kernel, but with randomization)
	kb.browsers[core.BrowserBrave] = &BrowserKnowledge{
		Family:             core.BrowserBrave,
		CommonCipherSuites: []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
		CommonExtensions:   []uint16{0x0000, 0x0017, 0xff01, 0x000a, 0x000b, 0x0023, 0x0016, 0x000d, 0x002b, 0x002d, 0x0033},
		CommonCurves:       []core.CurveID{core.CurveX25519, core.CurveP256, core.CurveP384},
		MarketShare:        0.02,
		Versions: []VersionKnowledge{
			{Version: "1.60", VersionMajor: 1, ReleasedYear: 2024, SupportedOS: desktopOSList(), TLSVersion: 0x0303,
				CipherSuites:        []uint16{0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f, 0xc02c, 0xc030, 0xcca9, 0xcca8},
				H2InitialWindowSize: 6291456, H2MaxConcurrentStreams: 1000, H2HeaderTableSize: 65536, ConnectionFlow: 15663105},
		},
	}
}

func desktopOSList() []core.OperatingSystem {
	return []core.OperatingSystem{core.OSWindows10, core.OSMacOS13, core.OSMacOS14, core.OSMacOS15, core.OSLinux}
}

func appleOSList() []core.OperatingSystem {
	return []core.OperatingSystem{core.OSMacOS13, core.OSMacOS14, core.OSMacOS15, core.OSiOS, core.OSiPadOS}
}

func (kb *KnowledgeBase) computeStats() {
	kb.stats.BrowserMarketShares = make(map[core.BrowserType]float64)
	kb.stats.OSMarketShares = map[string]float64{
		"windows": 0.72, "macos": 0.16, "linux": 0.04,
		"ios": 0.04, "android": 0.03, "other": 0.01,
	}
	totalVersions := 0
	for bt, bk := range kb.browsers {
		kb.stats.BrowserMarketShares[bt] = bk.MarketShare
		totalVersions += len(bk.Versions)
	}
	kb.stats.TotalKnownBrowsers = len(kb.browsers)
	kb.stats.TotalKnownVersions = totalVersions
	kb.stats.TotalKnownProfiles = totalVersions * 3 // Estimate × OS combinations
}

// ===================================================================
// Query interface
// ===================================================================

// GetBrowserKnowledge get knowledge of specified browser family
func (kb *KnowledgeBase) GetBrowserKnowledge(family core.BrowserType) *BrowserKnowledge {
	kb.mu.RLock()
	defer kb.mu.RUnlock()
	return kb.browsers[family]
}

// IsKnownCipherSuite check if cipher suite is known
func (kb *KnowledgeBase) IsKnownCipherSuite(suite uint16) bool {
	for _, s := range kb.tls.ValidTLS13Suites {
		if s == suite {
			return true
		}
	}
	for _, suites := range kb.tls.ValidTLS12Suites {
		for _, s := range suites {
			if s == suite {
				return true
			}
		}
	}
	return false
}

// IsGREASE check if value is a GREASE value
func (kb *KnowledgeBase) IsGREASE(value uint16) bool {
	for _, g := range kb.tls.GREASEValues {
		if value == g {
			return true
		}
	}
	return false
}

// GetExpectedTCPIP get expected TCP/IP parameters for specified OS
func (kb *KnowledgeBase) GetExpectedTCPIP(osFamily string) *TCPIPExpected {
	return kb.tcpip.OSFingerprints[osFamily]
}

// GetExpectedH2 get expected HTTP/2 parameters for specified browser
func (kb *KnowledgeBase) GetExpectedH2(family core.BrowserType) *H2Expected {
	return kb.http2.BrowserSettings[family]
}

// Stats return global statistics
func (kb *KnowledgeBase) Stats() *GlobalStats {
	return kb.stats
}

// FindClosestVersion find the closest matching version in specified browser
func (kb *KnowledgeBase) FindClosestVersion(family core.BrowserType, cipherSuites []uint16) *VersionKnowledge {
	kb.mu.RLock()
	defer kb.mu.RUnlock()

	bk, ok := kb.browsers[family]
	if !ok || len(bk.Versions) == 0 {
		return nil
	}

	bestScore := -1.0
	var bestVersion *VersionKnowledge

	for i := range bk.Versions {
		v := &bk.Versions[i]
		if len(v.CipherSuites) == 0 {
			continue
		}
		score := jaccardSimilarity(cipherSuites, v.CipherSuites)
		if score > bestScore {
			bestScore = score
			bestVersion = v
		}
	}

	return bestVersion
}

// jaccardSimilarity calculate Jaccard similarity of two uint16 sets
func jaccardSimilarity(a, b []uint16) float64 {
	setA := make(map[uint16]struct{}, len(a))
	for _, v := range a {
		setA[v] = struct{}{}
	}
	setB := make(map[uint16]struct{}, len(b))
	for _, v := range b {
		setB[v] = struct{}{}
	}

	intersection := 0
	for v := range setA {
		if _, ok := setB[v]; ok {
			intersection++
		}
	}

	union := len(setA)
	for v := range setB {
		if _, ok := setA[v]; !ok {
			union++
		}
	}

	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// sortedUint16 return sorted copy
func sortedUint16(in []uint16) []uint16 {
	out := make([]uint16, len(in))
	copy(out, in)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
