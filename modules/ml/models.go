// Package ml — models.go 提供指纹分析领域的专用神经网络模型库。
//
// ┌─────────────────────────────────────────────────────────────────────┐
// │              指纹智能分析模型库 — 领域驱动设计                        │
// ├─────────────────────────────────────────────────────────────────────┤
// │                                                                     │
// │  本项目的核心使命：                                                   │
// │    通过多层指纹分析（TLS + HTTP/2 + TCP/IP + JS + 行为）              │
// │    识别真实浏览器 vs 自动化工具 / 伪造客户端                          │
// │                                                                     │
// │  四个专用模型，各自定义清晰：                                         │
// │                                                                     │
// │  ┌──────────────────┐                                               │
// │  │ FingerprintEncoder│ 学什么：多层指纹的内在结构与浏览器独特性       │
// │  │  (指纹编码器)     │ 推理：将原始特征映射到32维稠密嵌入空间         │
// │  │                  │ 生成：嵌入向量——同浏览器聚近，异浏览器分远      │
// │  └────────┬─────────┘                                               │
// │           │ 32维嵌入                                                 │
// │           ▼                                                          │
// │  ┌──────────────────┐                                               │
// │  │ BrowserClassifier│ 学什么：嵌入空间到浏览器身份的映射关系          │
// │  │  (浏览器分类器)   │ 推理：从嵌入向量识别浏览器家族和版本           │
// │  │                  │ 生成：家族概率分布 + 版本概率分布 + 置信度      │
// │  └────────┬─────────┘                                               │
// │           │ 分类结果                                                 │
// │  ┌──────────────────┐                                               │
// │  │ ForgeryDetector  │ 学什么：真实指纹 vs 伪造指纹的区分特征         │
// │  │  (伪造检测器)     │ 推理：跨层一致性分析，检测指纹伪造/欺骗       │
// │  │                  │ 生成：伪造概率 + 伪造类型(正常/无头/反检测/代理)│
// │  └────────┬─────────┘                                               │
// │           │ 检测结果                                                 │
// │  ┌──────────────────┐                                               │
// │  │ ThreatAssessor   │ 学什么：综合所有信号后的最优安全响应            │
// │  │  (威胁评估器)     │ 推理：整合嵌入+分类+伪造+行为 → 威胁判定      │
// │  │                  │ 生成：威胁类别概率 + 推荐动作 + 置信度         │
// │  └──────────────────┘                                               │
// │                                                                     │
// │  训练策略：                                                          │
// │    FingerprintEncoder: 三元组损失(Triplet) — 从207个浏览器配置学习   │
// │    BrowserClassifier:  交叉熵损失 — 从配置标签学习                    │
// │    ForgeryDetector:    二元交叉熵 — 从真实+合成伪造数据学习           │
// │    ThreatAssessor:     交叉熵 — 初始从规则标签，后续从反馈学习       │
// │                                                                     │
// │  GPU 加速：所有模型的前向/反向传播基于 Tensor 运算，                  │
// │          通过 SetDevice(gpu) 自动切换到 GPU 后端                     │
// └─────────────────────────────────────────────────────────────────────┘
package ml

import (
	"math"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/profiles"
)

// =========================================================================
// 特征编码常量 — 定义输入维度
// =========================================================================

