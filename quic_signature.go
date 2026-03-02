package fingerprint

import nq "github.com/vistone/fingerprint/network/quic"

// QUICSignatureResult QUIC 签名结果（兼容别名）。
type QUICSignatureResult = nq.QUICSignatureResult

// QUICInitialData QUIC Initial 包数据（兼容别名）。
type QUICInitialData = nq.QUICInitialData

// QUICSignatureAnalyzer QUIC 签名分析器（兼容别名）。
type QUICSignatureAnalyzer = nq.QUICSignatureAnalyzer

// QUICClientProfile 已知 QUIC 客户端配置（兼容别名）。
type QUICClientProfile = nq.QUICClientProfile

// NewQUICSignatureAnalyzer 创建分析器（兼容入口）。
func NewQUICSignatureAnalyzer() *QUICSignatureAnalyzer {
	return nq.NewQUICSignatureAnalyzer()
}

// ComputeQUICSignature 便捷函数：计算 QUIC 签名（兼容入口）。
func ComputeQUICSignature(initial QUICInitialData) (*QUICSignatureResult, error) {
	return nq.ComputeQUICSignature(initial)
}
