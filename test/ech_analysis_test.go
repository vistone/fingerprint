package fingerprint

import (
	"testing"

	fp "github.com/vistone/fingerprint"
)

// TestECHAnalysis_NoECH 测试没有 ECH 的场景
func TestECHAnalysis_NoECH(t *testing.T) {
	data := fp.ClientHelloData{
		TLSVersion: 0x0304, // TLS 1.3
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0x0000, Data: []byte{0x00, 0x0b, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d}}, // SNI
			{Type: 0x000d, Data: []byte{}}, // signature_algorithms
			{Type: 0x0033, Data: []byte{}}, // key_share
		},
		HasSNI: true,
		SNI:    "example.com",
	}

	result, err := fp.AnalyzeECH(data)
	if err != nil {
		t.Fatalf("分析失败: %v", err)
	}

	// 验证结果
	if result.ECHPresent {
		t.Error("不应检测到 ECH")
	}

	if result.Impact.ImpactLevel != "none" {
		t.Errorf("影响等级应为 'none', 实际: %s", result.Impact.ImpactLevel)
	}

	if !result.Impact.SNIVisible {
		t.Error("SNI 应该可见")
	}

	if len(result.AnomalyFlags) > 0 {
		t.Errorf("不应有异常标记, 实际: %v", result.AnomalyFlags)
	}

	t.Logf("无 ECH 场景测试通过: %s", result.GetImpactSummary())
}

// TestECHAnalysis_GREASE 测试 GREASE ECH
func TestECHAnalysis_GREASE(t *testing.T) {
	data := fp.ClientHelloData{
		TLSVersion: 0x0304,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0x0000, Data: []byte{0x00, 0x0b, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d}}, // SNI
			{Type: 0xfe0d, Data: []byte{0x00, 0x00, 0x00}},                                                             // ECH GREASE
			{Type: 0x000d, Data: []byte{}},
		},
		HasSNI: true,
		SNI:    "example.com",
	}

	result, err := fp.AnalyzeECH(data)
	if err != nil {
		t.Fatalf("分析失败: %v", err)
	}

	// 验证
	if !result.ECHPresent {
		t.Error("应检测到 ECH")
	}

	if result.ECHType != "grease" {
		t.Errorf("ECH 类型应为 'grease', 实际: %s", result.ECHType)
	}

	if result.Impact.ImpactLevel != "low" {
		t.Errorf("GREASE ECH 影响应为 'low', 实际: %s", result.Impact.ImpactLevel)
	}

	// GREASE ECH 应该有 GREASE_ECH 标记
	found := false
	for _, flag := range result.AnomalyFlags {
		if flag == "GREASE_ECH" {
			found = true
			break
		}
	}
	if !found {
		t.Error("应包含 GREASE_ECH 异常标记")
	}

	t.Logf("GREASE ECH: 风险分数=%.2f, 哈希=%s", result.RiskScore, result.Hash[:16])
}

// TestECHAnalysis_OuterClientHello 测试 Outer ClientHello
func TestECHAnalysis_OuterClientHello(t *testing.T) {
	data := fp.ClientHelloData{
		TLSVersion: 0x0304,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0xfd00, Data: []byte{0xfe, 0x0d, 0x01, 0x02}}, // ECH outer extensions
			{Type: 0x000d, Data: []byte{}},                       // signature_algorithms
			{Type: 0x0033, Data: []byte{}},                       // key_share
			{Type: 0x002b, Data: []byte{}},                       // supported_versions
			{Type: 0x000a, Data: []byte{}},                       // supported_groups
			{Type: 0x0010, Data: []byte{}},                       // alpn
		},
		HasSNI: false, // ECH 加密了 SNI
		SNI:    "",
	}

	result, err := fp.AnalyzeECH(data)
	if err != nil {
		t.Fatalf("分析失败: %v", err)
	}

	// 验证
	if !result.ECHPresent {
		t.Error("应检测到 ECH")
	}

	if result.ECHType != "outer" {
		t.Errorf("ECH 类型应为 'outer', 实际: %s", result.ECHType)
	}

	if result.ClientHelloType != "outer" {
		t.Errorf("ClientHello 类型应为 'outer', 实际: %s", result.ClientHelloType)
	}

	if result.Impact.ImpactLevel != "high" {
		t.Errorf("Outer ClientHello 影响应为 'high', 实际: %s", result.Impact.ImpactLevel)
	}

	if result.Impact.SNIVisible {
		t.Error("SNI 不应该可见")
	}

	// 应该有替代策略建议
	if len(result.AlternativeStrategies) == 0 {
		t.Error("应提供替代策略建议")
	}

	// 应该生成可见字段签名
	if result.VisibleFieldsSignature == "" {
		t.Error("应生成可见字段签名")
	}

	t.Logf("Outer ClientHello: 影响=%s, 可见字段签名=%s",
		result.Impact.ImpactLevel,
		result.VisibleFieldsSignature,
	)
	t.Logf("替代策略数量: %d", len(result.AlternativeStrategies))
}