const (
	// FingerprintFeatureDim 原始指纹特征维度 (30维)
	//
	// TLS 层 (8维):
	//   [0] tls_version:       TLS 版本号归一化 (1.0=TLS1.0, 1.1, 1.2→0.75, 1.3→1.0)
	//   [1] cipher_count:      密码套件数量 / 20.0
	//   [2] tls13_ratio:       TLS 1.3 密码套件占比
	//   [3] extension_count:   扩展数量 / 20.0
	//   [4] has_sni:           是否包含 SNI (0/1)
	//   [5] has_alpn:          是否包含 ALPN (0/1)
	//   [6] curve_count:       椭圆曲线数量 / 6.0
	//   [7] grease_ratio:      GREASE 值占比
	//
	// HTTP/2 层 (6维):
	//   [8]  h2_window:        初始窗口大小 / 10M
	//   [9]  h2_streams:       最大并发流 / 1000
	//   [10] h2_header_table:  头表大小 / 100K
	//   [11] h2_pseudo_order:  伪头顺序编码 / 24 (有24种排列)
	//   [12] h2_priority:      优先级帧 (0/1)
	//   [13] h2_settings_cnt:  设置项数 / 10
	//
	// TCP/IP 层 (4维):
	//   [14] tcp_ttl:          TTL 归一化 (32→0.25, 64→0.5, 128→1.0)
	//   [15] tcp_window:       TCP 窗口大小 / 128K
	//   [16] tcp_mss:          MSS / 2000
	//   [17] tcp_timestamps:   TCP 时间戳 (0/1)
	//
	// JS 前端层 (8维):
	//   [18] canvas_entropy:   Canvas 指纹熵
	//   [19] webgl_score:      WebGL 得分
	//   [20] audio_entropy:    Audio 指纹熵
	//   [21] font_count:       字体数量 / 200
	//   [22] storage_score:    存储得分
	//   [23] webrtc_active:    WebRTC 活跃 (0/1)
	//   [24] hardware_cores:   CPU 核心数 / 16
	//   [25] headless_score:   无头浏览器得分
	//
	// 元特征层 (4维):
	//   [26] ua_entropy:       User-Agent 熵
	//   [27] config_entropy:   整体配置熵
	//   [28] tool_marker:      自动化工具标记
	//   [29] behavior_pattern: 行为模式特征
	FingerprintFeatureDim = 30

	// EmbeddingDim 嵌入向量维度
	EmbeddingDim = 32

	// CrossLayerFeatureDim 跨层一致性特征维度 (10维)
	//   [0] tls_h2_window_match:   TLS↔HTTP/2 窗口匹配分
	//   [1] tls_h2_pseudo_match:   TLS↔HTTP/2 伪头顺序匹配分
	//   [2] tls_tcp_ttl_match:     TLS↔TCP/IP TTL 匹配分
	//   [3] ua_tls_version_match:  UA↔TLS 版本匹配分
	//   [4] ua_h2_settings_match:  UA↔HTTP/2 设置匹配分
	//   [5] js_headless_indicator: JS 无头浏览器指示
	//   [6] canvas_webgl_consist:  Canvas↔WebGL 一致性
	//   [7] cipher_order_anomaly:  密码套件顺序异常分
	//   [8] ext_pattern_anomaly:   扩展模式异常分
	//   [9] contradiction_count:   跨层矛盾数量归一化
	CrossLayerFeatureDim = 10

	// BehaviorFeatureDim 行为特征维度 (8维)
	//   [0] fp_switch_rate:      指纹切换频率 / 10
	//   [1] request_rate:        请求速率 / 20
	//   [2] consistency_score:   一致性得分 [0,1]
	//   [3] risk_trend:          风险趋势 [-1,1] → [0,1]
	//   [4] observations_norm:   观测数量归一化
	//   [5] unique_fp_ratio:     唯一指纹占比
	//   [6] session_duration:    会话时长归一化
	//   [7] burst_indicator:     突发请求指示
	BehaviorFeatureDim = 8

	// NumBrowserFamilies 浏览器家族数量
	NumBrowserFamilies = 7 // chrome, firefox, safari, edge, opera, brave, samsung

	// NumForgeryTypes 伪造类型数量
	NumForgeryTypes = 4 // real, headless, antidetect, proxy

	// NumThreatClasses 威胁类别数量
	NumThreatClasses = 6 // none, bot, fingerprint_spoof, session_anomaly, behavioral_anomaly, evasion

	// NumActions 动作数量
	NumActions = 5 // allow, monitor, challenge, throttle, block
)

// =========================================================================
// 浏览器家族编码
// =========================================================================

var familyIndex = map[core.BrowserType]int{
	core.BrowserChrome:  0,
	core.BrowserFirefox: 1,
	core.BrowserSafari:  2,
	core.BrowserEdge:    3,
	core.BrowserOpera:   4,
	core.BrowserBrave:   5,
	core.BrowserSamsung: 6,
}

var indexFamily = [NumBrowserFamilies]core.BrowserType{
	core.BrowserChrome,
	core.BrowserFirefox,
	core.BrowserSafari,
	core.BrowserEdge,
	core.BrowserOpera,
	core.BrowserBrave,
	core.BrowserSamsung,
}

// =========================================================================
// Model 1: FingerprintEncoder — 指纹编码器
// =========================================================================
//
// 学什么：从原始多层指纹特征中学习浏览器的内在独特性表示。
//         同一浏览器（不同会话、不同网络环境）的指纹应该映射到
//         嵌入空间中的相近位置；不同浏览器应该映射到不同位置。
//
// 推理：  输入 30 维原始指纹向量 → 输出 32 维 L2 归一化嵌入向量
//
// 生成：  稠密嵌入向量，满足:
//         - 同家族浏览器: 余弦相似度 > 0.8
//         - 不同家族浏览器: 余弦相似度 < 0.3
//         - 伪造指纹: 不在任何已知浏览器聚类中心附近
//
// 训练：  三元组损失 (Triplet Margin Loss)
//         anchor=已知浏览器, positive=同浏览器变体, negative=其他浏览器
//         数据来源: 207 个真实浏览器配置文件 + 数据增强

