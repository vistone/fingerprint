package tcpip

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DeviceFingerprintingEngine is a device fingerprint identification engine
type DeviceFingerprintingEngine struct {
	deviceProfiles     map[string]*DeviceProfile
	behaviorAnalyzer   *NetworkBehaviorAnalyzer
	riskAssessment     *RiskAssessmentEngine
	osDatabase         map[string]*OSProfile
	lastAnalysisResult *DeviceFingerprintResult
}

// DeviceProfile represents a device profile
type DeviceProfile struct {
	Name              string
	DeviceType        string // phone, tablet, laptop, desktop, iot, router, server
	Manufacturer      string
	OS                string
	OSVersion         string
	BrowserSignatures map[string]string // browser name -> fingerprint
	NetworkSignatures map[string]string // signal -> value
	Applications      []string          // common applications
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Confidence        float64
}

// OSProfile represents an operating system profile
type OSProfile struct {
	Name              string
	Family            string // Windows, Linux, Darwin, iOS, Android
	Version           string
	DefaultTTL        int
	DefaultMSS        int
	DefaultWindowSize int
	TCPOptions        string
	Behavior          string
	Kernel            string
	BoostrapMSS       int
}

// DeviceFingerprintingConfig is the device fingerprint identification configuration
type DeviceFingerprintingConfig struct {
	EnableBehaviorAnalysis bool
	EnableRiskAssessment   bool
	EnableGeoIP            bool
	SamplingRate           float64 // 0.0-1.0
	CacheTTL               time.Duration
	MaxPacketsPerAnalysis  int
}

// NewDeviceFingerprintingEngine creates a new device fingerprint identification engine
func NewDeviceFingerprintingEngine() *DeviceFingerprintingEngine {
	return &DeviceFingerprintingEngine{
		deviceProfiles:   make(map[string]*DeviceProfile),
		behaviorAnalyzer: NewNetworkBehaviorAnalyzer(),
		riskAssessment:   NewRiskAssessmentEngine(),
		osDatabase:       initializeOSDatabase(),
	}
}

// RegisterDeviceProfile registers a device profile
func (dfe *DeviceFingerprintingEngine) RegisterDeviceProfile(profile *DeviceProfile) error {
	if profile.Name == "" {
		return fmt.Errorf("device profile name is required")
	}

	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()
	dfe.deviceProfiles[profile.Name] = profile

	return nil
}

// AnalyzeDevice analyzes a device
func (dfe *DeviceFingerprintingEngine) AnalyzeDevice(packets []*TCPPacket, behaviors *NetworkBehaviorResult) *DeviceFingerprintResult {
	result := &DeviceFingerprintResult{
		Timestamp:       time.Now(),
		AnalyzedPackets: len(packets),
		DeviceMatches:   make([]*DeviceMatch, 0),
		OSMatches:       make([]*OSMatch, 0),
		RiskIndicators:  make([]string, 0),
		Confidence:      0,
	}

	if len(packets) == 0 {
		return result
	}

	// Analyze packet features
	features := dfe.extractFeatures(packets)

	// Match device profiles
	result.DeviceMatches = dfe.matchDeviceProfiles(features)

	// Match operating systems
	result.OSMatches = dfe.matchOSProfiles(packets[0])

	// Perform risk assessment
	result.RiskIndicators = dfe.assessRisk(packets, behaviors)

	// Calculate overall confidence
	result.Confidence = dfe.calculateConfidence(result)

	// Add behavior analysis
	if behaviors != nil {
		result.BehaviorPattern = behaviors.SequenceNumberPattern
		result.NetworkType = behaviors.RTTAnalysis.NetworkType
	}

	dfe.lastAnalysisResult = result
	return result
}

// extractFeatures extracts features
func (dfe *DeviceFingerprintingEngine) extractFeatures(packets []*TCPPacket) map[string]interface{} {
	features := make(map[string]interface{})

	if len(packets) == 0 {
		return features
	}

	firstPacket := packets[0]

	// Extract IP features
	if firstPacket.IPHeader != nil {
		features["ttl"] = firstPacket.IPHeader.TimeToLive
		features["df"] = (firstPacket.IPHeader.Flags & 0x40) != 0
		features["protocol"] = firstPacket.IPHeader.Protocol
	}

	// Extract TCP features
	features["src_port"] = firstPacket.SourcePort
	features["dst_port"] = firstPacket.DestinationPort
	features["flags"] = firstPacket.Flags
	features["window_size"] = firstPacket.WindowSize

	// Extract TCP options
	if len(firstPacket.Options) > 0 {
		optStrings := make([]string, 0)
		for _, opt := range firstPacket.Options {
			optStrings = append(optStrings, fmt.Sprintf("opt_%d", opt))
		}
		features["tcp_options"] = strings.Join(optStrings, ",")
	}

	// MSS features
	mss := 0
	for _, opt := range firstPacket.Options {
		if opt == OptMSS {
			mss = 1460 // standard MSS
		}
	}
	if mss > 0 {
		features["mss"] = mss
	}

	// Packet count statistics
	features["packet_count"] = len(packets)

	// Calculate session duration
	if len(packets) > 1 {
		sessionDuration := packets[len(packets)-1].Timestamp.Sub(packets[0].Timestamp)
		features["session_duration"] = sessionDuration
	}

	return features
}

