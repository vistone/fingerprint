package http2

import fp "github.com/vistone/fingerprint"

// HTTP2SignatureResult HTTP/2 签名结果。
type HTTP2SignatureResult = fp.HTTP2SignatureResult

// HTTP2FrameData HTTP/2 帧数据。
type HTTP2FrameData = fp.HTTP2FrameData

// PriorityData 优先级信息。
type PriorityData = fp.PriorityData

// HTTP2SignatureAnalyzer HTTP/2 签名分析器。
type HTTP2SignatureAnalyzer = fp.HTTP2SignatureAnalyzer

// HTTP2ClientProfile 已知客户端配置。
type HTTP2ClientProfile = fp.HTTP2ClientProfile

// NewAnalyzer 创建 HTTP/2 签名分析器。
func NewAnalyzer() *HTTP2SignatureAnalyzer {
	return fp.NewHTTP2SignatureAnalyzer()
}

// Compute 便捷函数：计算 HTTP/2 签名。
func Compute(frames []HTTP2FrameData) (*HTTP2SignatureResult, error) {
	return fp.ComputeHTTP2Signature(frames)
}