// FingerprintEncoder 指纹编码器模型
type FingerprintEncoder struct {
	Net *Sequential // 内部神经网络
}

// NewFingerprintEncoder 创建指纹编码器。
// 架构: Input(30) → Dense(128,ReLU) → Dropout(0.1) → Dense(64,ReLU) → Dense(32)
func NewFingerprintEncoder() *FingerprintEncoder {
	return &FingerprintEncoder{
		Net: NewSequential(
			NewDenseLayer(FingerprintFeatureDim, 128),
			NewReLULayer(),
			NewDropoutLayer(0.1),
			NewDenseLayer(128, 64),
			NewReLULayer(),
			NewDenseLayer(64, EmbeddingDim),
		),
	}
}

// Encode 将原始指纹特征编码为嵌入向量。
// features: [batch × 30] → embedding: [batch × 32] (L2 归一化)
func (enc *FingerprintEncoder) Encode(features *Tensor) *Tensor {
	raw := enc.Net.Forward(features)
	// 逐行 L2 归一化
	rows := raw.Shape[0]
	cols := raw.Shape[1]
	for i := 0; i < rows; i++ {
		norm := 0.0
		for j := 0; j < cols; j++ {
			v := raw.At(i, j)
			norm += v * v
		}
		norm = math.Sqrt(norm)
		if norm < 1e-12 {
			norm = 1e-12
		}
		for j := 0; j < cols; j++ {
			raw.Set(i, j, raw.At(i, j)/norm)
		}
	}
	return raw
}

// EncodeSingle 编码单个指纹向量 (便捷方法)。
func (enc *FingerprintEncoder) EncodeSingle(features []float64) []float64 {
	input := FromSlice(features)
	return enc.Encode(input).ToSlice()
}

// =========================================================================
// Model 2: BrowserClassifier — 浏览器分类器
// =========================================================================
//
// 学什么：从嵌入向量到浏览器身份的映射关系。
//         不同浏览器家族有不同的 TLS 配置、HTTP/2 设置、TCP 行为——
//         这些差异已由编码器压缩为嵌入向量，分类器学习如何解读这些嵌入。
//
// 推理：  输入 32 维嵌入向量 → 输出浏览器家族概率分布(7类)
//
// 生成：  BrowserPrediction 包含:
//         - 家族概率分布 (Chrome 95%, Firefox 3%, Safari 2%, ...)
//         - 预测家族 (chrome)
//         - 置信度 (0.95)
//
// 训练：  交叉熵损失 + 从 207 个已知配置文件学习

// BrowserClassifier 浏览器家族分类器
type BrowserClassifier struct {
	Net *Sequential // 内部神经网络
}

// NewBrowserClassifier 创建浏览器分类器。
// 架构: Embedding(32) → Dense(64,ReLU) → Dropout(0.1) → Dense(7,Softmax)
func NewBrowserClassifier() *BrowserClassifier {
	return &BrowserClassifier{
		Net: NewSequential(
			NewDenseLayer(EmbeddingDim, 64),
			NewReLULayer(),
			NewDropoutLayer(0.1),
			NewDenseLayer(64, NumBrowserFamilies),
			NewSoftmaxLayer(),
		),
	}
}

// BrowserPrediction 浏览器分类预测结果
type BrowserPrediction struct {
	Family      core.BrowserType // 预测的浏览器家族
	Confidence  float64          // 置信度 [0,1]
	Probs       []float64        // 各家族概率分布
	FamilyNames []core.BrowserType // 家族名称，和 Probs 对应
}

// Classify 预测浏览器家族。
// embedding: [batch × 32] → 返回每行的预测结果
func (bc *BrowserClassifier) Classify(embedding *Tensor) []BrowserPrediction {
	probs := bc.Net.Forward(embedding)
	batch := probs.Shape[0]
	results := make([]BrowserPrediction, batch)
	for i := 0; i < batch; i++ {
		row := probs.Row(i)
		maxIdx := 0
		maxVal := row[0]
		probsCopy := make([]float64, NumBrowserFamilies)
		copy(probsCopy, row)
		for j := 1; j < NumBrowserFamilies; j++ {
			if row[j] > maxVal {
				maxVal = row[j]
				maxIdx = j
			}
		}
		names := make([]core.BrowserType, NumBrowserFamilies)
		copy(names, indexFamily[:])
		results[i] = BrowserPrediction{
			Family:      indexFamily[maxIdx],
			Confidence:  maxVal,
			Probs:       probsCopy,
			FamilyNames: names,
		}
	}
	return results
}

// ClassifySingle 分类单个嵌入向量 (便捷方法)。
func (bc *BrowserClassifier) ClassifySingle(embedding []float64) BrowserPrediction {
	input := FromSlice(embedding)
	return bc.Classify(input)[0]
}

