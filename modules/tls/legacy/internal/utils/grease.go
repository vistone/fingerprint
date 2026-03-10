package utils

import (
	tls "github.com/bogdanfinn/utls"
)

// IsGREASEValue checks whether a value is GREASE (RFC 8701)
// GREASE value format: 0xXAXA (hexadecimal)
func IsGREASEValue(v uint16) bool {
	return v&0x0f0f == 0x0a0a && (v>>8) == (v&0x00ff)
}

// FilterGREASEUint16 filters GREASE values (uint16 slice)
func FilterGREASEUint16(values []uint16) []uint16 {
	result := make([]uint16, 0, len(values))
	for _, v := range values {
		if !IsGREASEValue(v) {
			result = append(result, v)
		}
	}
	return result
}

// FilterGREASECurveID filters GREASE values (CurveID slice)
func FilterGREASECurveID(curves []tls.CurveID) []tls.CurveID {
	result := make([]tls.CurveID, 0, len(curves))
	for _, c := range curves {
		if !IsGREASEValue(uint16(c)) {
			result = append(result, c)
		}
	}
	return result
}
