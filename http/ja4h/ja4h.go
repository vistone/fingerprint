package ja4h

import fp "github.com/vistone/fingerprint"

// JA4HResult JA4H 指纹结果。
type JA4HResult = fp.JA4HResult

// HTTP2RequestData HTTP 请求信息。
type HTTP2RequestData = fp.HTTP2RequestData

// JA4HAnalyzer JA4H 分析器。
type JA4HAnalyzer = fp.JA4HAnalyzer

// ClientHintsAnalysis Client Hints 分析结果。
type ClientHintsAnalysis = fp.ClientHintsAnalysis

// HTTP2BrowserProfile 浏览器配置。
type HTTP2BrowserProfile = fp.HTTP2BrowserProfile

// NewAnalyzer 创建 JA4H 分析器。
func NewAnalyzer() *JA4HAnalyzer {
	return fp.NewJA4HAnalyzer()
}

// Compute 便捷函数：计算 JA4H。
func Compute(req HTTP2RequestData) (*JA4HResult, error) {
	return fp.ComputeJA4H(req)
}
