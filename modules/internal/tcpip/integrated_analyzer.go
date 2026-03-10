// translated comment
// translated comment
package tcpip

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

// translated comment
type IntegratedFingerprinter struct {
	// translated comment
	 tcpipAnalyzer *TCPIPAnalyzer
	 
	 // translated comment
	osUAMapping     map[string][]string // translated comment
	ipRegionDB      IPRegionDatabase    // translated comment
	osTCPSignatures map[string]OSSignature // translated comment
	 
	 // translated comment
	consistencyRules []ConsistencyRule
	 
	mu sync.RWMutex
}

// translated comment
type ConsistencyRule struct {
	Name        string
	Description string
	Check       func(ctx *FingerprintContext) *Inconsistency
}

// translated comment
type FingerprintContext struct {
	// translated comment
	TCPPacket *TCPPacket
	TCPSignature *TCPIPSignature
	 
	// translated comment
	UserAgent    string
	HTTPHeaders  map[string]string
	 
	// translated comment
	SourceIP     string
	DestIP       string
	 
	// translated comment
	GeoLocation  *GeoLocation
	 
	// translated comment
	DetectedOS   string
	DetectedDevice string
}

// translated comment
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

// translated comment
type IPRegionDatabase interface {
	Lookup(ip string) (*GeoLocation, error)
	GetRegionSignature(country, isp string) string
}

// translated comment
type Inconsistency struct {
	RuleName    string
	Severity    string // high, medium, low
	Description string
	Expected    string
	Actual      string
}

// translated comment
type IntegratedResult struct {
	// translated comment
	TCPResult      *TCPIPResult
	ParsedOSFromUA string
	GeoInfo        *GeoLocation
	 
	// translated comment
	OSCrossValidation OSCrossValidationResult
	IPUAConsistency   bool
	GeoUAConsistency  bool
	 
	// translated comment
	Inconsistencies []Inconsistency
	 
	// translated comment
	OverallConfidence float64
	RiskScore         float64
	FinalOS           string
	FinalDeviceType   string
	 
	// translated comment
	SourceIP      string
	UserAgent     string
	TCPSignature  string
}

// translated comment
type OSCrossValidationResult struct {
	OSFromTCP      string
	OSFromUA       string
	OSFromGeo      string
	ConsensusOS    string
	MatchScore     float64 // translated comment
}

// translated comment
func NewIntegratedFingerprinter() *IntegratedFingerprinter {
	ia := &IntegratedFingerprinter{
		tcpipAnalyzer:     NewTCPIPAnalyzer(),
		osUAMapping:       buildOSUAMapping(),
		osTCPSignatures:   BuildOSDatabase(),
		consistencyRules:  buildConsistencyRules(),
	}
	return ia
}

// translated comment
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
		SourceIP:      ctx.SourceIP,
		UserAgent:     userAgent,
		Inconsistencies: []Inconsistency{},
	}
	
	// translated comment
	if packet != nil {
		ia.analyzeTCPLayer(ctx, result)
	}
	
	// translated comment
	ia.parseUserAgent(ctx, result)
	
	// translated comment
	ia.analyzeIPGeolocation(ctx, result)
	
	// translated comment
	ia.crossValidate(ctx, result)
	
	// translated comment
	ia.checkConsistency(ctx, result)
	
	// translated comment
	ia.calculateOverallConfidence(result)
	
	return result, nil
}

// translated comment
func (ia *IntegratedFingerprinter) analyzeTCPLayer(ctx *FingerprintContext, result *IntegratedResult) {
	// translated comment
	ia.tcpipAnalyzer.AddPacket(ctx.TCPPacket)
	
	// translated comment
	tcpResult := ia.tcpipAnalyzer.AnalyzePacket(ctx.TCPPacket)
	result.TCPResult = tcpResult
	
	// translated comment
	if tcpResult != nil && tcpResult.OS != "" {
		ctx.DetectedOS = tcpResult.OS
		result.TCPSignature = fmt.Sprintf("TTL:%d,MSS:%d,Win:%d", 
			ctx.TCPPacket.IPHeader.TimeToLive,
			len(ctx.TCPPacket.Options),
			ctx.TCPPacket.WindowSize)
	}
}

