package quic

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// QUICSignatureResult represents a QUIC signature analysis result
type QUICSignatureResult struct {
	Hash string

	VersionSignature    string
	TransportParameters string
	FrameSequence       string

	RawSignature string

	RiskScore float64

	AnomalyFlags []string

	MatchedClients []string

	QUICVersion    string
	IsHTTP3        bool
	TransportLayer string
}

// QUICInitialData represents QUIC Initial packet data
type QUICInitialData struct {
	Version uint32

	TransportParams map[string]interface{}

	FrameTypes []uint64

	SourceConnectionID      []byte
	DestinationConnectionID []byte
	InitialMaxData          uint64
	InitialMaxStreamData    uint64
}

// QUICSignatureAnalyzer is a QUIC signature analyzer
type QUICSignatureAnalyzer struct {
	knownClientProfiles map[string]*QUICClientProfile
}

// QUICClientProfile represents a known QUIC client configuration
type QUICClientProfile struct {
	Name                   string
	ClientName             string
	ClientVersion          string
	QUICVersions           []uint32
	TypicalTransportParams map[string]interface{}
	FrameSequencePattern   string
	RiskScore              float64
}

// NewQUICSignatureAnalyzer creates a new QUIC signature analyzer
func NewQUICSignatureAnalyzer() *QUICSignatureAnalyzer {
	return &QUICSignatureAnalyzer{
		knownClientProfiles: initKnownQUICClientProfiles(),
	}
}

// AnalyzeQUICInitial analyzes a QUIC Initial packet
func (a *QUICSignatureAnalyzer) AnalyzeQUICInitial(initial QUICInitialData) (*QUICSignatureResult, error) {
	if initial.Version == 0 {
		return nil, fmt.Errorf("QUIC version required")
	}

	result := &QUICSignatureResult{
		AnomalyFlags: make([]string, 0, 8),
		IsHTTP3:      isHTTP3Version(initial.Version),
		QUICVersion:  formatQUICVersion(initial.Version),
	}

	result.VersionSignature = formatQUICVersion(initial.Version)
	transportSig := generateTransportParamsSignature(initial.TransportParams)
	result.TransportParameters = transportSig
	frameSig := generateFrameSequenceSignature(initial.FrameTypes)
	result.FrameSequence = frameSig

	fullSignature := fmt.Sprintf("%s_%s_%s",
		result.VersionSignature,
		result.TransportParameters,
		result.FrameSequence,
	)
	result.RawSignature = fullSignature

	hash := sha256.Sum256([]byte(fullSignature))
	result.Hash = hex.EncodeToString(hash[:])

	a.detectAnomalies(result, initial)
	result.MatchedClients = a.FindMatchingClients(result, 3)

	return result, nil
}

func (a *QUICSignatureAnalyzer) detectAnomalies(result *QUICSignatureResult, initial QUICInitialData) {
	baseScore := 0.0

	if isDraftVersion(initial.Version) {
		result.AnomalyFlags = append(result.AnomalyFlags, "DRAFT_VERSION")
		baseScore += 0.1
	}

	if !isKnownQUICVersion(initial.Version) {
		result.AnomalyFlags = append(result.AnomalyFlags, "UNKNOWN_VERSION")
		baseScore += 0.2
	}

	if hasAnomalousTransportParams(initial.TransportParams) {
		result.AnomalyFlags = append(result.AnomalyFlags, "ANOMALOUS_TRANSPORT_PARAMS")
		baseScore += 0.15
	}

	if hasAnomalousFrameSequence(initial.FrameTypes) {
		result.AnomalyFlags = append(result.AnomalyFlags, "ANOMALOUS_FRAME_SEQUENCE")
		baseScore += 0.15
	}

	if initial.InitialMaxData < 1024 || initial.InitialMaxStreamData < 256 {
		result.AnomalyFlags = append(result.AnomalyFlags, "SUSPICIOUS_LIMITS")
		baseScore += 0.2
	}

	if len(initial.SourceConnectionID) == 0 || len(initial.SourceConnectionID) > 20 {
		result.AnomalyFlags = append(result.AnomalyFlags, "INVALID_CONNECTION_ID")
		baseScore += 0.1
	}

	if baseScore > 1.0 {
		baseScore = 1.0
	}
	result.RiskScore = baseScore
}

// FindMatchingClients finds matching known clients for a QUIC signature
func (a *QUICSignatureAnalyzer) FindMatchingClients(result *QUICSignatureResult, maxResults int) []string {
	var matches []string

	for name, profile := range a.knownClientProfiles {
		versionMatch := false
		for _, ver := range profile.QUICVersions {
			if formatQUICVersion(ver) == result.VersionSignature {
				versionMatch = true
				break
			}
		}

		if versionMatch && profile.RiskScore < result.RiskScore+0.15 {
			matches = append(matches, name)
		}

		if len(matches) >= maxResults {
			break
		}
	}

	return matches
}

func formatQUICVersion(version uint32) string {
	switch version {
	case 0x00000001:
		return "v1"
	case 0x6b3343cf:
		return "v2"
	case 0xff00001d:
		return "draft-29"
	case 0xff00001e:
		return "draft-30"
	case 0xff00001f:
		return "draft-31"
	case 0xff000020:
		return "draft-32"
	default:
		return fmt.Sprintf("0x%08x", version)
	}
}

func isHTTP3Version(version uint32) bool {
	return version == 0x00000001 || version == 0x6b3343cf
}

