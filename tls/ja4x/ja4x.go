package ja4x

import (
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"strings"
)

// JA4XResult JA4X 指纹结果（X.509 证书指纹）
type JA4XResult struct {
	// JA4X 完整哈希
	Hash string

	// JA4X 原始字符串
	RawString string

	// JA4X_a: 颁发者 RDN（Distinguished Name）序列的 SHA256 前 12 字符
	JA4Xa string

	// JA4X_b: 主体 RDN 序列的 SHA256 前 12 字符
	JA4Xb string

	// JA4X_c: 扩展 OID 序列的 SHA256 前 12 字符
	JA4Xc string

	// 颁发者信息（可读形式）
	Issuer string

	// 主体信息（可读形式）
	Subject string

	// 异常标记
	AnomalyFlags []string

	// 风险评分 (0.0-1.0)
	RiskScore float64
}

// CertificateData 证书数据（用于不依赖 x509.Certificate 的场景）
type CertificateData struct {
	// 颁发者 RDN 字段（按顺序）
	IssuerRDNs []RDNField

	// 主体 RDN 字段（按顺序）
	SubjectRDNs []RDNField

	// 扩展 OID 列表（按顺序）
	ExtensionOIDs []string

	// 证书版本
	Version int

	// 签名算法 OID
	SignatureAlgorithm string

	// 有效期（天数）
	ValidityDays int

	// 是否自签名
	IsSelfSigned bool

	// 是否为 CA 证书
	IsCA bool
}

// RDNField 证书 RDN 字段
type RDNField struct {
	// OID (如 "2.5.4.3" 表示 CN)
	OID string

	// 值
	Value string
}

// sha256Hash12 计算 SHA256 哈希并返回前 12 个字符
func sha256Hash12(input string) string {
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)[:12]
}

// ComputeJA4X 从 x509.Certificate 计算 JA4X 指纹
func ComputeJA4X(cert *x509.Certificate) *JA4XResult {
	data := extractCertificateData(cert)
	return ComputeJA4XFromData(data)
}

// ComputeJA4XFromData 从 CertificateData 计算 JA4X 指纹
func ComputeJA4XFromData(data CertificateData) *JA4XResult {
	result := &JA4XResult{
		AnomalyFlags: []string{},
	}

	// JA4X_a: 颁发者 RDN OID 序列
	issuerOIDs := make([]string, len(data.IssuerRDNs))
	for i, rdn := range data.IssuerRDNs {
		issuerOIDs[i] = rdn.OID
	}
	issuerStr := strings.Join(issuerOIDs, ",")
	result.JA4Xa = sha256Hash12(issuerStr)

	// 颁发者可读形式
	issuerParts := make([]string, len(data.IssuerRDNs))
	for i, rdn := range data.IssuerRDNs {
		issuerParts[i] = fmt.Sprintf("%s=%s", oidName(rdn.OID), rdn.Value)
	}
	result.Issuer = strings.Join(issuerParts, ", ")

	// JA4X_b: 主体 RDN OID 序列
	subjectOIDs := make([]string, len(data.SubjectRDNs))
	for i, rdn := range data.SubjectRDNs {
		subjectOIDs[i] = rdn.OID
	}
	subjectStr := strings.Join(subjectOIDs, ",")
	result.JA4Xb = sha256Hash12(subjectStr)

	// 主体可读形式
	subjectParts := make([]string, len(data.SubjectRDNs))
	for i, rdn := range data.SubjectRDNs {
		subjectParts[i] = fmt.Sprintf("%s=%s", oidName(rdn.OID), rdn.Value)
	}
	result.Subject = strings.Join(subjectParts, ", ")

	// JA4X_c: 扩展 OID 序列
	extStr := strings.Join(data.ExtensionOIDs, ",")
	result.JA4Xc = sha256Hash12(extStr)

	// 原始字符串: issuerOIDs_subjectOIDs_extensionOIDs
	result.RawString = fmt.Sprintf("%s_%s_%s", issuerStr, subjectStr, extStr)

	// 完整哈希: JA4X_a_JA4X_b_JA4X_c
	result.Hash = fmt.Sprintf("%s_%s_%s", result.JA4Xa, result.JA4Xb, result.JA4Xc)

	// 异常检测
	detectCertAnomalies(data, result)

	return result
}

// ComputeJA4XChain 计算证书链的 JA4X 指纹列表
func ComputeJA4XChain(certs []*x509.Certificate) []*JA4XResult {
	results := make([]*JA4XResult, len(certs))
	for i, cert := range certs {
		results[i] = ComputeJA4X(cert)
	}
	return results
}