// translated comment
func (ia *IntegratedFingerprinter) parseUserAgent(ctx *FingerprintContext, result *IntegratedResult) {
	ua := ctx.UserAgent
	if ua == "" {
		return
	}
	
	uaLower := strings.ToLower(ua)
	
	// translated comment
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
	
	// translated comment
	if strings.Contains(uaLower, "mobile") {
		ctx.DetectedDevice = "Mobile"
	} else if strings.Contains(uaLower, "tablet") || strings.Contains(uaLower, "ipad") {
		ctx.DetectedDevice = "Tablet"
	} else {
		ctx.DetectedDevice = "Desktop"
	}
}

// translated comment
func (ia *IntegratedFingerprinter) analyzeIPGeolocation(ctx *FingerprintContext, result *IntegratedResult) {
	if ctx.SourceIP == "" {
		return
	}
	
	// translated comment
	ip := net.ParseIP(ctx.SourceIP)
	if ip == nil {
		return
	}
	
	// translated comment
	if isPrivateIP(ip) {
		result.GeoInfo = &GeoLocation{
			Country: "Private",
			Region:  "Local Network",
		}
		return
	}
	
	// translated comment
	if ia.ipRegionDB != nil {
		geo, err := ia.ipRegionDB.Lookup(ctx.SourceIP)
		if err == nil {
			result.GeoInfo = geo
			ctx.GeoLocation = geo
		}
	}
}

// translated comment
func (ia *IntegratedFingerprinter) crossValidate(ctx *FingerprintContext, result *IntegratedResult) {
	validation := OSCrossValidationResult{
		OSFromTCP: result.TCPResult.OS,
		OSFromUA:  result.ParsedOSFromUA,
	}
	
	// translated comment
	if result.GeoInfo != nil {
		validation.OSFromGeo = ia.inferOSFromGeography(result.GeoInfo)
	}
	
	// translated comment
	osVotes := make(map[string]int)
	if validation.OSFromTCP != "" {
		osVotes[validation.OSFromTCP]++
	}
	if validation.OSFromUA != "" {
		osVotes[validation.OSFromUA] += 2 // translated comment
	}
	if validation.OSFromGeo != "" {
		osVotes[validation.OSFromGeo]++
	}
	
	// translated comment
	maxVotes := 0
	for os, votes := range osVotes {
		if votes > maxVotes {
			maxVotes = votes
			validation.ConsensusOS = os
		}
	}
	
	// translated comment
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
	
	// translated comment
	result.IPUAConsistency = ia.checkIPUAConsistency(ctx, result)
	
	// translated comment
	result.GeoUAConsistency = ia.checkGeoUAConsistency(ctx, result)
}

// translated comment
func (ia *IntegratedFingerprinter) checkConsistency(ctx *FingerprintContext, result *IntegratedResult) {
	for _, rule := range ia.consistencyRules {
		if inconsistency := rule.Check(ctx); inconsistency != nil {
			result.Inconsistencies = append(result.Inconsistencies, *inconsistency)
		}
	}
}

// translated comment
func (ia *IntegratedFingerprinter) calculateOverallConfidence(result *IntegratedResult) {
	confidence := 0.5 // translated comment
	
	// translated comment
	if result.TCPResult != nil && result.TCPResult.Confidence > 0 {
		confidence += result.TCPResult.Confidence * 0.3
	}
	
	// translated comment
	if result.ParsedOSFromUA != "" {
		confidence += 0.3
	}
	
	// translated comment
	confidence += result.OSCrossValidation.MatchScore * 0.4
	
	// translated comment
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
	
	// translated comment
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}
	
	result.OverallConfidence = confidence
	
	// translated comment
	result.RiskScore = ia.calculateRiskScore(result)
}

