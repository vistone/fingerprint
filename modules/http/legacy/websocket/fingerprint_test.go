package websocket

import (
	"bytes"
	"net/http"
	"testing"
)

// TestAnalyzeRequest 测试请求分析
func TestAnalyzeRequest(t *testing.T) {
	analyzer := NewAnalyzer()

	t.Run("valid_websocket_request", func(t *testing.T) {
		req := createTestRequest()

		fp, err := analyzer.AnalyzeRequest(req)
		if err != nil {
			t.Fatalf("AnalyzeRequest() error = %v", err)
		}

		if fp.Version != "13" {
			t.Errorf("Version = %s, want 13", fp.Version)
		}

		if fp.Handshake.Method != "GET" {
			t.Errorf("Method = %s, want GET", fp.Handshake.Method)
		}

		if fp.Handshake.SecWebSocketVersion != "13" {
			t.Errorf("SecWebSocketVersion = %s, want 13", fp.Handshake.SecWebSocketVersion)
		}
	})

	t.Run("invalid_method", func(t *testing.T) {
		req := createTestRequest()
		req.Method = "POST"

		_, err := analyzer.AnalyzeRequest(req)
		if err == nil {
			t.Error("Expected error for non-GET method")
		}
	})

	t.Run("no_user_agent", func(t *testing.T) {
		req := createTestRequest()
		req.Header.Del("User-Agent")

		fp, err := analyzer.AnalyzeRequest(req)
		if err != nil {
			t.Fatalf("AnalyzeRequest() error = %v", err)
		}

		if fp.Handshake.UserAgent != "" {
			t.Error("Expected empty User-Agent")
		}
	})
}

// TestAnalyzeSecWebSocketKey 测试 Key 分析
func TestAnalyzeSecWebSocketKey(t *testing.T) {
	analyzer := NewAnalyzer()

	tests := []struct {
		name        string
		key         string
		wantStd     bool
		wantPattern bool
	}{
		{
			name:        "standard_key",
			key:         "dGhlIHNhbXBsZSBub25jZQ==", // 16 bytes base64
			wantStd:     true,
			wantPattern: false,
		},
		{
			name:        "invalid_base64",
			key:         "not-valid-base64!!!",
			wantStd:     false,
			wantPattern: false,
		},
		{
			name:        "wrong_length",
			key:         "dG9vLXNob3J0", // too short
			wantStd:     false,
			wantPattern: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest()
			req.Header.Set("Sec-Websocket-Key", tt.key)

			fp, err := analyzer.AnalyzeRequest(req)
			if err != nil {
				t.Fatalf("AnalyzeRequest() error = %v", err)
			}

			if fp.Handshake.SecWebSocketKeyCharacteristics.IsStandardBase64 != tt.wantStd {
				t.Errorf("IsStandardBase64 = %v, want %v",
					fp.Handshake.SecWebSocketKeyCharacteristics.IsStandardBase64, tt.wantStd)
			}
		})
	}
}

// TestAnalyzeExtensions 测试扩展分析
func TestAnalyzeExtensions(t *testing.T) {
	analyzer := NewAnalyzer()

	t.Run("permessage_deflate", func(t *testing.T) {
		req := createTestRequest()
		req.Header.Set("Sec-Websocket-Extensions", "permessage-deflate; client_max_window_bits")

		fp, err := analyzer.AnalyzeRequest(req)
		if err != nil {
			t.Fatalf("AnalyzeRequest() error = %v", err)
		}

		found := false
		for _, ext := range fp.Extensions {
			if ext == "permessage-deflate" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected permessage-deflate extension")
		}

		foundFlag := false
		for _, flag := range fp.FrameCharacteristics.ExtensionFlags {
			if flag == "compression" {
				foundFlag = true
				break
			}
		}
		if !foundFlag {
			t.Error("Expected compression flag")
		}
	})

	t.Run("multiple_extensions", func(t *testing.T) {
		req := createTestRequest()
		req.Header.Set("Sec-Websocket-Extensions", "permessage-deflate, x-webkit-deflate-frame")

		fp, err := analyzer.AnalyzeRequest(req)
		if err != nil {
			t.Fatalf("AnalyzeRequest() error = %v", err)
		}

		if len(fp.Extensions) != 2 {
			t.Errorf("len(Extensions) = %d, want 2", len(fp.Extensions))
		}
	})
}

// TestIdentifyBrowser 测试浏览器识别
func TestIdentifyBrowser(t *testing.T) {
	analyzer := NewAnalyzer()

	tests := []struct {
		name      string
		ua        string
		wantEmpty bool
	}{
		{
			name: "chrome",
			ua:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		},
		{
			name: "firefox",
			ua:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
		},
		{
			name:      "empty",
			ua:        "",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest()
			if tt.ua != "" {
				req.Header.Set("User-Agent", tt.ua)
			}

			fp, err := analyzer.AnalyzeRequest(req)
			if err != nil {
				t.Fatalf("AnalyzeRequest() error = %v", err)
			}

			browser, confidence := analyzer.IdentifyBrowser(fp)

			if tt.wantEmpty {
				if browser != "unknown" {
					t.Errorf("Browser = %s, want unknown", browser)
				}
			} else {
				if browser == "unknown" {
					t.Errorf("Expected browser identification, got unknown")
				}
				if confidence <= 0 {
					t.Error("Expected positive confidence")
				}
			}
		})
	}
}

