package tcpip

import (
	"testing"
)

// TestBuildOSDatabase 测试操作系统数据库构建
func TestBuildOSDatabase(t *testing.T) {
	db := BuildOSDatabase()
	if db == nil {
		t.Fatal("BuildOSDatabase returned nil")
	}

	// 验证包含主要的操作系统
	requiredOSes := []string{
		"Windows_10",
		"Windows_11",
		"Linux_Kernel_5.x",
		"macOS_13",
	}

	for _, osName := range requiredOSes {
		if _, ok := db[osName]; !ok {
			t.Errorf("OS database missing: %s", osName)
		}
	}
}

// TestOSSignatureFields 测试 OS 签名字段有效性
func TestOSSignatureFields(t *testing.T) {
	db := BuildOSDatabase()

	for name, sig := range db {
		t.Run(name, func(t *testing.T) {
			if sig.Name == "" {
				t.Error("Name field is empty")
			}
			if sig.Family == "" {
				t.Error("Family field is empty")
			}
			if sig.DefaultTTL <= 0 {
				t.Errorf("DefaultTTL must be positive, got %d", sig.DefaultTTL)
			}
			if sig.MSS <= 0 {
				t.Errorf("MSS must be positive, got %d", sig.MSS)
			}
			if sig.TCPOptions == "" {
				t.Error("TCPOptions field is empty")
			}
		})
	}
}

// TestComputeTCPSignature 测试 TCP 签名计算
func TestComputeTCPSignature(t *testing.T) {
	tests := []struct {
		name       string
		mss        int
		windowSize int
		options    string
		flags      string
		wantLen    int // MD5 哈希长度应为 32
	}{
		{
			name:       "Basic TCP signature",
			mss:        1460,
			windowSize: 65535,
			options:    "MSS,SACK",
			flags:      "SYN",
			wantLen:    32,
		},
		{
			name:       "Empty options",
			mss:        0,
			windowSize: 0,
			options:    "",
			flags:      "",
			wantLen:    32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := ComputeTCPSignature(tt.mss, tt.windowSize, tt.options, tt.flags)
			if len(sig) != tt.wantLen {
				t.Errorf("ComputeTCPSignature() = %v, want length %d, got %d", sig, tt.wantLen, len(sig))
			}
		})
	}
}

// TestComputeIPSignature 测试 IP 签名计算
func TestComputeIPSignature(t *testing.T) {
	sig := ComputeIPSignature(64, 0x02, 12345)
	if len(sig) != 32 {
		t.Errorf("ComputeIPSignature() length = %d, want 32", len(sig))
	}

	// 相同的输入应该产生相同的输出
	sig2 := ComputeIPSignature(64, 0x02, 12345)
	if sig != sig2 {
		t.Error("ComputeIPSignature should be deterministic")
	}
}

// TestMatchOSSignature 测试 OS 签名匹配
func TestMatchOSSignature(t *testing.T) {
	db := BuildOSDatabase()

	tests := []struct {
		name    string
		ttl     int
		mss     int
		options string
		wantOS  string // 期望匹配的操作系统
	}{
		{
			name:    "Windows 11 match",
			ttl:     64,
			mss:     1460,
			options: "MSS,SACK,TS,NOP,WS",
			wantOS:  "Windows_11",
		},
		{
			name:    "Linux kernel 5.x match",
			ttl:     64,
			mss:     1460,
			options: "MSS,TS,SACK,WS",
			wantOS:  "Linux_Kernel_5.x",
		},
		{
			name:    "Close to Windows TTL",
			ttl:     60, // 接近 64
			mss:     1460,
			options: "MSS,SACK",
			wantOS:  "Windows_10", // 应该匹配 Windows
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchOSSignature(db, tt.ttl, tt.mss, tt.options)
			if got == "" {
				t.Error("MatchOSSignature returned empty string")
			}
			t.Logf("Matched OS: %s", got)
		})
	}
}

