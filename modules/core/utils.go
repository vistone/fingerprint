// Package core 提供工具函数
package core

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// StringSliceToString 将字符串切片转换为逗号分隔的字符串
func StringSliceToString(slice []string) string {
	return strings.Join(slice, ",")
}

// Uint16SliceToString 将 uint16 切片转换为逗号分隔的字符串
func Uint16SliceToString(slice []uint16) string {
	var parts []string
	for _, v := range slice {
		parts = append(parts, strconv.Itoa(int(v)))
	}
	return strings.Join(parts, ",")
}

// HexSliceToString 将字节切片转换为十六进制字符串
func HexSliceToString(data []byte) string {
	return hex.EncodeToString(data)
}

// CalculateSHA256 计算 SHA256 哈希
func CalculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// CalculateJA3Hash 计算 JA3 哈希（简化版）
func CalculateJA3Hash(tlsVersion uint16, cipherSuites, extensions, curves, points []uint16) string {
	parts := []string{
		strconv.Itoa(int(tlsVersion)),
		Uint16SliceToString(cipherSuites),
		Uint16SliceToString(extensions),
		Uint16SliceToString(curves),
		Uint16SliceToString(points),
	}
	ja3String := strings.Join(parts, ",")
	return CalculateMD5([]byte(ja3String))
}

// CalculateMD5 计算 MD5 哈希
func CalculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// RandomChoice 从切片中随机选择一个元素
func RandomChoice[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	return slice[rand.Intn(len(slice))]
}

// RandomChoiceWithSeed 使用指定种子从切片中随机选择
func RandomChoiceWithSeed[T any](slice []T, seed int64) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	r := rand.New(rand.NewSource(seed))
	return slice[r.Intn(len(slice))]
}

// Shuffle 随机打乱切片顺序
func Shuffle[T any](slice []T) {
	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}

// Contains 检查切片是否包含元素
func Contains[T comparable](slice []T, elem T) bool {
	for _, v := range slice {
		if v == elem {
			return true
		}
	}
	return false
}

// Filter 过滤切片元素
func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Map 映射切片元素
func Map[T, U any](slice []T, mapper func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = mapper(v)
	}
	return result
}

// Unique 去重切片元素
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	var result []T
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// Min 返回两个整数中的较小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max 返回两个整数中的较大值
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Clamp 将值限制在范围内
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// GenerateRandomID 生成随机 ID
func GenerateRandomID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// ParseTLSVersion 解析 TLS 版本字符串
func ParseTLSVersion(version string) uint16 {
	switch version {
	case "1.0":
		return 0x0301
	case "1.1":
		return 0x0302
	case "1.2":
		return 0x0303
	case "1.3":
		return 0x0304
	default:
		return 0x0303 // 默认 TLS 1.2
	}
}

// TLSVersionToString 将 TLS 版本转换为字符串
func TLSVersionToString(version uint16) string {
	switch version {
	case 0x0301:
		return "1.0"
	case 0x0302:
		return "1.1"
	case 0x0303:
		return "1.2"
	case 0x0304:
		return "1.3"
	default:
		return "unknown"
	}
}

// NormalizeString 标准化字符串（小写、去空格）
func NormalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// TruncateString 截断字符串到指定长度
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// SafeString 安全获取字符串（处理 nil）
func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// PtrString 创建字符串指针
func PtrString(s string) *string {
	return &s
}

// PtrInt 创建 int 指针
func PtrInt(i int) *int {
	return &i
}

// PtrBool 创建 bool 指针
func PtrBool(b bool) *bool {
	return &b
}

// MergeMaps 合并多个 map
func MergeMaps(maps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// CopyMap 复制 map
func CopyMap(m map[string]string) map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
