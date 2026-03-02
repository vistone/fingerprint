# 行为信号分析 (Behavioral Signal Analysis)

## 概述

行为信号分析模块通过监测和分析 TLS 请求行为模式来识别可疑活动。该模块实现了三个核心分析功能：

1. **时序模式分析 (Temporal Pattern Analysis)**：检测请求间隔的规律性
2. **协议比例分析 (Protocol Proportion Analysis)**：分析 TLS/HTTP 版本和密码套件的分布
3. **连接复用行为分析 (Connection Reuse Analysis)**：检测异常的连接复用率

## 核心特性

### 时序模式分析

识别请求间隔的规律性，区分真实用户和自动化脚本：

```go
analyzer := fp.NewBehaviorAnalyzer(nil)

// 添加请求数据
for i := 0; i < 20; i++ {
    req := fp.RequestBehavior{
        Timestamp:   time.Now().Add(time.Duration(i*500) * time.Millisecond),
        TLSVersion:  "1.3",
        CipherSuite: "TLS_AES_256_GCM_SHA384",
        HTTPVersion: "2",
        SNI:        "example.com",
    }
    analyzer.AddRequest(req)
}

// 生成行为信号
signals := analyzer.GenerateBehaviorSignals("example.com")

// 分析时序模式
pattern := analyzer.AnalyzeTemporalPattern("example.com")
fmt.Printf("规律性指数: %.2f\n", pattern.RegularityIndex)
fmt.Printf("平均间隔: %.0f ms\n", pattern.MeanInterval)
fmt.Printf("标准差: %.2f\n", pattern.StdDev)
```

**规律性指数的含义：**
- 0.0 - 0.3：高度随机，典型的真实用户行为
- 0.4 - 0.6：中等规律，可能混合了真实和自动化流量
- 0.7 - 1.0：高度规律，可能是机器人或自动化脚本

### 协议比例分析

分析 TLS 版本、Cipher Suite 和 Extensions 的多样性：

```go
// 添加多个请求，使用不同的协议配置
for i := 0; i < 10; i++ {
    req := fp.RequestBehavior{
        Timestamp:  time.Now().Add(time.Duration(i*100) * time.Millisecond),
        TLSVersion: tlsVersions[i % len(tlsVersions)],
        CipherSuite: cipherSuites[i % len(cipherSuites)],
        HTTPVersion: "2",
        Extensions: []string{"supported_groups", "key_share"},
        SNI: "example.com",
    }
    analyzer.AddRequest(req)
}

// 分析协议分布
proportion := analyzer.AnalyzeProtocolProportion("example.com")
fmt.Printf("TLS版本数: %d\n", len(proportion.TLSVersions))
fmt.Printf("CipherSuite数: %d\n", len(proportion.TopCipherSuites))
fmt.Printf("熵值: %.2f\n", proportion.EntropyScore)
fmt.Printf("异常分布: %v\n", proportion.IsAnomalous)
```

**异常检测规则：**
- 所有请求使用相同的 TLS 版本或 Cipher Suite
- 协议版本熵值过低（数据过于一致）
- 分布不符合真实浏览器行为

### 连接复用行为

检测异常的连接复用率：

```go
// 连接复用率 = 复用连接的请求数 / 总请求数
// 真实浏览器：60-80% 的连接复用率
// 机器人脚本：95%+ 的连接复用率

signals := analyzer.GenerateBehaviorSignals("example.com")

for _, sig := range signals {
    if sig.SignalType == "CONNECTION_REUSE" {
        fmt.Printf("连接复用率异常: %s (风险: %s, 分数: %.2f)\n",
            sig.Description, sig.RiskLevel, sig.Score)
    }
}
```

## API 参考

### BehaviorAnalyzer

```go
// 创建分析器
analyzer := fp.NewBehaviorAnalyzer(config)

// 添加请求
analyzer.AddRequest(req)

// 生成行为信号
signals := analyzer.GenerateBehaviorSignals(origin)

// 获取分析结果
pattern := analyzer.AnalyzeTemporalPattern(origin)
proportion := analyzer.AnalyzeProtocolProportion(origin)

// 获取风险评分
riskScore := analyzer.GetRiskScore()

// 获取所有信号
allSignals := analyzer.GetAllSignals()

// 生成报告
summary := analyzer.GetAnalysisSummary()
```