// =========================================================================
// Model 3: ForgeryDetector — 伪造检测器
// =========================================================================
//
// 学什么：真实浏览器指纹 vs 伪造指纹的区分特征。
//         反检测工具（tls-client, curl-impersonate, Puppeteer）虽然能
//         模仿单层特征（如 TLS ClientHello），但跨层一致性难以完美伪造：
//         - Chrome TLS + Firefox HTTP/2 伪头顺序 → 矛盾
//         - Windows UA + Linux TTL(64) → 矛盾
//         - 缺少 Canvas/WebGL → 无头浏览器
//         伪造检测器从数据中学习这些复杂的跨层关联模式。
//
// 推理：  输入 30 维指纹 + 10 维跨层特征 = 40 维
//         → 输出伪造概率(0~1) + 伪造类型(4类)
//
// 生成：  ForgeryResult 包含:
//         - IsForgery: true/false
//         - ForgeryProb: [0, 1] 伪造概率
//         - ForgeryType: Real / Headless / AntiDetect / Proxy
//         - TypeProbs: 各类型概率
//
// 训练：  二元交叉熵(是否伪造) + 交叉熵(伪造类型)
//         训练数据: 真实配置 + 合成伪造样本(层交叉/噪声注入/缺失特征)

// ForgeryType 伪造类型
type ForgeryType int

const (
	ForgeryReal       ForgeryType = iota // 真实浏览器
	ForgeryHeadless                      // 无头浏览器 (Puppeteer/Selenium/PhantomJS)
	ForgeryAntiDetect                    // 反检测工具 (tls-client/curl-impersonate/GoLogin)
	ForgeryProxy                         // 代理/中间人 (误配特征)
)

var forgeryTypeNames = [NumForgeryTypes]string{
	"real", "headless", "antidetect", "proxy",
}

// ForgeryTypeName 返回伪造类型的字符串名称。
func (ft ForgeryType) String() string {
	if int(ft) < len(forgeryTypeNames) {
		return forgeryTypeNames[ft]
	}
	return "unknown"
}

// ForgeryDetector 伪造检测器模型
type ForgeryDetector struct {
	DetectorNet *Sequential // 伪造概率网络
	TypeNet     *Sequential // 伪造类型网络
}

// NewForgeryDetector 创建伪造检测器。
// 检测网络: Input(40) → Dense(64,ReLU) → Dense(32,ReLU) → Dense(1,Sigmoid)
// 分类网络: Input(40) → Dense(64,ReLU) → Dense(32,ReLU) → Dense(4,Softmax)
func NewForgeryDetector() *ForgeryDetector {
	inputDim := FingerprintFeatureDim + CrossLayerFeatureDim // 30 + 10 = 40
	return &ForgeryDetector{
		DetectorNet: NewSequential(
			NewDenseLayer(inputDim, 64),
			NewReLULayer(),
			NewDenseLayer(64, 32),
			NewReLULayer(),
			NewDenseLayer(32, 1),
			NewSigmoidLayer(),
		),
		TypeNet: NewSequential(
			NewDenseLayer(inputDim, 64),
			NewReLULayer(),
			NewDenseLayer(64, 32),
			NewReLULayer(),
			NewDenseLayer(32, NumForgeryTypes),
			NewSoftmaxLayer(),
		),
	}
}

// ForgeryResult 伪造检测结果
type ForgeryResult struct {
	IsForgery   bool        // 是否伪造（ForgeryProb > 0.5）
	ForgeryProb float64     // 伪造概率 [0,1]
	ForgeryType ForgeryType // 预测的伪造类型
	TypeProbs   []float64   // 各类型概率分布
	TypeNames   []string    // 类型名称，和 TypeProbs 对应
}

// Detect 检测指纹伪造。
// input: [batch × 40] (30维指纹 + 10维跨层特征)
func (fd *ForgeryDetector) Detect(input *Tensor) []ForgeryResult {
	probOut := fd.DetectorNet.Forward(input)
	typeOut := fd.TypeNet.Forward(input)
	batch := input.Shape[0]
	results := make([]ForgeryResult, batch)
	for i := 0; i < batch; i++ {
		prob := probOut.Data[i]
		typeRow := typeOut.Row(i)
		maxIdx := 0
		maxVal := typeRow[0]
		typeProbs := make([]float64, NumForgeryTypes)
		copy(typeProbs, typeRow)
		for j := 1; j < NumForgeryTypes; j++ {
			if typeRow[j] > maxVal {
				maxVal = typeRow[j]
				maxIdx = j
			}
		}
		names := make([]string, NumForgeryTypes)
		copy(names, forgeryTypeNames[:])
		results[i] = ForgeryResult{
			IsForgery:   prob > 0.5,
			ForgeryProb: prob,
			ForgeryType: ForgeryType(maxIdx),
			TypeProbs:   typeProbs,
			TypeNames:   names,
		}
	}
	return results
}

