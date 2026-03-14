package tcpip

import (
	"fmt"
	"net"
	"strings"
)

func (ia *IntegratedFingerprinter) checkConsistency(ctx *FingerprintContext, result *IntegratedResult) {
	for _, rule := range ia.consistencyRules {
		if inconsistency := rule.Check(ctx); inconsistency != nil {
			result.Inconsistencies = append(result.Inconsistencies, *inconsistency)
		}
	}
}

// calculateOverallConfidence calculates overall confidence
func (ia *IntegratedFingerprinter) calculateOverallConfidence(result *IntegratedResult) {
	confidence := 0.5 // base confidence

	// TCP layer contribution (0.3)
	if result.TCPResult != nil && result.TCPResult.Confidence > 0 {
		confidence += result.TCPResult.Confidence * 0.3
	}

	// UA layer contribution (0.3)
	if result.ParsedOSFromUA != "" {
		confidence += 0.3
	}

	// Cross-validation consistency contribution (0.4)
	confidence += result.OSCrossValidation.MatchScore * 0.4

	// If inconsistencies are severe, reduce confidence
	for _, inc := range result.Inconsistencies {
		switch inc.Severity {
		case "high":
			confidence -= 0.2
		case "medium":
			confidence -= 0.1
		case "low":
			confidence -= 0.05
		}
	}

	// Clamp to 0-1 range
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	result.OverallConfidence = confidence

	// Calculate risk score
	result.RiskScore = ia.calculateRiskScore(result)
}

// calculateRiskScore calculates risk score
func (ia *IntegratedFingerprinter) calculateRiskScore(result *IntegratedResult) float64 {
	risk := 0.0

	// Calculate risk based on inconsistencies
	for _, inc := range result.Inconsistencies {
		switch inc.Severity {
		case "high":
			risk += 0.3
		case "medium":
			risk += 0.15
		case "low":
			risk += 0.05
		}
	}

	// Low confidence increases risk
	if result.OverallConfidence < 0.3 {
		risk += 0.3
	} else if result.OverallConfidence < 0.5 {
		risk += 0.15
	}

	// IP-UA inconsistency increases risk
	if !result.IPUAConsistency {
		risk += 0.2
	}

	if risk > 1 {
		risk = 1
	}

	return risk
}

// Helper methods

func (ia *IntegratedFingerprinter) inferOSFromGeography(geo *GeoLocation) string {
	// Infer OS from geolocation (some regions have specific preferences)
	// For example: some enterprise networks may use Windows uniformly
	return ""
}

func (ia *IntegratedFingerprinter) checkIPUAConsistency(ctx *FingerprintContext, result *IntegratedResult) bool {
	// Check if the language settings in UA match IP geolocation
	if ctx.GeoLocation == nil || ctx.HTTPHeaders == nil {
		return true // unable to check, default to consistent
	}

	acceptLang := ctx.HTTPHeaders["Accept-Language"]
	if acceptLang == "" {
		return true
	}

	// Simple check: if IP is in China but UA language is only en-US, it may be inconsistent
	// Real applications need more complex language-region mapping
	return true
}

func (ia *IntegratedFingerprinter) checkGeoUAConsistency(ctx *FingerprintContext, result *IntegratedResult) bool {
	// Check if the timezone/region info in UA matches geolocation
	return true
}

func isPrivateIP(ip net.IP) bool {
	// Check if it is a private address
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}

	for _, cidr := range privateRanges {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil && ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

// buildOSUAMapping builds the OS to UA mapping
func buildOSUAMapping() map[string][]string {
	return map[string][]string{
		"Windows": {
			"Windows NT 10.0",
			"Windows NT 6.3",
			"Windows NT 6.2",
			"Windows NT 6.1",
		},
		"macOS": {
			"Macintosh; Intel Mac OS X",
			"Macintosh; PPC Mac OS X",
		},
		"Linux": {
			"X11; Linux",
			"X11; Ubuntu; Linux",
		},
		"Android": {
			"Android",
			"Linux; Android",
		},
		"iOS": {
			"iPhone; CPU iPhone OS",
			"iPad; CPU OS",
		},
	}
}

// buildConsistencyRules builds consistency check rules
func buildConsistencyRules() []ConsistencyRule {
	return []ConsistencyRule{
		{
			Name:        "TTL_OS_Mismatch",
			Description: "Check if TTL value matches the claimed OS",
			Check: func(ctx *FingerprintContext) *Inconsistency {
				if ctx.TCPPacket == nil || ctx.TCPPacket.IPHeader == nil {
					return nil
				}

				ttl := ctx.TCPPacket.IPHeader.TimeToLive
				uaOS := ""

				// Infer OS from UA
				uaLower := strings.ToLower(ctx.UserAgent)
				switch {
				case strings.Contains(uaLower, "windows"):
					uaOS = "Windows"
				case strings.Contains(uaLower, "mac"):
					uaOS = "macOS"
				case strings.Contains(uaLower, "linux"):
					uaOS = "Linux"
				}

				// Windows typically uses TTL 128, Linux/macOS uses 64
				if uaOS == "Windows" && ttl <= 64 {
					return &Inconsistency{
						RuleName:    "TTL_OS_Mismatch",
						Severity:    "medium",
						Description: "Windows typically uses TTL 128, but observed TTL <= 64",
						Expected:    "TTL ~128 for Windows",
						Actual:      fmt.Sprintf("TTL %d", ttl),
					}
				}

				return nil
			},
		},
		{
			Name:        "Mobile_WindowSize",
			Description: "Check if the window size for mobile devices is reasonable",
			Check: func(ctx *FingerprintContext) *Inconsistency {
				if ctx.TCPPacket == nil {
					return nil
				}

				isMobile := strings.Contains(strings.ToLower(ctx.UserAgent), "mobile")
				windowSize := ctx.TCPPacket.WindowSize

				// Mobile devices typically have smaller window sizes
				if isMobile && windowSize > 65000 {
					return &Inconsistency{
						RuleName:    "Mobile_WindowSize",
						Severity:    "low",
						Description: "Mobile devices typically have smaller window sizes",
						Expected:    "Window size < 65000 for mobile",
						Actual:      fmt.Sprintf("Window size %d", windowSize),
					}
				}

				return nil
			},
		},
	}
}

// SetIPRegionDB sets the IP region database
func (ia *IntegratedFingerprinter) SetIPRegionDB(db IPRegionDatabase) {
	ia.mu.Lock()
	defer ia.mu.Unlock()
	ia.ipRegionDB = db
}
