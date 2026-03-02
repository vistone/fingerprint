// Package fingerprint 提供 TCP/IP 层指纹识别
package fingerprint

import nt "github.com/vistone/fingerprint/network/tcp"

// TCPFlags TCP 标志位（兼容别名）。
type TCPFlags = nt.TCPFlags

// IPHeader IP 包头参数（兼容别名）。
type IPHeader = nt.IPHeader

// TCPOptions TCP 选项（兼容别名）。
type TCPOptions = nt.TCPOptions

// TCPPacket TCP 数据包参数（兼容别名）。
type TCPPacket = nt.TCPPacket

// TCPIPSignature TCP/IP 指纹签名（兼容别名）。
type TCPIPSignature = nt.TCPIPSignature

// OSFingerprint 操作系统指纹（兼容别名）。
type OSFingerprint = nt.OSFingerprint

// TCPIPAnalyzer TCP/IP 指纹分析器（兼容别名）。
type TCPIPAnalyzer = nt.TCPIPAnalyzer

// TCPIPResult TCP/IP 分析结果（兼容别名）。
type TCPIPResult = nt.TCPIPResult

// NewTCPIPAnalyzer 创建新的分析器（兼容入口）。
func NewTCPIPAnalyzer() *TCPIPAnalyzer {
	return nt.NewTCPIPAnalyzer()
}
