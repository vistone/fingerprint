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
