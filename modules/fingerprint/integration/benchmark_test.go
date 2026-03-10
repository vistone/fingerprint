package fingerprint_test

import (
	"testing"

	"github.com/vistone/fingerprint/modules/generator/random"
	"github.com/vistone/fingerprint/modules/http/legacy/headers"
	"github.com/vistone/fingerprint/modules/http/legacy/useragent"
	"github.com/vistone/fingerprint/modules/errors"
	"github.com/vistone/fingerprint/modules/profiles/legacy"
	"github.com/vistone/fingerprint/modules/core/types"
)

// BenchmarkGetRandomFingerprint 基准测试：随机get指纹
func BenchmarkGetRandomFingerprint(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := random.GetRandomFingerprint()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetRandomFingerprintWithOS 基准测试：随机get指纹（指定 OS）
func BenchmarkGetRandomFingerprintWithOS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := random.GetRandomFingerprintWithOS(types.OSWindows10)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetRandomFingerprintByBrowser 基准测试：按浏览器type随机get指纹
func BenchmarkGetRandomFingerprintByBrowser(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := random.GetRandomFingerprintByBrowser("chrome")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetUserAgentByProfileName 基准测试：根据 profile 名称get User-Agent
func BenchmarkGetUserAgentByProfileName(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := useragent.GetUserAgentByProfileName("chrome_133")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateHeaders 基准测试：generate HTTP Headers
func BenchmarkGenerateHeaders(b *testing.B) {
	b.ReportAllocs()
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
	for i := 0; i < b.N; i++ {
		_ = headers.GenerateHeaders(types.BrowserChrome, ua, false)
	}
}

// BenchmarkHeadersToMap 基准测试：Headers convert为 Map
func BenchmarkHeadersToMap(b *testing.B) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
	headers := headers.GenerateHeaders(types.BrowserChrome, ua, false)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = headers.ToMap()
	}
}

// BenchmarkHeadersClone 基准测试：Headers 克隆
func BenchmarkHeadersClone(b *testing.B) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
	headers := headers.GenerateHeaders(types.BrowserChrome, ua, false)
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
		_ = headers.RandomLanguage()
	}
}

// BenchmarkRandomOS 基准测试：随机选择操作系统
func BenchmarkRandomOS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = useragent.RandomOS()
	}
}

// BenchmarkGetClientHelloSpec 基准测试：get Client Hello Spec
func BenchmarkGetClientHelloSpec(b *testing.B) {
	profile := profiles.DefaultClientProfile
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
		// 1. get随机指纹
		result, err := random.GetRandomFingerprint()
		if err != nil {
			b.Fatal(err)
		}

		// 2. get Client Hello Spec
		_, err = result.Profile.GetClientHelloSpec()
		if err != nil && !errors.IsClientHelloSpecNotImplemented(err) {
			b.Fatal(err)
		}

		// 3. setting自定义 Headers
		result.Headers.Set("Cookie", "session_id=test")

		// 4. convert为 Map
		_ = result.Headers.ToMap()
	}
}

// BenchmarkParallelGetRandomFingerprint concurrent基准测试：随机get指纹
func BenchmarkParallelGetRandomFingerprint(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := random.GetRandomFingerprint()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkParallelRandomLanguage concurrent基准测试：随机选择语言
func BenchmarkParallelRandomLanguage(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = headers.RandomLanguage()
		}
	})
}

// BenchmarkParallelRandomOS concurrent基准测试：随机选择操作系统
func BenchmarkParallelRandomOS(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = useragent.RandomOS()
		}
	})
}
