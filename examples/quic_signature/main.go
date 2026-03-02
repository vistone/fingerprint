package main

import (
	"fmt"
	"log"

	"github.com/vistone/fingerprint"
)

func main() {
	fmt.Println("=== QUIC Signature Analysis Demo ===")

	// 创建分析器
	analyzer := fingerprint.NewQUICSignatureAnalyzer()

	// 示例 1: Chrome QUIC 请求
	fmt.Println("--- Example 1: Chrome QUIC Request ---")
	chromeInitial := fingerprint.QUICInitialData{
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

	result, err := analyzer.AnalyzeQUICInitial(chromeInitial)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	printResult(result)

	// 示例 2: 草稿版本 QUIC
	fmt.Println("\n--- Example 2: Draft Version QUIC ---")
	draftInitial := fingerprint.QUICInitialData{
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

	result, err = analyzer.AnalyzeQUICInitial(draftInitial)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	printResult(result)

	// 示例 3: 可疑的 QUIC 请求（缺少 CRYPTO 帧）
	fmt.Println("\n--- Example 3: Suspicious QUIC Request ---")
	suspiciousInitial := fingerprint.QUICInitialData{
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
			// 缺少 CRYPTO 帧 - 异常
		},
		SourceConnectionID:   []byte{0x01, 0x02, 0x03, 0x04},
		InitialMaxData:       100, // 异常小的值
		InitialMaxStreamData: 10,  // 异常小的值
	}

	result, err = analyzer.AnalyzeQUICInitial(suspiciousInitial)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	printResult(result)

	// 示例 4: Firefox QUIC
	fmt.Println("\n--- Example 4: Firefox QUIC Request ---")
	firefoxInitial := fingerprint.QUICInitialData{
		Version: 0x00000001,
		TransportParams: map[string]interface{}{
			"initial_max_data":                    15728640, // Firefox 使用更大的值
			"initial_max_stream_data_bidi_local":  524288,
			"initial_max_stream_data_bidi_remote": 524288,
			"initial_max_streams_bidi":            100,
			"max_idle_timeout":                    30000,
		},
		FrameTypes: []uint64{
			0x06, // CRYPTO
			0x00, // PADDING
		},
		SourceConnectionID:   []byte{0x01, 0x02, 0x03, 0x04},
		InitialMaxData:       15728640,
		InitialMaxStreamData: 524288,
	}

	result, err = analyzer.AnalyzeQUICInitial(firefoxInitial)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	printResult(result)

	// 示例 5: 使用便捷函数
	fmt.Println("\n--- Example 5: Using Convenience Function ---")
	result, err = fingerprint.ComputeQUICSignature(chromeInitial)
	if err != nil {
		log.Fatalf("ComputeQUICSignature failed: %v", err)
	}

	fmt.Printf("Quick QUIC Hash: %s\n", result.Hash)
	fmt.Printf("Version: %s, HTTP/3: %v\n", result.QUICVersion, result.IsHTTP3)
	fmt.Printf("Risk Score: %.2f\n", result.RiskScore)
}

func printResult(result *fingerprint.QUICSignatureResult) {
	fmt.Printf("QUIC Signature Hash: %s\n", result.Hash)
	fmt.Printf("Version: %s (HTTP/3: %v)\n", result.QUICVersion, result.IsHTTP3)
	fmt.Printf("Version Signature: %s\n", result.VersionSignature)
	fmt.Printf("Transport Parameters: %s\n", result.TransportParameters)
	fmt.Printf("Frame Sequence: %s\n", result.FrameSequence)
	fmt.Printf("Risk Score: %.2f\n", result.RiskScore)

	if len(result.AnomalyFlags) > 0 {
		fmt.Printf("⚠️  Anomaly Flags: %v\n", result.AnomalyFlags)
	} else {
		fmt.Println("✓ No anomalies detected")
	}

	if len(result.MatchedClients) > 0 {
		fmt.Printf("Matched Clients: %v\n", result.MatchedClients)
	}
}
