// Package tcpip 提供集成的 TCP/IP 指纹分析
// 结合 User-Agent、IP 地址和地理位置信息进行综合识别
package tcpip

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

// IntegratedFingerprinter 集成指纹分析器
type IntegratedFingerprinter struct {
	// 各层分析器
	 tcpipAnalyzer *TCPIPAnalyzer
	 
	 // 关联数据库
	osUAMapping     map[string][]string // OS -> 常见UA模式
	ipRegionDB      IPRegionDatabase    // IP地理位置数据库
	osTCPSignatures map[string]OSSignature // OS -> TCP签名特征
	 
	 // 一致性规则
	consistencyRules []ConsistencyRule
	 
	mu sync.RWMutex
}

// ConsistencyRule 一致性检查规则
type ConsistencyRule struct {
	Name        string
	Description string
	Check       func(ctx *FingerprintContext) *Inconsistency
}

// FingerprintContext 指纹分析上下文
type FingerprintContext struct {
	// TCP/IP 层信息
	TCPPacket *TCPPacket
	TCPSignature *TCPIPSignature
	 
	// 应用层信息
	UserAgent    string
	HTTPHeaders  map[string]string
	 
	// 网络层信息
	SourceIP     string
	DestIP       string
	 
	// 地理位置信息
	GeoLocation  *GeoLocation
	 
	// 推断结果
	DetectedOS   string
	DetectedDevice string
}

// GeoLocation 地理位置信息
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

// IPRegionDatabase IP区域数据库接口
type IPRegionDatabase interface {
	Lookup(ip string) (*GeoLocation, error)
	GetRegionSignature(country, isp string) string
}

// Inconsistency 不一致性报告
type Inconsistency struct {
	RuleName    string
	Severity    string // high, medium, low
	Description string
	Expected    string
	Actual      string
}

// IntegratedResult 集成分析结果
type IntegratedResult struct {
	// 各层独立识别结果
	TCPResult      *TCPIPResult
	ParsedOSFromUA string
	GeoInfo        *GeoLocation
	 
	// 交叉验证结果
	OSCrossValidation OSCrossValidationResult
	IPUAConsistency   bool
	GeoUAConsistency  bool
	 
	// 不一致性报告
	Inconsistencies []Inconsistency
	 
	// 综合评估
	OverallConfidence float64
	RiskScore         float64
	FinalOS           string
	FinalDeviceType   string
	 
	// 原始数据摘要
	SourceIP      string
	UserAgent     string
	TCPSignature  string
}

// OSCrossValidationResult OS交叉验证结果
type OSCrossValidationResult struct {
	OSFromTCP      string
	OSFromUA       string
	OSFromGeo      string
	ConsensusOS    string
	MatchScore     float64 // 0-1, 越高表示各层越一致
}

// NewIntegratedFingerprinter 创建集成指纹分析器
func NewIntegratedFingerprinter() *IntegratedFingerprinter {
	ia := &IntegratedFingerprinter{
		tcpipAnalyzer:     NewTCPIPAnalyzer(),
		osUAMapping:       buildOSUAMapping(),
		osTCPSignatures:   BuildOSDatabase(),
		consistencyRules:  buildConsistencyRules(),
	}
	return ia
}

// Analyze 执行集成指纹分析
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
	
	// 1. TCP/IP 层分析
	if packet != nil {
		ia.analyzeTCPLayer(ctx, result)
	}
	
	// 2. User-Agent 解析
	ia.parseUserAgent(ctx, result)
	
	// 3. IP 地理位置分析
	ia.analyzeIPGeolocation(ctx, result)
	
	// 4. 交叉验证
	ia.crossValidate(ctx, result)
	
	// 5. 一致性检查
	ia.checkConsistency(ctx, result)
	
	// 6. 综合评估
	ia.calculateOverallConfidence(result)
	
	return result, nil
}

