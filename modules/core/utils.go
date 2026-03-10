// Package core provides utility functions
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

// StringSliceToString converts string slice to comma-separated string
func StringSliceToString(slice []string) string {
	return strings.Join(slice, ",")
}

// Uint16SliceToString converts uint16 slice to comma-separated string
func Uint16SliceToString(slice []uint16) string {
	var parts []string
	for _, v := range slice {
		parts = append(parts, strconv.Itoa(int(v)))
	}
	return strings.Join(parts, ",")
}

// HexSliceToString converts byte slice to hexadecimal string
func HexSliceToString(data []byte) string {
	return hex.EncodeToString(data)
}

// CalculateSHA256 calculates SHA256 hash
func CalculateSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// CalculateJA3Hash calculates JA3 hash (simplified version)
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

// CalculateMD5 calculates MD5 hash
func CalculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// RandomChoice randomly selects one element from slice
func RandomChoice[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	return slice[rand.Intn(len(slice))]
}

// RandomChoiceWithSeed randomly selects from slice using specified seed
func RandomChoiceWithSeed[T any](slice []T, seed int64) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	r := rand.New(rand.NewSource(seed))
	return slice[r.Intn(len(slice))]
}

// Shuffle randomly shuffles slice order
func Shuffle[T any](slice []T) {
	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}

// Contains checks if slice contains element
func Contains[T comparable](slice []T, elem T) bool {
	for _, v := range slice {
		if v == elem {
			return true
		}
	}
	return false
}

// Filter filters slice elements
func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Map maps slice elements
func Map[T, U any](slice []T, mapper func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = mapper(v)
	}
	return result
}

// Unique deduplicates slice elements
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

// Min returns smaller of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns larger of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Clamp clamps value within range
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// GenerateRandomID generates random ID
func GenerateRandomID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// ParseTLSVersion parses TLS version string
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
		return 0x0303 // default TLS 1.2
	}
}

// TLSVersionToString converts TLS version to string
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

// NormalizeString normalizes string (lowercase, trim spaces)
func NormalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// TruncateString truncates string to specified length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// SafeString safely gets string (handles nil)
func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// PtrString creates string pointer
func PtrString(s string) *string {
	return &s
}

// PtrInt creates int pointer
func PtrInt(i int) *int {
	return &i
}

// PtrBool creates bool pointer
func PtrBool(b bool) *bool {
	return &b
}

// MergeMaps merges multiple maps
func MergeMaps(maps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// CopyMap copies map
func CopyMap(m map[string]string) map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
