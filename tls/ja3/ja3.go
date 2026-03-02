package ja3

import (
	"crypto/md5"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"

	tls "github.com/bogdanfinn/utls"
)

// JA3Result JA3 指纹结果
type JA3Result struct {
	// JA3 fingerprint hash (MD5)
	Hash string
	// JA3 原始字符串（可读形式）
	RawString string
	// TLS 版本
	TLSVersion uint16
	// 密码套件列表（已过滤 GREASE）
	CipherSuites []uint16
	// 扩展列表（已过滤 GREASE）
	Extensions []uint16
	// 椭圆曲线列表（已过滤 GREASE）
	EllipticCurves []tls.CurveID
	// 椭圆曲线点格式列表
	EllipticCurvePointFormats []uint8
}

// ClientProfile 客户端指纹配置
type ClientProfile interface {
	GetClientHelloSpec() (tls.ClientHelloSpec, error)
}

var (
	ja3ProfileIndexOnce sync.Once
	ja3ProfileIndex     map[string][]string
	MappedTLSClients    map[string]ClientProfile
)

// InitMappedTLSClients 由根包调用以初始化客户端映射表
// 使用 interface{} 避免类型匹配问题
func InitMappedTLSClients(clients interface{}) {
	if m, ok := clients.(map[string]ClientProfile); ok {
		MappedTLSClients = m
		return
	}
	// 尝试转换为 map[string]interface{}
	if m, ok := clients.(map[string]interface{}); ok {
		MappedTLSClients = make(map[string]ClientProfile)
		for k, v := range m {
			if cp, ok := v.(ClientProfile); ok {
				MappedTLSClients[k] = cp
			}
		}
	}
}

// InitMappedTLSClientsRaw 从根包接收任何 ClientProfile map 并进行类型转换
func InitMappedTLSClientsRaw(clients interface{}) {
	// 使用反射来导出和转换底层 map
	clientsVal := reflect.ValueOf(clients)
	if clientsVal.Kind() != reflect.Map {
		return
	}

	// 创建新的 map
	MappedTLSClients = make(map[string]ClientProfile)

	// 遍历原始 map 的所有键值对
	for _, keyVal := range clientsVal.MapKeys() {
		key := fmt.Sprintf("%v", keyVal.Interface())
		valInterface := clientsVal.MapIndex(keyVal).Interface()

		// 尝试将值转换为 ClientProfile
		// 由于值实现了 GetClientHelloSpec 方法，它应该满足 ClientProfile 接口
		if cp, ok := valInterface.(ClientProfile); ok {
			MappedTLSClients[key] = cp
		} else {
			// 如果直接转换不成功，使用类型断言来获取 Getter 方法
			// 这里我们检查该值是否有 GetClientHelloSpec 方法
			refVal := reflect.ValueOf(valInterface)
			method := refVal.MethodByName("GetClientHelloSpec")
			if method.IsValid() {
				// 创建一个包装器来实现 ClientProfile 接口
				wrapper := &dynamicClientProfile{value: valInterface}
				MappedTLSClients[key] = wrapper
			}
		}
	}
}

// dynamicClientProfile 动态包装器，用于实现 ClientProfile 接口
type dynamicClientProfile struct {
	value interface{}
}

// GetClientHelloSpec 实现 ClientProfile 接口
func (d *dynamicClientProfile) GetClientHelloSpec() (tls.ClientHelloSpec, error) {
	refVal := reflect.ValueOf(d.value)
	method := refVal.MethodByName("GetClientHelloSpec")
	if !method.IsValid() {
		return tls.ClientHelloSpec{}, fmt.Errorf("对象未实现 GetClientHelloSpec 方法")
	}

	results := method.Call(nil)
	if len(results) >= 2 {
		if spec, ok := results[0].Interface().(tls.ClientHelloSpec); ok {
			var err error
			if !results[1].IsNil() {
				err = results[1].Interface().(error)
			}
			return spec, err
		}
	}
	return tls.ClientHelloSpec{}, fmt.Errorf("方法调用失败")
}