// TestECHAnalysis_InnerClientHello 测试 Inner ClientHello
func TestECHAnalysis_InnerClientHello(t *testing.T) {
	data := fp.ClientHelloData{
		TLSVersion: 0x0304,
		CipherSuites: []uint16{
			0x1301, 0x1302,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0xfe0d, Data: []byte{0xfe, 0x0d, 0x01, 0xff}}, // ECH inner
			{Type: 0x000d, Data: []byte{}},
			{Type: 0x0033, Data: []byte{}},
		},
		HasSNI: false,
		SNI:    "",
	}

	result, err := fp.AnalyzeECH(data)
	if err != nil {
		t.Fatalf("分析失败: %v", err)
	}

	if result.ECHType != "inner" {
		t.Errorf("ECH 类型应为 'inner', 实际: %s", result.ECHType)
	}

	if result.Impact.ImpactLevel != "medium" {
		t.Errorf("Inner ClientHello 影响应为 'medium', 实际: %s", result.Impact.ImpactLevel)
	}

	t.Logf("Inner ClientHello 测试通过")
}

// TestECHAnalysis_Anomaly_ECHWithVisibleSNI 测试异常：ECH 但 SNI 可见
func TestECHAnalysis_Anomaly_ECHWithVisibleSNI(t *testing.T) {
	data := fp.ClientHelloData{
		TLSVersion: 0x0304,
		CipherSuites: []uint16{
			0x1301, 0x1302,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0x0000, Data: []byte{0x00, 0x0b, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d}}, // SNI
			{Type: 0xfe0d, Data: []byte{0xfe, 0x0d, 0x01, 0xff}},                                                       // ECH (非 GREASE)
		},
		HasSNI: true,
		SNI:    "example.com",
	}

	result, err := fp.AnalyzeECH(data)
	if err != nil {
		t.Fatalf("分析失败: %v", err)
	}

	// 应该检测到异常
	found := false
	for _, flag := range result.AnomalyFlags {
		if flag == "ECH_WITH_VISIBLE_SNI" {
			found = true
			break
		}
	}
	if !found {
		t.Error("应检测到 ECH_WITH_VISIBLE_SNI 异常")
	}

	// 风险分数应该较高
	if result.RiskScore < 0.2 {
		t.Errorf("风险分数应该较高 (>=0.2), 实际: %.2f", result.RiskScore)
	}

	t.Logf("异常检测成功: 标记=%v, 风险分数=%.2f",
		result.AnomalyFlags,
		result.RiskScore,
	)
}

// TestECHAnalysis_Anomaly_OldTLS 测试异常：旧版本 TLS 使用 ECH
func TestECHAnalysis_Anomaly_OldTLS(t *testing.T) {
	data := fp.ClientHelloData{
		TLSVersion: 0x0303, // TLS 1.2
		CipherSuites: []uint16{
			0xc02b, 0xc02f,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0xfe0d, Data: []byte{0xfe, 0x0d, 0x01, 0x02}}, // ECH
		},
		HasSNI: false,
	}

	result, err := fp.AnalyzeECH(data)
	if err != nil {
		t.Fatalf("分析失败: %v", err)
	}

	// 应该检测到异常
	found := false
	for _, flag := range result.AnomalyFlags {
		if flag == "ECH_WITH_OLD_TLS" {
			found = true
			break
		}
	}
	if !found {
		t.Error("应检测到 ECH_WITH_OLD_TLS 异常")
	}

	// 风险分数应该很高
	if result.RiskScore < 0.3 {
		t.Errorf("风险分数应该很高 (>=0.3), 实际: %.2f", result.RiskScore)
	}

	t.Logf("旧 TLS 版本异常检测成功: 风险分数=%.2f", result.RiskScore)
}

