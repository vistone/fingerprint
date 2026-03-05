package utils

import (
	"strconv"
	"strings"

	tls "github.com/bogdanfinn/utls"
)

// Uint16SliceToString 将 uint16 切片转换为连字符分隔字符串
func Uint16SliceToString(values []uint16) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.Itoa(int(v))
	}
	return strings.Join(parts, "-")
}

// CurveIDSliceToString 将 CurveID 切片转换为连字符分隔字符串
func CurveIDSliceToString(curves []tls.CurveID) string {
	parts := make([]string, len(curves))
	for i, c := range curves {
		parts[i] = strconv.Itoa(int(c))
	}
	return strings.Join(parts, "-")
}

// Uint8SliceToString 将 uint8 切片转换为连字符分隔字符串
func Uint8SliceToString(values []uint8) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.Itoa(int(v))
	}
	return strings.Join(parts, "-")
}
