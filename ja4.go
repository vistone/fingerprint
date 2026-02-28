package fingerprint

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	tls "github.com/bogdanfinn/utls"
)

// JA4Result JA4 指纹结果
type JA4Result struct {
	// JA4 完整指纹（哈希形式）
	Hash string
	// JA4_r 完整原始字符串形式
	RawString string
	// JA4_a: 协议标识 + TLS版本 + SNI + 密码套件数 + 扩展数 + ALPN首尾字符
	JA4A string
	// JA4_b: 密码套件（排序后，逗号分隔的4位十六进制，前12位SHA256）
	JA4B string
	// JA4_c: 扩展（排序后，逗号分隔的4位十六进制）+ 签名算法（前12位SHA256）
	JA4C string
}

// tlsVersionToJA4 将 TLS 版本转换为 JA4 格式字符串
func tlsVersionToJA4(version uint16) string {
	switch version {
	case tls.VersionTLS13:
		return "13"
	case tls.VersionTLS12:
		return "12"
	case tls.VersionTLS11:
		return "11"
	case tls.VersionTLS10:
		return "10"
	default:
		return "00"
	}
}

// firstLastALPN 从 ALPN 字符串提取首尾字符
// 非 ASCII 字符替换为 '9'
func firstLastALPN(s string) (byte, byte) {
	if len(s) == 0 {
		return '0', '0'
	}
	replaceNonASCII := func(b byte) byte {
		if b < 128 {
			return b
		}
		return '9'
	}
	first := replaceNonASCII(s[0])
	if len(s) == 1 {
		return first, '0'
	}
	last := replaceNonASCII(s[len(s)-1])
	return first, last
}

// sha256Hash12 计算 SHA256 哈希并返回前12个字符
func sha256Hash12(input string) string {
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)[:12]
}

// JA4Signature JA4 指纹签名输入
type JA4Signature struct {
	// TLS 版本
	TLSVersion uint16
	// 密码套件（包含 GREASE）
	CipherSuites []uint16
	// 扩展（包含 GREASE）
	Extensions []uint16
	// 签名算法（包含 GREASE）
	SignatureAlgorithms []uint16
	// SNI（可选）
	SNI string
	// ALPN 首个协议（可选）
	ALPN string
}

// ComputeJA4 计算 JA4 指纹（排序版本）
func (s *JA4Signature) ComputeJA4() *JA4Result {
	return s.computeJA4WithOrder(false)
}

// ComputeJA4Original 计算 JA4 指纹（保持原始顺序版本，即 JA4_o）
func (s *JA4Signature) ComputeJA4Original() *JA4Result {
	return s.computeJA4WithOrder(true)
}

// computeJA4WithOrder 计算 JA4 指纹（指定顺序）
func (s *JA4Signature) computeJA4WithOrder(originalOrder bool) *JA4Result {
	// 过滤 GREASE 值
	filteredCiphers := filterGREASEUint16(s.CipherSuites)
	filteredExtensions := filterGREASEUint16(s.Extensions)
	filteredSigAlgs := filterGREASEUint16(s.SignatureAlgorithms)

	// 协议标识（TLS 为 't'，QUIC 为 'q'）
	protocol := "t"

	// TLS 版本字符串
	tlsVersionStr := tlsVersionToJA4(s.TLSVersion)

	// SNI 指示符：存在 SNI 为 'd'，不存在为 'i'
	sniIndicator := "i"
	if s.SNI != "" {
		sniIndicator = "d"
	}

	// 密码套件数（2位十进制，最大99）- 使用原始数量（过滤前）
	cipherCount := fmt.Sprintf("%02d", min99(len(s.CipherSuites)))

	// 扩展数（2位十进制，最大99）- 使用原始数量（过滤前）
	extensionCount := fmt.Sprintf("%02d", min99(len(s.Extensions)))

	// ALPN 首尾字符
	alpnFirst, alpnLast := byte('0'), byte('0')
	if s.ALPN != "" {
		alpnFirst, alpnLast = firstLastALPN(s.ALPN)
	}

	// JA4_a 格式：protocol + version + sni + cipher_count + extension_count + alpn_first + alpn_last
	ja4a := fmt.Sprintf("%s%s%s%s%s%c%c", protocol, tlsVersionStr, sniIndicator, cipherCount, extensionCount, alpnFirst, alpnLast)

	// JA4_b: 密码套件（排序或保持原始顺序，逗号分隔，4位十六进制）- 过滤 GREASE
	ciphersForB := make([]uint16, len(filteredCiphers))
	copy(ciphersForB, filteredCiphers)
	if !originalOrder {
		sort.Slice(ciphersForB, func(i, j int) bool { return ciphersForB[i] < ciphersForB[j] })
	}
	ja4bParts := make([]string, len(ciphersForB))
	for i, c := range ciphersForB {
		ja4bParts[i] = fmt.Sprintf("%04x", c)
	}
	ja4bRaw := strings.Join(ja4bParts, ",")

	// JA4_c: 扩展（排序或保持原始顺序）+ 签名算法
	extensionsForC := make([]uint16, len(filteredExtensions))
	copy(extensionsForC, filteredExtensions)

	// 排序版本：移除 SNI (0x0000) 和 ALPN (0x0010) 并排序
	// 原始顺序版本：保留 SNI/ALPN 并保持原始顺序
	if !originalOrder {
		filtered := extensionsForC[:0]
		for _, ext := range extensionsForC {
			if ext != 0x0000 && ext != 0x0010 {
				filtered = append(filtered, ext)
			}
		}
		extensionsForC = filtered
		sort.Slice(extensionsForC, func(i, j int) bool { return extensionsForC[i] < extensionsForC[j] })
	}

	extParts := make([]string, len(extensionsForC))
	for i, e := range extensionsForC {
		extParts[i] = fmt.Sprintf("%04x", e)
	}
	extensionsStr := strings.Join(extParts, ",")

	// 签名算法不排序，但过滤 GREASE
	sigAlgParts := make([]string, len(filteredSigAlgs))
	for i, s := range filteredSigAlgs {
		sigAlgParts[i] = fmt.Sprintf("%04x", s)
	}
	sigAlgsStr := strings.Join(sigAlgParts, ",")

	// 根据规范，如果没有签名算法，字符串不以下划线结尾
	var ja4cRaw string
	if sigAlgsStr == "" {
		ja4cRaw = extensionsStr
	} else if extensionsStr == "" {
		ja4cRaw = sigAlgsStr
	} else {
		ja4cRaw = extensionsStr + "_" + sigAlgsStr
	}

	// 生成 JA4_b 和 JA4_c 哈希（SHA256 前12个字符）
	ja4bHash := sha256Hash12(ja4bRaw)
	ja4cHash := sha256Hash12(ja4cRaw)

	// JA4 完整哈希：ja4_a + "_" + ja4_b_hash + "_" + ja4_c_hash
	ja4Hash := fmt.Sprintf("%s_%s_%s", ja4a, ja4bHash, ja4cHash)

	// JA4_r 原始：ja4_a + "_" + ja4_b_raw + "_" + ja4_c_raw
	ja4Raw := fmt.Sprintf("%s_%s_%s", ja4a, ja4bRaw, ja4cRaw)

	return &JA4Result{
		Hash:      ja4Hash,
		RawString: ja4Raw,
		JA4A:      ja4a,
		JA4B:      ja4bRaw,
		JA4C:      ja4cRaw,
	}
}