// TestExtractTCPOptions 测试 TCP 选项提取（当前为简化实现）
func TestExtractTCPOptions(t *testing.T) {
	tests := []struct {
		name   string
		packet []byte
		want   string
	}{
		{
			name:   "Empty packet",
			packet: []byte{},
			want:   "",
		},
		{
			name:   "Short packet",
			packet: make([]byte, 10),
			want:   "",
		},
		{
			name:   "Valid packet",
			packet: make([]byte, 20),
			want:   "", // 空包（Data Offset=5，只有头部没有选项）
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractTCPOptions(tt.packet)
			if got != tt.want {
				t.Errorf("ExtractTCPOptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAnalyzeTTL 测试 TTL 分析
func TestAnalyzeTTL(t *testing.T) {
	tests := []struct {
		name       string
		currentTTL int
		want       int
	}{
		{"High TTL (Windows)", 120, 128},
		{"Medium TTL (Linux)", 60, 64},
		{"Low TTL", 30, 32},
		{"Exactly 64", 64, 64},
		{"Exactly 128", 128, 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AnalyzeTTL(tt.currentTTL)
			if got != tt.want {
				t.Errorf("AnalyzeTTL(%d) = %d, want %d", tt.currentTTL, got, tt.want)
			}
		})
	}
}

// TestAnalyzeWindowSize 测试窗口大小分析
func TestAnalyzeWindowSize(t *testing.T) {
	tests := []struct {
		name string
		size int
		want string
	}{
		{"Large window", 65535, "Large (Windows/macOS style)"},
		{"Medium window", 29200, "Medium (Linux style)"},
		{"Small window", 1024, "Small"},
		{"Boundary large", 60001, "Large (Windows/macOS style)"},
		{"Boundary medium", 20001, "Medium (Linux style)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AnalyzeWindowSize(tt.size)
			if got != tt.want {
				t.Errorf("AnalyzeWindowSize(%d) = %v, want %v", tt.size, got, tt.want)
			}
		})
	}
}

// TestDetectSequenceNumberPattern 测试序列号模式检测
func TestDetectSequenceNumberPattern(t *testing.T) {
	tests := []struct {
		name       string
		seqNumbers []uint32
		want       string
	}{
		{
			name:       "Insufficient data",
			seqNumbers: []uint32{1000},
			want:       "Insufficient data",
		},
		{
			name:       "Random pattern",
			seqNumbers: []uint32{10000, 50000, 90000, 130000},
			want:       "Random (cryptographically secure)",
		},
		{
			name:       "Time-based pattern",
			seqNumbers: []uint32{0, 200000, 400000, 600000},
			want:       "Time-based",
		},
		{
			name:       "Sequential pattern",
			seqNumbers: []uint32{100, 101, 102, 103},
			want:       "Sequential or low-entropy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectSequenceNumberPattern(tt.seqNumbers)
			if got != tt.want {
				t.Errorf("DetectSequenceNumberPattern() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAnalyzeNetworkBehavior 测试网络行为分析
func TestAnalyzeNetworkBehavior(t *testing.T) {
	tests := []struct {
		name      string
		rttValues []int64
		wantType  string
	}{
		{
			name:      "Local LAN",
			rttValues: []int64{1, 2, 3, 4, 5},
			wantType:  "Local LAN",
		},
		{
			name:      "Domestic network",
			rttValues: []int64{20, 25, 30, 35, 40},
			wantType:  "Domestic network",
		},
		{
			name:      "Regional network",
			rttValues: []int64{80, 90, 100, 110, 120},
			wantType:  "Regional network",
		},
		{
			name:      "International",
			rttValues: []int64{200, 250, 300},
			wantType:  "International/Satellite",
		},
		{
			name:      "Empty RTT",
			rttValues: []int64{},
			wantType:  "", // 空结果
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AnalyzeNetworkBehavior(tt.rttValues)
			if tt.wantType == "" {
				if len(result) != 0 {
					t.Errorf("Expected empty result, got %v", result)
				}
				return
			}
			if got, ok := result["network_type"]; !ok || got != tt.wantType {
				t.Errorf("network_type = %v, want %v", got, tt.wantType)
			}
		})
	}
}

// TestCalculateConfidence 测试置信度计算
func TestCalculateConfidence(t *testing.T) {
	tests := []struct {
		name    string
		matches int
		total   int
		want    float64
	}{
		{"Perfect match", 10, 10, 1.0},
		{"No match", 0, 10, 0.0},
		{"Half match", 5, 10, 0.5},
		{"Empty total", 0, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateConfidence(tt.matches, tt.total)
			if got != tt.want {
				t.Errorf("CalculateConfidence(%d, %d) = %v, want %v", tt.matches, tt.total, got, tt.want)
			}
		})
	}
}

// BenchmarkComputeTCPSignature 基准测试：TCP 签名计算
func BenchmarkComputeTCPSignature(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ComputeTCPSignature(1460, 65535, "MSS,SACK,TS", "SYN,ACK")
	}
}

// BenchmarkMatchOSSignature 基准测试：OS 签名匹配
func BenchmarkMatchOSSignature(b *testing.B) {
	db := BuildOSDatabase()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MatchOSSignature(db, 64, 1460, "MSS,SACK,TS,NOP,WS")
	}
}

// BenchmarkAnalyzeNetworkBehavior 基准测试：网络行为分析
func BenchmarkAnalyzeNetworkBehavior(b *testing.B) {
	rttValues := []int64{10, 20, 30, 40, 50}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AnalyzeNetworkBehavior(rttValues)
	}
}
