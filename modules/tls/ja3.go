// Package tls 提供 TLS 指纹生成功能
package tls

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// JA3Result JA3 指纹结果
type JA3Result struct {
	// Hash JA3 MD5 哈希
	Hash string
	// RawString JA3 原始字符串
	RawString string
	// TLSVersion TLS 版本
	TLSVersion uint16
	// CipherSuites 密码套件列表
	CipherSuites []uint16
	// Extensions 扩展列表
	Extensions []uint16
	// EllipticCurves 椭圆曲线列表
	EllipticCurves []core.CurveID
	// EllipticCurvePointFormats 椭圆曲线点格式列表
	EllipticCurvePointFormats []uint8
}

// CalculateJA3 从 ClientHello 规范计算 JA3 指纹
func CalculateJA3(spec core.ClientHelloSpec) *JA3Result {
	// 过滤 GREASE 值
	cipherSuites := filterGREASEUint16(spec.CipherSuites)
	extensions := filterGREASEExtensions(spec.Extensions)
	curves := filterGREASECurves(spec.SupportedCurves)

	// 构建 JA3 字符串
	parts := []string{
		strconv.Itoa(int(spec.TLSVersion)),
		joinUint16(cipherSuites),
		joinExtensions(extensions),
		joinCurves(curves),
		joinUint8(spec.SupportedPoints),
	}

	rawString := strings.Join(parts, ",")

	// 计算 MD5 哈希
	hash := md5.Sum([]byte(rawString))

	return &JA3Result{
		Hash:                      hex.EncodeToString(hash[:]),
		RawString:                 rawString,
		TLSVersion:                spec.TLSVersion,
		CipherSuites:              cipherSuites,
		Extensions:                extensionTypes(extensions),
		EllipticCurves:            curves,
		EllipticCurvePointFormats: spec.SupportedPoints,
	}
}

// CalculateJA3FromProfile 从客户端配置计算 JA3 指纹
func CalculateJA3FromProfile(profile profiles.ClientProfile) *JA3Result {
	spec := core.ClientHelloSpec{
		TLSVersion:      profile.TLSVersion,
		CipherSuites:    profile.CipherSuites,
		Extensions:      profile.Extensions,
		SupportedCurves: profile.SupportedCurves,
		SupportedPoints: profile.SupportedPoints,
	}
	return CalculateJA3(spec)
}

// IsGREASEUint16 检查是否为 GREASE 值
func IsGREASEUint16(v uint16) bool {
	// GREASE 值模式: 0x0A0A, 0x1A1A, 0x2A2A, ..., 0xFAFA
	return (v&0x0F0F) == 0x0A0A && ((v>>4)&0xFF) == ((v>>12)&0xFF)
}

// filterGREASEUint16 过滤 GREASE 值
func filterGREASEUint16(values []uint16) []uint16 {
	var result []uint16
	for _, v := range values {
		if !IsGREASEUint16(v) {
			result = append(result, v)
		}
	}
	return result
}

// filterGREASEExtensions 过滤 GREASE 扩展
func filterGREASEExtensions(exts []core.TLSExtension) []core.TLSExtension {
	var result []core.TLSExtension
	for _, e := range exts {
		if !IsGREASEUint16(e.Type) {
			result = append(result, e)
		}
	}
	return result
}

// filterGREASECurves 过滤 GREASE 曲线
func filterGREASECurves(curves []core.CurveID) []core.CurveID {
	var result []core.CurveID
	for _, c := range curves {
		if !IsGREASEUint16(uint16(c)) {
			result = append(result, c)
		}
	}
	return result
}

// joinUint16 将 uint16 切片连接为字符串
func joinUint16(values []uint16) string {
	var parts []string
	for _, v := range values {
		parts = append(parts, strconv.Itoa(int(v)))
	}
	return strings.Join(parts, "-")
}

// joinUint8 将 uint8 切片连接为字符串
func joinUint8(values []uint8) string {
	var parts []string
	for _, v := range values {
		parts = append(parts, strconv.Itoa(int(v)))
	}
	return strings.Join(parts, "-")
}

// joinExtensions 将扩展连接为字符串
func joinExtensions(exts []core.TLSExtension) string {
	var parts []string
	for _, e := range exts {
		parts = append(parts, strconv.Itoa(int(e.Type)))
	}
	return strings.Join(parts, "-")
}

// extensionTypes 提取扩展类型
func extensionTypes(exts []core.TLSExtension) []uint16 {
	var result []uint16
	for _, e := range exts {
		result = append(result, e.Type)
	}
	return result
}

// JA4Result JA4 指纹结果
type JA4Result struct {
	// Fingerprint JA4 指纹
	Fingerprint string
	// TLSVersion TLS 版本
	TLSVersion uint16
	// CipherSuitesCount 密码套件数量
	CipherSuitesCount int
	// ExtensionsCount 扩展数量
	ExtensionsCount int
}

// CalculateJA4 从 ClientHello 规范计算 JA4 指纹（简化版）
func CalculateJA4(spec core.ClientHelloSpec) *JA4Result {
	// JA4 格式: t<version><sni><cipher_count><extension_count><algo>
	// 例如: t13d1516h2_8daaf6152771_b1ff...a5

	// 简化实现
	version := core.TLSVersionToString(spec.TLSVersion)
	cipherCount := len(filterGREASEUint16(spec.CipherSuites))
	extCount := len(filterGREASEExtensions(spec.Extensions))

	// 构建简化的 JA4 指纹
	fingerprint := "t" + version + "d" +
		strconv.Itoa(cipherCount) +
		strconv.Itoa(extCount)

	return &JA4Result{
		Fingerprint:       fingerprint,
		TLSVersion:        spec.TLSVersion,
		CipherSuitesCount: cipherCount,
		ExtensionsCount:   extCount,
	}
}

// TLSVersionToString 将 TLS 版本转换为 JA4 格式
func TLSVersionToString(version uint16) string {
	switch version {
	case 0x0301:
		return "10"
	case 0x0302:
		return "11"
	case 0x0303:
		return "12"
	case 0x0304:
		return "13"
	default:
		return "00"
	}
}

// Analyzer TLS 分析器
type Analyzer struct {
	profile *profiles.ClientProfile
}

// NewAnalyzer 创建新的 TLS 分析器
func NewAnalyzer(profile *profiles.ClientProfile) *Analyzer {
	return &Analyzer{profile: profile}
}

// AnalyzeJA3 分析 JA3 指纹
func (a *Analyzer) AnalyzeJA3() *JA3Result {
	if a.profile == nil {
		return nil
	}
	return CalculateJA3FromProfile(*a.profile)
}

// AnalyzeJA4 分析 JA4 指纹
func (a *Analyzer) AnalyzeJA4() *JA4Result {
	if a.profile == nil {
		return nil
	}
	spec := core.ClientHelloSpec{
		TLSVersion:   a.profile.TLSVersion,
		CipherSuites: a.profile.CipherSuites,
		Extensions:   a.profile.Extensions,
	}
	return CalculateJA4(spec)
}

// joinCurves 将曲线 ID 切片连接为字符串
func joinCurves(curves []core.CurveID) string {
	var parts []string
	for _, c := range curves {
		parts = append(parts, strconv.Itoa(int(c)))
	}
	return strings.Join(parts, "-")
}

// Fingerprint 生成完整的 TLS 指纹
func (a *Analyzer) Fingerprint() map[string]interface{} {
	return map[string]interface{}{
		"ja3": a.AnalyzeJA3(),
		"ja4": a.AnalyzeJA4(),
	}
}
