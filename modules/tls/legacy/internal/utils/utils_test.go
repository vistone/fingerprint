package utils

import (
	"reflect"
	"testing"

	tls "github.com/bogdanfinn/utls"
)

// translated comment
func TestIsGREASEValue(t *testing.T) {
	tests := []struct {
		name     string
		value    uint16
		expected bool
	}{
		{
			name:     "GREASE value 0x0A0A",
			value:    0x0A0A,
			expected: true,
		},
		{
			name:     "GREASE value 0x1A1A",
			value:    0x1A1A,
			expected: true,
		},
		{
			name:     "GREASE value 0x2A2A",
			value:    0x2A2A,
			expected: true,
		},
		{
			name:     "GREASE value 0x3A3A",
			value:    0x3A3A,
			expected: true,
		},
		{
			name:     "GREASE value 0x4A4A",
			value:    0x4A4A,
			expected: true,
		},
		{
			name:     "GREASE value 0x5A5A",
			value:    0x5A5A,
			expected: true,
		},
		{
			name:     "GREASE value 0x6A6A",
			value:    0x6A6A,
			expected: true,
		},
		{
			name:     "GREASE value 0x7A7A",
			value:    0x7A7A,
			expected: true,
		},
		{
			name:     "GREASE value 0x8A8A",
			value:    0x8A8A,
			expected: true,
		},
		{
			name:     "GREASE value 0x9A9A",
			value:    0x9A9A,
			expected: true,
		},
		{
			name:     "GREASE value 0xAAAA",
			value:    0xAAAA,
			expected: true,
		},
		{
			name:     "GREASE value 0xBABA",
			value:    0xBABA,
			expected: true,
		},
		{
			name:     "GREASE value 0xCACA",
			value:    0xCACA,
			expected: true,
		},
		{
			name:     "GREASE value 0xDADA",
			value:    0xDADA,
			expected: true,
		},
		{
			name:     "GREASE value 0xEAEA",
			value:    0xEAEA,
			expected: true,
		},
		{
			name:     "GREASE value 0xFAFA",
			value:    0xFAFA,
			expected: true,
		},
		{
			name:     "TLS 1.3 cipher suite 0x1301",
			value:    0x1301,
			expected: false,
		},
		{
			name:     "TLS cipher suite 0x002f",
			value:    0x002f,
			expected: false,
		},
		{
			name:     "Non-GREASE value 0x0000",
			value:    0x0000,
			expected: false,
		},
		{
			name:     "Non-GREASE value 0x0A0B",
			value:    0x0A0B,
			expected: false,
		},
		{
			name:     "Non-GREASE value 0x0B0A",
			value:    0x0B0A,
			expected: false,
		},
		{
			name:     "Non-GREASE value 0xFFFF",
			value:    0xFFFF,
			expected: false,
		},
		{
			name:     "Non-GREASE value 0x1234",
			value:    0x1234,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGREASEValue(tt.value)
			if result != tt.expected {
				t.Errorf("IsGREASEValue(0x%04X) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

// translated comment
func TestFilterGREASEUint16(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint16
		expected []uint16
	}{
		{
			name:     "nil slice",
			input:    nil,
			expected: []uint16{},
		},
		{
			name:     "empty slice",
			input:    []uint16{},
			expected: []uint16{},
		},
		{
			name:     "no GREASE values",
			input:    []uint16{0x1301, 0x1302, 0x1303},
			expected: []uint16{0x1301, 0x1302, 0x1303},
		},
		{
			name:     "all GREASE values",
			input:    []uint16{0x0A0A, 0x1A1A, 0x2A2A},
			expected: []uint16{},
		},
		{
			name:     "mixed GREASE and non-GREASE",
			input:    []uint16{0x0A0A, 0x1301, 0x1A1A, 0x002f, 0x2A2A},
			expected: []uint16{0x1301, 0x002f},
		},
		{
			name:     "GREASE at start",
			input:    []uint16{0x0A0A, 0x1301, 0x1302},
			expected: []uint16{0x1301, 0x1302},
		},
		{
			name:     "GREASE at end",
			input:    []uint16{0x1301, 0x1302, 0xFAFA},
			expected: []uint16{0x1301, 0x1302},
		},
		{
			name:     "single GREASE value",
			input:    []uint16{0x0A0A},
			expected: []uint16{},
		},
		{
			name:     "single non-GREASE value",
			input:    []uint16{0x1301},
			expected: []uint16{0x1301},
		},
		{
			name:     "multiple GREASE values between non-GREASE",
			input:    []uint16{0x1301, 0x0A0A, 0x1A1A, 0x1302, 0x2A2A, 0x1303},
			expected: []uint16{0x1301, 0x1302, 0x1303},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterGREASEUint16(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("FilterGREASEUint16(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// translated comment
func TestFilterGREASECurveID(t *testing.T) {
	tests := []struct {
		name     string
		input    []tls.CurveID
		expected []tls.CurveID
	}{
		{
			name:     "nil slice",
			input:    nil,
			expected: []tls.CurveID{},
		},
		{
			name:     "empty slice",
			input:    []tls.CurveID{},
			expected: []tls.CurveID{},
		},
		{
			name:     "no GREASE values",
			input:    []tls.CurveID{tls.X25519, tls.CurveP256, tls.CurveP384},
			expected: []tls.CurveID{tls.X25519, tls.CurveP256, tls.CurveP384},
		},
		{
			name:     "all GREASE values",
			input:    []tls.CurveID{0x0A0A, 0x1A1A, 0x2A2A},
			expected: []tls.CurveID{},
		},
		{
			name:     "mixed GREASE and non-GREASE",
			input:    []tls.CurveID{0x0A0A, tls.X25519, 0x1A1A, tls.CurveP256, 0x2A2A},
			expected: []tls.CurveID{tls.X25519, tls.CurveP256},
		},
		{
			name:     "GREASE at start",
			input:    []tls.CurveID{0x0A0A, tls.X25519, tls.CurveP256},
			expected: []tls.CurveID{tls.X25519, tls.CurveP256},
		},
		{
			name:     "GREASE at end",
			input:    []tls.CurveID{tls.X25519, tls.CurveP256, 0xFAFA},
			expected: []tls.CurveID{tls.X25519, tls.CurveP256},
		},
		{
			name:     "single GREASE value",
			input:    []tls.CurveID{0x0A0A},
			expected: []tls.CurveID{},
		},
		{
			name:     "single non-GREASE value",
			input:    []tls.CurveID{tls.X25519},
			expected: []tls.CurveID{tls.X25519},
		},
		{
			name:     "multiple GREASE values between curves",
			input:    []tls.CurveID{tls.X25519, 0x0A0A, 0x1A1A, tls.CurveP256, 0x2A2A, tls.CurveP384},
			expected: []tls.CurveID{tls.X25519, tls.CurveP256, tls.CurveP384},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterGREASECurveID(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("FilterGREASECurveID(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// translated comment
func TestUint16SliceToString(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint16
		expected string
	}{
		{
			name:     "nil slice",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty slice",
			input:    []uint16{},
			expected: "",
		},
		{
			name:     "single value",
			input:    []uint16{0x1301},
			expected: "4865",
		},
		{
			name:     "two values",
			input:    []uint16{0x1301, 0x1302},
			expected: "4865-4866",
		},
		{
			name:     "multiple values",
			input:    []uint16{0x1301, 0x1302, 0x1303, 0x002f},
			expected: "4865-4866-4867-47",
		},
		{
			name:     "GREASE values included",
			input:    []uint16{0x0A0A, 0x1301},
			expected: "2570-4865",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Uint16SliceToString(tt.input)
			if result != tt.expected {
				t.Errorf("Uint16SliceToString(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// translated comment
func TestCurveIDSliceToString(t *testing.T) {
	tests := []struct {
		name     string
		input    []tls.CurveID
		expected string
	}{
		{
			name:     "nil slice",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty slice",
			input:    []tls.CurveID{},
			expected: "",
		},
		{
			name:     "single curve",
			input:    []tls.CurveID{tls.X25519},
			expected: "29",
		},
		{
			name:     "two curves",
			input:    []tls.CurveID{tls.X25519, tls.CurveP256},
			expected: "29-23",
		},
		{
			name:     "multiple curves",
			input:    []tls.CurveID{tls.X25519, tls.CurveP256, tls.CurveP384, tls.CurveP521},
			expected: "29-23-24-25",
		},
		{
			name:     "GREASE curve IDs included",
			input:    []tls.CurveID{0x0A0A, tls.X25519},
			expected: "2570-29",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CurveIDSliceToString(tt.input)
			if result != tt.expected {
				t.Errorf("CurveIDSliceToString(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// translated comment
func TestUint8SliceToString(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint8
		expected string
	}{
		{
			name:     "nil slice",
			input:    nil,
			expected: "",
		},
		{
			name:     "empty slice",
			input:    []uint8{},
			expected: "",
		},
		{
			name:     "single value",
			input:    []uint8{10},
			expected: "10",
		},
		{
			name:     "two values",
			input:    []uint8{10, 20},
			expected: "10-20",
		},
		{
			name:     "multiple values",
			input:    []uint8{0, 127, 128, 255},
			expected: "0-127-128-255",
		},
		{
			name:     "consecutive values",
			input:    []uint8{1, 2, 3, 4, 5},
			expected: "1-2-3-4-5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Uint8SliceToString(tt.input)
			if result != tt.expected {
				t.Errorf("Uint8SliceToString(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// translated comment
func TestInternalIsGREASEValue(t *testing.T) {
	tests := []struct {
		name     string
		value    uint16
		expected bool
	}{
		{
			name:     "internal GREASE value 0x0A0A",
			value:    0x0A0A,
			expected: true,
		},
		{
			name:     "internal GREASE value 0x5A5A",
			value:    0x5A5A,
			expected: true,
		},
		{
			name:     "internal GREASE value 0xFAFA",
			value:    0xFAFA,
			expected: true,
		},
		{
			name:     "internal non-GREASE value 0x1301",
			value:    0x1301,
			expected: false,
		},
		{
			name:     "internal non-GREASE value 0x0000",
			value:    0x0000,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGREASEValue(tt.value)
			if result != tt.expected {
				t.Errorf("isGREASEValue(0x%04X) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

// translated comment
func TestInternalFilterGREASEUint16(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint16
		expected []uint16
	}{
		{
			name:     "internal nil slice",
			input:    nil,
			expected: []uint16{},
		},
		{
			name:     "internal empty slice",
			input:    []uint16{},
			expected: []uint16{},
		},
		{
			name:     "internal no GREASE values",
			input:    []uint16{0x1301, 0x1302},
			expected: []uint16{0x1301, 0x1302},
		},
		{
			name:     "internal all GREASE values",
			input:    []uint16{0x0A0A, 0x1A1A},
			expected: []uint16{},
		},
		{
			name:     "internal mixed values",
			input:    []uint16{0x0A0A, 0x1301, 0x1A1A, 0x1302},
			expected: []uint16{0x1301, 0x1302},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterGREASEUint16(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("filterGREASEUint16(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// translated comment
func TestInternalFilterGREASECurveID(t *testing.T) {
	tests := []struct {
		name     string
		input    []tls.CurveID
		expected []tls.CurveID
	}{
		{
			name:     "internal nil slice",
			input:    nil,
			expected: []tls.CurveID{},
		},
		{
			name:     "internal empty slice",
			input:    []tls.CurveID{},
			expected: []tls.CurveID{},
		},
		{
			name:     "internal no GREASE values",
			input:    []tls.CurveID{tls.X25519, tls.CurveP256},
			expected: []tls.CurveID{tls.X25519, tls.CurveP256},
		},
		{
			name:     "internal all GREASE values",
			input:    []tls.CurveID{0x0A0A, 0x1A1A},
			expected: []tls.CurveID{},
		},
		{
			name:     "internal mixed values",
			input:    []tls.CurveID{0x0A0A, tls.X25519, 0x1A1A, tls.CurveP256},
			expected: []tls.CurveID{tls.X25519, tls.CurveP256},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterGREASECurveID(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("filterGREASECurveID(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// translated comment
func TestInternalUint16SliceToString(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint16
		expected string
	}{
		{
			name:     "internal nil slice",
			input:    nil,
			expected: "",
		},
		{
			name:     "internal empty slice",
			input:    []uint16{},
			expected: "",
		},
		{
			name:     "internal single value",
			input:    []uint16{0x1301},
			expected: "4865",
		},
		{
			name:     "internal multiple values",
			input:    []uint16{0x1301, 0x1302, 0x1303},
			expected: "4865-4866-4867",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uint16SliceToString(tt.input)
			if result != tt.expected {
				t.Errorf("uint16SliceToString(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// translated comment
func TestInternalCurveIDSliceToString(t *testing.T) {
	tests := []struct {
		name     string
		input    []tls.CurveID
		expected string
	}{
		{
			name:     "internal nil slice",
			input:    nil,
			expected: "",
		},
		{
			name:     "internal empty slice",
			input:    []tls.CurveID{},
			expected: "",
		},
		{
			name:     "internal single curve",
			input:    []tls.CurveID{tls.X25519},
			expected: "29",
		},
		{
			name:     "internal multiple curves",
			input:    []tls.CurveID{tls.X25519, tls.CurveP256, tls.CurveP384},
			expected: "29-23-24",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := curveIDSliceToString(tt.input)
			if result != tt.expected {
				t.Errorf("curveIDSliceToString(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// translated comment
func TestInternalUint8SliceToString(t *testing.T) {
	tests := []struct {
		name     string
		input    []uint8
		expected string
	}{
		{
			name:     "internal nil slice",
			input:    nil,
			expected: "",
		},
		{
			name:     "internal empty slice",
			input:    []uint8{},
			expected: "",
		},
		{
			name:     "internal single value",
			input:    []uint8{10},
			expected: "10",
		},
		{
			name:     "internal multiple values",
			input:    []uint8{10, 20, 30},
			expected: "10-20-30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uint8SliceToString(tt.input)
			if result != tt.expected {
				t.Errorf("uint8SliceToString(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