// TestCompareFingerprints 测试指纹比较
func TestCompareFingerprints(t *testing.T) {
	fp1 := &WebSocketFingerprint{
		Handshake: WebSocketHandshake{
			HTTPVersion: "HTTP/1.1",
			HeaderOrder: []string{"Host", "Upgrade"},
		},
		Extensions: []string{"permessage-deflate"},
	}

	fp2 := &WebSocketFingerprint{
		Handshake: WebSocketHandshake{
			HTTPVersion: "HTTP/1.1",
			HeaderOrder: []string{"Host", "Upgrade"},
		},
		Extensions: []string{"permessage-deflate"},
	}

	fp3 := &WebSocketFingerprint{
		Handshake: WebSocketHandshake{
			HTTPVersion: "HTTP/2.0",
			HeaderOrder: []string{"Upgrade", "Host"},
		},
		Extensions: []string{"x-webkit-deflate-frame"},
	}

	t.Run("identical", func(t *testing.T) {
		similarity := CompareFingerprints(fp1, fp2)
		if similarity != 1.0 {
			t.Errorf("Similarity = %f, want 1.0", similarity)
		}
	})

	t.Run("different", func(t *testing.T) {
		similarity := CompareFingerprints(fp1, fp3)
		if similarity > 0.5 {
			t.Errorf("Similarity = %f, expected <= 0.5", similarity)
		}
	})

	t.Run("nil", func(t *testing.T) {
		similarity := CompareFingerprints(nil, fp1)
		if similarity != 0.0 {
			t.Errorf("Similarity = %f, want 0.0", similarity)
		}
	})
}

// TestGenerateAcceptKey 测试 Accept Key 生成
func TestGenerateAcceptKey(t *testing.T) {
	// RFC 6455 测试向量
	key := "dGhlIHNhbXBsZSBub25jZQ=="
	expected := "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="

	accept, err := GenerateAcceptKey(key)
	if err != nil {
		t.Fatalf("GenerateAcceptKey() error = %v", err)
	}

	if accept != expected {
		t.Errorf("GenerateAcceptKey() = %s, want %s", accept, expected)
	}
}

