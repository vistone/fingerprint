package fingerprint_test

import (
	"testing"

	"github.com/vistone/fingerprint"
)

// BenchmarkGetRandomFingerprint 基准测试：随机获取指纹
func BenchmarkGetRandomFingerprint(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := fingerprint.GetRandomFingerprint()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetRandomFingerprintWithOS 基准测试：随机获取指纹（指定 OS）
func BenchmarkGetRandomFingerprintWithOS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := fingerprint.GetRandomFingerprintWithOS(fingerprint.OSWindows10)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetRandomFingerprintByBrowser 基准测试：按浏览器类型随机获取指纹
func BenchmarkGetRandomFingerprintByBrowser(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := fingerprint.GetRandomFingerprintByBrowser("chrome")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetUserAgentByProfileName 基准测试：根据 profile 名称获取 User-Agent
func BenchmarkGetUserAgentByProfileName(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := fingerprint.GetUserAgentByProfileName("chrome_133")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateHeaders 基准测试：生成 HTTP Headers
func BenchmarkGenerateHeaders(b *testing.B) {
	b.ReportAllocs()
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
	for i := 0; i < b.N; i++ {
		_ = fingerprint.GenerateHeaders(fingerprint.BrowserChrome, ua, false)
	}
}

// BenchmarkHeadersToMap 基准测试：Headers 转换为 Map
func BenchmarkHeadersToMap(b *testing.B) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
	headers := fingerprint.GenerateHeaders(fingerprint.BrowserChrome, ua, false)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = headers.ToMap()
	}
}

// BenchmarkHeadersClone 基准测试：Headers 克隆
func BenchmarkHeadersClone(b *testing.B) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
	headers := fingerprint.GenerateHeaders(fingerprint.BrowserChrome, ua, false)
	headers.Set("Cookie", "session_id=abc123")
	headers.Set("Authorization", "Bearer token")
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = headers.Clone()
	}
}

// BenchmarkRandomLanguage 基准测试：随机选择语言
func BenchmarkRandomLanguage(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fingerprint.RandomLanguage()
	}
}

// BenchmarkRandomOS 基准测试：随机选择操作系统
func BenchmarkRandomOS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fingerprint.RandomOS()
	}
}

// BenchmarkGetClientHelloSpec 基准测试：获取 Client Hello Spec
func BenchmarkGetClientHelloSpec(b *testing.B) {
	profile := fingerprint.DefaultClientProfile
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := profile.GetClientHelloSpec()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFullWorkflow 基准测试：完整工作流程
func BenchmarkFullWorkflow(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// 1. 获取随机指纹
		result, err := fingerprint.GetRandomFingerprint()
		if err != nil {
			b.Fatal(err)
		}
		
		// 2. 获取 Client Hello Spec
		_, err = result.Profile.GetClientHelloSpec()
		if err != nil && err.Error() != "please implement this method" {
			b.Fatal(err)
		}
		
		// 3. 设置自定义 Headers
		result.Headers.Set("Cookie", "session_id=test")
		
		// 4. 转换为 Map
		_ = result.Headers.ToMap()
	}
}

// BenchmarkParallelGetRandomFingerprint 并发基准测试：随机获取指纹
func BenchmarkParallelGetRandomFingerprint(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := fingerprint.GetRandomFingerprint()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkParallelRandomLanguage 并发基准测试：随机选择语言
func BenchmarkParallelRandomLanguage(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = fingerprint.RandomLanguage()
		}
	})
}

// BenchmarkParallelRandomOS 并发基准测试：随机选择操作系统
func BenchmarkParallelRandomOS(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = fingerprint.RandomOS()
		}
	})
}