// translated comment
func (ia *IntegratedFingerprinter) calculateRiskScore(result *IntegratedResult) float64 {
	risk := 0.0
	
	// translated comment
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
	
	// translated comment
	if result.OverallConfidence < 0.3 {
		risk += 0.3
	} else if result.OverallConfidence < 0.5 {
		risk += 0.15
	}
	
	// translated comment
	if !result.IPUAConsistency {
		risk += 0.2
	}
	
	if risk > 1 {
		risk = 1
	}
	
	return risk
}

// translated comment

func (ia *IntegratedFingerprinter) inferOSFromGeography(geo *GeoLocation) string {
	// translated comment
	// translated comment
	return ""
}

func (ia *IntegratedFingerprinter) checkIPUAConsistency(ctx *FingerprintContext, result *IntegratedResult) bool {
	// translated comment
	if ctx.GeoLocation == nil || ctx.HTTPHeaders == nil {
		return true // translated comment
	}
	
	acceptLang := ctx.HTTPHeaders["Accept-Language"]
	if acceptLang == "" {
		return true
	}
	
	// translated comment
	// translated comment
	return true
}

func (ia *IntegratedFingerprinter) checkGeoUAConsistency(ctx *FingerprintContext, result *IntegratedResult) bool {
	// translated comment
	return true
}

func isPrivateIP(ip net.IP) bool {
	// translated comment
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

// translated comment
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

// translated comment
func buildConsistencyRules() []ConsistencyRule {
	return []ConsistencyRule{
		{
			Name:        "TTL_OS_Mismatch",
			Description: "检查 TTL 值与声称的 OS 是否匹配",
			Check: func(ctx *FingerprintContext) *Inconsistency {
				if ctx.TCPPacket == nil || ctx.TCPPacket.IPHeader == nil {
					return nil
				}
				
				ttl := ctx.TCPPacket.IPHeader.TimeToLive
				uaOS := ""
				
				// translated comment
				uaLower := strings.ToLower(ctx.UserAgent)
				switch {
				case strings.Contains(uaLower, "windows"):
					uaOS = "Windows"
				case strings.Contains(uaLower, "mac"):
					uaOS = "macOS"
				case strings.Contains(uaLower, "linux"):
					uaOS = "Linux"
				}
				
				// translated comment
				if uaOS == "Windows" && ttl <= 64 {
					return &Inconsistency{
						RuleName:    "TTL_OS_Mismatch",
						Severity:    "medium",
						Description: "Windows 通常使用 TTL 128，但观察到 TTL <= 64",
						Expected:    "TTL ~128 for Windows",
						Actual:      fmt.Sprintf("TTL %d", ttl),
					}
				}
				
				return nil
			},
		},
		{
			Name:        "Mobile_WindowSize",
			Description: "检查移动设备的窗口大小是否合理",
			Check: func(ctx *FingerprintContext) *Inconsistency {
				if ctx.TCPPacket == nil {
					return nil
				}
				
				isMobile := strings.Contains(strings.ToLower(ctx.UserAgent), "mobile")
				windowSize := ctx.TCPPacket.WindowSize
				
				// translated comment
				if isMobile && windowSize > 65000 {
					return &Inconsistency{
						RuleName:    "Mobile_WindowSize",
						Severity:    "low",
						Description: "移动设备通常窗口大小较小",
						Expected:    "Window size < 65000 for mobile",
						Actual:      fmt.Sprintf("Window size %d", windowSize),
					}
				}
				
				return nil
			},
		},
	}
}

// translated comment
func (ia *IntegratedFingerprinter) SetIPRegionDB(db IPRegionDatabase) {
	ia.mu.Lock()
	defer ia.mu.Unlock()
	ia.ipRegionDB = db
}