// TestIsValidWebSocketRequest 测试请求验证
func TestIsValidWebSocketRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *http.Request
		want bool
	}{
		{
			name: "valid",
			req:  createTestRequest(),
			want: true,
		},
		{
			name: "wrong_method",
			req: func() *http.Request {
				r := createTestRequest()
				r.Method = "POST"
				return r
			}(),
			want: false,
		},
		{
			name: "no_upgrade",
			req: func() *http.Request {
				r := createTestRequest()
				r.Header.Del("Upgrade")
				return r
			}(),
			want: false,
		},
		{
			name: "no_key",
			req: func() *http.Request {
				r := createTestRequest()
				r.Header.Del("Sec-Websocket-Key")
				return r
			}(),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidWebSocketRequest(tt.req)
			if got != tt.want {
				t.Errorf("IsValidWebSocketRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseFrame 测试帧解析
func TestParseFrame(t *testing.T) {
	t.Run("text_frame", func(t *testing.T) {
		// FIN=1, Opcode=1 (text), MASK=1, Length=5
		frame := []byte{
			0x81,                   // FIN=1, Opcode=1
			0x85,                   // MASK=1, Length=5
			0x00, 0x00, 0x00, 0x00, // Masking key (all zeros for test)
			'H', 'e', 'l', 'l', 'o',
		}

		f, err := ParseFrame(frame)
		if err != nil {
			t.Fatalf("ParseFrame() error = %v", err)
		}

		if !f.FIN {
			t.Error("Expected FIN=1")
		}
		if f.Opcode != OpCodeText {
			t.Errorf("Opcode = %d, want %d", f.Opcode, OpCodeText)
		}
		if !f.MASK {
			t.Error("Expected MASK=1")
		}
		if f.PayloadLength != 5 {
			t.Errorf("PayloadLength = %d, want 5", f.PayloadLength)
		}

		expected := []byte("Hello")
		if !bytes.Equal(f.Payload, expected) {
			t.Errorf("Payload = %v, want %v", f.Payload, expected)
		}
	})

	t.Run("ping_frame", func(t *testing.T) {
		// FIN=1, Opcode=9 (ping)
		frame := []byte{
			0x89, // FIN=1, Opcode=9
			0x00, // MASK=0, Length=0
		}

		f, err := ParseFrame(frame)
		if err != nil {
			t.Fatalf("ParseFrame() error = %v", err)
		}

		if f.Opcode != OpCodePing {
			t.Errorf("Opcode = %d, want %d", f.Opcode, OpCodePing)
		}
	})

	t.Run("extended_length_16", func(t *testing.T) {
		// FIN=1, Opcode=2 (binary), Length=126 (indicates 16-bit length)
		payload := make([]byte, 100)
		frame := []byte{
			0x82,       // FIN=1, Opcode=2
			0x7e,       // MASK=0, Length=126
			0x00, 0x64, // Actual length: 100
		}
		frame = append(frame, payload...)

		f, err := ParseFrame(frame)
		if err != nil {
			t.Fatalf("ParseFrame() error = %v", err)
		}

		if f.PayloadLength != 100 {
			t.Errorf("PayloadLength = %d, want 100", f.PayloadLength)
		}
	})

	t.Run("truncated", func(t *testing.T) {
		frame := []byte{0x81} // Only 1 byte

		_, err := ParseFrame(frame)
		if err == nil {
			t.Error("Expected error for truncated frame")
		}
	})
}

// TestAnalyzeFrame 测试帧分析
func TestAnalyzeFrame(t *testing.T) {
	frame := &Frame{
		FIN:           true,
		Opcode:        OpCodeText,
		MASK:          true,
		PayloadLength: 100,
	}

	features := AnalyzeFrame(frame)

	if features["opcode"] != uint8(OpCodeText) {
		t.Errorf("opcode = %v, want %d", features["opcode"], OpCodeText)
	}

	if features["frame_type"] != "text" {
		t.Errorf("frame_type = %v, want text", features["frame_type"])
	}
}

// TestAnalyzeFrameStream 测试帧流分析
func TestAnalyzeFrameStream(t *testing.T) {
	frames := []*Frame{
		{Opcode: OpCodeText, MASK: true, PayloadLength: 100},
		{Opcode: OpCodeText, MASK: true, PayloadLength: 200},
		{Opcode: OpCodeBinary, MASK: true, PayloadLength: 50},
		{Opcode: OpCodePing, MASK: true, PayloadLength: 0},
		{Opcode: OpCodePong, MASK: false, PayloadLength: 0},
	}

	features := AnalyzeFrameStream(frames)

	if features["total_frames"] != 5 {
		t.Errorf("total_frames = %v, want 5", features["total_frames"])
	}

	if features["masked_ratio"] != 0.8 {
		t.Errorf("masked_ratio = %v, want 0.8", features["masked_ratio"])
	}

	if features["max_payload"] != uint64(200) {
		t.Errorf("max_payload = %v, want 200", features["max_payload"])
	}
}

// BenchmarkAnalyzeRequest 基准测试
func BenchmarkAnalyzeRequest(b *testing.B) {
	analyzer := NewAnalyzer()
	req := createTestRequest()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeRequest(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseFrame 帧解析基准测试
func BenchmarkParseFrame(b *testing.B) {
	frame := []byte{
		0x81, 0x85,
		0x00, 0x00, 0x00, 0x00,
		'H', 'e', 'l', 'l', 'o',
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ParseFrame(frame)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// createTestRequest 创建测试请求
func createTestRequest() *http.Request {
	req, _ := http.NewRequest("GET", "ws://example.com/ws", nil)
	req.Header.Set("Host", "example.com")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-Websocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Sec-Websocket-Version", "13")
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Test)")
	return req
}

// TestJaccardSimilarity 测试 Jaccard 相似度
func TestJaccardSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want float64
	}{
		{
			name: "identical",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b", "c"},
			want: 1.0,
		},
		{
			name: "half_match",
			a:    []string{"a", "b"},
			b:    []string{"a", "c"},
			want: 0.333333,
		},
		{
			name: "empty",
			a:    []string{},
			b:    []string{},
			want: 1.0,
		},
		{
			name: "one_empty",
			a:    []string{"a"},
			b:    []string{},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := jaccardSimilarity(tt.a, tt.b)
			if abs(got-tt.want) > 0.001 {
				t.Errorf("jaccardSimilarity() = %f, want %f", got, tt.want)
			}
		})
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// TestParseExtensionHeader 测试扩展头部解析
func TestParseExtensionHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   []string
	}{
		{
			name:   "single",
			header: "permessage-deflate",
			want:   []string{"permessage-deflate"},
		},
		{
			name:   "with_params",
			header: "permessage-deflate; client_max_window_bits",
			want:   []string{"permessage-deflate"},
		},
		{
			name:   "multiple",
			header: "permessage-deflate, x-webkit-deflate-frame",
			want:   []string{"permessage-deflate", "x-webkit-deflate-frame"},
		},
		{
			name:   "empty",
			header: "",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseExtensionHeader(tt.header)
			if len(got) != len(tt.want) {
				t.Errorf("len = %d, want %d", len(got), len(tt.want))
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("got[%d] = %s, want %s", i, got[i], tt.want[i])
				}
			}
		})
	}
}
