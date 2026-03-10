package tcpip

// translated comment
const (
	// translated comment
	FlagFIN = 1 << iota
	FlagSYN
	FlagRST
	FlagPSH
	FlagACK
	FlagURG
	FlagECE
	FlagCWR
)

// translated comment
const (
	OptMSS      = 2  // translated comment
	OptWS       = 3  // translated comment
	OptSACK     = 4  // translated comment
	OptTS       = 8  // translated comment
	OptNOP      = 1  // No Operation
	OptEOL      = 0  // translated comment
	OptSACKPerm = 4  // SACK Permitted
	OptAltSum   = 14 // translated comment
	OptMD5      = 19 // TCP MD5
	OptFastOpen = 34 // TCP Fast Open
	OptMptcp    = 30 // MPTCP
)

// translated comment
const (
	DefaultTTLLinux   = 64
	DefaultTTLWindows = 128
	DefaultTTLMacOS   = 64
	DefaultTTLiOS     = 64
	DefaultTTLAndroid = 64
)

// translated comment
const (
	IPFlagMF = 0x20 // translated comment
	IPFlagDF = 0x40 // translated comment
	IPFlagRF = 0x80 // translated comment
)

// translated comment
const (
	MinWindowSize = 512
	MaxWindowSize = 1073741824 // 1GB
	MinMSS        = 536        // translated comment
	MaxMSS        = 65495
)

// translated comment
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

// translated comment
type DeviceSignature struct {
	Name       string // translated comment
	Vendor     string // translated comment
	Device     string // translated comment
	HTTPOS     string // translated comment
	DefaultTTL int    // translated comment
	Behavior   string // translated comment
}

// translated comment
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

// translated comment
var VPNSignatures = []string{
	"non-standard_ttl",
	"consistent_ip_id_counter",
	"unusual_window_scaling",
	"mtu_smaller_than_standard",
	"specific_cipher_patterns",
}

// translated comment
var ProxySignatures = []string{
	"time_sync_issues",
	"inconsistent_tcp_options",
	"path_mtu_discovery_failure",
	"sack_sack_perm_mismatch",
	"duplicate_sequence_numbers",
}

// translated comment
var BotSignatures = []string{
	"abnormal_syn_timing",
	"identical_tcp_fingerprints",
	"no_variance_in_rtt",
	"exact_packet_timing",
	"missing_human_delay_pattern",
	"invalid_version_headers",
	"impossible_os_combinations",
}

// translated comment
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

// translated comment
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