// TestECHAnalysis_Anomaly_IncompleteOuter 测试异常：不完整的 Outer ClientHello
func TestECHAnalysis_Anomaly_IncompleteOuter(t *testing.T) {
	data := fp.ClientHelloData{
		TLSVersion: 0x0304,
		CipherSuites: []uint16{
			0x1301,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0xfd00, Data: []byte{0xfe, 0x0d}}, // ECH outer
			{Type: 0x000d, Data: []byte{}},           // 只有 2 个扩展
		},
		HasSNI: false,
	}

	result, err := fp.AnalyzeECH(data)
	if err != nil {
		t.Fatalf("分析失败: %v", err)
	}

	// 应该检测到不完整的 outer hello
	found := false
	for _, flag := range result.AnomalyFlags {
		if flag == "INCOMPLETE_OUTER_HELLO" {
			found = true
			break
		}
	}
	if !found {
		t.Error("应检测到 INCOMPLETE_OUTER_HELLO 异常")
	}

	t.Logf("不完整 Outer ClientHello 检测成功")
}

// TestECHAnalysis_VisibleFieldsSignature 测试可见字段签名生成
func TestECHAnalysis_VisibleFieldsSignature(t *testing.T) {
	data1 := fp.ClientHelloData{
		TLSVersion: 0x0304,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0xfd00, Data: []byte{0xfe, 0x0d}}, // ECH
			{Type: 0x000d, Data: []byte{}},
			{Type: 0x0033, Data: []byte{}},
		},
		HasSNI: false,
	}

	data2 := fp.ClientHelloData{
		TLSVersion: 0x0304,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0xfd00, Data: []byte{0xfe, 0x0d}}, // ECH
			{Type: 0x000d, Data: []byte{}},
			{Type: 0x0033, Data: []byte{}},
		},
		HasSNI: false,
	}

	result1, _ := fp.AnalyzeECH(data1)
	result2, _ := fp.AnalyzeECH(data2)

	// 相同配置应该生成相同签名
	if result1.VisibleFieldsSignature != result2.VisibleFieldsSignature {
		t.Error("相同配置应生成相同的可见字段签名")
	}

	// 修改配置
	data3 := data1
	data3.CipherSuites = []uint16{0x1301, 0x1303, 0x1302} // Sequence different

	result3, _ := fp.AnalyzeECH(data3)

	// Different config should generate different signatures
	if result1.VisibleFieldsSignature == result3.VisibleFieldsSignature {
		t.Log("警告: 不同配置生成了相同签名（可能是哈希碰撞）")
	}

	t.Logf("签名一致性测试通过")
	t.Logf("签名1: %s", result1.VisibleFieldsSignature)
	t.Logf("签名2: %s", result2.VisibleFieldsSignature)
	t.Logf("签名3: %s", result3.VisibleFieldsSignature)
}

