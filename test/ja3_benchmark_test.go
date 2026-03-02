package fingerprint_test

import (
	"testing"

	"github.com/vistone/fingerprint"
)

// BenchmarkFindProfileByJA3NoCopy 基准测试：内部零拷贝查找路径
func BenchmarkFindProfileByJA3NoCopy(b *testing.B) {
	// Use a fixed JA3 hash to avoid relying on a missing helper.
	const ja3Hash = "d41d8cd98f00b204e9800998ecf8427e"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fingerprint.FindProfileByJA3(ja3Hash)
	}
}