func buildJA3ProfileIndex() {
	index := make(map[string][]string)

	if MappedTLSClients == nil {
		ja3ProfileIndex = index
		return
	}

	for name, profile := range MappedTLSClients {
		result, err := ComputeJA3FromProfile(profile)
		if err != nil || result == nil || result.Hash == "" {
			continue
		}

		hash := strings.ToLower(result.Hash)
		index[hash] = append(index[hash], name)
	}

	for hash := range index {
		sort.Strings(index[hash])
	}

	ja3ProfileIndex = index
}

// findProfileByJA3NoCopy 根据 JA3 哈希查找匹配 Profile（内部零拷贝路径）
// 返回值直接引用内部缓存，仅供只读场景使用。
func findProfileByJA3NoCopy(ja3Hash string) []string {
	ja3ProfileIndexOnce.Do(buildJA3ProfileIndex)
	if ja3ProfileIndex == nil {
		return nil
	}
	return ja3ProfileIndex[strings.ToLower(ja3Hash)]
}

// isGREASEValue 检查是否为 GREASE 值（RFC 8701）
// GREASE 值格式：0xXAXA（十六进制）
func isGREASEValue(v uint16) bool {
	return v&0x0f0f == 0x0a0a && (v>>8) == (v&0x00ff)
}

// filterGREASEUint16 过滤 GREASE 值（uint16 切片）
func filterGREASEUint16(values []uint16) []uint16 {
	result := make([]uint16, 0, len(values))
	for _, v := range values {
		if !isGREASEValue(v) {
			result = append(result, v)
		}
	}
	return result
}

// filterGREASECurveID 过滤 GREASE 值（CurveID 切片）
func filterGREASECurveID(curves []tls.CurveID) []tls.CurveID {
	result := make([]tls.CurveID, 0, len(curves))
	for _, c := range curves {
		if !isGREASEValue(uint16(c)) {
			result = append(result, c)
		}
	}
	return result
}

// uint16SliceToString 将 uint16 切片转换为逗号分隔字符串
func uint16SliceToString(values []uint16) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.Itoa(int(v))
	}
	return strings.Join(parts, "-")
}

// curveIDSliceToString 将 CurveID 切片转换为逗号分隔字符串
func curveIDSliceToString(curves []tls.CurveID) string {
	parts := make([]string, len(curves))
	for i, c := range curves {
		parts[i] = strconv.Itoa(int(c))
	}
	return strings.Join(parts, "-")
}

// uint8SliceToString 将 uint8 切片转换为逗号分隔字符串
func uint8SliceToString(values []uint8) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.Itoa(int(v))
	}
	return strings.Join(parts, "-")
}