// DetectSingle 检测单个样本 (便捷方法)。
func (fd *ForgeryDetector) DetectSingle(fpFeatures, crossLayerFeatures []float64) ForgeryResult {
	combined := make([]float64, FingerprintFeatureDim+CrossLayerFeatureDim)
	copy(combined, fpFeatures)
	copy(combined[FingerprintFeatureDim:], crossLayerFeatures)
	input := FromSlice(combined)
	return fd.Detect(input)[0]
}

// =========================================================================
// Model 4: ThreatAssessor — 威胁评估器
// =========================================================================
//
// 学什么：综合所有模型输出和行为特征后的最优安全响应。
//         这是整个智能系统的「决策大脑」，整合：
//         - 指纹嵌入(32维) → 客户端身份表示
//         - 伪造检测(5维) → 真实性评估
//         - 行为特征(8维) → 时序行为模式
//         学习在不同威胁场景下选择最优安全动作。
//
// 推理：  输入 45 维综合向量 → 输出威胁类别分布(6类) + 推荐动作分布(5类)
//
// 生成：  ThreatPrediction 包含:
//         - ThreatClass: None/Bot/FingerprintSpoof/SessionAnomaly/BehavioralAnomaly/Evasion
//         - ThreatProb: 威胁存在的概率
//         - Action: Allow/Monitor/Challenge/Throttle/Block
//         - ActionConfidence: 动作置信度
//         - ClassProbs: 各威胁类别概率
//         - ActionProbs: 各动作概率
//
// 训练：  交叉熵损失 (威胁类别) + 交叉熵损失 (推荐动作)
//         初始训练: 从既有规则引擎的标签学习 (策略蒸馏)
//         持续学习: 从 ReportReward 反馈更新 (在线微调)

// ThreatClass 威胁类别
type ThreatClass int

const (
	ThreatNone              ThreatClass = iota // 无威胁
	ThreatBot                                  // 机器人/爬虫
	ThreatFingerprintSpoof                     // 指纹伪造
	ThreatSessionAnomaly                       // 会话异常
	ThreatBehavioralAnomaly                    // 行为异常
	ThreatEvasion                              // 规避行为
)

var threatClassNames = [NumThreatClasses]string{
	"none", "bot", "fingerprint_spoof", "session_anomaly", "behavioral_anomaly", "evasion",
}

// String 返回威胁类别的字符串名称。
func (tc ThreatClass) String() string {
	if int(tc) < len(threatClassNames) {
		return threatClassNames[tc]
	}
	return "unknown"
}

// ActionClass 安全动作类别
type ActionClass int

const (
	ActAllow     ActionClass = iota // 放行
	ActMonitor                      // 监控
	ActChallenge                    // 挑战验证
	ActThrottle                     // 限流
	ActBlock                        // 阻断
)

var actionClassNames = [NumActions]string{
	"allow", "monitor", "challenge", "throttle", "block",
}

// String 返回动作的字符串名称。
func (a ActionClass) String() string {
	if int(a) < len(actionClassNames) {
		return actionClassNames[a]
	}
	return "unknown"
}

// ThreatAssessor 威胁评估器模型
type ThreatAssessor struct {
	ThreatNet *Sequential // 威胁分类网络
	ActionNet *Sequential // 动作推荐网络
}

// NewThreatAssessor 创建威胁评估器。
// 输入维度: 嵌入(32) + 伪造输出(5: prob+4_type_probs) + 行为(8) = 45
// 威胁网络: Input(45) → Dense(64,ReLU) → Dense(32,ReLU) → Dense(6,Softmax)
// 动作网络: Input(45) → Dense(64,ReLU) → Dense(32,ReLU) → Dense(5,Softmax)
func NewThreatAssessor() *ThreatAssessor {
	inputDim := EmbeddingDim + 1 + NumForgeryTypes + BehaviorFeatureDim // 32+1+4+8 = 45
	return &ThreatAssessor{
		ThreatNet: NewSequential(
			NewDenseLayer(inputDim, 64),
			NewReLULayer(),
			NewDenseLayer(64, 32),
			NewReLULayer(),
			NewDenseLayer(32, NumThreatClasses),
			NewSoftmaxLayer(),
		),
		ActionNet: NewSequential(
			NewDenseLayer(inputDim, 64),
			NewReLULayer(),
			NewDenseLayer(64, 32),
			NewReLULayer(),
			NewDenseLayer(32, NumActions),
			NewSoftmaxLayer(),
		),
	}
}

