package ja4s

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// JA4SResult JA4S 指纹结果
type JA4SResult struct {
	// 完整 JA4S SHA256 哈希
	Hash string

	// JA4S_a: TLS版本+密码套件+扩展特征
	JA4Sa string

	// JA4S_r: 原始顺序版本（未排序的扩展列表）
	JA4Sr string

	// 完整签名字符串（用于调试）
	RawString string

	// TLS 版本字符串（如 "1.2", "1.3"）
	TLSVersion string

	// 异常判定分数 (0.0-1.0)
	RiskScore float64

	// 异常标记列表
	AnomalyFlags []string

	// 匹配的已知服务端指纹
	MatchedProfiles []string
}

// JA4SAnalyzer JA4S 分析器
type JA4SAnalyzer struct {
	// 已知的服务端指纹库（可后续扩展）
	knownProfiles map[string]*ServerProfileInfo
}

// ServerProfileInfo 已知的服务端配置信息
type ServerProfileInfo struct {
	Name        string   // 服务端名称
	TLSVersions []string // 支持的 TLS 版本
	Ciphers     []string // 常用密码套件
	Extensions  []string // 常用扩展
	RiskScore   float64  // 基线风险分数
}

// NewJA4SAnalyzer 创建 JA4S 分析器
func NewJA4SAnalyzer() *JA4SAnalyzer {
	return &JA4SAnalyzer{
		knownProfiles: initKnownServerProfiles(),
	}
}

// ServerHelloData 导出的 ServerHello 数据结构，用于从结构化数据计算 JA4S
type ServerHelloData struct {
	TLSVersion     uint16   // TLS 版本（如 0x0303=TLS1.2, 0x0304=TLS1.3）
	CipherSuite    uint16   // 选择的密码套件
	Extensions     []uint16 // 扩展列表
	Compression    uint8    // 压缩方法
	ServerName     string   // 服务器名称
	SelectedALPN   string   // 选择的 ALPN 协议
}

// AnalyzeServerHello 从结构化 ServerHello 数据分析指纹
func (a *JA4SAnalyzer) AnalyzeServerHello(data ServerHelloData) (*JA4SResult, error) {
	sh := &serverHelloData{
		Version:           data.TLSVersion,
		CipherSuite:       data.CipherSuite,
		CompressionMethod: data.Compression,
		Extensions:        data.Extensions,
	}

	result := &JA4SResult{
		RawString:    sh.String(),
		TLSVersion:   tlsVersionString(sh.Version),
		AnomalyFlags: make([]string, 0, 8),
	}

	tlsVersionCode := formatTLSVersion(sh.Version)
	cipherCode := formatCipherCode(sh.CipherSuite)
	extensionCount := fmt.Sprintf("%d", len(sh.Extensions))
	compressionCode := formatCompressionCode(sh.CompressionMethod)

	result.JA4Sa = fmt.Sprintf("%s,%s,%s,%s",
		tlsVersionCode,
		cipherCode,
		extensionCount,
		compressionCode,
	)

	var extensionCodes []string
	for _, ext := range sh.Extensions {
		extensionCodes = append(extensionCodes, fmt.Sprintf("%d", ext))
	}
	result.JA4Sr = fmt.Sprintf("%s,%s",
		result.JA4Sa,
		strings.Join(extensionCodes, ","),
	)

	hash := sha256.Sum256([]byte(result.JA4Sr))
	result.Hash = hex.EncodeToString(hash[:])

	a.detectAnomalies(result, sh)

	return result, nil
}

