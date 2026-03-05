package tcpip

// TCP 协议常量
const (
	// TCP 标志位
	FlagFIN = 1 << iota
	FlagSYN
	FlagRST
	FlagPSH
	FlagACK
	FlagURG
	FlagECE
	FlagCWR
)

// TCP 选项类型
const (
	OptMSS      = 2  // 最大段大小
	OptWS       = 3  // 窗口缩放
	OptSACK     = 4  // 选择性确认
	OptTS       = 8  // 时间戳
	OptNOP      = 1  // No Operation
	OptEOL      = 0  // 选项列表结束
	OptSACKPerm = 4  // SACK Permitted
	OptAltSum   = 14 // 替代校验和
	OptMD5      = 19 // TCP MD5
	OptFastOpen = 34 // TCP Fast Open
	OptMptcp    = 30 // MPTCP
)

// TTL 推荐值（基于操作系统）
const (
	DefaultTTLLinux   = 64
	DefaultTTLWindows = 128
	DefaultTTLMacOS   = 64
	DefaultTTLiOS     = 64
	DefaultTTLAndroid = 64
)

// IP 标志位
const (
	IPFlagMF = 0x20 // 更多分片
	IPFlagDF = 0x40 // "不要分片"标志
	IPFlagRF = 0x80 // 保留位
)

// TCP 特征阈值
const (
	MinWindowSize = 512
	MaxWindowSize = 1073741824 // 1GB
	MinMSS        = 536        // 最小可行 MSS
	MaxMSS        = 65495
)

// 网络设备特征
var NetworkDeviceSignatures = map[string]DeviceSignature{
	"FritzBox_Router": {
		Name:       "AVM FritzBox Router",
		Vendor:     "AVM",
		Device:     "Router",
		HTTPOS:     "Custom Linux",
		DefaultTTL: 64,
		Behavior:   "Specific port scanning behavior",
	},
	"Cisco_Router": {
		Name:       "Cisco IOS Router",
		Vendor:     "Cisco",
		Device:     "Router",
		HTTPOS:     "IOS",
		DefaultTTL: 255,
		Behavior:   "BGP/OSPF signatures",
	},
	"Juniper_Switch": {
		Name:       "Juniper Junos Switch",
		Vendor:     "Juniper",
		Device:     "Switch",
		HTTPOS:     "Junos",
		DefaultTTL: 64,
		Behavior:   "LACP/LLDP signatures",
	},
	"HP_Printer": {
		Name:       "HP Network Printer",
		Vendor:     "HP",
		Device:     "Printer",
		HTTPOS:     "Proprietary",
		DefaultTTL: 255,
		Behavior:   "LPD/IPP protocols",
	},
	"QNAP_NAS": {
		Name:       "QNAP NAS",
		Vendor:     "QNAP",
		Device:     "NAS",
		HTTPOS:     "Linux",
		DefaultTTL: 64,
		Behavior:   "SMB/NFS/AFP signatures",
	},
}

// DeviceSignature 设备签名
type DeviceSignature struct {
	Name       string // 设备名称
	Vendor     string // 厂商
	Device     string // 设备类型
	HTTPOS     string // HTTP 服务的 OS 特征
	DefaultTTL int    // 默认 TTL
	Behavior   string // 典型行为
}

// TCPIPAnomalyType TCP/IP 异常类型
type TCPIPAnomalyType string

const (
	AnomalyInvalidTTL        TCPIPAnomalyType = "invalid_ttl"
	AnomalyInvalidMSS        TCPIPAnomalyType = "invalid_mss"
	AnomalyInvalidWindowSize TCPIPAnomalyType = "invalid_window_size"
	AnomalyUnusualOptions    TCPIPAnomalyType = "unusual_options"
	AnomalyZeroACK           TCPIPAnomalyType = "zero_ack"
	AnomalySpoofedIP         TCPIPAnomalyType = "spoofed_ip"
	AnomalyNATDetected       TCPIPAnomalyType = "nat_detected"
	AnomalyVPNDetected       TCPIPAnomalyType = "vpn_detected"
	AnomalyProxyDetected     TCPIPAnomalyType = "proxy_detected"
	AnomalyFragmentation     TCPIPAnomalyType = "fragmentation_detected"
	AnomalyTTLDecrement      TCPIPAnomalyType = "unexpected_ttl_decrement"
)

// VPNSignature VPN 特征
var VPNSignatures = []string{
	"non-standard_ttl",
	"consistent_ip_id_counter",
	"unusual_window_scaling",
	"mtu_smaller_than_standard",
	"specific_cipher_patterns",
}

// ProxySignature 代理特征
var ProxySignatures = []string{
	"time_sync_issues",
	"inconsistent_tcp_options",
	"path_mtu_discovery_failure",
	"sack_sack_perm_mismatch",
	"duplicate_sequence_numbers",
}

// BotSignature 机器人特征
var BotSignatures = []string{
	"abnormal_syn_timing",
	"identical_tcp_fingerprints",
	"no_variance_in_rtt",
	"exact_packet_timing",
	"missing_human_delay_pattern",
	"invalid_version_headers",
	"impossible_os_combinations",
}

// RiskIndicators 风险指标
type RiskIndicators struct {
	IsBot      bool
	IsScanner  bool
	IsVPN      bool
	IsProxy    bool
	IsNAT      bool
	IsSpoofed  bool
	Suspicious []string
	RiskScore  float64
}

// CalculateRiskScore 计算总体风险评分
func CalculateRiskScore(indicators RiskIndicators) float64 {
	score := 0.0

	if indicators.IsBot {
		score += 0.3
	}
	if indicators.IsScanner {
		score += 0.2
	}
	if indicators.IsVPN {
		score += 0.1
	}
	if indicators.IsProxy {
		score += 0.1
	}
	if indicators.IsNAT {
		score += 0.05
	}
	if indicators.IsSpoofed {
		score += 0.25
	}

	suspiciousPenalty := float64(len(indicators.Suspicious)) * 0.02
	score += suspiciousPenalty

	if score > 1.0 {
		score = 1.0
	}

	return score
}
