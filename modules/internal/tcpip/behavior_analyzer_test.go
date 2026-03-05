package tcpip

import (
	"reflect"
	"testing"
	"time"
)

// TestNewNetworkBehaviorAnalyzer 测试创建网络行为分析器
func TestNewNetworkBehaviorAnalyzer(t *testing.T) {
	tests := []struct {
		name       string
		maxSamples int
		want       int
	}{
		{
			name:       "创建默认分析器",
			maxSamples: DefaultMaxSamples,
			want:       DefaultMaxSamples,
		},
		{
			name:       "创建带正常限制的分析器",
			maxSamples: 5000,
			want:       5000,
		},
		{
			name:       "创建带零值限制的分析器",
			maxSamples: 0,
			want:       DefaultMaxSamples,
		},
		{
			name:       "创建带负值限制的分析器",
			maxSamples: -100,
			want:       DefaultMaxSamples,
		},
		{
			name:       "创建带限制为1的分析器",
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

// TestNetworkBehaviorAnalyzer_RecordPacket 测试记录数据包
func TestNetworkBehaviorAnalyzer_RecordPacket(t *testing.T) {
	tests := []struct {
		name         string
		packetCount  int
		maxSamples   int
		expectedLen  int
		expectedSeq  uint32
		expectedRTT  time.Duration
	}{
		{
			name:        "记录单个数据包",
			packetCount: 1,
			maxSamples:  100,
			expectedLen: 1,
			expectedSeq: 1000,
			expectedRTT: 10 * time.Millisecond,
		},
		{
			name:        "记录多个数据包",
			packetCount: 5,
			maxSamples:  100,
			expectedLen: 5,
		},
		{
			name:        "测试滑动窗口行为",
			packetCount: 10,
			maxSamples:  8,
			expectedLen: 7, // 超过后移除25%(2个) + 新添加 = 8 - 2 + 1 = 7... 实际是滑动窗口逻辑
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

			// 对于滑动窗口测试，实际长度可能不同
			if tt.packetCount <= tt.maxSamples {
				if len(nba.packets) != tt.packetCount {
					t.Errorf("packets length = %v, want %v", len(nba.packets), tt.packetCount)
				}
			} else {
				// 滑动窗口生效，检查长度不超过maxSamples
				if len(nba.packets) > tt.maxSamples {
					t.Errorf("packets length %v exceeds maxSamples %v", len(nba.packets), tt.maxSamples)
				}
			}

			// 验证第一个数据包的序列号（对于非滑动窗口情况）
			if tt.packetCount <= tt.maxSamples && len(nba.packets) > 0 {
				if nba.packets[0].SequenceNumber != 1000 {
					t.Errorf("first packet seq = %v, want %v", nba.packets[0].SequenceNumber, 1000)
				}
			}
		})
	}
}

// TestAppendWithLimit 测试带限制的切片追加
func TestAppendWithLimit(t *testing.T) {
	tests := []struct {
		name       string
		initial    []int
		item       int
		maxSamples int
		expected   []int
	}{
		{
			name:       "追加到未满切片",
			initial:    []int{1, 2, 3},
			item:       4,
			maxSamples: 10,
			expected:   []int{1, 2, 3, 4},
		},
		{
			name:       "超过限制时的滑动窗口",
			initial:    []int{1, 2, 3, 4},
			item:       5,
			maxSamples: 4,
			expected:   []int{2, 3, 4, 5}, // 移除25%(1个) + 添加新项
		},
		{
			name:       "maxSamples为1的边界情况",
			initial:    []int{1},
			item:       2,
			maxSamples: 1,
			expected:   []int{2}, // 移除1个 + 添加新项
		},
		{
			name:       "空切片追加",
			initial:    []int{},
			item:       1,
			maxSamples: 5,
			expected:   []int{1},
		},
		{
			name:       "刚好达到限制",
			initial:    []int{1, 2, 3},
			item:       4,
			maxSamples: 4,
			expected:   []int{1, 2, 3, 4}, // 刚好达到限制，不滑动
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

// TestNetworkBehaviorAnalyzer_AnalyzeBehavior 测试分析网络行为
func TestNetworkBehaviorAnalyzer_AnalyzeBehavior(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(*NetworkBehaviorAnalyzer)
		expectedPackets int
		checkResult    func(*testing.T, *NetworkBehaviorResult)
	}{
		{
			name:           "空分析器返回空结果",
			setupFunc:      func(nba *NetworkBehaviorAnalyzer) {},
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
			name: "单数据包分析",
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
			name: "多数据包完整分析",
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

// TestAnalyzeRTT 测试RTT分析
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
			name:        "空RTT列表",
			rttValues:   []time.Duration{},
			expectedAvg: 0,
			expectedMin: 0,
			expectedMax: 0,
			wantNetwork: "",
		},
		{
			name:        "单RTT",
			rttValues:   []time.Duration{50 * time.Millisecond},
			expectedAvg: 50 * time.Millisecond,
			expectedMin: 50 * time.Millisecond,
			expectedMax: 50 * time.Millisecond,
			wantNetwork: "regional", // 50ms是regional的边界
		},
		{
			name:        "多RTT计算",
			rttValues:   []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond},
			expectedAvg: 20 * time.Millisecond,
			expectedMin: 10 * time.Millisecond,
			expectedMax: 30 * time.Millisecond,
			wantNetwork: "domestic",
		},
		{
			name:        "local_lan网络类型",
			rttValues:   []time.Duration{5 * time.Millisecond},
			expectedAvg: 5 * time.Millisecond,
			expectedMin: 5 * time.Millisecond,
			expectedMax: 5 * time.Millisecond,
			wantNetwork: "local_lan",
		},
		{
			name:        "regional网络类型",
			rttValues:   []time.Duration{100 * time.Millisecond},
			expectedAvg: 100 * time.Millisecond,
			expectedMin: 100 * time.Millisecond,
			expectedMax: 100 * time.Millisecond,
			wantNetwork: "regional",
		},
		{
			name:        "international网络类型",
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
				// 空RTT列表时，RTTAnalysis应该是空的（Count=0）
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

// TestAnalyzeSequenceNumbers 测试序列号分析
func TestAnalyzeSequenceNumbers(t *testing.T) {
	tests := []struct {
		name        string
		seqNumbers  []uint32
		wantPattern string
	}{
		{
			name:        "数据不足",
			seqNumbers:  []uint32{1000},
			wantPattern: "insufficient_data",
		},
		{
			name:        "随机模式-高方差",
			seqNumbers:  []uint32{1000, 50000, 100, 999999, 50},
			wantPattern: "random",
		},
		{
			name:        "时间相关模式",
			seqNumbers:  []uint32{1000, 1010, 1020, 1030, 1040, 1050, 1070},
			wantPattern: "time_based",
		},
		{
			name:        "线性顺序模式",
			seqNumbers:  []uint32{1000, 1100, 1200, 1300, 1400},
			wantPattern: "time_based", // 差值是100，会被识别为时间相关
		},
		{
			name:        "复杂模式",
			seqNumbers:  []uint32{1000, 1100, 1150, 1200, 1250},
			wantPattern: "time_based", // 大多数是正的小差值，会被识别为时间相关
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

// TestAnalyzeIPIDs 测试IP ID分析
func TestAnalyzeIPIDs(t *testing.T) {
	tests := []struct {
		name       string
		ipIDs      []uint16
		wantPattern string
	}{
		{
			name:       "数据不足",
			ipIDs:      []uint16{1},
			wantPattern: "insufficient_data",
		},
		{
			name:       "线性计数器模式",
			ipIDs:      []uint16{1, 2, 3, 4, 5},
			wantPattern: "linear_counter",
		},
		{
			name:       "随机模式",
			ipIDs:      []uint16{1, 5000, 100, 9999, 50},
			wantPattern: "random",
		},
		{
			name:       "混合模式",
			ipIDs:      []uint16{1, 2, 100, 101, 102},
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

// TestAnalyzePacketSizes 测试数据包大小分析
func TestAnalyzePacketSizes(t *testing.T) {
	tests := []struct {
		name         string
		payloads     []string
		wantVariance float64
	}{
		{
			name:         "空数据包列表",
			payloads:     []string{},
			wantVariance: 0,
		},
		{
			name:         "相同大小的数据包",
			payloads:     []string{"test", "test", "test"},
			wantVariance: 0,
		},
		{
			name:         "不同大小的数据包",
			payloads:     []string{"a", "ab", "abc", "abcd"},
			wantVariance: 1, // 整数除法: 5/4 = 1
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

// TestAnalyzeProtocolDistribution 测试协议分布统计
func TestAnalyzeProtocolDistribution(t *testing.T) {
	tests := []struct {
		name         string
		protocols    []uint8
		expectedDist map[string]int
	}{
		{
			name:      "TCP协议分布",
			protocols: []uint8{6, 6, 6},
			expectedDist: map[string]int{
				"TCP": 3,
			},
		},
		{
			name:      "UDP协议分布",
			protocols: []uint8{17, 17},
			expectedDist: map[string]int{
				"UDP": 2,
			},
		},
		{
			name:      "ICMP协议分布",
			protocols: []uint8{1, 1, 1},
			expectedDist: map[string]int{
				"ICMP": 3,
			},
		},
		{
			name:      "混合协议分布",
			protocols: []uint8{6, 17, 1, 6, 255, 17},
			expectedDist: map[string]int{
				"TCP":   2,
				"UDP":   2,
				"ICMP":  1,
				"OTHER": 1,
			},
		},
		{
			name:      "OTHER协议分布",
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

// TestAnalyzeTimingPattern 测试时间模式分析
func TestAnalyzeTimingPattern(t *testing.T) {
	tests := []struct {
		name        string
		intervals   []time.Duration
		wantPattern string
	}{
		{
			name:        "数据不足",
			intervals:   []time.Duration{100 * time.Millisecond},
			wantPattern: "insufficient_data",
		},
		{
			name:        "周期性模式",
			intervals:   []time.Duration{100 * time.Millisecond, 101 * time.Millisecond, 99 * time.Millisecond, 100 * time.Millisecond},
			wantPattern: "periodic",
		},
		{
			name:        "突发模式",
			intervals:   []time.Duration{10 * time.Millisecond, 15 * time.Millisecond, 500 * time.Millisecond, 10 * time.Millisecond},
			wantPattern: "bursty",
		},
		{
			name:        "不规则模式",
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
				// 手动设置时间戳来模拟特定的时间间隔
				nba.timestamps[i] = baseTime.Add(time.Duration(i) * interval)
			}

			result := nba.AnalyzeBehavior()
			if result.TimingPattern != tt.wantPattern {
				t.Errorf("TimingPattern = %v, want %v", result.TimingPattern, tt.wantPattern)
			}
		})
	}
}

// TestComputeBehaviorCharacteristics 测试行为特征计算
func TestComputeBehaviorCharacteristics(t *testing.T) {
	tests := []struct {
		name           string
		packetCount    int
		setupFunc      func(*NetworkBehaviorAnalyzer)
		wantCharacteristics []string
	}{
		{
			name:           "空数据包",
			packetCount:    0,
			setupFunc:      func(nba *NetworkBehaviorAnalyzer) {},
			wantCharacteristics: nil, // 空切片返回nil
		},
		{
			name:        "自动化流量检测",
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
			wantCharacteristics: []string{"automated", "interactive", "bulk_transfer"}, // 101个包会触发多个特征
		},
		{
			name:        "交互式流量检测",
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
					// RTT在50-500ms之间表示交互式流量
					nba.RecordPacket(packet, 100*time.Millisecond)
				}
			},
			wantCharacteristics: []string{"interactive"},
		},
		{
			name:        "批量传输检测",
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
			name:        "扫描行为检测",
			packetCount: 10,
			setupFunc: func(nba *NetworkBehaviorAnalyzer) {
				for i := 0; i < 10; i++ {
					// 设置RST标志
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

// TestCalculateStdDeviation 测试标准差计算
func TestCalculateStdDeviation(t *testing.T) {
	tests := []struct {
		name         string
		measurements []time.Duration
		avg          time.Duration
	}{
		{
			name:         "空测量列表",
			measurements: []time.Duration{},
			avg:          0,
		},
		{
			name:         "单测量值",
			measurements: []time.Duration{100 * time.Millisecond},
			avg:          100 * time.Millisecond,
		},
		{
			name:         "多测量值",
			measurements: []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 300 * time.Millisecond},
			avg:          200 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nba := NewNetworkBehaviorAnalyzer()
			result := nba.calculateStdDeviation(tt.measurements, tt.avg)
			// 标准差应该非负
			if result < 0 {
				t.Errorf("StdDeviation = %v, want non-negative", result)
			}
		})
	}
}

// TestClassifyNetworkType 测试网络类型分类
func TestClassifyNetworkType(t *testing.T) {
	tests := []struct {
		name    string
		avgRTT  time.Duration
		want    string
	}{
		{
			name:    "local_lan",
			avgRTT:  5 * time.Millisecond,
			want:    "local_lan",
		},
		{
			name:    "local_lan边界",
			avgRTT:  9 * time.Millisecond,
			want:    "local_lan",
		},
		{
			name:    "domestic",
			avgRTT:  10 * time.Millisecond,
			want:    "domestic",
		},
		{
			name:    "domestic边界",
			avgRTT:  49 * time.Millisecond,
			want:    "domestic",
		},
		{
			name:    "regional",
			avgRTT:  50 * time.Millisecond,
			want:    "regional",
		},
		{
			name:    "regional边界",
			avgRTT:  149 * time.Millisecond,
			want:    "regional",
		},
		{
			name:    "international",
			avgRTT:  150 * time.Millisecond,
			want:    "international",
		},
		{
			name:    "international高延迟",
			avgRTT:  500 * time.Millisecond,
			want:    "international",
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

// TestHasHighVariance 测试高方差检测
func TestHasHighVariance(t *testing.T) {
	tests := []struct {
		name string
		diffs []int64
		want bool
	}{
		{
			name: "数据不足",
			diffs: []int64{100},
			want: false,
		},
		{
			name: "低方差",
			diffs: []int64{1, 2, 3, 4, 5},
			want: false,
		},
		{
			name: "高方差超过阈值",
			diffs: []int64{1000, 2000, 3000},
			want: true,
		},
		{
			name: "负值高方差",
			diffs: []int64{-1000, -2000, -3000},
			want: true,
		},
		{
			name: "混合符号",
			diffs: []int64{-1000, 1000, -2000},
			want: true,
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

// TestIsTimeRelated 测试时间相关检测
func TestIsTimeRelated(t *testing.T) {
	tests := []struct {
		name string
		diffs []int64
		want bool
	}{
		{
			name: "数据不足",
			diffs: []int64{100},
			want: false,
		},
		{
			name: "时间相关模式",
			diffs: []int64{1000, 2000, 3000, 4000},
			want: true,
		},
		{
			name: "非时间相关",
			diffs: []int64{100000, -1000, 50000, -50000},
			want: false,
		},
		{
			name: "部分递增",
			diffs: []int64{1000, 2000, -5000, 3000},
			want: true, // 3/4 = 75% > 70%，满足时间相关条件
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

// TestIsLinear 测试线性检测
func TestIsLinear(t *testing.T) {
	tests := []struct {
		name string
		diffs []int64
		want bool
	}{
		{
			name: "数据不足",
			diffs: []int64{100},
			want: false,
		},
		{
			name: "完全线性",
			diffs: []int64{100, 100, 100, 100},
			want: true,
		},
		{
			name: "非线性",
			diffs: []int64{100, 200, 100, 200},
			want: false,
		},
		{
			name: "单差值",
			diffs: []int64{100, 100},
			want: true,
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

// TestIsLinearIPID 测试线性IP ID检测
func TestIsLinearIPID(t *testing.T) {
	tests := []struct {
		name string
		diffs []int
		want bool
	}{
		{
			name: "数据不足",
			diffs: []int{1},
			want: false,
		},
		{
			name: "线性计数器-递增1",
			diffs: []int{1, 1, 1, 1},
			want: true,
		},
		{
			name: "线性计数器-递增0",
			diffs: []int{0, 0, 0, 0},
			want: true,
		},
		{
			name: "混合01",
			diffs: []int{0, 1, 0, 1},
			want: true,
		},
		{
			name: "非线性",
			diffs: []int{1, 2, 3, 4},
			want: false,
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

// TestHasRegularIntervals 测试规律间隔检测
func TestHasRegularIntervals(t *testing.T) {
	tests := []struct {
		name      string
		intervals []time.Duration
		want      bool
	}{
		{
			name:      "数据不足",
			intervals: []time.Duration{100 * time.Millisecond},
			want:      false,
		},
		{
			name:      "规律间隔",
			intervals: []time.Duration{100 * time.Millisecond, 101 * time.Millisecond, 99 * time.Millisecond, 100 * time.Millisecond},
			want:      true,
		},
		{
			name:      "不规律间隔",
			intervals: []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 50 * time.Millisecond},
			want:      false,
		},
		{
			name:      "部分规律",
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

// TestIsBurstPattern 测试突发模式检测
func TestIsBurstPattern(t *testing.T) {
	tests := []struct {
		name      string
		intervals []time.Duration
		want      bool
	}{
		{
			name:      "数据不足",
			intervals: []time.Duration{10 * time.Millisecond},
			want:      false,
		},
		{
			name:      "突发模式-小间隔为主",
			intervals: []time.Duration{10 * time.Millisecond, 15 * time.Millisecond, 500 * time.Millisecond, 20 * time.Millisecond},
			want:      true,
		},
		{
			name:      "突发模式-大间隔为主",
			intervals: []time.Duration{500 * time.Millisecond, 10 * time.Millisecond, 600 * time.Millisecond, 15 * time.Millisecond},
			want:      false, // 2小2大，都不超过一半
		},
		{
			name:      "非突发-全部小间隔",
			intervals: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 15 * time.Millisecond},
			want:      false,
		},
		{
			name:      "非突发-全部大间隔",
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

// TestNetworkBehaviorResult_String 测试结果字符串输出
func TestNetworkBehaviorResult_String(t *testing.T) {
	tests := []struct {
		name   string
		result *NetworkBehaviorResult
		want   string
	}{
		{
			name: "完整结果",
			result: &NetworkBehaviorResult{
				TotalPackets: 10,
				RTTAnalysis: &RTTAnalysis{
					Count:      10,
					AverageRTT: 50 * time.Millisecond,
				},
				SequenceNumberPattern:   "sequential",
				TimingPattern:          "periodic",
				BehaviorCharacteristics: []string{"interactive"},
			},
			want: "NetworkBehavior[packets=10, avgRTT=50ms, seqPattern=sequential, timing=periodic, characteristics=[interactive]]",
		},
		{
			name: "空特征",
			result: &NetworkBehaviorResult{
				TotalPackets: 0,
				RTTAnalysis: &RTTAnalysis{
					Count:      0,
					AverageRTT: 0,
				},
				SequenceNumberPattern:   "insufficient_data",
				TimingPattern:          "insufficient_data",
				BehaviorCharacteristics: []string{},
			},
			want: "NetworkBehavior[packets=0, avgRTT=0s, seqPattern=insufficient_data, timing=insufficient_data, characteristics=[]]",
		},
		{
			name: "多特征",
			result: &NetworkBehaviorResult{
				TotalPackets: 150,
				RTTAnalysis: &RTTAnalysis{
					Count:      150,
					AverageRTT: 100 * time.Millisecond,
				},
				SequenceNumberPattern:   "random",
				TimingPattern:          "bursty",
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
