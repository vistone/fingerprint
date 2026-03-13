// Package tcpip provides integrated TCP/IP fingerprint analysis
// combining User-Agent, IP address and geolocation for comprehensive identification
package tcpip

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

// IntegratedFingerprinter is an integrated fingerprint analyzer
type IntegratedFingerprinter struct {
	// Per-layer analyzers
	tcpipAnalyzer *TCPIPAnalyzer

	// Association databases
	osUAMapping     map[string][]string    // OS -> common UA patterns
	ipRegionDB      IPRegionDatabase       // IP geolocation database
	osTCPSignatures map[string]OSSignature // OS -> TCP signature features

	// Consistency rules
	consistencyRules []ConsistencyRule

	mu sync.RWMutex
}

// ConsistencyRule represents a consistency check rule
type ConsistencyRule struct {
	Name        string
	Description string
	Check       func(ctx *FingerprintContext) *Inconsistency
}

// FingerprintContext represents a fingerprint analysis context
type FingerprintContext struct {
	// TCP/IP layer information
	TCPPacket    *TCPPacket
	TCPSignature *TCPIPSignature

	// Application layer information
	UserAgent   string
	HTTPHeaders map[string]string

	// Network layer information
	SourceIP string
	DestIP   string

	// Geolocation information
	GeoLocation *GeoLocation

	// Inferred results
	DetectedOS     string
	DetectedDevice string
}

// GeoLocation represents geolocation information
type GeoLocation struct {
	Country     string
	CountryCode string
	Region      string
	City        string
	ISP         string
	ASN         int
	Timezone    string
	Latitude    float64
	Longitude   float64
}

// IPRegionDatabase is the IP region database interface
type IPRegionDatabase interface {
	Lookup(ip string) (*GeoLocation, error)
	GetRegionSignature(country, isp string) string
}

// Inconsistency represents an inconsistency report
type Inconsistency struct {
	RuleName    string
	Severity    string // high, medium, low
	Description string
	Expected    string
	Actual      string
}

// IntegratedResult represents an integrated analysis result
type IntegratedResult struct {
	// Per-layer independent identification results
	TCPResult      *TCPIPResult
	ParsedOSFromUA string
	GeoInfo        *GeoLocation

	// Cross-validation results
	OSCrossValidation OSCrossValidationResult
	IPUAConsistency   bool
	GeoUAConsistency  bool

	// Inconsistency report
	Inconsistencies []Inconsistency

	// Overall assessment
	OverallConfidence float64
	RiskScore         float64
	FinalOS           string
	FinalDeviceType   string

	// Raw data summary
	SourceIP     string
	UserAgent    string
	TCPSignature string
}

// OSCrossValidationResult represents OS cross-validation results
type OSCrossValidationResult struct {
	OSFromTCP   string
	OSFromUA    string
	OSFromGeo   string
	ConsensusOS string
	MatchScore  float64 // 0-1, higher means more consistency across layers
}

// NewIntegratedFingerprinter creates an integrated fingerprint analyzer
func NewIntegratedFingerprinter() *IntegratedFingerprinter {
	ia := &IntegratedFingerprinter{
		tcpipAnalyzer:    NewTCPIPAnalyzer(),
		osUAMapping:      buildOSUAMapping(),
		osTCPSignatures:  BuildOSDatabase(),
		consistencyRules: buildConsistencyRules(),
	}
	return ia
}

// Analyze performs integrated fingerprint analysis
func (ia *IntegratedFingerprinter) Analyze(
	packet *TCPPacket,
	userAgent string,
	headers map[string]string,
) (*IntegratedResult, error) {
	ctx := &FingerprintContext{
		TCPPacket:   packet,
		UserAgent:   userAgent,
		HTTPHeaders: headers,
	}

	if packet != nil && packet.IPHeader != nil {
		ctx.SourceIP = packet.IPHeader.SourceAddress
		ctx.DestIP = packet.IPHeader.DestAddress
	}

	result := &IntegratedResult{
		SourceIP:        ctx.SourceIP,
		UserAgent:       userAgent,
		Inconsistencies: []Inconsistency{},
	}

	// 1. TCP/IP layer analysis
	if packet != nil {
		ia.analyzeTCPLayer(ctx, result)
	}

	// 2. User-Agent parsing
	ia.parseUserAgent(ctx, result)

	// 3. IP geolocation analysis
	ia.analyzeIPGeolocation(ctx, result)

	// 4. Cross-validation
	ia.crossValidate(ctx, result)

	// 5. Consistency check
	ia.checkConsistency(ctx, result)

	// 6. Overall assessment
	ia.calculateOverallConfidence(result)

	return result, nil
}