// ThreatPrediction 威胁评估预测结果
type ThreatPrediction struct {
	ThreatClass      ThreatClass // 预测的威胁类别
	ThreatProb       float64     // 存在威胁的概率 (1 - P(none))
	Action           ActionClass // 推荐的安全动作
	ActionConfidence float64     // 动作置信度
	ClassProbs       []float64   // 各威胁类别概率分布
	ActionProbs      []float64   // 各动作概率分布
}

// Assess 评估威胁并推荐动作。
// input: [batch × 45] (嵌入 + 伪造输出 + 行为特征)
func (ta *ThreatAssessor) Assess(input *Tensor) []ThreatPrediction {
	threatOut := ta.ThreatNet.Forward(input)
	actionOut := ta.ActionNet.Forward(input)
	batch := input.Shape[0]
	results := make([]ThreatPrediction, batch)
	for i := 0; i < batch; i++ {
		tRow := threatOut.Row(i)
		aRow := actionOut.Row(i)
		// 找到最高概率的威胁类别
		tMaxIdx, tMaxVal := 0, tRow[0]
		classProbs := make([]float64, NumThreatClasses)
		copy(classProbs, tRow)
		for j := 1; j < NumThreatClasses; j++ {
			if tRow[j] > tMaxVal {
				tMaxVal = tRow[j]
				tMaxIdx = j
			}
		}
		// 找到最高概率的动作
		aMaxIdx, aMaxVal := 0, aRow[0]
		actionProbs := make([]float64, NumActions)
		copy(actionProbs, aRow)
		for j := 1; j < NumActions; j++ {
			if aRow[j] > aMaxVal {
				aMaxVal = aRow[j]
				aMaxIdx = j
			}
		}
		results[i] = ThreatPrediction{
			ThreatClass:      ThreatClass(tMaxIdx),
			ThreatProb:       1.0 - tRow[0], // P(threat) = 1 - P(none)
			Action:           ActionClass(aMaxIdx),
			ActionConfidence: aMaxVal,
			ClassProbs:       classProbs,
			ActionProbs:      actionProbs,
		}
	}
	return results
}

// AssessSingle 评估单个样本 (便捷方法)。
func (ta *ThreatAssessor) AssessSingle(embedding []float64, forgeryResult *ForgeryResult, behavior []float64) ThreatPrediction {
	input := buildThreatInput(embedding, forgeryResult, behavior)
	return ta.Assess(FromSlice(input))[0]
}

// buildThreatInput 构建威胁评估器输入向量。
func buildThreatInput(embedding []float64, fr *ForgeryResult, behavior []float64) []float64 {
	dim := EmbeddingDim + 1 + NumForgeryTypes + BehaviorFeatureDim
	vec := make([]float64, dim)
	// 嵌入 (32维)
	copy(vec[:EmbeddingDim], embedding)
	// 伪造概率 (1维) + 伪造类型概率 (4维)
	off := EmbeddingDim
	if fr != nil {
		vec[off] = fr.ForgeryProb
		if len(fr.TypeProbs) == NumForgeryTypes {
			copy(vec[off+1:off+1+NumForgeryTypes], fr.TypeProbs)
		}
	}
	// 行为特征 (8维)
	off += 1 + NumForgeryTypes
	if len(behavior) >= BehaviorFeatureDim {
		copy(vec[off:off+BehaviorFeatureDim], behavior[:BehaviorFeatureDim])
	}
	return vec
}

// =========================================================================
// 特征编码函数 — 从原始数据构建模型输入
// =========================================================================