// analyzeTCPLayer 分析 TCP/IP 层
func (ia *IntegratedFingerprinter) analyzeTCPLayer(ctx *FingerprintContext, result *IntegratedResult) {
	// 添加数据包到分析器
	ia.tcpipAnalyzer.AddPacket(ctx.TCPPacket)
	
	// 分析 TCP 特征
	tcpResult := ia.tcpipAnalyzer.AnalyzePacket(ctx.TCPPacket)
	result.TCPResult = tcpResult
	
	// 推断 OS
	if tcpResult != nil && tcpResult.OS != "" {
		ctx.DetectedOS = tcpResult.OS
		result.TCPSignature = fmt.Sprintf("TTL:%d,MSS:%d,Win:%d", 
			ctx.TCPPacket.IPHeader.TimeToLive,
			len(ctx.TCPPacket.Options),
			ctx.TCPPacket.WindowSize)
	}
}

// parseUserAgent 解析 User-Agent
func (ia *IntegratedFingerprinter) parseUserAgent(ctx *FingerprintContext, result *IntegratedResult) {
	ua := ctx.UserAgent
	if ua == "" {
		return
	}
	
	uaLower := strings.ToLower(ua)
	
	// 提取 OS 信息
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
	
	// 提取设备类型
	if strings.Contains(uaLower, "mobile") {
		ctx.DetectedDevice = "Mobile"
	} else if strings.Contains(uaLower, "tablet") || strings.Contains(uaLower, "ipad") {
		ctx.DetectedDevice = "Tablet"
	} else {
		ctx.DetectedDevice = "Desktop"
	}
}

// analyzeIPGeolocation 分析 IP 地理位置
func (ia *IntegratedFingerprinter) analyzeIPGeolocation(ctx *FingerprintContext, result *IntegratedResult) {
	if ctx.SourceIP == "" {
		return
	}
	
	// 解析 IP
	ip := net.ParseIP(ctx.SourceIP)
	if ip == nil {
		return
	}
	
	// 检查是否为私有地址
	if isPrivateIP(ip) {
		result.GeoInfo = &GeoLocation{
			Country: "Private",
			Region:  "Local Network",
		}
		return
	}
	
	// 如果有 IP 区域数据库，进行查询
	if ia.ipRegionDB != nil {
		geo, err := ia.ipRegionDB.Lookup(ctx.SourceIP)
		if err == nil {
			result.GeoInfo = geo
			ctx.GeoLocation = geo
		}
	}
}

// crossValidate 交叉验证各层结果
func (ia *IntegratedFingerprinter) crossValidate(ctx *FingerprintContext, result *IntegratedResult) {
	validation := OSCrossValidationResult{
		OSFromTCP: result.TCPResult.OS,
		OSFromUA:  result.ParsedOSFromUA,
	}
	
	// 从地理位置推断 OS（某些地区/ISP有特定偏好）
	if result.GeoInfo != nil {
		validation.OSFromGeo = ia.inferOSFromGeography(result.GeoInfo)
	}
	
	// 计算共识 OS
	osVotes := make(map[string]int)
	if validation.OSFromTCP != "" {
		osVotes[validation.OSFromTCP]++
	}
	if validation.OSFromUA != "" {
		osVotes[validation.OSFromUA] += 2 // UA 权重更高
	}
	if validation.OSFromGeo != "" {
		osVotes[validation.OSFromGeo]++
	}
	
	// 找出得票最高的 OS
	maxVotes := 0
	for os, votes := range osVotes {
		if votes > maxVotes {
			maxVotes = votes
			validation.ConsensusOS = os
		}
	}
	
	// 计算匹配分数
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
	
	// IP-UA 一致性：检查 UA 中的语言/地区是否与 IP 地理位置匹配
	result.IPUAConsistency = ia.checkIPUAConsistency(ctx, result)
	
	// 地理位置-UA 一致性
	result.GeoUAConsistency = ia.checkGeoUAConsistency(ctx, result)
}