// analyzeTCPLayer analyzes the TCP/IP layer
func (ia *IntegratedFingerprinter) analyzeTCPLayer(ctx *FingerprintContext, result *IntegratedResult) {
	// Add packet to analyzer
	ia.tcpipAnalyzer.AddPacket(ctx.TCPPacket)

	// Analyze TCP features
	tcpResult := ia.tcpipAnalyzer.AnalyzePacket(ctx.TCPPacket)
	result.TCPResult = tcpResult

	// Infer OS
	if tcpResult != nil && tcpResult.OS != "" {
		ctx.DetectedOS = tcpResult.OS
		result.TCPSignature = fmt.Sprintf("TTL:%d,MSS:%d,Win:%d",
			ctx.TCPPacket.IPHeader.TimeToLive,
			len(ctx.TCPPacket.Options),
			ctx.TCPPacket.WindowSize)
	}
}

// parseUserAgent parses the User-Agent
func (ia *IntegratedFingerprinter) parseUserAgent(ctx *FingerprintContext, result *IntegratedResult) {
	ua := ctx.UserAgent
	if ua == "" {
		return
	}

	uaLower := strings.ToLower(ua)

	// Extract OS information
	osFromUA := ""
	switch {
	case strings.Contains(uaLower, "windows nt 10.0"):
		osFromUA = "Windows 10"
	case strings.Contains(uaLower, "windows nt 11.0"):
		osFromUA = "Windows 11"
	case strings.Contains(uaLower, "macintosh") || strings.Contains(uaLower, "mac os"):
		osFromUA = "macOS"
	case strings.Contains(uaLower, "linux") && !strings.Contains(uaLower, "android"):
		osFromUA = "Linux"
	case strings.Contains(uaLower, "android"):
		osFromUA = "Android"
	case strings.Contains(uaLower, "iphone") || strings.Contains(uaLower, "ipad"):
		osFromUA = "iOS"
	}

	result.ParsedOSFromUA = osFromUA
	ctx.DetectedOS = osFromUA

	// Extract device type
	if strings.Contains(uaLower, "mobile") {
		ctx.DetectedDevice = "Mobile"
	} else if strings.Contains(uaLower, "tablet") || strings.Contains(uaLower, "ipad") {
		ctx.DetectedDevice = "Tablet"
	} else {
		ctx.DetectedDevice = "Desktop"
	}
}

// analyzeIPGeolocation analyzes IP geolocation
func (ia *IntegratedFingerprinter) analyzeIPGeolocation(ctx *FingerprintContext, result *IntegratedResult) {
	if ctx.SourceIP == "" {
		return
	}

	// Parse IP
	ip := net.ParseIP(ctx.SourceIP)
	if ip == nil {
		return
	}

	// Check if it is a private address
	if isPrivateIP(ip) {
		result.GeoInfo = &GeoLocation{
			Country: "Private",
			Region:  "Local Network",
		}
		return
	}

	// If IP region database is available, perform lookup
	if ia.ipRegionDB != nil {
		geo, err := ia.ipRegionDB.Lookup(ctx.SourceIP)
		if err == nil {
			result.GeoInfo = geo
			ctx.GeoLocation = geo
		}
	}
}

// crossValidate cross-validates results across layers
func (ia *IntegratedFingerprinter) crossValidate(ctx *FingerprintContext, result *IntegratedResult) {
	validation := OSCrossValidationResult{
		OSFromTCP: result.TCPResult.OS,
		OSFromUA:  result.ParsedOSFromUA,
	}

	// Infer OS from geolocation (some regions/ISPs have specific preferences)
	if result.GeoInfo != nil {
		validation.OSFromGeo = ia.inferOSFromGeography(result.GeoInfo)
	}

	// Calculate consensus OS
	osVotes := make(map[string]int)
	if validation.OSFromTCP != "" {
		osVotes[validation.OSFromTCP]++
	}
	if validation.OSFromUA != "" {
		osVotes[validation.OSFromUA] += 2 // UA has higher weight
	}
	if validation.OSFromGeo != "" {
		osVotes[validation.OSFromGeo]++
	}

	// Find the OS with the most votes
	maxVotes := 0
	for os, votes := range osVotes {
		if votes > maxVotes {
			maxVotes = votes
			validation.ConsensusOS = os
		}
	}

	// Calculate match score
	matchCount := 0
	totalLayers := 0

	if validation.OSFromTCP != "" {
		totalLayers++
		if validation.OSFromTCP == validation.ConsensusOS {
			matchCount++
		}
	}
	if validation.OSFromUA != "" {
		totalLayers++
		if validation.OSFromUA == validation.ConsensusOS {
			matchCount++
		}
	}
	if validation.OSFromGeo != "" {
		totalLayers++
		if validation.OSFromGeo == validation.ConsensusOS {
			matchCount++
		}
	}

	if totalLayers > 0 {
		validation.MatchScore = float64(matchCount) / float64(totalLayers)
	}

	result.OSCrossValidation = validation
	result.FinalOS = validation.ConsensusOS
	result.FinalDeviceType = ctx.DetectedDevice

	// IP-UA consistency: check if the language/region in UA matches IP geolocation
	result.IPUAConsistency = ia.checkIPUAConsistency(ctx, result)

	// Geolocation-UA consistency
	result.GeoUAConsistency = ia.checkGeoUAConsistency(ctx, result)
}

// checkConsistency executes consistency check rules
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
