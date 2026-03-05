package tcpip

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DeviceFingerprintingEngine 设备指纹识别引擎
type DeviceFingerprintingEngine struct {
	deviceProfiles     map[string]*DeviceProfile
	behaviorAnalyzer   *NetworkBehaviorAnalyzer
	riskAssessment     *RiskAssessmentEngine
	osDatabase         map[string]*OSProfile
	lastAnalysisResult *DeviceFingerprintResult
}

// DeviceProfile 设备轮廓
type DeviceProfile struct {
	Name              string
	DeviceType        string // phone, tablet, laptop, desktop, iot, router, server
	Manufacturer      string
	OS                string
	OSVersion         string
	BrowserSignatures map[string]string // browser name -> fingerprint
	NetworkSignatures map[string]string // signal -> value
	Applications      []string          // 常见应用
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Confidence        float64
}

// OSProfile 操作系统轮廓
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

// DeviceFingerprintingConfig 设备指纹识别配置
type DeviceFingerprintingConfig struct {
	EnableBehaviorAnalysis bool
	EnableRiskAssessment   bool
	EnableGeoIP            bool
	SamplingRate           float64 // 0.0-1.0
	CacheTTL               time.Duration
	MaxPacketsPerAnalysis  int
}

// NewDeviceFingerprintingEngine 创建新的设备指纹识别引擎
func NewDeviceFingerprintingEngine() *DeviceFingerprintingEngine {
	return &DeviceFingerprintingEngine{
		deviceProfiles:   make(map[string]*DeviceProfile),
		behaviorAnalyzer: NewNetworkBehaviorAnalyzer(),
		riskAssessment:   NewRiskAssessmentEngine(),
		osDatabase:       initializeOSDatabase(),
	}
}

// RegisterDeviceProfile 注册设备轮廓
func (dfe *DeviceFingerprintingEngine) RegisterDeviceProfile(profile *DeviceProfile) error {
	if profile.Name == "" {
		return fmt.Errorf("device profile name is required")
	}

	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()
	dfe.deviceProfiles[profile.Name] = profile

	return nil
}

// AnalyzeDevice 分析设备
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

	// 分析数据包特征
	features := dfe.extractFeatures(packets)

	// 匹配设备轮廓
	result.DeviceMatches = dfe.matchDeviceProfiles(features)

	// 匹配操作系统
	result.OSMatches = dfe.matchOSProfiles(packets[0])

	// 进行风险评估
	result.RiskIndicators = dfe.assessRisk(packets, behaviors)

	// 计算总体信心度
	result.Confidence = dfe.calculateConfidence(result)

	// 添加行为分析
	if behaviors != nil {
		result.BehaviorPattern = behaviors.SequenceNumberPattern
		result.NetworkType = behaviors.RTTAnalysis.NetworkType
	}

	dfe.lastAnalysisResult = result
	return result
}

// extractFeatures 提取特征
func (dfe *DeviceFingerprintingEngine) extractFeatures(packets []*TCPPacket) map[string]interface{} {
	features := make(map[string]interface{})

	if len(packets) == 0 {
		return features
	}

	firstPacket := packets[0]

	// 提取 IP 特征
	if firstPacket.IPHeader != nil {
		features["ttl"] = firstPacket.IPHeader.TimeToLive
		features["df"] = (firstPacket.IPHeader.Flags & 0x40) != 0
		features["protocol"] = firstPacket.IPHeader.Protocol
	}

	// 提取 TCP 特征
	features["src_port"] = firstPacket.SourcePort
	features["dst_port"] = firstPacket.DestinationPort
	features["flags"] = firstPacket.Flags
	features["window_size"] = firstPacket.WindowSize

	// 提取 TCP 选项
	if len(firstPacket.Options) > 0 {
		optStrings := make([]string, 0)
		for _, opt := range firstPacket.Options {
			optStrings = append(optStrings, fmt.Sprintf("opt_%d", opt))
		}
		features["tcp_options"] = strings.Join(optStrings, ",")
	}

	// MSS 特征
	mss := 0
	for _, opt := range firstPacket.Options {
		if opt == OptMSS {
			mss = 1460 // 标准 MSS
		}
	}
	if mss > 0 {
		features["mss"] = mss
	}

	// 数据包数量统计
	features["packet_count"] = len(packets)

	// 计算会话时间
	if len(packets) > 1 {
		sessionDuration := packets[len(packets)-1].Timestamp.Sub(packets[0].Timestamp)
		features["session_duration"] = sessionDuration
	}

	return features
}

// matchDeviceProfiles 匹配设备轮廓
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

	// 按分数排序
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].MatchScore > matches[i].MatchScore {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	return matches
}