// ComputeJA3FromSpec 从 TLS ClientHello 规范计算 JA3 指纹
// JA3 算法：MD5(TLSVersion,Ciphers,Extensions,EllipticCurves,EllipticCurvePointFormats)
func ComputeJA3FromSpec(spec tls.ClientHelloSpec) (*JA3Result, error) {
	result := &JA3Result{}

	// 提取 TLS 版本（默认为 TLS 1.2）
	result.TLSVersion = tls.VersionTLS12

	// 提取密码套件（过滤 GREASE）
	ciphers := filterGREASEUint16(spec.CipherSuites)
	result.CipherSuites = ciphers

	// 提取扩展信息
	extensions := make([]uint16, 0)
	var curves []tls.CurveID
	var pointFormats []uint8

	for _, ext := range spec.Extensions {
		switch e := ext.(type) {
		case *tls.SupportedVersionsExtension:
			// 提取最高 TLS 版本
			for _, v := range e.Versions {
				if !isGREASEValue(v) && v > result.TLSVersion {
					result.TLSVersion = v
				}
			}
			extensions = append(extensions, 43) // extension_type_supported_versions

		case *tls.SupportedCurvesExtension:
			curves = filterGREASECurveID(e.Curves)
			extensions = append(extensions, 10) // extension_type_supported_groups

		case *tls.SupportedPointsExtension:
			pointFormats = e.SupportedPoints
			extensions = append(extensions, 11) // extension_type_ec_point_formats

		case *tls.SNIExtension:
			extensions = append(extensions, 0) // extension_type_server_name

		case *tls.StatusRequestExtension:
			extensions = append(extensions, 5) // extension_type_status_request

		case *tls.SessionTicketExtension:
			extensions = append(extensions, 35) // extension_type_session_ticket

		case *tls.ALPNExtension:
			extensions = append(extensions, 16) // extension_type_alpn

		case *tls.SignatureAlgorithmsExtension:
			extensions = append(extensions, 13) // extension_type_signature_algorithms

		case *tls.SCTExtension:
			extensions = append(extensions, 18) // extension_type_signed_certificate_timestamp

		case *tls.KeyShareExtension:
			extensions = append(extensions, 51) // extension_type_key_share

		case *tls.PSKKeyExchangeModesExtension:
			extensions = append(extensions, 45) // extension_type_psk_key_exchange_modes

		case *tls.ExtendedMasterSecretExtension:
			extensions = append(extensions, 23) // extension_type_extended_master_secret

		case *tls.RenegotiationInfoExtension:
			extensions = append(extensions, 65281) // extension_type_renegotiation_info (0xff01)

		case *tls.UtlsCompressCertExtension:
			extensions = append(extensions, 27) // extension_type_compress_certificate

		case *tls.ApplicationSettingsExtension:
			extensions = append(extensions, 17513) // extension_type_application_settings

		case *tls.ApplicationSettingsExtensionNew:
			extensions = append(extensions, 17613) // extension_type_application_settings_new

		case *tls.UtlsGREASEExtension:
			// 跳过 GREASE 扩展

		default:
			// 对于未知扩展，忽略
			_ = e
		}
	}

	// 过滤扩展中的 GREASE 值
	result.Extensions = filterGREASEUint16(extensions)
	result.EllipticCurves = curves
	result.EllipticCurvePointFormats = pointFormats

	// 构建 JA3 原始字符串
	// 格式：TLSVersion,CipherSuites,Extensions,EllipticCurves,EllipticCurvePointFormats
	rawParts := []string{
		strconv.Itoa(int(result.TLSVersion)),
		uint16SliceToString(result.CipherSuites),
		uint16SliceToString(result.Extensions),
		curveIDSliceToString(result.EllipticCurves),
		uint8SliceToString(result.EllipticCurvePointFormats),
	}
	result.RawString = strings.Join(rawParts, ",")

	// 计算 MD5 哈希
	hash := md5.Sum([]byte(result.RawString))
	result.Hash = fmt.Sprintf("%x", hash)

	return result, nil
}

// ComputeJA3FromProfile 从 ClientProfile 计算 JA3 指纹
func ComputeJA3FromProfile(profile ClientProfile) (*JA3Result, error) {
	spec, err := profile.GetClientHelloSpec()
	if err != nil {
		return nil, fmt.Errorf("获取 ClientHelloSpec 失败: %w", err)
	}
	return ComputeJA3FromSpec(spec)
}

// ComputeJA3ByProfileName 根据指纹名称计算 JA3 指纹
func ComputeJA3ByProfileName(profileName string) (*JA3Result, error) {
	profile, ok := MappedTLSClients[profileName]
	if !ok {
		return nil, fmt.Errorf("指纹 %s 不存在", profileName)
	}
	return ComputeJA3FromProfile(profile)
}

// MatchJA3 检查两个 JA3 哈希是否匹配
func MatchJA3(hash1, hash2 string) bool {
	return strings.EqualFold(hash1, hash2)
}

// FindProfileByJA3 根据 JA3 哈希查找匹配的 ClientProfile 名称
// 返回所有匹配的 profile 名称列表（可能有多个）
func FindProfileByJA3(ja3Hash string) []string {
	matches := findProfileByJA3NoCopy(ja3Hash)
	if len(matches) == 0 {
		return nil
	}

	result := make([]string, len(matches))
	copy(result, matches)
	return result
}
