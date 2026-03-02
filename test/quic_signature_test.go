package fingerprint

import (
	"testing"

	"github.com/vistone/fingerprint"
)

func TestQUICSignature_BasicFunctionality(t *testing.T) {
	analyzer := fingerprint.NewQUICSignatureAnalyzer()

	initial := fingerprint.QUICInitialData{
		Version: 0x00000001, // QUIC v1
		TransportParams: map[string]interface{}{
			"initial_max_data":                    10485760,
			"initial_max_stream_data_bidi_local":  1048576,
			"initial_max_stream_data_bidi_remote": 1048576,
			"initial_max_streams_bidi":            100,
			"max_idle_timeout":                    30000,
		},
		FrameTypes: []uint64{
			0x06, // CRYPTO
			0x02, // ACK
		},
		SourceConnectionID:      []byte{0x01, 0x02, 0x03, 0x04},
		DestinationConnectionID: []byte{0x05, 0x06, 0x07, 0x08},
		InitialMaxData:          10485760,
		InitialMaxStreamData:    1048576,
	}

	result, err := analyzer.AnalyzeQUICInitial(initial)
	if err != nil {
		t.Fatalf("AnalyzeQUICInitial failed: %v", err)
	}

	if result.Hash == "" {
		t.Error("Expected QUIC signature hash to be generated")
	}

	if result.VersionSignature != "v1" {
		t.Errorf("Expected version signature 'v1', got %s", result.VersionSignature)
	}

	if !result.IsHTTP3 {
		t.Error("QUIC v1 should be recognized as HTTP/3")
	}

	if result.RiskScore > 0.3 {
		t.Errorf("Expected low risk score, got %.2f with flags: %v",
			result.RiskScore, result.AnomalyFlags)
	}

	t.Logf("QUIC Hash: %s", result.Hash)
	t.Logf("Version: %s", result.VersionSignature)
	t.Logf("Transport Params: %s", result.TransportParameters)
	t.Logf("Frame Sequence: %s", result.FrameSequence)
	t.Logf("Risk Score: %.2f", result.RiskScore)
}

func TestQUICSignature_DraftVersion(t *testing.T) {
	analyzer := fingerprint.NewQUICSignatureAnalyzer()

	initial := fingerprint.QUICInitialData{
		Version: 0xff00001d, // draft-29
		TransportParams: map[string]interface{}{
			"initial_max_data":                    10485760,
			"initial_max_stream_data_bidi_local":  1048576,
			"initial_max_stream_data_bidi_remote": 1048576,
			"initial_max_streams_bidi":            100,
		},
		FrameTypes: []uint64{
			0x06, // CRYPTO
		},
		SourceConnectionID:   []byte{0x01, 0x02, 0x03, 0x04},
		InitialMaxData:       10485760,
		InitialMaxStreamData: 1048576,
	}

	result, err := analyzer.AnalyzeQUICInitial(initial)
	if err != nil {
		t.Fatalf("AnalyzeQUICInitial failed: %v", err)
	}

	// 应该检测到草稿版本
	hasDraftFlag := false
	for _, flag := range result.AnomalyFlags {
		if flag == "DRAFT_VERSION" {
			hasDraftFlag = true
			break
		}
	}

	if !hasDraftFlag {
		t.Errorf("Expected DRAFT_VERSION flag, got flags: %v", result.AnomalyFlags)
	}

	if result.VersionSignature != "draft-29" {
		t.Errorf("Expected version signature 'draft-29', got %s", result.VersionSignature)
	}

	t.Logf("Draft version detected: %s", result.VersionSignature)
	t.Logf("Anomaly flags: %v", result.AnomalyFlags)
}

func TestQUICSignature_MissingCryptoFrame(t *testing.T) {
	analyzer := fingerprint.NewQUICSignatureAnalyzer()

	initial := fingerprint.QUICInitialData{
		Version: 0x00000001,
		TransportParams: map[string]interface{}{
			"initial_max_data":                    10485760,
			"initial_max_stream_data_bidi_local":  1048576,
			"initial_max_stream_data_bidi_remote": 1048576,
			"initial_max_streams_bidi":            100,
		},
		FrameTypes: []uint64{
			0x00, // PADDING
			0x01, // PING
			// 缺少 CRYPTO 帧
		},
		SourceConnectionID:   []byte{0x01, 0x02, 0x03, 0x04},
		InitialMaxData:       10485760,
		InitialMaxStreamData: 1048576,
	}

	result, err := analyzer.AnalyzeQUICInitial(initial)
	if err != nil {
		t.Fatalf("AnalyzeQUICInitial failed: %v", err)
	}

	// 应该检测到异常帧序列
	hasAnomalyFlag := false
	for _, flag := range result.AnomalyFlags {
		if flag == "ANOMALOUS_FRAME_SEQUENCE" {
			hasAnomalyFlag = true
			break
		}
	}

	if !hasAnomalyFlag {
		t.Errorf("Expected ANOMALOUS_FRAME_SEQUENCE flag, got flags: %v", result.AnomalyFlags)
	}

	if result.RiskScore <= 0.1 {
		t.Errorf("Expected higher risk score for missing CRYPTO frame, got %.2f", result.RiskScore)
	}

	t.Logf("Anomaly detected - Risk Score: %.2f", result.RiskScore)
	t.Logf("Flags: %v", result.AnomalyFlags)
}

