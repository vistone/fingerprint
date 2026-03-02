package clienthints

import fp "github.com/vistone/fingerprint"

// NegotiationState Client Hints 协商状态。
type NegotiationState = fp.NegotiationState

// ServerPreferences 服务器偏好。
type ServerPreferences = fp.ServerPreferences

// ClientCapabilities 客户端能力声明。
type ClientCapabilities = fp.ClientCapabilities

// NegotiationStrategy 协商策略。
type NegotiationStrategy = fp.NegotiationStrategy

// NegotiationRecord 协商记录。
type NegotiationRecord = fp.NegotiationRecord

// CHNegotiationAnalyzer 协商分析器。
type CHNegotiationAnalyzer = fp.CHNegotiationAnalyzer

const (
	NEGOTIATION_INIT      = fp.NEGOTIATION_INIT
	NEGOTIATION_REQUESTED = fp.NEGOTIATION_REQUESTED
	NEGOTIATION_ACCEPTED  = fp.NEGOTIATION_ACCEPTED
	NEGOTIATION_REJECTED  = fp.NEGOTIATION_REJECTED
	NEGOTIATION_DELEGATED = fp.NEGOTIATION_DELEGATED
)

// NewNegotiationAnalyzer 创建协商分析器。
func NewNegotiationAnalyzer() *CHNegotiationAnalyzer {
	return fp.NewCHNegotiationAnalyzer()
}
