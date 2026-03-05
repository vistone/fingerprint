package gateway

import (
	"testing"

	"github.com/vistone/fingerprint/modules/internal/testhelpers"
	
)

func TestConvertToGatewayRequest(t *testing.T) {
	tests := []struct {
		name     string
		grpcReq  *GRPCAnalyzeRequest
		wantErr  bool
	}{
		{
			name: "valid request conversion",
			grpcReq: &GRPCAnalyzeRequest{
				ClientID:        "test-client",
				TLSVersion:      0x0303,
				CipherSuites:    []uint32{0x1301, 0x1302, 0x1303},
				Extensions:      []uint32{0, 23, 65281},
				SupportedCurves: []uint32{23, 24, 25},
				Headers: map[string]string{
					"user-agent":      "Mozilla/5.0",
					"accept":          "text/html",
					"accept-language": "en-US",
				},
				ClientIP: "192.168.1.1",
			},
			wantErr: false,
		},
		{
			name: "empty request",
			grpcReq: &GRPCAnalyzeRequest{
				ClientID: "empty-client",
				Headers:  make(map[string]string),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToGatewayRequest(tt.grpcReq)
			testhelpers.AssertNotNil(t, result)
			testhelpers.AssertEqual(t, result.TLSVersion, uint16(tt.grpcReq.TLSVersion))
			testhelpers.AssertEqual(t, len(result.CipherSuites), len(tt.grpcReq.CipherSuites))
			testhelpers.AssertEqual(t, result.ClientIP, tt.grpcReq.ClientIP)
		})
	}
}

func TestConvertFromGatewayResponse(t *testing.T) {
	t.Run("convert response", func(t *testing.T) {
		// Create a simple response with minimal fields
		gatewayResp := &AnalyzeResponse{
			FingerprintHash: "abc123",
			JA3: &JA3Info{
				Hash: "ja3_hash_123",
			},
			JA4: &JA4Info{
				Fingerprint: "ja4_fingerprint_456",
			},
			Cached:           true,
			ProcessingTimeMs: 15,
		}

		result := ConvertFromGatewayResponse(gatewayResp)

		testhelpers.AssertEqual(t, result.FingerprintHash, gatewayResp.FingerprintHash)
		testhelpers.AssertEqual(t, result.JA3Hash, gatewayResp.JA3.Hash)
		testhelpers.AssertEqual(t, result.JA4Fingerprint, gatewayResp.JA4.Fingerprint)
		testhelpers.AssertEqual(t, result.Cached, gatewayResp.Cached)
	})
}

func TestGRPCServer(t *testing.T) {
	t.Run("create gRPC server", func(t *testing.T) {
		// Create a minimal gateway for testing
		gateway := &Gateway{}
		
		server := NewGRPCServer(gateway)
		testhelpers.AssertNotNil(t, server)
		testhelpers.AssertNotNil(t, server.server)
		testhelpers.AssertEqual(t, server.gateway, gateway)
	})
}

func TestHybridServer(t *testing.T) {
	t.Run("create hybrid server", func(t *testing.T) {
		gateway := &Gateway{}
		
		hybrid := NewHybridServer(gateway, 8080, 9090)
		testhelpers.AssertNotNil(t, hybrid)
		testhelpers.AssertEqual(t, hybrid.httpPort, 8080)
		testhelpers.AssertEqual(t, hybrid.grpcPort, 9090)
	})

	t.Run("get server info", func(t *testing.T) {
		gateway := &Gateway{}
		
		hybrid := NewHybridServer(gateway, 8080, 9090)
		info := hybrid.GetServerInfo()

		testhelpers.AssertEqual(t, len(info), 2)
		testhelpers.AssertEqual(t, info[0].Protocol, ProtocolHTTP)
		testhelpers.AssertEqual(t, info[0].Port, 8080)
		testhelpers.AssertEqual(t, info[1].Protocol, ProtocolGRPC)
		testhelpers.AssertEqual(t, info[1].Port, 9090)
	})
}

func TestGRPCMessageTypes(t *testing.T) {
	t.Run("GRPCAnalyzeRequest", func(t *testing.T) {
		req := GRPCAnalyzeRequest{
			ClientID:        "test",
			TLSVersion:      0x0303,
			CipherSuites:    []uint32{0x1301},
			Extensions:      []uint32{0},
			SupportedCurves: []uint32{23},
			Headers:         map[string]string{"key": "value"},
			ClientIP:        "127.0.0.1",
		}
		testhelpers.AssertEqual(t, req.ClientID, "test")
		testhelpers.AssertEqual(t, req.TLSVersion, uint32(0x0303))
	})

	t.Run("GRPCAnalyzeResponse", func(t *testing.T) {
		resp := GRPCAnalyzeResponse{
			FingerprintHash:  "hash123",
			Protocol:         "TLS",
			BrowserFamily:    "Chrome",
			BrowserVersion:   "120",
			Confidence:       0.95,
			RiskLevel:        "low",
			RiskScore:        0.1,
			JA3Hash:          "ja3abc",
			JA4Fingerprint:   "ja4def",
			DefenseHints:     []string{"hint1"},
			Cached:           false,
			ProcessingTimeMs: 10,
		}
		testhelpers.AssertEqual(t, resp.FingerprintHash, "hash123")
		testhelpers.AssertEqual(t, resp.Confidence, 0.95)
	})
}

func TestProtocolTypes(t *testing.T) {
	testhelpers.AssertEqual(t, string(ProtocolHTTP), "http")
	testhelpers.AssertEqual(t, string(ProtocolGRPC), "grpc")
}

func TestServerInfo(t *testing.T) {
	info := ServerInfo{
		Protocol:    ProtocolHTTP,
		Port:        8080,
		Status:      "running",
		Connections: 42,
	}
	
	testhelpers.AssertEqual(t, info.Protocol, ProtocolHTTP)
	testhelpers.AssertEqual(t, info.Port, 8080)
	testhelpers.AssertEqual(t, info.Status, "running")
	testhelpers.AssertEqual(t, info.Connections, 42)
}
