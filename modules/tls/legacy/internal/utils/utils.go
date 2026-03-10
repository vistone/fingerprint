package utils

import (
	"strconv"
	"strings"

	tls "github.com/bogdanfinn/utls"
)

// isGREASEValue checks whether a value is GREASE (RFC 8701)
// GREASE value format: 0xXAXA (hexadecimal)
func isGREASEValue(v uint16) bool {
	return v&0x0f0f == 0x0a0a && (v>>8) == (v&0x00ff)
}

// filterGREASEUint16 filters GREASE values (uint16 slice)
func filterGREASEUint16(values []uint16) []uint16 {
	result := make([]uint16, 0, len(values))
	for _, v := range values {
		if !isGREASEValue(v) {
			result = append(result, v)
		}
	}
	return result
}

// filterGREASECurveID filters GREASE values (CurveID slice)
func filterGREASECurveID(curves []tls.CurveID) []tls.CurveID {
	result := make([]tls.CurveID, 0, len(curves))
	for _, c := range curves {
		if !isGREASEValue(uint16(c)) {
			result = append(result, c)
		}
	}
	return result
}

// uint16SliceToString converts a uint16 slice to a comma-separated string
func uint16SliceToString(values []uint16) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.Itoa(int(v))
	}
	return strings.Join(parts, "-")
}

// curveIDSliceToString converts a CurveID slice to a comma-separated string
func curveIDSliceToString(curves []tls.CurveID) string {
	parts := make([]string, len(curves))
	for i, c := range curves {
		parts[i] = strconv.Itoa(int(c))
	}
	return strings.Join(parts, "-")
}

// uint8SliceToString converts a uint8 slice to a comma-separated string
func uint8SliceToString(values []uint8) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.Itoa(int(v))
	}
	return strings.Join(parts, "-")
}
