package fingerprint

import (
	hh "github.com/vistone/fingerprint/http/http2"
)

// HTTP2SignatureResult HTTP/2 签名结果（兼容别名）。
type HTTP2SignatureResult = hh.HTTP2SignatureResult

// HTTP2FrameData HTTP/2 帧数据（兼容别名）。
type HTTP2FrameData = hh.HTTP2FrameData

// PriorityData 优先级数据（兼容别名）。
type PriorityData = hh.PriorityData

// HTTP2SignatureAnalyzer HTTP/2 签名分析器（兼容别名）。
type HTTP2SignatureAnalyzer = hh.HTTP2SignatureAnalyzer

// HTTP2ClientProfile HTTP/2 客户端配置文件（兼容别名）。
type HTTP2ClientProfile = hh.HTTP2ClientProfile

// NewHTTP2SignatureAnalyzer 创建新的 HTTP/2 签名分析器（兼容入口）。
func NewHTTP2SignatureAnalyzer() *HTTP2SignatureAnalyzer {
	return hh.NewHTTP2SignatureAnalyzer()
}

// ComputeHTTP2Signature 计算 HTTP/2 签名（兼容入口）。
func ComputeHTTP2Signature(frames []HTTP2FrameData) (*HTTP2SignatureResult, error) {
	return hh.ComputeHTTP2Signature(frames)
}