// matchDeviceProfiles matches device profiles
func (dfe *DeviceFingerprintingEngine) matchDeviceProfiles(features map[string]interface{}) []*DeviceMatch {
	matches := make([]*DeviceMatch, 0)

	for profileName, profile := range dfe.deviceProfiles {
		score := dfe.calculateProfileScore(features, profile)

		if score > 0.5 {
			matches = append(matches, &DeviceMatch{
				DeviceName:    profileName,
				DeviceType:    profile.DeviceType,
				Manufacturer:  profile.Manufacturer,
				MatchScore:    score,
				MatchedFields: extractMatchedFields(features, profile),
			})
		}
	}

	// Sort by score
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].MatchScore > matches[i].MatchScore {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	return matches
}

// matchOSProfiles matches operating system profiles
func (dfe *DeviceFingerprintingEngine) matchOSProfiles(packet *TCPPacket) []*OSMatch {
	matches := make([]*OSMatch, 0)

	if packet.IPHeader == nil {
		return matches
	}

	for osName, osProfile := range dfe.osDatabase {
		score := dfe.calculateOSScore(packet, osProfile)

		if score > 0.5 {
			matches = append(matches, &OSMatch{
				OSName:     osName,
				OSFamily:   osProfile.Family,
				MatchScore: score,
				DefaultTTL: osProfile.DefaultTTL,
				DefaultMSS: osProfile.DefaultMSS,
				TCPOptions: osProfile.TCPOptions,
			})
		}
	}

	return matches
}

// assessRisk assesses risk
func (dfe *DeviceFingerprintingEngine) assessRisk(packets []*TCPPacket, behaviors *NetworkBehaviorResult) []string {
	riskIndicators := make([]string, 0)

	// Check for anomalies
	for _, packet := range packets {
		if packet.IPHeader != nil {
			// Check for non-standard TTL
			if !dfe.isStandardTTL(packet.IPHeader.TimeToLive) {
				riskIndicators = append(riskIndicators, "non_standard_ttl")
				break
			}

			// Check for suspicious IP
			if IsReservedIP(packet.IPHeader.SourceAddress) {
				riskIndicators = append(riskIndicators, "reserved_ip")
				break
			}
		}

		// Check for anomalous TCP options
		if len(packet.Options) == 0 {
			riskIndicators = append(riskIndicators, "missing_tcp_options")
			break
		}
	}

	// Behavior-based risk assessment
	if behaviors != nil {
		for _, characteristic := range behaviors.BehaviorCharacteristics {
			if characteristic == "scanning" {
				riskIndicators = append(riskIndicators, "scanning_behavior")
			} else if characteristic == "automated" {
				riskIndicators = append(riskIndicators, "automated_traffic")
			}
		}
	}

	return riskIndicators
}

// calculateProfileScore calculates profile match score
func (dfe *DeviceFingerprintingEngine) calculateProfileScore(features map[string]interface{}, profile *DeviceProfile) float64 {
	score := 0.0
	matches := 0
	total := 0

	// Check TTL
	if ttl, ok := features["ttl"]; ok {
		total++
		// Device profiles are typically associated with a specific OS, which has a specific TTL
		if ttlInt, ok := ttl.(int); ok && ttlInt == 64 {
			matches++
		}
	}

	// Check TCP options
	if opts, ok := features["tcp_options"]; ok {
		total++
		if optStr, ok := opts.(string); ok {
			for _, netSig := range profile.NetworkSignatures {
				if strings.Contains(optStr, netSig) {
					matches++
					break
				}
			}
		}
	}

	if total > 0 {
		score = float64(matches) / float64(total)
	}

	return score
}