// AnalyzeServerHelloBytes 分析 TLS ServerHello 数据包
// serverHelloBytes: 完整的 ServerHello 字节数据
func (a *JA4SAnalyzer) AnalyzeServerHelloBytes(serverHelloBytes []byte) (*JA4SResult, error) {
	if len(serverHelloBytes) < 43 {
		return nil, fmt.Errorf("ServerHello too short: %d bytes", len(serverHelloBytes))
	}

	// 解析 ServerHello 结构
	sh, err := parseServerHello(serverHelloBytes)
	if err != nil {
		return nil, err
	}

	// 构建签名字符串
	result := &JA4SResult{
		RawString:    sh.String(),
		TLSVersion:   tlsVersionString(sh.Version),
		AnomalyFlags: make([]string, 0, 8), // 预分配容量
	}

	// 生成 JA4S_a: TLS版本,密码套件,扩展计数,压缩方法
	tlsVersionCode := formatTLSVersion(sh.Version)
	cipherCode := formatCipherCode(sh.CipherSuite)
	extensionCount := fmt.Sprintf("%d", len(sh.Extensions))
	compressionCode := formatCompressionCode(sh.CompressionMethod)

	result.JA4Sa = fmt.Sprintf("%s,%s,%s,%s",
		tlsVersionCode,
		cipherCode,
		extensionCount,
		compressionCode,
	)

	// 生成 JA4S_r: 原始扩展列表（不排序，保持顺序）
	var extensionCodes []string
	for _, ext := range sh.Extensions {
		extensionCodes = append(extensionCodes, fmt.Sprintf("%d", ext))
	}
	result.JA4Sr = fmt.Sprintf("%s,%s",
		result.JA4Sa,
		strings.Join(extensionCodes, ","),
	)

	// 计算 SHA256 哈希
	hash := sha256.Sum256([]byte(result.JA4Sr))
	result.Hash = hex.EncodeToString(hash[:])

	// 异常检测与评分
	a.detectAnomalies(result, sh)

	return result, nil
}

// AnalyzeServerHelloProfile 从指纹 Profile 分析（对于客户端模拟）
// 这用于生成与真实服务器高度一致的虚拟 ServerHello
func (a *JA4SAnalyzer) GenerateServerHelloSignature(
	tlsVersion uint16,
	cipherSuite uint16,
	extensions []uint16,
	compressionMethod uint8,
) (*JA4SResult, error) {

	// 构建虚拟 ServerHello 结构
	sh := &serverHelloData{
		Version:           tlsVersion,
		CipherSuite:       cipherSuite,
		CompressionMethod: compressionMethod,
		Extensions:        extensions,
	}

	result := &JA4SResult{
		RawString:    sh.String(),
		TLSVersion:   tlsVersionString(sh.Version),
		AnomalyFlags: make([]string, 0, 8), // 预分配容量
	}
	tlsVersionCode := formatTLSVersion(sh.Version)
	cipherCode := formatCipherCode(sh.CipherSuite)
	extensionCount := fmt.Sprintf("%d", len(sh.Extensions))
	compressionCode := formatCompressionCode(sh.CompressionMethod)

	result.JA4Sa = fmt.Sprintf("%s,%s,%s,%s",
		tlsVersionCode,
		cipherCode,
		extensionCount,
		compressionCode,
	)

	var extensionCodes []string
	for _, ext := range sh.Extensions {
		extensionCodes = append(extensionCodes, fmt.Sprintf("%d", ext))
	}
	result.JA4Sr = fmt.Sprintf("%s,%s",
		result.JA4Sa,
		strings.Join(extensionCodes, ","),
	)

	hash := sha256.Sum256([]byte(result.JA4Sr))
	result.Hash = hex.EncodeToString(hash[:])

	a.detectAnomalies(result, sh)

	return result, nil
}