### RequestBehavior

```go
req := fp.RequestBehavior{
    Timestamp:        time.Now(),
    TLSVersion:      "1.3",
    CipherSuite:     "TLS_AES_256_GCM_SHA384",
    Extensions:      []string{"supported_groups", "key_share"},
    HTTPVersion:     "2",
    RequestSize:     1024,
    ResponseSize:    2048,
    Latency:         100 * time.Millisecond,
    ReusingConnection: true,
    SourceIP:        "192.168.1.1",
    DestinationIP:   "10.0.0.1",
    DestinationPort: 443,
    SNI:             "example.com",
}
```

### BehaviorSignal

```go
type BehaviorSignal struct {
    SignalType   string  // "TEMPORAL_PATTERN", "PROTOCOL_PROPORTION", "CONNECTION_REUSE"
    Score        float64 // 0-1 范围
    Description  string  // 易读的信号描述
    Timestamp    time.Time
    RiskLevel    string  // "safe", "low", "medium", "high", "critical"
}
```

## 配置选项

```go
config := &fp.BehaviorAnalysisConfig{
    TimeWindowSize:                 60,  // 时间窗口大小（秒）
    MinRequestsForAnalysis:         5,   // 最小请求数
    RegularityThreshold:            0.3, // 规律性阈值
    EntropyThreshold:               0.5, // 熵值阈值
    AnomalousIntervalRateThreshold: 0.2, // 异常间隔率阈值
}

analyzer := fp.NewBehaviorAnalyzer(config)
```

## 使用案例

### 检测机器人流量

```go
// 机器人通常表现出：
// 1. 高规律性的请求间隔
// 2. 单一的 TLS/Cipher 配置
// 3. 极高的连接复用率

analyzer := fp.NewBehaviorAnalyzer(nil)

// ... 添加请求 ...

signals := analyzer.GenerateBehaviorSignals(origin)
riskScore := analyzer.GetRiskScore()

if riskScore > 0.7 {
    fmt.Println("检测到高度可疑的行为 - 可能是机器人")
}
```

### 区分浏览器类型

```go
// Chrome 浏览器通常特征：
// - 规律性较低
// - 多样的 Cipher Suite 使用
// - 合理的连接复用率

// Firefox 浏览器特征：
// - 不同的 Extension 签名
// - 不同的 TLS 版本偏好
// - 不同的连接行为

// 通过行为信号分析可以补充现有的指纹识别方法
```

## 性能优化

模块已针对高性能进行优化：

- **零分配** 的关键操作路径
- **增量计算** 支持，不需要重新处理所有数据
- **并发安全** 的操作（每个分析器实例）
- **基准测试** 显示：
  - 添加请求：< 1μs
  - 时序分析：< 10μs
  - 协议分析：< 50μs
  - 信号生成：< 100μs

## 与其他模块的集成

行为信号分析可以与其他指纹识别模块协同工作：

```go
// 结合 JA3 指纹和行为信号
ja3Fp := fp.GenerateJA3(...)
behaviorSignals := analyzer.GenerateBehaviorSignals(origin)

// 结合综合风险评分
riskInput := fp.RiskInput{
    JA3Hash: ja3Fp,
    Context: fp.RiskContext{
        BehaviorSignals: behaviorSignals,
        // ... 其他上下文信息
    },
}
riskResult, _ := fp.CalculateRisk(riskInput)
```

## 限制和注意事项

1. **数据量要求**：至少需要 5-10 个请求才能进行有效的行为分析
2. **时间依赖**：需要准确的时间戳，时间同步至关重要
3. **协议特定**：不同的协议版本（HTTP/1.1 vs HTTP/2）有不同的行为特性
4. **VPN 影响**：VPN 和代理可能改变连接行为模式
5. **网络条件**：高延迟或丢包可能导致异常的间隔分布

## 相关资源

- JA3/JA4 指纹识别
- 综合风险评分系统
- HTTP/2 签名分析
- 权限策略和 Client Hints 管理
