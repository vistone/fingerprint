package fingerprint

import (
	te "github.com/vistone/fingerprint/tls/ech"
)

// ECHAnalysisResult ECH 分析结果（兼容别名）。
type ECHAnalysisResult = te.ECHAnalysisResult

// ECHImpact ECH 影响度分析（兼容别名）。
type ECHImpact = te.ECHImpact

// ClientHelloData ClientHello 数据（兼容别名）。
type ClientHelloData = te.ClientHelloData

// ExtensionData 扩展数据（兼容别名）。
type ExtensionData = te.ExtensionData

// ECHAnalyzer ECH 分析器（兼容别名）。
type ECHAnalyzer = te.ECHAnalyzer

// ECHConfig ECH 配置（兼容别名）。
type ECHConfig = te.ECHConfig

// NewECHAnalyzer 创建新的 ECH 分析器（兼容入口）。
func NewECHAnalyzer() *ECHAnalyzer {
	return te.NewECHAnalyzer()
}

// AnalyzeECH 分析 ECH 配置（兼容入口）。
func AnalyzeECH(data ClientHelloData) (*ECHAnalysisResult, error) {
	return te.AnalyzeECH(data)
}
