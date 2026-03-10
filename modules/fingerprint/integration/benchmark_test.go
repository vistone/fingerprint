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

// translated comment
func BenchmarkGetRandomFingerprint(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := random.GetRandomFingerprint()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// translated comment
func BenchmarkGetRandomFingerprintWithOS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := random.GetRandomFingerprintWithOS(types.OSWindows10)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// translated comment
func BenchmarkGetRandomFingerprintByBrowser(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := random.GetRandomFingerprintByBrowser("chrome")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// translated comment
func BenchmarkGetUserAgentByProfileName(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := useragent.GetUserAgentByProfileName("chrome_133")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// translated comment
func BenchmarkGenerateHeaders(b *testing.B) {
	b.ReportAllocs()
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
	for i := 0; i < b.N; i++ {
		_ = headers.GenerateHeaders(types.BrowserChrome, ua, false)
	}
}

// translated comment
func BenchmarkHeadersToMap(b *testing.B) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
	headers := headers.GenerateHeaders(types.BrowserChrome, ua, false)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = headers.ToMap()
	}
}

// translated comment
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

// translated comment
func BenchmarkRandomLanguage(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = headers.RandomLanguage()
	}
}

// translated comment
func BenchmarkRandomOS(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = useragent.RandomOS()
	}
}

// translated comment
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

// translated comment
func BenchmarkFullWorkflow(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// translated comment
		result, err := random.GetRandomFingerprint()
		if err != nil {
			b.Fatal(err)
		}

		// 2. get Client Hello Spec
		_, err = result.Profile.GetClientHelloSpec()
		if err != nil && !errors.IsClientHelloSpecNotImplemented(err) {
			b.Fatal(err)
		}

		// translated comment
		result.Headers.Set("Cookie", "session_id=test")

		// translated comment
		_ = result.Headers.ToMap()
	}
}

// translated comment
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

// translated comment
func BenchmarkParallelRandomLanguage(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = headers.RandomLanguage()
		}
	})
}

// translated comment
func BenchmarkParallelRandomOS(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = useragent.RandomOS()
		}
	})
}
