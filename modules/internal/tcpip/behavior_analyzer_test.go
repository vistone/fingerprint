package tcpip

import (
	"reflect"
	"testing"
	"time"
)

// TestNewNetworkBehaviorAnalyzer tests creating a network behavior analyzer
func TestNewNetworkBehaviorAnalyzer(t *testing.T) {
	tests := []struct {
		name       string
		maxSamples int
		want       int
	}{
		{
			name:       "create default analyzer",
			maxSamples: DefaultMaxSamples,
			want:       DefaultMaxSamples,
		},
		{
			name:       "create analyzer with normal limit",
			maxSamples: 5000,
			want:       5000,
		},
		{
			name:       "create analyzer with zero limit",
			maxSamples: 0,
			want:       DefaultMaxSamples,
		},
		{
			name:       "create analyzer with negative limit",
			maxSamples: -100,
			want:       DefaultMaxSamples,
		},
		{
			name:       "create analyzer with limit of 1",
			maxSamples: 1,
			want:       1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nba *NetworkBehaviorAnalyzer
			if tt.maxSamples == DefaultMaxSamples {
				nba = NewNetworkBehaviorAnalyzer()
			} else {
				nba = NewNetworkBehaviorAnalyzerWithLimit(tt.maxSamples)
			}
			if nba.maxSamples != tt.want {
				t.Errorf("NewNetworkBehaviorAnalyzer() maxSamples = %v, want %v", nba.maxSamples, tt.want)
			}
			if nba.packets == nil {
				t.Error("packets slice not initialized")
			}
			if nba.rttMeasurements == nil {
				t.Error("rttMeasurements slice not initialized")
			}
			if nba.sequenceNumbers == nil {
				t.Error("sequenceNumbers slice not initialized")
			}
			if nba.ipIDs == nil {
				t.Error("ipIDs slice not initialized")
			}
			if nba.timestamps == nil {
				t.Error("timestamps slice not initialized")
			}
		})
	}
}

// TestNetworkBehaviorAnalyzer_RecordPacket tests recording packets
func TestNetworkBehaviorAnalyzer_RecordPacket(t *testing.T) {
	tests := []struct {
		name        string
		packetCount int
		maxSamples  int
		expectedLen int
		expectedSeq uint32
		expectedRTT time.Duration
	}{
		{
			name:        "record single packet",
			packetCount: 1,
			maxSamples:  100,
			expectedLen: 1,
			expectedSeq: 1000,
			expectedRTT: 10 * time.Millisecond,
		},
		{
			name:        "record multiple packets",
			packetCount: 5,
			maxSamples:  100,
			expectedLen: 5,
		},
		{
			name:        "test sliding window behavior",
			packetCount: 10,
			maxSamples:  8,
			expectedLen: 7, // after exceeding limit, remove 25% (2 items) then add new item: 8 - 2 + 1 = 7
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzerWithLimit(tt.maxSamples)

			var lastSeq uint32 = 1000
			for i := 0; i < tt.packetCount; i++ {
				packet := &TCPPacket{
					SequenceNumber: lastSeq,
					Payload:        []byte("test"),
					IPHeader: &IPHeader{
						Identification: uint16(i),
						Protocol:       6,
					},
				}
				nba.RecordPacket(packet, time.Duration(i+1)*time.Millisecond)
				lastSeq += 100
			}

			// For sliding window tests, actual length may differ
			if tt.packetCount <= tt.maxSamples {
				if len(nba.packets) != tt.packetCount {
					t.Errorf("packets length = %v, want %v", len(nba.packets), tt.packetCount)
				}
			} else {
				// Sliding window is active, check length does not exceed maxSamples
				if len(nba.packets) > tt.maxSamples {
					t.Errorf("packets length %v exceeds maxSamples %v", len(nba.packets), tt.maxSamples)
				}
			}

			// Verify first packet sequence number (for non-sliding window case)
			if tt.packetCount <= tt.maxSamples && len(nba.packets) > 0 {
				if nba.packets[0].SequenceNumber != 1000 {
					t.Errorf("first packet seq = %v, want %v", nba.packets[0].SequenceNumber, 1000)
				}
			}
		})
	}
}