// matchOSProfiles 匹配操作系统轮廓
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

// assessRisk 评估风险
func (dfe *DeviceFingerprintingEngine) assessRisk(packets []*TCPPacket, behaviors *NetworkBehaviorResult) []string {
	riskIndicators := make([]string, 0)

	// 检查异常
	for _, packet := range packets {
		if packet.IPHeader != nil {
			// 检查非标准 TTL
			if !dfe.isStandardTTL(packet.IPHeader.TimeToLive) {
				riskIndicators = append(riskIndicators, "non_standard_ttl")
				break
			}

			// 检查可疑 IP
			if IsReservedIP(packet.IPHeader.SourceAddress) {
				riskIndicators = append(riskIndicators, "reserved_ip")
				break
			}
		}

		// 检查异常 TCP 选项
		if len(packet.Options) == 0 {
			riskIndicators = append(riskIndicators, "missing_tcp_options")
			break
		}
	}

	// 基于行为的风险评估
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

// calculateProfileScore 计算轮廓匹配分数
func (dfe *DeviceFingerprintingEngine) calculateProfileScore(features map[string]interface{}, profile *DeviceProfile) float64 {
	score := 0.0
	matches := 0
	total := 0

	// 检查 TTL
	if ttl, ok := features["ttl"]; ok {
		total++
		// 设备轮廓通常关联特定的 OS，而 OS 有特定的 TTL
		if ttlInt, ok := ttl.(int); ok && ttlInt == 64 {
			matches++
		}
	}

	// 检查 TCP 选项
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

// calculateOSScore 计算操作系统匹配分数
func (dfe *DeviceFingerprintingEngine) calculateOSScore(packet *TCPPacket, osProfile *OSProfile) float64 {
	matches := 0
	total := 3

	// 检查 TTL
	if packet.IPHeader.TimeToLive == uint8(osProfile.DefaultTTL) {
		matches++
	}

	// 检查窗口大小
	if packet.WindowSize == uint16(osProfile.DefaultWindowSize) {
		matches++
	}

	// 检查 DF 标志
	if (packet.IPHeader.Flags & 0x40) != 0 {
		matches++
	}

	return float64(matches) / float64(total)
}

// calculateConfidence 计算总体信心度
func (dfe *DeviceFingerprintingEngine) calculateConfidence(result *DeviceFingerprintResult) float64 {
	if len(result.DeviceMatches) == 0 {
		return 0
	}

	avgScore := 0.0
	for _, match := range result.DeviceMatches {
		avgScore += match.MatchScore
	}

	avgScore /= float64(len(result.DeviceMatches))

	// 根据风险调整信心度
	riskPenalty := float64(len(result.RiskIndicators)) * 0.1
	confidence := avgScore - riskPenalty

	if confidence < 0 {
		confidence = 0
	}

	return confidence
}

// isStandardTTL 检查是否为标准 TTL
func (dfe *DeviceFingerprintingEngine) isStandardTTL(ttl uint8) bool {
	standardTTLs := []uint8{32, 64, 128, 255}
	for _, std := range standardTTLs {
		if ttl == std {
			return true
		}
	}
	return false
}

// DeviceFingerprintResult 设备指纹识别结果
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

// DeviceMatch 设备匹配结果
type DeviceMatch struct {
	DeviceName    string
	DeviceType    string
	Manufacturer  string
	MatchScore    float64
	MatchedFields []string
}

// OSMatch 操作系统匹配结果
type OSMatch struct {
	OSName     string
	OSFamily   string
	MatchScore float64
	DefaultTTL int
	DefaultMSS int
	TCPOptions string
}

// String 返回易读的指纹识别结果
func (dfr *DeviceFingerprintResult) String() string {
	data, _ := json.MarshalIndent(dfr, "", "  ")
	return string(data)
}

// extractMatchedFields 提取匹配的字段
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

// RiskAssessmentEngine 风险评估引擎
type RiskAssessmentEngine struct {
	rules map[string]RiskRule
}

// RiskRule 风险规则
type RiskRule struct {
	Name        string
	Description string
	Severity    string // low, medium, high, critical
	CheckFunc   func(*TCPPacket) bool
}

// NewRiskAssessmentEngine 创建新的风险评估引擎
func NewRiskAssessmentEngine() *RiskAssessmentEngine {
	return &RiskAssessmentEngine{
		rules: initializeRiskRules(),
	}
}

// initializeRiskRules 初始化风险规则
func initializeRiskRules() map[string]RiskRule {
	return map[string]RiskRule{}
}

// initializeOSDatabase 初始化操作系统数据库
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