// calculateOSScore calculates operating system match score
func (dfe *DeviceFingerprintingEngine) calculateOSScore(packet *TCPPacket, osProfile *OSProfile) float64 {
	matches := 0
	total := 3

	// Check TTL
	if packet.IPHeader.TimeToLive == uint8(osProfile.DefaultTTL) {
		matches++
	}

	// Check window size
	if packet.WindowSize == uint16(osProfile.DefaultWindowSize) {
		matches++
	}

	// Check DF flag
	if (packet.IPHeader.Flags & 0x40) != 0 {
		matches++
	}

	return float64(matches) / float64(total)
}

// calculateConfidence calculates overall confidence
func (dfe *DeviceFingerprintingEngine) calculateConfidence(result *DeviceFingerprintResult) float64 {
	if len(result.DeviceMatches) == 0 {
		return 0
	}

	avgScore := 0.0
	for _, match := range result.DeviceMatches {
		avgScore += match.MatchScore
	}

	avgScore /= float64(len(result.DeviceMatches))

	// Adjust confidence based on risk
	riskPenalty := float64(len(result.RiskIndicators)) * 0.1
	confidence := avgScore - riskPenalty

	if confidence < 0 {
		confidence = 0
	}

	return confidence
}

// isStandardTTL checks whether the TTL is a standard value
func (dfe *DeviceFingerprintingEngine) isStandardTTL(ttl uint8) bool {
	standardTTLs := []uint8{32, 64, 128, 255}
	for _, std := range standardTTLs {
		if ttl == std {
			return true
		}
	}
	return false
}

// DeviceFingerprintResult represents device fingerprint identification results
type DeviceFingerprintResult struct {
	Timestamp       time.Time
	AnalyzedPackets int
	DeviceMatches   []*DeviceMatch
	OSMatches       []*OSMatch
	RiskIndicators  []string
	Confidence      float64
	BehaviorPattern string
	NetworkType     string
	DetailedReport  string
}

// DeviceMatch represents a device match result
type DeviceMatch struct {
	DeviceName    string
	DeviceType    string
	Manufacturer  string
	MatchScore    float64
	MatchedFields []string
}

// OSMatch represents an operating system match result
type OSMatch struct {
	OSName     string
	OSFamily   string
	MatchScore float64
	DefaultTTL int
	DefaultMSS int
	TCPOptions string
}

// String returns a human-readable fingerprint identification result
func (dfr *DeviceFingerprintResult) String() string {
	data, _ := json.MarshalIndent(dfr, "", "  ")
	return string(data)
}

// extractMatchedFields extracts matched fields
func extractMatchedFields(features map[string]interface{}, profile *DeviceProfile) []string {
	fields := make([]string, 0)

	for netSig := range profile.NetworkSignatures {
		for featureKey := range features {
			if strings.Contains(featureKey, "network") || strings.Contains(featureKey, "tcp") {
				fields = append(fields, netSig)
			}
		}
	}

	return fields
}

// RiskAssessmentEngine is a risk assessment engine
type RiskAssessmentEngine struct {
	rules map[string]RiskRule
}

// RiskRule represents a risk rule
type RiskRule struct {
	Name        string
	Description string
	Severity    string // low, medium, high, critical
	CheckFunc   func(*TCPPacket) bool
}

// NewRiskAssessmentEngine creates a new risk assessment engine
func NewRiskAssessmentEngine() *RiskAssessmentEngine {
	return &RiskAssessmentEngine{
		rules: initializeRiskRules(),
	}
}

// initializeRiskRules initializes risk rules
func initializeRiskRules() map[string]RiskRule {
	return map[string]RiskRule{}
}

// initializeOSDatabase initializes the operating system database
func initializeOSDatabase() map[string]*OSProfile {
	return map[string]*OSProfile{
		"Windows11": {
			Name:              "Windows 11",
			Family:            "Windows",
			Version:           "11",
			DefaultTTL:        128,
			DefaultMSS:        1460,
			DefaultWindowSize: 65535,
			TCPOptions:        "MSS,SACK,TS,NOP,WS",
			Behavior:          "Standard Windows TCP behavior",
		},
		"Linux": {
			Name:              "Linux Kernel 5.x",
			Family:            "Linux",
			Version:           "5.x",
			DefaultTTL:        64,
			DefaultMSS:        1460,
			DefaultWindowSize: 29200,
			TCPOptions:        "MSS,TS,SACK,WS",
			Behavior:          "Standard Linux TCP behavior",
		},
		"macOS": {
			Name:              "macOS 13",
			Family:            "Darwin",
			Version:           "13",
			DefaultTTL:        64,
			DefaultMSS:        1460,
			DefaultWindowSize: 65535,
			TCPOptions:        "MSS,NOP,WS,NOP,NOP,TS",
			Behavior:          "Standard macOS TCP behavior",
		},
	}
}
