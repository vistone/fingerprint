package fingerprint

import (
	hj "github.com/vistone/fingerprint/http/ja4h"
)

// JA4HResult JA4H 指纹结果（兼容别名）。
type JA4HResult = hj.JA4HResult

// HTTP2RequestData HTTP 请求信息（兼容别名）。
type HTTP2RequestData = hj.HTTP2RequestData

// JA4HAnalyzer JA4H 分析器（兼容别名）。
type JA4HAnalyzer = hj.JA4HAnalyzer

// HTTP2BrowserProfile HTTP/2 浏览器配置文件（兼容别名）。
type HTTP2BrowserProfile = hj.HTTP2BrowserProfile

// ClientHintsAnalysis 客户端提示分析结果（兼容别名）。
type ClientHintsAnalysis = hj.ClientHintsAnalysis

// NewJA4HAnalyzer 创建新的 JA4H 分析器（兼容入口）。
func NewJA4HAnalyzer() *JA4HAnalyzer {
	return hj.NewJA4HAnalyzer()
}