// EncodeFingerprint 从 ClientProfile 提取 30 维指纹特征向量。
// 这是连接原始数据和神经网络的桥梁。
func EncodeFingerprint(profile *profiles.ClientProfile) []float64 {
	vec := make([]float64, FingerprintFeatureDim)

	// TLS 层 (索引 0-7)
	vec[0] = normalizeTLSVersion(profile.TLSVersion)
	vec[1] = float64(len(profile.CipherSuites)) / 20.0
	vec[2] = tls13Ratio(profile.CipherSuites)
	vec[3] = float64(len(profile.Extensions)) / 20.0
	vec[4] = boolToFloat(hasSNI(profile.Extensions))
	vec[5] = boolToFloat(hasALPN(profile.Extensions))
	vec[6] = float64(len(profile.SupportedCurves)) / 6.0
	vec[7] = greaseRatio(profile.CipherSuites)

	// HTTP/2 层 (索引 8-13)
	h2 := profile.HTTP2Settings
	vec[8] = float64(h2.InitialWindowSize) / 10_000_000.0
	vec[9] = float64(h2.MaxConcurrentStreams) / 1000.0
	vec[10] = float64(h2.HeaderTableSize) / 100_000.0
	vec[11] = encodePseudoHeaderOrder(profile.PseudoHeaderOrder) / 24.0
	if h2.EnablePush > 0 {
		vec[12] = 1.0
	}
	vec[13] = float64(countH2Settings(h2)) / 10.0

	// TCP/IP 层 (索引 14-17)
	if tcp := profile.TCPIP; tcp != nil {
		vec[14] = float64(tcp.TTL) / 128.0
		vec[15] = float64(tcp.WindowSize) / 131072.0
		vec[16] = float64(tcp.MSS) / 2000.0
		vec[17] = boolToFloat(tcp.Timestamps)
	}

	// JS 前端层: 在 profile 中不可用 (0 值), 运行时从 Frontend SDK 填充
	// 索引 18-25 保持 0

	// 元特征层 (索引 26-29)
	if profile.Headers != nil {
		vec[26] = stringEntropy(profile.Headers.UserAgent) / 5.0
	}
	vec[27] = profileEntropy(profile) / 5.0
	// 索引 28-29 运行时填充

	// 裁剪到 [0, 1] 范围
	for i := range vec {
		vec[i] = math.Max(0, math.Min(1, vec[i]))
	}
	return vec
}

// EncodeFingerprintFromFeatureVector 从已提取的 FeatureVector 构建 30 维向量。
// 这允许与现有 FeatureExtractor 管线集成。
func EncodeFingerprintFromFeatureVector(fv *core.FeatureVector) []float64 {
	vec := make([]float64, FingerprintFeatureDim)
	if fv == nil {
		return vec
	}
	// 映射现有特征到新维度
	vec[0] = normalizeFeatureValue(fv.Get(core.FeatureTLSVersion), 0x0304)
	vec[1] = fv.Get(core.FeatureCipherSuites) / 20.0
	vec[3] = fv.Get(core.FeatureExtensions) / 20.0
	vec[8] = fv.Get(core.FeatureHTTP2Settings) / 10_000_000.0
	vec[10] = fv.Get(core.FeatureHTTPHeaders) / 100.0
	vec[18] = fv.Get(core.FeatureCanvas) / 100.0
	vec[19] = fv.Get(core.FeatureWebGL) / 100.0
	vec[20] = fv.Get(core.FeatureAudio) / 100.0
	vec[21] = fv.Get(core.FeatureFonts) / 200.0
	vec[22] = fv.Get(core.FeatureStorage) / 100.0
	vec[23] = fv.Get(core.FeatureWebRTC) // already 0/1
	vec[24] = fv.Get(core.FeatureHardware) / 16.0
	vec[25] = fv.Get(core.FeatureHeadlessBrowser)
	vec[26] = fv.Get(core.FeatureUserAgent) / 200.0
	vec[27] = fv.Get(core.FeatureEntropy) / 5.0
	vec[28] = fv.Get(core.FeatureToolMarker)
	vec[29] = fv.Get(core.FeatureBehaviorPattern)

	for i := range vec {
		vec[i] = math.Max(0, math.Min(1, vec[i]))
	}
	return vec
}

// ComputeCrossLayerFeatures 计算跨层一致性特征 (10维)。
// 这些特征捕获多层指纹之间的矛盾程度——伪造检测器的关键信号。
func ComputeCrossLayerFeatures(fp []float64) []float64 {
	cross := make([]float64, CrossLayerFeatureDim)
	if len(fp) < FingerprintFeatureDim {
		return cross
	}

	// [0] TLS↔HTTP/2 窗口匹配: Chrome 6.2M, Firefox 131K, Safari 2M
	// 窗口大小与密码套件数应该对应同一浏览器
	cross[0] = 1.0 - math.Abs(fp[1]-fp[8])*2.0 // cipher_count vs h2_window 的协调性

	// [1] TLS↔HTTP/2 伪头顺序匹配
	cross[1] = 1.0 - math.Abs(fp[2]-fp[11]) // tls13_ratio vs pseudo_order 的协调性

	// [2] TLS↔TCP/IP TTL 匹配
	// Chrome/Firefox/Safari 在不同 OS 上应有对应的 TTL
	cross[2] = 1.0 - math.Abs(fp[0]-fp[14]) // tls_version vs ttl 的协调性

	// [3] UA↔TLS 版本匹配
	cross[3] = 1.0 - math.Abs(fp[26]-fp[0]) // ua_entropy vs tls_version

	// [4] UA↔HTTP/2 设置匹配
	cross[4] = 1.0 - math.Abs(fp[26]-fp[8]) // ua_entropy vs h2_window

	// [5] JS 无头浏览器指示 (直接使用)
	cross[5] = fp[25] // headless_score

	// [6] Canvas↔WebGL 一致性 (两者应同时存在或同时缺失)
	if fp[18] > 0.1 && fp[19] > 0.1 {
		cross[6] = 1.0 // 两者都存在
	} else if fp[18] < 0.1 && fp[19] < 0.1 {
		cross[6] = 0.8 // 两者都缺失（可能是隐私模式或无头）
	} else {
		cross[6] = 0.2 // 不一致——可疑
	}

	// [7] 密码套件顺序异常分
	// TLS 1.3 浏览器应有较高的 tls13_ratio
	if fp[0] > 0.8 { // TLS 1.3
		cross[7] = fp[2] // tls13_ratio 应该高
	} else {
		cross[7] = 1.0 - fp[2] // TLS 1.2 不应有太多 1.3 套件
	}

	// [8] 扩展模式异常分
	// 扩展数与密码套件数应在合理比例
	if fp[1] > 0 {
		ratio := fp[3] / fp[1] // ext_count / cipher_count
		cross[8] = 1.0 - math.Abs(ratio-1.0) // 理想比例约 1:1
	}

	// [9] 跨层矛盾数量归一化 (从各项一致性分计算)
	contradictions := 0.0
	for i := 0; i < 9; i++ {
		if cross[i] < 0.3 {
			contradictions++
		}
	}
	cross[9] = contradictions / 9.0

	for i := range cross {
		cross[i] = math.Max(0, math.Min(1, cross[i]))
	}
	return cross
}