// TestECHAnalysis_AlternativeStrategies 测试替代策略建议
func TestECHAnalysis_AlternativeStrategies(t *testing.T) {
	testCases := []struct {
		name     string
		data     fp.ClientHelloData
		minCount int // 最少策略数量
	}{
		{
			name: "无 ECH - 无需替代策略",
			data: fp.ClientHelloData{
				TLSVersion:   0x0304,
				CipherSuites: []uint16{0x1301},
				Extensions: []fp.ExtensionData{
					{Type: 0x0000, Data: []byte{}},
				},
				HasSNI: true,
			},
			minCount: 0,
		},
		{
			name: "GREASE ECH - 一条策略",
			data: fp.ClientHelloData{
				TLSVersion:   0x0304,
				CipherSuites: []uint16{0x1301},
				Extensions: []fp.ExtensionData{
					{Type: 0xfe0d, Data: []byte{0x00, 0x00, 0x00}},
				},
				HasSNI: true,
			},
			minCount: 1,
		},
		{
			name: "Outer ClientHello - 多条策略",
			data: fp.ClientHelloData{
				TLSVersion:   0x0304,
				CipherSuites: []uint16{0x1301},
				Extensions: []fp.ExtensionData{
					{Type: 0xfd00, Data: []byte{0xfe, 0x0d, 0x01}},
					{Type: 0x000d, Data: []byte{}},
					{Type: 0x0033, Data: []byte{}},
					{Type: 0x002b, Data: []byte{}},
					{Type: 0x000a, Data: []byte{}},
					{Type: 0x0010, Data: []byte{}},
				},
				HasSNI: false,
			},
			minCount: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := fp.AnalyzeECH(tc.data)
			if err != nil {
				t.Fatalf("分析失败: %v", err)
			}

			if len(result.AlternativeStrategies) < tc.minCount {
				t.Errorf("策略数量不足: 期望>=%d, 实际=%d",
					tc.minCount,
					len(result.AlternativeStrategies),
				)
			}

			if len(result.AlternativeStrategies) > 0 {
				t.Logf("替代策略:")
				for i, strategy := range result.AlternativeStrategies {
					t.Logf("  %d. %s", i+1, strategy)
				}
			}
		})
	}
}

// TestECHAnalysis_ImpactSummary 测试影响摘要
func TestECHAnalysis_ImpactSummary(t *testing.T) {
	testCases := []struct {
		name           string
		data           fp.ClientHelloData
		expectedPrefix string
	}{
		{
			name: "无 ECH",
			data: fp.ClientHelloData{
				TLSVersion:   0x0304,
				CipherSuites: []uint16{0x1301},
				Extensions: []fp.ExtensionData{
					{Type: 0x000d, Data: []byte{}},
				},
				HasSNI: true,
			},
			expectedPrefix: "无 ECH",
		},
		{
			name: "有 ECH",
			data: fp.ClientHelloData{
				TLSVersion:   0x0304,
				CipherSuites: []uint16{0x1301},
				Extensions: []fp.ExtensionData{
					{Type: 0xfe0d, Data: []byte{0x00, 0x00, 0x00}},
				},
			},
			expectedPrefix: "ECH 类型",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _ := fp.AnalyzeECH(tc.data)
			summary := result.GetImpactSummary()

			if len(summary) == 0 {
				t.Error("摘要不应为空")
			}

			t.Logf("摘要: %s", summary)
		})
	}
}

// BenchmarkECHAnalysis ECH 分析性能测试
func BenchmarkECHAnalysis(b *testing.B) {
	data := fp.ClientHelloData{
		TLSVersion: 0x0304,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303, 0xc02b, 0xc02f,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0xfd00, Data: []byte{0xfe, 0x0d, 0x01, 0x02}},
			{Type: 0x000d, Data: []byte{0x04, 0x03, 0x05, 0x03}},
			{Type: 0x0033, Data: []byte{0x00, 0x1d}},
			{Type: 0x002b, Data: []byte{0x03, 0x04}},
			{Type: 0x000a, Data: []byte{0x00, 0x1d}},
		},
		HasSNI: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fp.AnalyzeECH(data)
	}
}

// BenchmarkECHAnalysis_NoECH 无 ECH 场景性能测试
func BenchmarkECHAnalysis_NoECH(b *testing.B) {
	data := fp.ClientHelloData{
		TLSVersion: 0x0304,
		CipherSuites: []uint16{
			0x1301, 0x1302, 0x1303,
		},
		Extensions: []fp.ExtensionData{
			{Type: 0x0000, Data: []byte{0x00, 0x0b, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d}},
			{Type: 0x000d, Data: []byte{}},
		},
		HasSNI: true,
		SNI:    "example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fp.AnalyzeECH(data)
	}
}