// min99 返回 n 和 99 中的较小值
func min99(n int) int {
	if n > 99 {
		return 99
	}
	return n
}

// ComputeJA4FromSpec 从 TLS ClientHello 规范计算 JA4 指纹
func ComputeJA4FromSpec(spec tls.ClientHelloSpec) (*JA4Result, error) {
	sig := &JA4Signature{
		TLSVersion: tls.VersionTLS12,
	}

	// 提取密码套件
	sig.CipherSuites = make([]uint16, len(spec.CipherSuites))
	copy(sig.CipherSuites, spec.CipherSuites)

	// 提取扩展信息
	extensions := make([]uint16, 0)
	var sigAlgs []uint16

	for _, ext := range spec.Extensions {
		switch e := ext.(type) {
		case *tls.SupportedVersionsExtension:
			for _, v := range e.Versions {
				if !isGREASEValue(v) && v > sig.TLSVersion {
					sig.TLSVersion = v
				}
			}
			extensions = append(extensions, 43)

		case *tls.SNIExtension:
			extensions = append(extensions, 0)

		case *tls.ALPNExtension:
			if len(e.AlpnProtocols) > 0 {
				sig.ALPN = e.AlpnProtocols[0]
			}
			extensions = append(extensions, 16)

		case *tls.SignatureAlgorithmsExtension:
			for _, sa := range e.SupportedSignatureAlgorithms {
				sigAlgs = append(sigAlgs, uint16(sa))
			}
			extensions = append(extensions, 13)

		case *tls.SupportedCurvesExtension:
			extensions = append(extensions, 10)

		case *tls.SupportedPointsExtension:
			extensions = append(extensions, 11)

		case *tls.StatusRequestExtension:
			extensions = append(extensions, 5)

		case *tls.SessionTicketExtension:
			extensions = append(extensions, 35)

		case *tls.SCTExtension:
			extensions = append(extensions, 18)

		case *tls.KeyShareExtension:
			extensions = append(extensions, 51)

		case *tls.PSKKeyExchangeModesExtension:
			extensions = append(extensions, 45)

		case *tls.ExtendedMasterSecretExtension:
			extensions = append(extensions, 23)

		case *tls.RenegotiationInfoExtension:
			extensions = append(extensions, 65281)

		case *tls.UtlsCompressCertExtension:
			extensions = append(extensions, 27)

		case *tls.ApplicationSettingsExtension:
			extensions = append(extensions, 17513)

		case *tls.ApplicationSettingsExtensionNew:
			extensions = append(extensions, 17613)

		case *tls.UtlsGREASEExtension:
			// 跳过 GREASE 扩展

		default:
			_ = e
		}
	}

	sig.Extensions = extensions
	sig.SignatureAlgorithms = sigAlgs

	return sig.ComputeJA4(), nil
}

// ComputeJA4FromProfile 从 ClientProfile 计算 JA4 指纹
func ComputeJA4FromProfile(profile ClientProfile) (*JA4Result, error) {
	spec, err := profile.GetClientHelloSpec()
	if err != nil {
		return nil, fmt.Errorf("获取 ClientHelloSpec 失败: %w", err)
	}
	return ComputeJA4FromSpec(spec)
}

// ComputeJA4ByProfileName 根据指纹名称计算 JA4 指纹
func ComputeJA4ByProfileName(profileName string) (*JA4Result, error) {
	profile, ok := MappedTLSClients[profileName]
	if !ok {
		return nil, fmt.Errorf("指纹 %s 不存在", profileName)
	}
	return ComputeJA4FromProfile(profile)
}
