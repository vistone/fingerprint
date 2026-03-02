package ech

import fp "github.com/vistone/fingerprint"

// ECHAnalysisResult ECH 分析结果。
type ECHAnalysisResult = fp.ECHAnalysisResult

// ECHImpact ECH 对指纹识别的影响。
type ECHImpact = fp.ECHImpact

// ClientHelloData ClientHello 数据。
type ClientHelloData = fp.ClientHelloData

// ExtensionData 扩展数据。
type ExtensionData = fp.ExtensionData

// ECHAnalyzer ECH 分析器。
type ECHAnalyzer = fp.ECHAnalyzer

// ECHConfig 已知 ECH 配置。
type ECHConfig = fp.ECHConfig

// NewAnalyzer 创建 ECH 分析器。
func NewAnalyzer() *ECHAnalyzer {
	return fp.NewECHAnalyzer()
}

// Analyze 便捷函数：分析 ECH。
func Analyze(data ClientHelloData) (*ECHAnalysisResult, error) {
	return fp.AnalyzeECH(data)
}