// detectAnomalies 检测异常特征
func (a *JA4SAnalyzer) detectAnomalies(result *JA4SResult, sh *serverHelloData) {
	baseScore := 0.0

	// 异常检测 1: TLS 版本检查
	if isDeprecatedTLSVersion(sh.Version) {
		// TLS 1.0/1.1/SSL 3.0 已弃用（RFC 8996），存在安全风险
		result.AnomalyFlags = append(result.AnomalyFlags, "DEPRECATED_TLS_VERSION")
		baseScore += 0.2
	} else if !isSupportedTLSVersion(sh.Version) {
		result.AnomalyFlags = append(result.AnomalyFlags, "UNSUPPORTED_TLS_VERSION")
		baseScore += 0.3
	}

	// 异常检测 2: 已知弱密码套件
	if isWeakCipherSuite(sh.CipherSuite) {
		result.AnomalyFlags = append(result.AnomalyFlags, "WEAK_CIPHER_SUITE")
		baseScore += 0.25
	}

	// 异常检测 3: 异常的扩展组合
	if len(sh.Extensions) < 3 {
		result.AnomalyFlags = append(result.AnomalyFlags, "MINIMAL_EXTENSIONS")
		baseScore += 0.2
	}
	if len(sh.Extensions) > 30 {
		result.AnomalyFlags = append(result.AnomalyFlags, "EXCESSIVE_EXTENSIONS")
		baseScore += 0.15
	}

	// 异常检测 4: 扩展列表异常（检测重复扩展）
	if !hasValidExtensionOrder(sh.Extensions) {
		result.AnomalyFlags = append(result.AnomalyFlags, "DUPLICATE_EXTENSIONS")
		baseScore += 0.2
	}

	// 异常检测 5: 压缩方法异常（TLS 压缩存在 CRIME 攻击风险，仅 null=0 是安全的）
	if sh.CompressionMethod != 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "UNSAFE_COMPRESSION")
		baseScore += 0.2
	}

	// 标准化评分
	if baseScore > 1.0 {
		baseScore = 1.0
	}
	result.RiskScore = baseScore
}

// FindMatchingProfiles 查找匹配的已知服务端配置
func (a *JA4SAnalyzer) FindMatchingProfiles(result *JA4SResult, maxResults int) []string {
	// 简单的 hash 匹配（可后续扩展为相似度匹配）
	var matches []string

	for name, profile := range a.knownProfiles {
		// 计算相似度（当前为简单匹配，可改进）
		if profile.RiskScore < result.RiskScore-0.1 {
			matches = append(matches, name)
		}

		if len(matches) >= maxResults {
			break
		}
	}

	result.MatchedProfiles = matches
	return matches
}

// ============ 辅助函数 ============

type serverHelloData struct {
	Version           uint16
	CipherSuite       uint16
	CompressionMethod uint8
	Extensions        []uint16
}

func (sh *serverHelloData) String() string {
	var extStr []string
	for _, e := range sh.Extensions {
		extStr = append(extStr, fmt.Sprintf("%d", e))
	}
	return fmt.Sprintf("TLS%x,Cipher%x,Comp%d,Ext[%s]",
		sh.Version,
		sh.CipherSuite,
		sh.CompressionMethod,
		strings.Join(extStr, ","),
	)
}

// parseServerHello 解析 ServerHello 字节数据
func parseServerHello(data []byte) (*serverHelloData, error) {
	if len(data) < 43 {
		return nil, fmt.Errorf("data too short")
	}

	sh := &serverHelloData{}

	// 偏移量：HandshakeType(1) + Length(3) = 4 字节头部
	// Version(2) 在偏移 4-5
	sh.Version = uint16(data[4])<<8 | uint16(data[5])

	// Random(32) 在偏移 6-37
	// Session ID Length(1) 在偏移 38
	sessionIDLen := int(data[38])
	offset := 39 + sessionIDLen

	// Cipher Suite(2) 紧跟在 Session ID 之后
	if len(data) < offset+3 {
		return nil, fmt.Errorf("data too short for cipher suite: need %d, have %d", offset+3, len(data))
	}
	sh.CipherSuite = uint16(data[offset])<<8 | uint16(data[offset+1])
	offset += 2

	// Compression Method(1)
	sh.CompressionMethod = data[offset]
	offset++

	// 解析扩展列表
	if len(data) > offset+2 {
		extensionsLen := int(data[offset])<<8 | int(data[offset+1])
		offset += 2

		endOffset := offset + extensionsLen
		if endOffset > len(data) {
			endOffset = len(data)
		}

		for offset+4 <= endOffset {
			extType := uint16(data[offset])<<8 | uint16(data[offset+1])
			extLen := int(data[offset+2])<<8 | int(data[offset+3])
			if offset+4+extLen > endOffset {
				break // 扩展数据被截断，停止解析
			}
			sh.Extensions = append(sh.Extensions, extType)
			offset += 4 + extLen
		}
	}

	return sh, nil
}

