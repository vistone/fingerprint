package utils

import (
	tls "github.com/bogdanfinn/utls"
)

// IsGREASEValue 检查是否为 GREASE 值（RFC 8701）
// GREASE 值格式：0xXAXA（十六进制）
func IsGREASEValue(v uint16) bool {
	return v&0x0f0f == 0x0a0a && (v>>8) == (v&0x00ff)
}

// FilterGREASEUint16 过滤 GREASE 值（uint16 切片）
func FilterGREASEUint16(values []uint16) []uint16 {
	result := make([]uint16, 0, len(values))
	for _, v := range values {
		if !IsGREASEValue(v) {
			result = append(result, v)
		}
	}
	return result
}

// FilterGREASECurveID 过滤 GREASE 值（CurveID 切片）
func FilterGREASECurveID(curves []tls.CurveID) []tls.CurveID {
	result := make([]tls.CurveID, 0, len(curves))
	for _, c := range curves {
		if !IsGREASEValue(uint16(c)) {
			result = append(result, c)
		}
	}
	return result
}