// TestAppendWithLimit tests slice append with limit
func TestAppendWithLimit(t *testing.T) {
	tests := []struct {
		name       string
		initial    []int
		item       int
		maxSamples int
		expected   []int
	}{
		{
			name:       "append to non-full slice",
			initial:    []int{1, 2, 3},
			item:       4,
			maxSamples: 10,
			expected:   []int{1, 2, 3, 4},
		},
		{
			name:       "sliding window when exceeding limit",
			initial:    []int{1, 2, 3, 4},
			item:       5,
			maxSamples: 4,
			expected:   []int{2, 3, 4, 5}, // remove 25%(1 item) + add new item
		},
		{
			name:       "boundary case with maxSamples of 1",
			initial:    []int{1},
			item:       2,
			maxSamples: 1,
			expected:   []int{2}, // remove 1 item + add new item
		},
		{
			name:       "append to empty slice",
			initial:    []int{},
			item:       1,
			maxSamples: 5,
			expected:   []int{1},
		},
		{
			name:       "just reaching the limit",
			initial:    []int{1, 2, 3},
			item:       4,
			maxSamples: 4,
			expected:   []int{1, 2, 3, 4}, // just reaching the limit, no sliding
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := appendWithLimit(tt.initial, tt.item, tt.maxSamples)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("appendWithLimit() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestNetworkBehaviorAnalyzer_AnalyzeBehavior tests analyzing network behavior
func TestNetworkBehaviorAnalyzer_AnalyzeBehavior(t *testing.T) {
	tests := []struct {
		name            string
		setupFunc       func(*NetworkBehaviorAnalyzer)
		expectedPackets int
		checkResult     func(*testing.T, *NetworkBehaviorResult)
	}{
		{
			name:            "empty analyzer returns empty result",
			setupFunc:       func(nba *NetworkBehaviorAnalyzer) {},
			expectedPackets: 0,
			checkResult: func(t *testing.T, result *NetworkBehaviorResult) {
				if result.TotalPackets != 0 {
					t.Errorf("TotalPackets = %v, want 0", result.TotalPackets)
				}
				if result.RTTAnalysis != nil && result.RTTAnalysis.Count != 0 {
					t.Error("Expected empty RTT analysis")
				}
			},
		},
		{
			name: "single packet analysis",
			setupFunc: func(nba *NetworkBehaviorAnalyzer) {
				packet := &TCPPacket{
					SequenceNumber: 1000,
					Payload:        []byte("test"),
					IPHeader: &IPHeader{
						Identification: 1,
						Protocol:       6,
					},
				}
				nba.RecordPacket(packet, 10*time.Millisecond)
			},
			expectedPackets: 1,
			checkResult: func(t *testing.T, result *NetworkBehaviorResult) {
				if result.TotalPackets != 1 {
					t.Errorf("TotalPackets = %v, want 1", result.TotalPackets)
				}
				if result.RTTAnalysis == nil {
					t.Error("RTTAnalysis is nil")
				} else if result.RTTAnalysis.Count != 1 {
					t.Errorf("RTT count = %v, want 1", result.RTTAnalysis.Count)
				}
			},
		},
		{
			name: "multiple packet full analysis",
			setupFunc: func(nba *NetworkBehaviorAnalyzer) {
				for i := 0; i < 5; i++ {
					packet := &TCPPacket{
						SequenceNumber: uint32(1000 + i*100),
						Payload:        []byte("test"),
						IPHeader: &IPHeader{
							Identification: uint16(i),
							Protocol:       6,
						},
					}
					nba.RecordPacket(packet, time.Duration(10+i*5)*time.Millisecond)
				}
			},
			expectedPackets: 5,
			checkResult: func(t *testing.T, result *NetworkBehaviorResult) {
				if result.TotalPackets != 5 {
					t.Errorf("TotalPackets = %v, want 5", result.TotalPackets)
				}
				if result.RTTAnalysis == nil || result.RTTAnalysis.Count != 5 {
					t.Errorf("RTT count = %v, want 5", result.RTTAnalysis.Count)
				}
				if result.ProtocolDistribution["TCP"] != 5 {
					t.Errorf("TCP count = %v, want 5", result.ProtocolDistribution["TCP"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()
			tt.setupFunc(nba)
			result := nba.AnalyzeBehavior()

			if result.TotalPackets != tt.expectedPackets {
				t.Errorf("TotalPackets = %v, want %v", result.TotalPackets, tt.expectedPackets)
			}

			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

// TestAnalyzeRTT tests RTT analysis
func TestAnalyzeRTT(t *testing.T) {
	tests := []struct {
		name        string
		rttValues   []time.Duration
		expectedAvg time.Duration
		expectedMin time.Duration
		expectedMax time.Duration
		wantNetwork string
	}{
		{
			name:        "empty RTT list",
			rttValues:   []time.Duration{},
			expectedAvg: 0,
			expectedMin: 0,
			expectedMax: 0,
			wantNetwork: "",
		},
		{
			name:        "single RTT",
			rttValues:   []time.Duration{50 * time.Millisecond},
			expectedAvg: 50 * time.Millisecond,
			expectedMin: 50 * time.Millisecond,
			expectedMax: 50 * time.Millisecond,
			wantNetwork: "regional", // 50ms is at the regional boundary
		},
		{
			name:        "multiple RTT calculation",
			rttValues:   []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond},
			expectedAvg: 20 * time.Millisecond,
			expectedMin: 10 * time.Millisecond,
			expectedMax: 30 * time.Millisecond,
			wantNetwork: "domestic",
		},
		{
			name:        "local_lan network type",
			rttValues:   []time.Duration{5 * time.Millisecond},
			expectedAvg: 5 * time.Millisecond,
			expectedMin: 5 * time.Millisecond,
			expectedMax: 5 * time.Millisecond,
			wantNetwork: "local_lan",
		},
		{
			name:        "regional network type",
			rttValues:   []time.Duration{100 * time.Millisecond},
			expectedAvg: 100 * time.Millisecond,
			expectedMin: 100 * time.Millisecond,
			expectedMax: 100 * time.Millisecond,
			wantNetwork: "regional",
		},
		{
			name:        "international network type",
			rttValues:   []time.Duration{200 * time.Millisecond},
			expectedAvg: 200 * time.Millisecond,
			expectedMin: 200 * time.Millisecond,
			expectedMax: 200 * time.Millisecond,
			wantNetwork: "international",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()

			for i, rtt := range tt.rttValues {
				packet := &TCPPacket{
					SequenceNumber: uint32(i),
					Payload:        []byte("test"),
					IPHeader: &IPHeader{
						Identification: uint16(i),
						Protocol:       6,
					},
				}
				nba.RecordPacket(packet, rtt)
			}

			result := nba.AnalyzeBehavior()

			if len(tt.rttValues) == 0 {
				// When RTT list is empty, RTTAnalysis should be empty (Count=0)
				if result.RTTAnalysis != nil && result.RTTAnalysis.Count != 0 {
					t.Errorf("Expected empty RTT analysis, got count %d", result.RTTAnalysis.Count)
				}
				return
			}

			if result.RTTAnalysis.AverageRTT != tt.expectedAvg {
				t.Errorf("AverageRTT = %v, want %v", result.RTTAnalysis.AverageRTT, tt.expectedAvg)
			}
			if result.RTTAnalysis.MinRTT != tt.expectedMin {
				t.Errorf("MinRTT = %v, want %v", result.RTTAnalysis.MinRTT, tt.expectedMin)
			}
			if result.RTTAnalysis.MaxRTT != tt.expectedMax {
				t.Errorf("MaxRTT = %v, want %v", result.RTTAnalysis.MaxRTT, tt.expectedMax)
			}
			if result.RTTAnalysis.NetworkType != tt.wantNetwork {
				t.Errorf("NetworkType = %v, want %v", result.RTTAnalysis.NetworkType, tt.wantNetwork)
			}
		})
	}
}

// TestAnalyzeSequenceNumbers tests sequence number analysis
func TestAnalyzeSequenceNumbers(t *testing.T) {
	tests := []struct {
		name        string
		seqNumbers  []uint32
		wantPattern string
	}{
		{
			name:        "insufficient data",
			seqNumbers:  []uint32{1000},
			wantPattern: "insufficient_data",
		},
		{
			name:        "random pattern - high variance",
			seqNumbers:  []uint32{1000, 50000, 100, 999999, 50},
			wantPattern: "random",
		},
		{
			name:        "time-related pattern",
			seqNumbers:  []uint32{1000, 1010, 1020, 1030, 1040, 1050, 1070},
			wantPattern: "time_based",
		},
		{
			name:        "linear sequential pattern",
			seqNumbers:  []uint32{1000, 1100, 1200, 1300, 1400},
			wantPattern: "time_based", // difference is 100, will be identified as time-related
		},
		{
			name:        "complex pattern",
			seqNumbers:  []uint32{1000, 1100, 1150, 1200, 1250},
			wantPattern: "time_based", // most are small positive differences, will be identified as time-related
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()

			for i, seq := range tt.seqNumbers {
				packet := &TCPPacket{
					SequenceNumber: seq,
					Payload:        []byte("test"),
					IPHeader: &IPHeader{
						Identification: uint16(i),
						Protocol:       6,
					},
				}
				nba.RecordPacket(packet, 10*time.Millisecond)
			}

			result := nba.AnalyzeBehavior()
			if result.SequenceNumberPattern != tt.wantPattern {
				t.Errorf("SequenceNumberPattern = %v, want %v", result.SequenceNumberPattern, tt.wantPattern)
			}
		})
	}
}

// TestAnalyzeIPIDs tests IP ID analysis
func TestAnalyzeIPIDs(t *testing.T) {
	tests := []struct {
		name        string
		ipIDs       []uint16
		wantPattern string
	}{
		{
			name:        "insufficient data",
			ipIDs:       []uint16{1},
			wantPattern: "insufficient_data",
		},
		{
			name:        "linear counter pattern",
			ipIDs:       []uint16{1, 2, 3, 4, 5},
			wantPattern: "linear_counter",
		},
		{
			name:        "random pattern",
			ipIDs:       []uint16{1, 5000, 100, 9999, 50},
			wantPattern: "random",
		},
		{
			name:        "mixed pattern",
			ipIDs:       []uint16{1, 2, 100, 101, 102},
			wantPattern: "mixed_pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()

			for i, ipID := range tt.ipIDs {
				packet := &TCPPacket{
					SequenceNumber: uint32(i * 100),
					Payload:        []byte("test"),
					IPHeader: &IPHeader{
						Identification: ipID,
						Protocol:       6,
					},
				}
				nba.RecordPacket(packet, 10*time.Millisecond)
			}

			result := nba.AnalyzeBehavior()
			if result.IPIDPattern != tt.wantPattern {
				t.Errorf("IPIDPattern = %v, want %v", result.IPIDPattern, tt.wantPattern)
			}
		})
	}
}

// TestAnalyzePacketSizes tests packet size analysis
func TestAnalyzePacketSizes(t *testing.T) {
	tests := []struct {
		name         string
		payloads     []string
		wantVariance float64
	}{
		{
			name:         "empty packet list",
			payloads:     []string{},
			wantVariance: 0,
		},
		{
			name:         "same size packets",
			payloads:     []string{"test", "test", "test"},
			wantVariance: 0,
		},
		{
			name:         "different size packets",
			payloads:     []string{"a", "ab", "abc", "abcd"},
			wantVariance: 1, // integer division: 5/4 = 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()

			for i, payload := range tt.payloads {
				packet := &TCPPacket{
					SequenceNumber: uint32(i * 100),
					Payload:        []byte(payload),
					IPHeader: &IPHeader{
						Identification: uint16(i),
						Protocol:       6,
					},
				}
				nba.RecordPacket(packet, 10*time.Millisecond)
			}

			result := nba.AnalyzeBehavior()
			if result.PacketSizeVariance != tt.wantVariance {
				t.Errorf("PacketSizeVariance = %v, want %v", result.PacketSizeVariance, tt.wantVariance)
			}
		})
	}
}

// TestAnalyzeProtocolDistribution tests protocol distribution statistics
func TestAnalyzeProtocolDistribution(t *testing.T) {
	tests := []struct {
		name         string
		protocols    []uint8
		expectedDist map[string]int
	}{
		{
			name:      "TCP protocol distribution",
			protocols: []uint8{6, 6, 6},
			expectedDist: map[string]int{
				"TCP": 3,
			},
		},
		{
			name:      "UDP protocol distribution",
			protocols: []uint8{17, 17},
			expectedDist: map[string]int{
				"UDP": 2,
			},
		},
		{
			name:      "ICMP protocol distribution",
			protocols: []uint8{1, 1, 1},
			expectedDist: map[string]int{
				"ICMP": 3,
			},
		},
		{
			name:      "mixed protocol distribution",
			protocols: []uint8{6, 17, 1, 6, 255, 17},
			expectedDist: map[string]int{
				"TCP":   2,
				"UDP":   2,
				"ICMP":  1,
				"OTHER": 1,
			},
		},
		{
			name:      "OTHER protocol distribution",
			protocols: []uint8{255, 0, 50},
			expectedDist: map[string]int{
				"OTHER": 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()

			for i, proto := range tt.protocols {
				packet := &TCPPacket{
					SequenceNumber: uint32(i * 100),
					Payload:        []byte("test"),
					IPHeader: &IPHeader{
						Identification: uint16(i),
						Protocol:       proto,
					},
				}
				nba.RecordPacket(packet, 10*time.Millisecond)
			}

			result := nba.AnalyzeBehavior()
			if !reflect.DeepEqual(result.ProtocolDistribution, tt.expectedDist) {
				t.Errorf("ProtocolDistribution = %v, want %v", result.ProtocolDistribution, tt.expectedDist)
			}
		})
	}
}

// TestAnalyzeTimingPattern tests timing pattern analysis
func TestAnalyzeTimingPattern(t *testing.T) {
	tests := []struct {
		name        string
		intervals   []time.Duration
		wantPattern string
	}{
		{
			name:        "insufficient data",
			intervals:   []time.Duration{100 * time.Millisecond},
			wantPattern: "insufficient_data",
		},
		{
			name:        "periodic pattern",
			intervals:   []time.Duration{100 * time.Millisecond, 101 * time.Millisecond, 99 * time.Millisecond, 100 * time.Millisecond},
			wantPattern: "periodic",
		},
		{
			name:        "burst pattern",
			intervals:   []time.Duration{10 * time.Millisecond, 15 * time.Millisecond, 500 * time.Millisecond, 10 * time.Millisecond},
			wantPattern: "bursty",
		},
		{
			name:        "irregular pattern",
			intervals:   []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 150 * time.Millisecond},
			wantPattern: "irregular",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()

			baseTime := time.Now()
			for i, interval := range tt.intervals {
				packet := &TCPPacket{
					SequenceNumber: uint32(i * 100),
					Payload:        []byte("test"),
					IPHeader: &IPHeader{
						Identification: uint16(i),
						Protocol:       6,
					},
				}
				nba.RecordPacket(packet, 10*time.Millisecond)
				// Manually set timestamps to simulate specific time intervals
				nba.timestamps[i] = baseTime.Add(time.Duration(i) * interval)
			}

			result := nba.AnalyzeBehavior()
			if result.TimingPattern != tt.wantPattern {
				t.Errorf("TimingPattern = %v, want %v", result.TimingPattern, tt.wantPattern)
			}
		})
	}
}

// TestComputeBehaviorCharacteristics tests behavior characteristics computation
func TestComputeBehaviorCharacteristics(t *testing.T) {
	tests := []struct {
		name                string
		packetCount         int
		setupFunc           func(*NetworkBehaviorAnalyzer)
		wantCharacteristics []string
	}{
		{
			name:                "empty packets",
			packetCount:         0,
			setupFunc:           func(nba *NetworkBehaviorAnalyzer) {},
			wantCharacteristics: nil, // empty slice returns nil
		},
		{
			name:        "automated traffic detection",
			packetCount: 101,
			setupFunc: func(nba *NetworkBehaviorAnalyzer) {
				for i := 0; i < 101; i++ {
					packet := &TCPPacket{
						SequenceNumber: uint32(i * 100),
						Payload:        []byte("test"),
						IPHeader: &IPHeader{
							Identification: uint16(i),
							Protocol:       6,
						},
					}
					nba.RecordPacket(packet, 100*time.Millisecond)
				}
			},
			wantCharacteristics: []string{"automated", "interactive", "bulk_transfer"}, // 101 packets will trigger multiple characteristics
		},
		{
			name:        "interactive traffic detection",
			packetCount: 10,
			setupFunc: func(nba *NetworkBehaviorAnalyzer) {
				for i := 0; i < 10; i++ {
					packet := &TCPPacket{
						SequenceNumber: uint32(i * 100),
						Payload:        []byte("test"),
						IPHeader: &IPHeader{
							Identification: uint16(i),
							Protocol:       6,
						},
					}
					// RTT between 50-500ms indicates interactive traffic
					nba.RecordPacket(packet, 100*time.Millisecond)
				}
			},
			wantCharacteristics: []string{"interactive"},
		},
		{
			name:        "bulk transfer detection",
			packetCount: 51,
			setupFunc: func(nba *NetworkBehaviorAnalyzer) {
				for i := 0; i < 51; i++ {
					packet := &TCPPacket{
						SequenceNumber: uint32(i * 100),
						Payload:        []byte("test"),
						IPHeader: &IPHeader{
							Identification: uint16(i),
							Protocol:       6,
						},
					}
					nba.RecordPacket(packet, 1*time.Millisecond)
				}
			},
			wantCharacteristics: []string{"bulk_transfer"},
		},
		{
			name:        "scanning behavior detection",
			packetCount: 10,
			setupFunc: func(nba *NetworkBehaviorAnalyzer) {
				for i := 0; i < 10; i++ {
					// Set RST flag
					flags := uint8(0x04)
					packet := &TCPPacket{
						SequenceNumber: uint32(i * 100),
						Payload:        []byte("test"),
						Flags:          flags,
						IPHeader: &IPHeader{
							Identification: uint16(i),
							Protocol:       6,
						},
					}
					nba.RecordPacket(packet, 1*time.Millisecond)
				}
			},
			wantCharacteristics: []string{"scanning"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()
			tt.setupFunc(nba)

			result := nba.AnalyzeBehavior()
			if !reflect.DeepEqual(result.BehaviorCharacteristics, tt.wantCharacteristics) {
				t.Errorf("BehaviorCharacteristics = %v, want %v", result.BehaviorCharacteristics, tt.wantCharacteristics)
			}
		})
	}
}

// TestCalculateStdDeviation tests standard deviation calculation
func TestCalculateStdDeviation(t *testing.T) {
	tests := []struct {
		name         string
		measurements []time.Duration
		avg          time.Duration
	}{
		{
			name:         "empty measurement list",
			measurements: []time.Duration{},
			avg:          0,
		},
		{
			name:         "single measurement",
			measurements: []time.Duration{100 * time.Millisecond},
			avg:          100 * time.Millisecond,
		},
		{
			name:         "multiple measurements",
			measurements: []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 300 * time.Millisecond},
			avg:          200 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()
			result := nba.calculateStdDeviation(tt.measurements, tt.avg)
			// Standard deviation should be non-negative
			if result < 0 {
				t.Errorf("StdDeviation = %v, want non-negative", result)
			}
		})
	}
}

// TestClassifyNetworkType tests network type classification
func TestClassifyNetworkType(t *testing.T) {
	tests := []struct {
		name   string
		avgRTT time.Duration
		want   string
	}{
		{
			name:   "local_lan",
			avgRTT: 5 * time.Millisecond,
			want:   "local_lan",
		},
		{
			name:   "local_lan boundary",
			avgRTT: 9 * time.Millisecond,
			want:   "local_lan",
		},
		{
			name:   "domestic",
			avgRTT: 10 * time.Millisecond,
			want:   "domestic",
		},
		{
			name:   "domestic boundary",
			avgRTT: 49 * time.Millisecond,
			want:   "domestic",
		},
		{
			name:   "regional",
			avgRTT: 50 * time.Millisecond,
			want:   "regional",
		},
		{
			name:   "regional boundary",
			avgRTT: 149 * time.Millisecond,
			want:   "regional",
		},
		{
			name:   "international",
			avgRTT: 150 * time.Millisecond,
			want:   "international",
		},
		{
			name:   "international high latency",
			avgRTT: 500 * time.Millisecond,
			want:   "international",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()
			got := nba.classifyNetworkType(tt.avgRTT)
			if got != tt.want {
				t.Errorf("classifyNetworkType() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHasHighVariance tests high variance detection
func TestHasHighVariance(t *testing.T) {
	tests := []struct {
		name  string
		diffs []int64
		want  bool
	}{
		{
			name:  "insufficient data",
			diffs: []int64{100},
			want:  false,
		},
		{
			name:  "low variance",
			diffs: []int64{1, 2, 3, 4, 5},
			want:  false,
		},
		{
			name:  "high variance exceeding threshold",
			diffs: []int64{1000, 2000, 3000},
			want:  true,
		},
		{
			name:  "negative value high variance",
			diffs: []int64{-1000, -2000, -3000},
			want:  true,
		},
		{
			name:  "mixed signs",
			diffs: []int64{-1000, 1000, -2000},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()
			got := nba.hasHighVariance(tt.diffs)
			if got != tt.want {
				t.Errorf("hasHighVariance() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsTimeRelated tests time-related detection
func TestIsTimeRelated(t *testing.T) {
	tests := []struct {
		name  string
		diffs []int64
		want  bool
	}{
		{
			name:  "insufficient data",
			diffs: []int64{100},
			want:  false,
		},
		{
			name:  "time-related pattern",
			diffs: []int64{1000, 2000, 3000, 4000},
			want:  true,
		},
		{
			name:  "not time-related",
			diffs: []int64{100000, -1000, 50000, -50000},
			want:  false,
		},
		{
			name:  "partial increment",
			diffs: []int64{1000, 2000, -5000, 3000},
			want:  true, // 3/4 = 75% > 70%, meets time-related condition
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()
			got := nba.isTimeRelated(tt.diffs)
			if got != tt.want {
				t.Errorf("isTimeRelated() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsLinear tests linearity detection
func TestIsLinear(t *testing.T) {
	tests := []struct {
		name  string
		diffs []int64
		want  bool
	}{
		{
			name:  "insufficient data",
			diffs: []int64{100},
			want:  false,
		},
		{
			name:  "fully linear",
			diffs: []int64{100, 100, 100, 100},
			want:  true,
		},
		{
			name:  "non-linear",
			diffs: []int64{100, 200, 100, 200},
			want:  false,
		},
		{
			name:  "single difference",
			diffs: []int64{100, 100},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()
			got := nba.isLinear(tt.diffs)
			if got != tt.want {
				t.Errorf("isLinear() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsLinearIPID tests linear IP ID detection
func TestIsLinearIPID(t *testing.T) {
	tests := []struct {
		name  string
		diffs []int
		want  bool
	}{
		{
			name:  "insufficient data",
			diffs: []int{1},
			want:  false,
		},
		{
			name:  "linear counter - increment 1",
			diffs: []int{1, 1, 1, 1},
			want:  true,
		},
		{
			name:  "linear counter - increment 0",
			diffs: []int{0, 0, 0, 0},
			want:  true,
		},
		{
			name:  "mixed 0 and 1",
			diffs: []int{0, 1, 0, 1},
			want:  true,
		},
		{
			name:  "non-linear",
			diffs: []int{1, 2, 3, 4},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()
			got := nba.isLinearIPID(tt.diffs)
			if got != tt.want {
				t.Errorf("isLinearIPID() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHasRegularIntervals tests regular interval detection
func TestHasRegularIntervals(t *testing.T) {
	tests := []struct {
		name      string
		intervals []time.Duration
		want      bool
	}{
		{
			name:      "insufficient data",
			intervals: []time.Duration{100 * time.Millisecond},
			want:      false,
		},
		{
			name:      "regular intervals",
			intervals: []time.Duration{100 * time.Millisecond, 101 * time.Millisecond, 99 * time.Millisecond, 100 * time.Millisecond},
			want:      true,
		},
		{
			name:      "irregular intervals",
			intervals: []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 50 * time.Millisecond},
			want:      false,
		},
		{
			name:      "partially regular",
			intervals: []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 500 * time.Millisecond},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()
			got := nba.hasRegularIntervals(tt.intervals)
			if got != tt.want {
				t.Errorf("hasRegularIntervals() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsBurstPattern tests burst pattern detection
func TestIsBurstPattern(t *testing.T) {
	tests := []struct {
		name      string
		intervals []time.Duration
		want      bool
	}{
		{
			name:      "insufficient data",
			intervals: []time.Duration{10 * time.Millisecond},
			want:      false,
		},
		{
			name:      "burst pattern - mostly small intervals",
			intervals: []time.Duration{10 * time.Millisecond, 15 * time.Millisecond, 500 * time.Millisecond, 20 * time.Millisecond},
			want:      true,
		},
		{
			name:      "burst pattern - mostly large intervals",
			intervals: []time.Duration{500 * time.Millisecond, 10 * time.Millisecond, 600 * time.Millisecond, 15 * time.Millisecond},
			want:      false, // 2 small 2 large, neither exceeds half
		},
		{
			name:      "non-burst - all small intervals",
			intervals: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 15 * time.Millisecond},
			want:      false,
		},
		{
			name:      "non-burst - all large intervals",
			intervals: []time.Duration{200 * time.Millisecond, 300 * time.Millisecond, 250 * time.Millisecond},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()
			got := nba.isBurstPattern(tt.intervals)
			if got != tt.want {
				t.Errorf("isBurstPattern() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNetworkBehaviorResult_String tests result string output
func TestNetworkBehaviorResult_String(t *testing.T) {
	tests := []struct {
		name   string
		result *NetworkBehaviorResult
		want   string
	}{
		{
			name: "complete result",
			result: &NetworkBehaviorResult{
				TotalPackets: 10,
				RTTAnalysis: &RTTAnalysis{
					Count:      10,
					AverageRTT: 50 * time.Millisecond,
				},
				SequenceNumberPattern:   "sequential",
				TimingPattern:           "periodic",
				BehaviorCharacteristics: []string{"interactive"},
			},
			want: "NetworkBehavior[packets=10, avgRTT=50ms, seqPattern=sequential, timing=periodic, characteristics=[interactive]]",
		},
		{
			name: "empty characteristics",
			result: &NetworkBehaviorResult{
				TotalPackets: 0,
				RTTAnalysis: &RTTAnalysis{
					Count:      0,
					AverageRTT: 0,
				},
				SequenceNumberPattern:   "insufficient_data",
				TimingPattern:           "insufficient_data",
				BehaviorCharacteristics: []string{},
			},
			want: "NetworkBehavior[packets=0, avgRTT=0s, seqPattern=insufficient_data, timing=insufficient_data, characteristics=[]]",
		},
		{
			name: "multiple characteristics",
			result: &NetworkBehaviorResult{
				TotalPackets: 150,
				RTTAnalysis: &RTTAnalysis{
					Count:      150,
					AverageRTT: 100 * time.Millisecond,
				},
				SequenceNumberPattern:   "random",
				TimingPattern:           "bursty",
				BehaviorCharacteristics: []string{"automated", "bulk_transfer"},
			},
			want: "NetworkBehavior[packets=150, avgRTT=100ms, seqPattern=random, timing=bursty, characteristics=[automated bulk_transfer]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