func formatTLSVersion(v uint16) string {
	switch v {
	case 0x0303:
		return "773" // TLS 1.2
	case 0x0304:
		return "774" // TLS 1.3
	default:
		return fmt.Sprintf("%d", v)
	}
}

func tlsVersionString(v uint16) string {
	switch v {
	case 0x0301:
		return "1.0"
	case 0x0302:
		return "1.1"
	case 0x0303:
		return "1.2"
	case 0x0304:
		return "1.3"
	default:
		return fmt.Sprintf("0x%04x", v)
	}
}

func formatCipherCode(cipher uint16) string {
	// 返回简化的密码套件代码
	// 可扩展为完整的映射表
	switch cipher {
	case 0x002f:
		return "1" // TLS_RSA_WITH_AES_128_CBC_SHA
	case 0x007c:
		return "2" // TLS_RSA_WITH_AES_256_CBC_SHA
	case 0x1301:
		return "3" // TLS_AES_128_GCM_SHA256
	case 0x1302:
		return "4" // TLS_AES_256_GCM_SHA384
	default:
		// 返回原始值作为后备方案
		return fmt.Sprintf("%d", cipher)
	}
}

func formatCompressionCode(c uint8) string {
	if c == 0 {
		return "0" // null compression
	}
	return fmt.Sprintf("%d", c)
}

func isSupportedTLSVersion(v uint16) bool {
	supportedVersions := map[uint16]bool{
		0x0303: true, // TLS 1.2
		0x0304: true, // TLS 1.3
	}
	return supportedVersions[v]
}

// isDeprecatedTLSVersion 检查是否为已弃用的 TLS 版本（RFC 8996）
func isDeprecatedTLSVersion(v uint16) bool {
	deprecatedVersions := map[uint16]bool{
		0x0300: true, // SSL 3.0
		0x0301: true, // TLS 1.0
		0x0302: true, // TLS 1.1
	}
	return deprecatedVersions[v]
}

func isWeakCipherSuite(cipher uint16) bool {
	// 已知弱密码的列表（RFC 7457, CRIME, BEAST, SWEET32 等攻击向量）
	weakCiphers := map[uint16]bool{
		0x0000: true, // TLS_NULL_WITH_NULL_NULL
		0x0001: true, // TLS_RSA_WITH_NULL_MD5
		0x0002: true, // TLS_RSA_WITH_NULL_SHA
		0x0003: true, // TLS_RSA_EXPORT_WITH_RC4_40_MD5
		0x0004: true, // TLS_RSA_WITH_RC4_128_MD5
		0x0005: true, // TLS_RSA_WITH_RC4_128_SHA
		0x0006: true, // TLS_RSA_EXPORT_WITH_RC2_CBC_40_MD5
		0x0008: true, // TLS_RSA_EXPORT_WITH_DES40_CBC_SHA
		0x0009: true, // TLS_RSA_WITH_DES_CBC_SHA
		0x000A: true, // TLS_RSA_WITH_3DES_EDE_CBC_SHA (SWEET32)
		0x0011: true, // TLS_DHE_DSS_EXPORT_WITH_DES40_CBC_SHA
		0x0012: true, // TLS_DHE_DSS_WITH_DES_CBC_SHA
		0x0014: true, // TLS_DHE_RSA_EXPORT_WITH_DES40_CBC_SHA
		0x0015: true, // TLS_DHE_RSA_WITH_DES_CBC_SHA
		0x0017: true, // TLS_DH_anon_EXPORT_WITH_RC4_40_MD5
		0x0018: true, // TLS_DH_anon_WITH_RC4_128_MD5
		0x0019: true, // TLS_DH_anon_EXPORT_WITH_DES40_CBC_SHA
		0x001A: true, // TLS_DH_anon_WITH_DES_CBC_SHA
		0x003B: true, // TLS_RSA_WITH_NULL_SHA256
		0xC00A: true, // TLS_ECDHE_ECDSA_WITH_3DES_EDE_CBC_SHA (SWEET32)
		0xC014: true, // TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA (SWEET32)
		0xC011: true, // TLS_ECDHE_RSA_WITH_RC4_128_SHA
		0xC007: true, // TLS_ECDHE_ECDSA_WITH_RC4_128_SHA
	}
	return weakCiphers[cipher]
}