// MatchJA4X 比较两个 JA4X 哈希是否匹配
func MatchJA4X(hash1, hash2 string) bool {
	return hash1 != "" && hash2 != "" && hash1 == hash2
}

// extractCertificateData 从 x509.Certificate 提取数据
func extractCertificateData(cert *x509.Certificate) CertificateData {
	data := CertificateData{
		Version:      cert.Version,
		IsSelfSigned: cert.Issuer.String() == cert.Subject.String(),
		IsCA:         cert.IsCA,
	}

	// 签名算法
	data.SignatureAlgorithm = cert.SignatureAlgorithm.String()

	// 有效期
	if !cert.NotBefore.IsZero() && !cert.NotAfter.IsZero() {
		data.ValidityDays = int(cert.NotAfter.Sub(cert.NotBefore).Hours() / 24)
	}

	// 颁发者 RDN
	data.IssuerRDNs = extractRDNs(cert.Issuer.String())

	// 主体 RDN
	data.SubjectRDNs = extractRDNs(cert.Subject.String())

	// 扩展 OID
	for _, ext := range cert.Extensions {
		data.ExtensionOIDs = append(data.ExtensionOIDs, ext.Id.String())
	}

	return data
}

// extractRDNs 从 DN 字符串中提取 RDN 字段
func extractRDNs(dn string) []RDNField {
	var rdns []RDNField
	if dn == "" {
		return rdns
	}

	parts := strings.Split(dn, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		eqIdx := strings.Index(part, "=")
		if eqIdx > 0 {
			key := strings.TrimSpace(part[:eqIdx])
			value := strings.TrimSpace(part[eqIdx+1:])
			rdns = append(rdns, RDNField{
				OID:   rdnKeyToOID(key),
				Value: value,
			})
		}
	}
	return rdns
}

// rdnKeyToOID 将 RDN 键名转换为 OID
func rdnKeyToOID(key string) string {
	switch strings.ToUpper(key) {
	case "CN":
		return "2.5.4.3"
	case "O":
		return "2.5.4.10"
	case "OU":
		return "2.5.4.11"
	case "C":
		return "2.5.4.6"
	case "ST", "S":
		return "2.5.4.8"
	case "L":
		return "2.5.4.7"
	case "SERIALNUMBER":
		return "2.5.4.5"
	case "STREET":
		return "2.5.4.9"
	case "POSTALCODE":
		return "2.5.4.17"
	default:
		// 如果已经是 OID 格式，直接返回
		if strings.Contains(key, ".") {
			return key
		}
		return key
	}
}

// oidName 将 OID 转换为可读名称
func oidName(oid string) string {
	switch oid {
	case "2.5.4.3":
		return "CN"
	case "2.5.4.10":
		return "O"
	case "2.5.4.11":
		return "OU"
	case "2.5.4.6":
		return "C"
	case "2.5.4.8":
		return "ST"
	case "2.5.4.7":
		return "L"
	case "2.5.4.5":
		return "SERIALNUMBER"
	case "2.5.4.9":
		return "STREET"
	case "2.5.4.17":
		return "POSTALCODE"
	default:
		return oid
	}
}

// detectCertAnomalies 检测证书异常
func detectCertAnomalies(data CertificateData, result *JA4XResult) {
	baseScore := 0.0

	// 异常 1: 自签名证书
	if data.IsSelfSigned {
		result.AnomalyFlags = append(result.AnomalyFlags, "SELF_SIGNED")
		baseScore += 0.2
	}

	// 异常 2: 过长的有效期（超过 398 天，违反 CA/Browser Forum 要求）
	if data.ValidityDays > 398 && !data.IsCA {
		result.AnomalyFlags = append(result.AnomalyFlags, "LONG_VALIDITY")
		baseScore += 0.15
	}

	// 异常 3: 证书版本不是 v3
	if data.Version != 3 && data.Version != 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "NON_V3_CERT")
		baseScore += 0.1
	}

	// 异常 4: 无扩展
	if len(data.ExtensionOIDs) == 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "NO_EXTENSIONS")
		baseScore += 0.15
	}

	// 异常 5: 主体为空
	if len(data.SubjectRDNs) == 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "EMPTY_SUBJECT")
		baseScore += 0.1
	}

	// 异常 6: 颁发者为空
	if len(data.IssuerRDNs) == 0 {
		result.AnomalyFlags = append(result.AnomalyFlags, "EMPTY_ISSUER")
		baseScore += 0.15
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}
	result.RiskScore = baseScore
}