// =========================================================================
// 辅助函数
// =========================================================================

func normalizeTLSVersion(v uint16) float64 {
	switch v {
	case 0x0304:
		return 1.0 // TLS 1.3
	case 0x0303:
		return 0.75 // TLS 1.2
	case 0x0302:
		return 0.5 // TLS 1.1
	case 0x0301:
		return 0.25 // TLS 1.0
	default:
		return 0.0
	}
}

func normalizeFeatureValue(val, maxVal float64) float64 {
	if maxVal == 0 {
		return 0
	}
	return math.Min(1.0, val/maxVal)
}

func tls13Ratio(suites []uint16) float64 {
	if len(suites) == 0 {
		return 0
	}
	count := 0
	for _, s := range suites {
		if s == 0x1301 || s == 0x1302 || s == 0x1303 {
			count++
		}
	}
	return float64(count) / float64(len(suites))
}

func hasSNI(exts []core.TLSExtension) bool {
	for _, e := range exts {
		if e.Type == 0x0000 { // SNI extension
			return true
		}
	}
	return false
}

func hasALPN(exts []core.TLSExtension) bool {
	for _, e := range exts {
		if e.Type == 0x0010 { // ALPN extension
			return true
		}
	}
	return false
}

func greaseRatio(suites []uint16) float64 {
	if len(suites) == 0 {
		return 0
	}
	count := 0
	for _, s := range suites {
		if isGREASE(s) {
			count++
		}
	}
	return float64(count) / float64(len(suites))
}

func isGREASE(v uint16) bool {
	return (v & 0x0f0f) == 0x0a0a
}

func encodePseudoHeaderOrder(order []string) float64 {
	// 将伪头顺序编码为唯一整数
	// Chrome: :method, :authority, :scheme, :path → 0
	// Firefox: :method, :path, :authority, :scheme → 1
	// Safari: :method, :scheme, :path, :authority → 2
	if len(order) < 4 {
		return 0
	}
	hash := 0.0
	for i, h := range order {
		switch h {
		case ":method":
			hash += float64(i) * 1
		case ":authority":
			hash += float64(i) * 4
		case ":scheme":
			hash += float64(i) * 16
		case ":path":
			hash += float64(i) * 64
		}
	}
	return hash
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func countH2Settings(h2 core.HTTP2Settings) int {
	count := 0
	if h2.HeaderTableSize != 0 {
		count++
	}
	if h2.EnablePush != 0 {
		count++
	}
	if h2.MaxConcurrentStreams != 0 {
		count++
	}
	if h2.InitialWindowSize != 0 {
		count++
	}
	if h2.MaxFrameSize != 0 {
		count++
	}
	if h2.MaxHeaderListSize != 0 {
		count++
	}
	return count
}

func stringEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	freq := make(map[rune]int)
	for _, c := range s {
		freq[c]++
	}
	entropy := 0.0
	n := float64(len([]rune(s)))
	for _, count := range freq {
		p := float64(count) / n
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}

func profileEntropy(p *profiles.ClientProfile) float64 {
	// 用配置参数的多样性衡量熵
	e := 0.0
	e += float64(len(p.CipherSuites)) * 0.1
	e += float64(len(p.Extensions)) * 0.1
	e += float64(len(p.SupportedCurves)) * 0.2
	return e
}
