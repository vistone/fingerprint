package ja3

import (
	"testing"

	tls "github.com/bogdanfinn/utls"
	"github.com/vistone/fingerprint/profiles"
)

// BenchmarkComputeJA3FromSpec_Simple 基准测试：简单 ClientHelloSpec
func BenchmarkComputeJA3FromSpec_Simple(b *testing.B) {
	spec := tls.ClientHelloSpec{
		CipherSuites: []uint16{0x1301},
		Extensions: []tls.TLSExtension{
			&tls.SNIExtension{},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := ComputeJA3FromSpec(spec)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkComputeJA3FromSpec_Complex 基准测试：复杂 ClientHelloSpec
func BenchmarkComputeJA3FromSpec_Complex(b *testing.B) {
	spec := tls.ClientHelloSpec{
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xC02C, 0xC02B, 0xC02F,
			0xC030, 0xC031, 0xC032, 0xCCA8, 0xCCA9,
		},
		Extensions: []tls.TLSExtension{
			&tls.SNIExtension{},
			&tls.SupportedVersionsExtension{Versions: []uint16{0x0304, 0x0303, 0x0302}},
			&tls.SupportedCurvesExtension{Curves: []tls.CurveID{tls.X25519, tls.CurveP256, tls.CurveP384}},
			&tls.SupportedPointsExtension{SupportedPoints: []byte{0x00}},
			&tls.KeyShareExtension{KeyShares: []tls.KeyShare{}},
			&tls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
			&tls.UtlsCompressCertExtension{},
			&tls.UtlsGREASEExtension{},
			&tls.StatusRequestExtension{},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := ComputeJA3FromSpec(spec)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkIsGREASEValue 基准测试：GREASE 值检测
func BenchmarkIsGREASEValue(b *testing.B) {
	values := []uint16{0x0A0A, 0x1301, 0x1A1A, 0x1302, 0x2A2A}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, v := range values {
			_ = isGREASEValue(v)
		}
	}
}

// BenchmarkFilterGREASEUint16 基准测试：GREASE 过滤
func BenchmarkFilterGREASEUint16(b *testing.B) {
	values := []uint16{0x1301, 0x0A0A, 0x1302, 0x1A1A, 0x1303, 0x2A2A}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = filterGREASEUint16(values)
	}
}

// BenchmarkMatchJA3 基准测试：JA3 哈希匹配
func BenchmarkMatchJA3(b *testing.B) {
	hash1 := "769,47-53-5-10-49161-49162-49171-49172-50-56-19-4,0-10-11,23-24-25,0"
	hash2 := "769,47-53-5-10-49161-49162-49171-49172-50-56-19-4,0-10-11,23-24-25,0"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = MatchJA3(hash1, hash2)
	}
}

// BenchmarkComputeJA3FromProfile 基准测试：从 Profile 计算 JA3
func BenchmarkComputeJA3FromProfile(b *testing.B) {
	// 尝试初始化（忽略可能已初始化的错误）
	InitMappedTLSClientsRaw(profiles.MappedTLSClients)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ComputeJA3ByProfileName("chrome_133")
	}
}

// BenchmarkParallelComputeJA3FromSpec 并发基准测试
func BenchmarkParallelComputeJA3FromSpec(b *testing.B) {
	spec := tls.ClientHelloSpec{
		CipherSuites: []uint16{0x1301, 0x1302, 0x1303},
		Extensions: []tls.TLSExtension{
			&tls.SNIExtension{},
			&tls.SupportedVersionsExtension{Versions: []uint16{0x0304, 0x0303}},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := ComputeJA3FromSpec(spec)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
