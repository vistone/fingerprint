package fingerprint

import (
	tj "github.com/vistone/fingerprint/tls/ja4s"
)

// JA4SResult JA4S 指纹结果（兼容别名）。
type JA4SResult = tj.JA4SResult

// JA4SAnalyzer JA4S 分析器（兼容别名）。
type JA4SAnalyzer = tj.JA4SAnalyzer

// ServerProfileInfo 已知服务端配置信息（兼容别名）。
type ServerProfileInfo = tj.ServerProfileInfo

// NewJA4SAnalyzer 创建 JA4S 分析器（兼容入口）。
func NewJA4SAnalyzer() *JA4SAnalyzer {
	return tj.NewJA4SAnalyzer()
}

// ComputeJA4S 从 ServerHello 字节数据计算 JA4S（兼容入口）。
func ComputeJA4S(serverHelloBytes []byte) (*JA4SResult, error) {
	return tj.ComputeJA4S(serverHelloBytes)
}

// ComputeJA4SFromProfileData 从 Profile 数据计算 JA4S（兼容入口）。
func ComputeJA4SFromProfileData(tlsVersion uint16, cipherSuite uint16, extensions []uint16) (*JA4SResult, error) {
	return tj.ComputeJA4SFromProfileData(tlsVersion, cipherSuite, extensions)
}
