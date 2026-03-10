package core

import (
	"testing"
	"time"
)

// TestRiskLevelFromScore 验证风险等级计算
func TestRiskLevelFromScore(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		expected RiskLevel
	}{
		{"none", 0.0, RiskLevelNone},
		{"none_boundary", 0.09, RiskLevelNone},
		{"low", 0.1, RiskLevelLow},
		{"low_mid", 0.25, RiskLevelLow},
		{"medium", 0.4, RiskLevelMedium},
		{"medium_mid", 0.55, RiskLevelMedium},
		{"high", 0.7, RiskLevelHigh},
		{"high_mid", 0.8, RiskLevelHigh},
		{"critical", 0.9, RiskLevelCritical},
		{"critical_max", 1.0, RiskLevelCritical},
		{"negative", -0.5, RiskLevelNone}, // 边界情况
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RiskLevelFromScore(tt.score)
			if result != tt.expected {
				t.Errorf("RiskLevelFromScore(%v) = %v, want %v", tt.score, result, tt.expected)
			}
		})
	}
}

// TestRiskLevelString 验证 RiskLevel 字符串表示
func TestRiskLevelString(t *testing.T) {
	tests := []struct {
		level    RiskLevel
		expected string
	}{
		{RiskLevelNone, "none"},
		{RiskLevelLow, "low"},
		{RiskLevelMedium, "medium"},
		{RiskLevelHigh, "high"},
		{RiskLevelCritical, "critical"},
		{RiskLevel(999), "unknown"}, // 无效值
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("RiskLevel(%d).String() = %q, want %q", tt.level, result, tt.expected)
			}
		})
	}
}

// TestConstantValues 验证关键常量值
func TestConstantValues(t *testing.T) {
	// 时间常量
	if DefaultTimeout != 30*time.Second {
		t.Errorf("DefaultTimeout = %v, want %v", DefaultTimeout, 30*time.Second)
	}
	if DefaultDialTimeout != 10*time.Second {
		t.Errorf("DefaultDialTimeout = %v, want %v", DefaultDialTimeout, 10*time.Second)
	}
	if DefaultTLSTimeout != 15*time.Second {
		t.Errorf("DefaultTLSTimeout = %v, want %v", DefaultTLSTimeout, 15*time.Second)
	}

	// 缓存常量
	if DefaultCacheSize != 10000 {
		t.Errorf("DefaultCacheSize = %d, want %d", DefaultCacheSize, 10000)
	}
	if DefaultCacheTTL != 5*time.Minute {
		t.Errorf("DefaultCacheTTL = %v, want %v", DefaultCacheTTL, 5*time.Minute)
	}

	// 限流常量
	if DefaultRateLimit != 1000 {
		t.Errorf("DefaultRateLimit = %d, want %d", DefaultRateLimit, 1000)
	}
	if DefaultRateLimitBurst != 2000 {
		t.Errorf("DefaultRateLimitBurst = %d, want %d", DefaultRateLimitBurst, 2000)
	}

	// 风险阈值
	if RiskThresholdLow != 0.1 {
		t.Errorf("RiskThresholdLow = %v, want %v", RiskThresholdLow, 0.1)
	}
	if RiskThresholdMedium != 0.4 {
		t.Errorf("RiskThresholdMedium = %v, want %v", RiskThresholdMedium, 0.4)
	}
	if RiskThresholdHigh != 0.7 {
		t.Errorf("RiskThresholdHigh = %v, want %v", RiskThresholdHigh, 0.7)
	}
	if RiskThresholdCritical != 0.9 {
		t.Errorf("RiskThresholdCritical = %v, want %v", RiskThresholdCritical, 0.9)
	}

	// 大小常量
	if Size1KB != 1024 {
		t.Errorf("Size1KB = %d, want %d", Size1KB, 1024)
	}
	if Size1MB != 1024*1024 {
		t.Errorf("Size1MB = %d, want %d", Size1MB, 1024*1024)
	}
	if MaxRequestBodySize != 5*1024*1024 {
		t.Errorf("MaxRequestBodySize = %d, want %d", MaxRequestBodySize, 5*1024*1024)
	}
}

// TestLoggerAdapters 验证 Logger 适配器
func TestLoggerAdapters(t *testing.T) {
	t.Run("NoOpLogger", func(t *testing.T) {
		logger := NoOpLogger{}
		// 不应 panic
		logger.Debug("debug message", "key", "value")
		logger.Info("info message", "key", "value")
		logger.Warn("warn message", "key", "value")
		logger.Error("error message", "key", "value")
	})

	t.Run("SlogAdapter", func(t *testing.T) {
		logger := NewDefaultLogger("debug")
		// 不应 panic
		logger.Debug("debug message", "key", "value")
		logger.Info("info message", "key", "value")
		logger.Warn("warn message", "key", "value")
		logger.Error("error message", "key", "value")
	})
}

// TestTLSVersionConstants 验证 TLS 版本常量
func TestTLSVersionConstants(t *testing.T) {
	if TLSVersion10 != 0x0301 {
		t.Errorf("TLSVersion10 = 0x%04x, want 0x0301", TLSVersion10)
	}
	if TLSVersion11 != 0x0302 {
		t.Errorf("TLSVersion11 = 0x%04x, want 0x0302", TLSVersion11)
	}
	if TLSVersion12 != 0x0303 {
		t.Errorf("TLSVersion12 = 0x%04x, want 0x0303", TLSVersion12)
	}
	if TLSVersion13 != 0x0304 {
		t.Errorf("TLSVersion13 = 0x%04x, want 0x0304", TLSVersion13)
	}
}

// BenchmarkRiskLevelFromScore 性能基准测试
func BenchmarkRiskLevelFromScore(b *testing.B) {
	scores := []float64{0.05, 0.15, 0.45, 0.75, 0.95}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, score := range scores {
			RiskLevelFromScore(score)
		}
	}
}