func isDraftVersion(version uint32) bool {
	return (version & 0xff000000) == 0xff000000
}

func isKnownQUICVersion(version uint32) bool {
	knownVersions := []uint32{
		0x00000001,
		0x6b3343cf,
		0xff00001d,
		0xff00001e,
		0xff00001f,
		0xff000020,
	}

	for _, known := range knownVersions {
		if version == known {
			return true
		}
	}
	return false
}

func generateTransportParamsSignature(params map[string]interface{}) string {
	if len(params) == 0 {
		return "empty"
	}

	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		v := params[k]
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}

	signature := strings.Join(parts, ",")
	hash := sha256.Sum256([]byte(signature))
	return hex.EncodeToString(hash[:8])
}

func generateFrameSequenceSignature(frameTypes []uint64) string {
	if len(frameTypes) == 0 {
		return "none"
	}

	var parts []string
	for _, ft := range frameTypes {
		parts = append(parts, frameTypeName(ft))
	}

	sequence := strings.Join(parts, "-")
	hash := sha256.Sum256([]byte(sequence))
	return hex.EncodeToString(hash[:8])
}

func frameTypeName(frameType uint64) string {
	switch frameType {
	case 0x00:
		return "padding"
	case 0x01:
		return "ping"
	case 0x02, 0x03:
		return "ack"
	case 0x04:
		return "reset_stream"
	case 0x05:
		return "stop_sending"
	case 0x06:
		return "crypto"
	case 0x07:
		return "new_token"
	case 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f:
		return "stream"
	case 0x10:
		return "max_data"
	case 0x11:
		return "max_stream_data"
	case 0x12, 0x13:
		return "max_streams"
	case 0x14:
		return "data_blocked"
	case 0x15:
		return "stream_data_blocked"
	case 0x16, 0x17:
		return "streams_blocked"
	case 0x18:
		return "new_connection_id"
	case 0x19:
		return "retire_connection_id"
	case 0x1a:
		return "path_challenge"
	case 0x1b:
		return "path_response"
	case 0x1c, 0x1d:
		return "connection_close"
	case 0x1e:
		return "handshake_done"
	default:
		return fmt.Sprintf("0x%02x", frameType)
	}
}

func hasAnomalousTransportParams(params map[string]interface{}) bool {
	if len(params) == 0 {
		return true
	}

	requiredParams := []string{
		"initial_max_data",
		"initial_max_stream_data_bidi_local",
		"initial_max_stream_data_bidi_remote",
		"initial_max_streams_bidi",
	}

	for _, param := range requiredParams {
		if _, exists := params[param]; !exists {
			return true
		}
	}

	return false
}

func hasAnomalousFrameSequence(frameTypes []uint64) bool {
	if len(frameTypes) == 0 {
		return true
	}

	hasCrypto := false
	for _, ft := range frameTypes {
		if ft == 0x06 {
			hasCrypto = true
			break
		}
	}

	return !hasCrypto
}

func initKnownQUICClientProfiles() map[string]*QUICClientProfile {
	return map[string]*QUICClientProfile{
		"chrome_quic": {
			Name:          "Chrome QUIC",
			ClientName:    "Chrome",
			ClientVersion: "120+",
			QUICVersions:  []uint32{0x00000001},
			TypicalTransportParams: map[string]interface{}{
				"initial_max_data":                    10485760,
				"initial_max_stream_data_bidi_local":  1048576,
				"initial_max_stream_data_bidi_remote": 1048576,
				"initial_max_streams_bidi":            100,
				"max_idle_timeout":                    30000,
			},
			FrameSequencePattern: "crypto-ack",
			RiskScore:            0.05,
		},
		"firefox_quic": {
			Name:          "Firefox QUIC",
			ClientName:    "Firefox",
			ClientVersion: "120+",
			QUICVersions:  []uint32{0x00000001},
			TypicalTransportParams: map[string]interface{}{
				"initial_max_data":                    15728640,
				"initial_max_stream_data_bidi_local":  524288,
				"initial_max_stream_data_bidi_remote": 524288,
				"initial_max_streams_bidi":            100,
				"max_idle_timeout":                    30000,
			},
			FrameSequencePattern: "crypto-padding",
			RiskScore:            0.08,
		},
		"curl_quic": {
			Name:          "curl with QUIC",
			ClientName:    "curl",
			ClientVersion: "8.0+",
			QUICVersions:  []uint32{0x00000001},
			TypicalTransportParams: map[string]interface{}{
				"initial_max_data":                    10485760,
				"initial_max_stream_data_bidi_local":  1048576,
				"initial_max_stream_data_bidi_remote": 1048576,
				"initial_max_streams_bidi":            100,
			},
			FrameSequencePattern: "crypto",
			RiskScore:            0.1,
		},
	}
}

// ComputeQUICSignature is a convenience function for computing a QUIC signature
func ComputeQUICSignature(initial QUICInitialData) (*QUICSignatureResult, error) {
	analyzer := NewQUICSignatureAnalyzer()
	return analyzer.AnalyzeQUICInitial(initial)
}

// NewAnalyzer creates a QUIC signature analyzer (module-unified naming).
func NewAnalyzer() *QUICSignatureAnalyzer {
	return NewQUICSignatureAnalyzer()
}

// Compute is a convenience function for computing a QUIC signature (module-unified naming).
func Compute(initial QUICInitialData) (*QUICSignatureResult, error) {
	return ComputeQUICSignature(initial)
}
