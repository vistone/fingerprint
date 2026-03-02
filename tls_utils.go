package fingerprint

import (
	"strconv"
	"strings"

	tls "github.com/bogdanfinn/utls"
)

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
