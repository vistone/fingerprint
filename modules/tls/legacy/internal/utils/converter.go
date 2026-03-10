package utils

import (
	"strconv"
	"strings"

	tls "github.com/bogdanfinn/utls"
)

// Uint16SliceToString converts a uint16 slice to a hyphen-separated string
func Uint16SliceToString(values []uint16) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.Itoa(int(v))
	}
	return strings.Join(parts, "-")
}

// CurveIDSliceToString converts a CurveID slice to a hyphen-separated string
func CurveIDSliceToString(curves []tls.CurveID) string {
	parts := make([]string, len(curves))
	for i, c := range curves {
		parts[i] = strconv.Itoa(int(c))
	}
	return strings.Join(parts, "-")
}

// Uint8SliceToString converts a uint8 slice to a hyphen-separated string
func Uint8SliceToString(values []uint8) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.Itoa(int(v))
	}
	return strings.Join(parts, "-")
}