func TestQUICSignature_SuspiciousLimits(t *testing.T) {
	analyzer := fingerprint.NewQUICSignatureAnalyzer()

	initial := fingerprint.QUICInitialData{
		Version: 0x00000001,
		TransportParams: map[string]interface{}{
			"initial_max_data":                    10485760,
			"initial_max_stream_data_bidi_local":  1048576,
			"initial_max_stream_data_bidi_remote": 1048576,
			"initial_max_streams_bidi":            100,
		},
		FrameTypes: []uint64{
			0x06, // CRYPTO
		},
		SourceConnectionID:   []byte{0x01, 0x02, 0x03, 0x04},
		InitialMaxData:       100, // 异常小
		InitialMaxStreamData: 10,  // 异常小
	}

	result, err := analyzer.AnalyzeQUICInitial(initial)
	if err != nil {
		t.Fatalf("AnalyzeQUICInitial failed: %v", err)
	}

	// 应该检测到可疑的限制值
	hasSuspiciousFlag := false
	for _, flag := range result.AnomalyFlags {
		if flag == "SUSPICIOUS_LIMITS" {
			hasSuspiciousFlag = true
			break
		}
	}

	if !hasSuspiciousFlag {
		t.Errorf("Expected SUSPICIOUS_LIMITS flag, got flags: %v", result.AnomalyFlags)
	}

	t.Logf("Suspicious limits detected - Risk Score: %.2f", result.RiskScore)
}

func TestQUICSignature_InvalidConnectionID(t *testing.T) {
	analyzer := fingerprint.NewQUICSignatureAnalyzer()

	initial := fingerprint.QUICInitialData{
		Version: 0x00000001,
		TransportParams: map[string]interface{}{
			"initial_max_data":                    10485760,
			"initial_max_stream_data_bidi_local":  1048576,
			"initial_max_stream_data_bidi_remote": 1048576,
			"initial_max_streams_bidi":            100,
		},
		FrameTypes: []uint64{
			0x06, // CRYPTO
		},
		SourceConnectionID:   []byte{}, // 空连接 ID（异常）
		InitialMaxData:       10485760,
		InitialMaxStreamData: 1048576,
	}

	result, err := analyzer.AnalyzeQUICInitial(initial)
	if err != nil {
		t.Fatalf("AnalyzeQUICInitial failed: %v", err)
	}

	// 应该检测到无效的连接 ID
	hasInvalidIDFlag := false
	for _, flag := range result.AnomalyFlags {
		if flag == "INVALID_CONNECTION_ID" {
			hasInvalidIDFlag = true
			break
		}
	}

	if !hasInvalidIDFlag {
		t.Errorf("Expected INVALID_CONNECTION_ID flag, got flags: %v", result.AnomalyFlags)
	}

	t.Logf("Invalid connection ID detected")
}

func TestQUICSignature_MatchKnownClients(t *testing.T) {
	analyzer := fingerprint.NewQUICSignatureAnalyzer()

	// Chrome-like QUIC 配置
	initial := fingerprint.QUICInitialData{
		Version: 0x00000001,
		TransportParams: map[string]interface{}{
			"initial_max_data":                    10485760,
			"initial_max_stream_data_bidi_local":  1048576,
			"initial_max_stream_data_bidi_remote": 1048576,
			"initial_max_streams_bidi":            100,
			"max_idle_timeout":                    30000,
		},
		FrameTypes: []uint64{
			0x06, // CRYPTO
			0x02, // ACK
		},
		SourceConnectionID:   []byte{0x01, 0x02, 0x03, 0x04},
		InitialMaxData:       10485760,
		InitialMaxStreamData: 1048576,
	}

	result, err := analyzer.AnalyzeQUICInitial(initial)
	if err != nil {
		t.Fatalf("AnalyzeQUICInitial failed: %v", err)
	}

	if len(result.MatchedClients) == 0 {
		t.Error("Expected to match at least one known client")
	}

	t.Logf("Matched clients: %v", result.MatchedClients)
}

func BenchmarkQUICSignature_Analysis(b *testing.B) {
	analyzer := fingerprint.NewQUICSignatureAnalyzer()

	initial := fingerprint.QUICInitialData{
		Version: 0x00000001,
		TransportParams: map[string]interface{}{
			"initial_max_data":                    10485760,
			"initial_max_stream_data_bidi_local":  1048576,
			"initial_max_stream_data_bidi_remote": 1048576,
			"initial_max_streams_bidi":            100,
		},
		FrameTypes: []uint64{
			0x06, // CRYPTO
			0x02, // ACK
		},
		SourceConnectionID:   []byte{0x01, 0x02, 0x03, 0x04},
		InitialMaxData:       10485760,
		InitialMaxStreamData: 1048576,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeQUICInitial(initial)
		if err != nil {
			b.Fatal(err)
		}
	}
}