// checkConsistency 执行一致性检查规则
func (ia *IntegratedFingerprinter) checkConsistency(ctx *FingerprintContext, result *IntegratedResult) {
	for _, rule := range ia.consistencyRules {
		if inconsistency := rule.Check(ctx); inconsistency != nil {
			result.Inconsistencies = append(result.Inconsistencies, *inconsistency)
		}
	}
}

// calculateOverallConfidence 计算综合置信度
func (ia *IntegratedFingerprinter) calculateOverallConfidence(result *IntegratedResult) {
	confidence := 0.5 // 基础置信度
	
	// TCP 层贡献 (0.3)
	if result.TCPResult != nil && result.TCPResult.Confidence > 0 {
		confidence += result.TCPResult.Confidence * 0.3
	}
	
	// UA 层贡献 (0.3)
	if result.ParsedOSFromUA != "" {
		confidence += 0.3
	}
	
	// 交叉验证一致性贡献 (0.4)
	confidence += result.OSCrossValidation.MatchScore * 0.4
	
	// 如果不一致性严重，降低置信度
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
	
	// 限制在 0-1 范围
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}
	
	result.OverallConfidence = confidence
	
	// 计算风险分数
	result.RiskScore = ia.calculateRiskScore(result)
}

// calculateRiskScore 计算风险分数
func (ia *IntegratedFingerprinter) calculateRiskScore(result *IntegratedResult) float64 {
	risk := 0.0
	
	// 基于不一致性计算风险
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
	
	// 低置信度增加风险
	if result.OverallConfidence < 0.3 {
		risk += 0.3
	} else if result.OverallConfidence < 0.5 {
		risk += 0.15
	}
	
	// IP-UA 不一致增加风险
	if !result.IPUAConsistency {
		risk += 0.2
	}
	
	if risk > 1 {
		risk = 1
	}
	
	return risk
}

// Helper 方法

func (ia *IntegratedFingerprinter) inferOSFromGeography(geo *GeoLocation) string {
	// 基于地理位置推断 OS（某些地区有特定偏好）
	// 例如：某些企业网络可能统一使用 Windows
	return ""
}

func (ia *IntegratedFingerprinter) checkIPUAConsistency(ctx *FingerprintContext, result *IntegratedResult) bool {
	// 检查 UA 中的语言设置是否与 IP 地理位置匹配
	if ctx.GeoLocation == nil || ctx.HTTPHeaders == nil {
		return true // 无法检查，默认为一致
	}
	
	acceptLang := ctx.HTTPHeaders["Accept-Language"]
	if acceptLang == "" {
		return true
	}
	
	// 简单检查：如果 IP 在中国但 UA 语言只有 en-US，可能不一致
	// 实际应用需要更复杂的语言-地区映射
	return true
}

func (ia *IntegratedFingerprinter) checkGeoUAConsistency(ctx *FingerprintContext, result *IntegratedResult) bool {
	// 检查 UA 中的时区/地区信息是否与地理位置匹配
	return true
}

func isPrivateIP(ip net.IP) bool {
	// 检查是否为私有地址
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

// buildOSUAMapping 构建 OS 到 UA 的映射
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

// buildConsistencyRules 构建一致性检查规则
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
				
				// 从 UA 推断 OS
				uaLower := strings.ToLower(ctx.UserAgent)
				switch {
				case strings.Contains(uaLower, "windows"):
					uaOS = "Windows"
				case strings.Contains(uaLower, "mac"):
					uaOS = "macOS"
				case strings.Contains(uaLower, "linux"):
					uaOS = "Linux"
				}
				
				// Windows 通常使用 TTL 128，Linux/macOS 使用 64
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
				
				// 移动设备通常有较小的窗口大小
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

// SetIPRegionDB 设置 IP 区域数据库
func (ia *IntegratedFingerprinter) SetIPRegionDB(db IPRegionDatabase) {
	ia.mu.Lock()
	defer ia.mu.Unlock()
	ia.ipRegionDB = db
}