func hasValidExtensionOrder(extensions []uint16) bool {
	// ServerHello 扩展没有规范要求的排列顺序，
	// 因此不应基于排序来判定异常。
	// 仅检测重复扩展（重复扩展是不合法的）。
	if len(extensions) < 2 {
		return true
	}

	seen := make(map[uint16]bool, len(extensions))
	for _, ext := range extensions {
		if seen[ext] {
			return false // 重复扩展是异常的
		}
		seen[ext] = true
	}
	return true
}

// initKnownServerProfiles 初始化已知的服务端配置库
func initKnownServerProfiles() map[string]*ServerProfileInfo {
	return map[string]*ServerProfileInfo{
		"nginx_default": {
			Name:        "Nginx (Default)",
			TLSVersions: []string{"TLS 1.2", "TLS 1.3"},
			Ciphers:     []string{"AES_128_GCM", "AES_256_GCM", "CHACHA20"},
			Extensions:  []string{"0", "10", "11", "16", "23", "35"},
			RiskScore:   0.1,
		},
		"apache_default": {
			Name:        "Apache (Default)",
			TLSVersions: []string{"TLS 1.2", "TLS 1.3"},
			Ciphers:     []string{"AES_128_GCM", "AES_256_GCM"},
			Extensions:  []string{"0", "10", "11", "16", "23", "35"},
			RiskScore:   0.15,
		},
		"cloudflare": {
			Name:        "Cloudflare",
			TLSVersions: []string{"TLS 1.3"},
			Ciphers:     []string{"AES_256_GCM", "CHACHA20"},
			Extensions:  []string{"0", "10", "11", "16", "23", "35", "43"},
			RiskScore:   0.05,
		},
	}
}

// ComputeJA4S 便捷函数：从 ServerHelloData 结构计算 JA4S
func ComputeJA4S(data ServerHelloData) (*JA4SResult, error) {
	analyzer := NewJA4SAnalyzer()
	return analyzer.AnalyzeServerHello(data)
}

// ComputeJA4SFromBytes 便捷函数：直接从字节数据计算 JA4S
func ComputeJA4SFromBytes(serverHelloBytes []byte) (*JA4SResult, error) {
	analyzer := NewJA4SAnalyzer()
	return analyzer.AnalyzeServerHelloBytes(serverHelloBytes)
}

// MatchJA4S 比较两个 JA4S 哈希值是否匹配
func MatchJA4S(hash1, hash2 string) bool {
	return len(hash1) == 64 && len(hash2) == 64 && hash1 == hash2
}

// ComputeJA4SFromProfileData 从 Profile 数据计算 JA4S（用于客户端模拟）
func ComputeJA4SFromProfileData(
	tlsVersion uint16,
	cipherSuite uint16,
	extensions []uint16,
) (*JA4SResult, error) {
	analyzer := NewJA4SAnalyzer()
	return analyzer.GenerateServerHelloSignature(tlsVersion, cipherSuite, extensions, 0)
}
