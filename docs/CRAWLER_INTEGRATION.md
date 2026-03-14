# 爬虫与反爬系统集成指南

## 架构概览

```
┌─────────────────────────────────────────────────────────────────────┐
│                        数据回流与训练闭环                            │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────────────┐  │
│  │ Active Crawler│───▶│ Data Feedback│───▶│ ML Model Training    │  │
│  │  (测试/训练)  │    │ (数据回流)   │    │ (模型增强)           │  │
│  └──────────────┘    └──────────────┘    └──────────────────────┘  │
│         │                                            │               │
│         │                                            ▼               │
│         │                                    ┌──────────────┐       │
│         │                                    │ Model Store  │       │
│         │                                    └──────────────┘       │
│         │                                            │               │
│         ▼                                            ▼               │
│  ┌──────────────────────────────────────────────────────────┐       │
│  │                    Production WAF                         │       │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐        │       │
│  │  │ L1      │ │ L2      │ │ L3      │ │ L4      │        │       │
│  │  │ Network │ │ TLS     │ │ HTTP    │ │ Behavior│        │       │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘        │       │
│  │                           │                             │       │
│  │                    ┌──────▼──────┐                       │       │
│  │                    │ Agent       │                       │       │
│  │                    │ (自治决策)  │                       │       │
│  │                    └──────┬──────┘                       │       │
│  │                           │                             │       │
│  │              ┌────────────┼────────────┐                │       │
│  │              ▼            ▼            ▼                │       │
│  │         [Allow]    [Challenge]    [Block]              │       │
│  └──────────────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────────────┘
```

## 核心组件

### 1. 主动爬虫 (modules/crawler)

**用途**: 
- 测试自身反爬系统检测能力
- 收集真实流量样本训练模型
- 验证指纹伪装效果

**核心特性**:
- 200+ 真实浏览器指纹轮换
- 智能代理池管理
- 人类行为模拟
- 自动拦截检测
- 数据回流支持

### 2. WAF 防护层 (modules/waf)

**5 层防护架构**:

| 层级 | 名称 | 检测内容 | 响应方式 |
|------|------|----------|----------|
| L1 | Network Layer | TCP/IP 指纹、IP 信誉 | Block/Monitor |
| L2 | TLS Layer | JA3/JA4 指纹、TLS 异常 | Block/Challenge |
| L3 | HTTP Layer | Header 指纹、请求模式 | Challenge/Throttle |
| L4 | Behavior Layer | 行为序列、频率分析 | Throttle/Monitor |
| L5 | Device Layer | 设备指纹、长期追踪 | Monitor/Block |

### 3. 自治代理 (modules/agent)

**OADA 循环**:
- **Observe**: 持续收集请求特征
- **Analyze**: 行为分析和异常检测
- **Decide**: 基于策略和学习的决策
- **Act**: 执行防护动作

### 4. 机器学习 (modules/ml)

**4 阶段管道**:
1. Feature Extraction - 特征提取
2. Hierarchical Classification - 分层分类
3. Threat Detection - 威胁检测
4. Risk Assessment - 风险评估

---

## 快速开始

### 场景 1: 启动测试爬虫

```go
package main

import (
    "github.com/vistone/fingerprint/modules/crawler"
    "github.com/vistone/fingerprint/modules/profiles"
)

func main() {
    // 配置爬虫
    config := &crawler.CrawlerConfig{
        Name: "test_crawler",
        TargetURLs: []string{
            "https://target-site.com/products",
            "https://target-site.com/api/data",
        },
        Workers:         3,
        RateLimit:       2 * time.Second,
        ProfileStrategy: crawler.ProfileStrategyRotate,
        ProfilePool: []string{
            "chrome_133_windows",
            "firefox_135_macos",
            "safari_17_ios",
        },
        ProxyList: []string{
            "http://proxy1:8080",
            "http://proxy2:8080",
        },
        HumanLike:    true,
        StealthMode:  true,
        CollectMode:  crawler.CollectModeBlocked,
        FeedbackURL:  "http://localhost:8080/api/feedback",
    }

    // 启动爬虫
    c := crawler.NewCrawler(config)
    if err := c.Start(); err != nil {
        panic(err)
    }

    // 处理结果
    for result := range c.GetResults() {
        if result.Blocked {
            fmt.Printf("🚫 Blocked: %s - %s\n", result.URL, result.BlockReason)
            fmt.Printf("   Fingerprint: %s\n", result.Fingerprint.ID)
            fmt.Printf("   Detection: %v\n", result.DetectionInfo)
        } else {
            fmt.Printf("✅ Success: %s - %d bytes\n", result.URL, result.ContentLength)
        }
    }

    // 统计
    stats := c.GetStats()
    fmt.Printf("\n总成功率: %.2f%%\n", stats.SuccessRate()*100)
    fmt.Printf("拦截率: %.2f%%\n", stats.BlockRate()*100)
}
```

### 场景 2: 集成 WAF 到网关

```go
package main

import (
    "github.com/vistone/fingerprint/modules/gateway"
    "github.com/vistone/fingerprint/modules/waf"
)

func main() {
    // 创建 WAF
    wafConfig := &waf.WAFConfig{
        Enabled: true,
        Mode:    waf.WAFModeProtection, // learning/detection/protection/aggressive
        
        // 各层防护开关
        NetworkLayerEnabled:  true,
        TLSLayerEnabled:      true,
        HTTPLayerEnabled:     true,
        BehaviorLayerEnabled: true,
        DeviceLayerEnabled:   true,
        
        // 阈值配置
        RiskThreshold:  0.7,
        RateLimitRPS:   100,
        BlockDuration:  1 * time.Hour,
        
        // 黑名单
        BlacklistJA3: []string{
            "769,47-53-5-10-61-...,0-23-65281-10-11,...",
            // 已知的爬虫工具 JA3
        },
        BlacklistIPs: []string{
            "192.168.1.100",
        },
        
        // ML 配置
        MLEnabled:        true,
        MLClassifierPath: "./models",
        
        // 代理配置
        AgentEnabled: true,
    }
    
    w := waf.NewWAF(wafConfig)
    defer w.Stop()

    // 创建 Gateway
    gwConfig := &gateway.GatewayConfig{
        Port:         8080,
        CacheEnabled: true,
        AgentEnabled: true,
    }
    gw := gateway.NewGateway(gwConfig)

    // 将 WAF 作为中间件
    http.Handle("/", w.Middleware(gw))
    
    fmt.Println("Server started on :8080")
    http.ListenAndServe(":8080", nil)
}
```

### 场景 3: 数据回流与模型训练

```go
package main

import (
    "github.com/vistone/fingerprint/modules/ml"
    "github.com/vistone/fingerprint/modules/crawler"
)

func main() {
    // 1. 启动带数据回流的爬虫
    crawlerConfig := &crawler.CrawlerConfig{
        TargetURLs:  []string{"https://test-target.com"},
        CollectMode: crawler.CollectModeFull,
        FeedbackURL: "http://localhost:9090/api/feedback",
    }
    
    c := crawler.NewCrawler(crawlerConfig)
    go c.Start()

    // 2. 启动 ML 服务接收数据
    mlConfig := &ml.ServiceConfig{
        ModelStorePath:   "./models",
        AutoLoadLatest:   true,
        EvolveInterval:   24 * time.Hour,
        DriftThreshold:   0.05,
    }
    
    svc, _ := ml.NewMLService(mlConfig)

    // 3. 处理回流数据并反馈给模型
    http.HandleFunc("/api/feedback", func(w http.ResponseWriter, r *http.Request) {
        var entries []crawler.FeedbackEntry
        json.NewDecoder(r.Body).Decode(&entries)
        
        for _, entry := range entries {
            // 转换为训练样本
            features := extractFeatures(entry)
            
            // 标记: 被拦截的请求是正样本 (需要学习检测)
            label := 1.0
            if !entry.Blocked {
                label = 0.0
            }
            
            // 反馈给模型
            svc.Feedback(features, label)
        }
        
        w.WriteHeader(http.StatusOK)
    })

    http.ListenAndServe(":9090", nil)
}
```

---

## 实战配置

### 配置 A: 生产环境反爬 (高防护)

```yaml
# waf-production.yaml
waf:
  mode: protection
  
  # 全层防护
  layers:
    network: true
    tls: true
    http: true
    behavior: true
    device: true
  
  thresholds:
    risk: 0.6           # 较低阈值，更严格
    rate_limit_rps: 50  # 每秒 50 请求
    rate_limit_burst: 80
    block_duration: 24h
  
  # 黑名单
  blacklist:
    ja3:
      - "769,47-53-5-10-49161-49162-49171-49172-50-56-19-4,0-10-11,23-24-25,0"
      - "771,49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21,29-23-24-25-256-257,0"
    ips:
      - "10.0.0.100"
      - "10.0.0.101"
    paths:
      - "/api/internal"
  
  ml:
    enabled: true
    model_path: /data/models/production
    
  agent:
    enabled: true
    session_window: 30m
    min_observations_to_learn: 10
```

### 配置 B: 测试环境爬虫 (学习模式)

```yaml
# crawler-learning.yaml
crawler:
  name: learning_crawler
  
  targets:
    - https://staging-api.example.com/products
    - https://staging-api.example.com/search
  
  workers: 5
  rate_limit: 3s
  rate_jitter: 0.4
  
  # 指纹策略
  profile_strategy: adaptive
  profile_pool:
    - chrome_133_windows
    - firefox_135_macos
    - safari_17_macos
    - edge_120_windows
  
  rotate_interval: 60s
  
  # 代理
  proxy_strategy: rotate
  proxy_list:
    - http://proxy1:8080
    - http://proxy2:8080
    - http://proxy3:8080
  
  # 行为模拟
  human_like: true
  scroll_delay: 800ms
  
  # 数据收集
  collect_mode: full
  feedback_url: http://ml-service:9090/api/feedback
```

### 配置 C: 对抗测试 (红队)

```yaml
# crawler-redteam.yaml
crawler:
  name: redteam_crawler
  
  # 激进测试
  targets:
    - https://target.com/api/v1/data
    - https://target.com/api/v2/search
  
  workers: 20              # 高并发
  rate_limit: 500ms        # 快速请求
  burst_limit: 10
  
  # 指纹快速轮换
  profile_strategy: random
  rotate_interval: 10s
  
  # 代理池
  proxy_strategy: random
  proxy_list:
    - http://residential-proxy-1
    - http://residential-proxy-2
    # ... 更多代理
  
  # 高级隐身
  stealth_mode: true
  block_avoidance: true
  
  # 只收集被拦截的请求 (负样本)
  collect_mode: blocked
  feedback_url: http://localhost:9090/api/blocked
```

---

## 数据流与反馈闭环

```
┌─────────────────┐
│  Active Crawler │
│  (主动测试)     │
└────────┬────────┘
         │ 爬取结果 (成功/被拦截)
         ▼
┌─────────────────┐
│  Data Feedback  │
│  (数据回流)     │
│                 │
│  ┌───────────┐  │
│  │ 正样本    │  │ ← 被拦截的请求
│  │ (Blocked) │  │   (需要检测)
│  ├───────────┤  │
│  │ 负样本    │  │ ← 成功的请求
│  │ (Clean)   │  │   (正常流量)
│  └───────────┘  │
└────────┬────────┘
         │ 训练数据
         ▼
┌─────────────────┐
│  ML Pipeline    │
│  (模型训练)     │
│                 │
│  1. Feature Ext │
│  2. Classify    │
│  3. Detect      │
│  4. Evaluate    │
└────────┬────────┘
         │ 新模型
         ▼
┌─────────────────┐
│  Model Store    │
│  (模型存储)     │
└────────┬────────┘
         │ 模型更新
         ▼
┌─────────────────┐
│  Production WAF │
│  (生产防护)     │
└─────────────────┘
```

---

## API 参考

### 爬虫 API

```go
// 创建爬虫
crawler := crawler.NewCrawler(config)

// 启动/停止
crawler.Start()
crawler.Stop()

// 获取统计
stats := crawler.GetStats()

// 获取结果通道
for result := range crawler.GetResults() {
    // result.URL, result.Blocked, result.BlockReason
}
```

### WAF API

```go
// 创建 WAF
w := waf.NewWAF(config)

// 分析请求
result := w.Analyze(ctx, request)

// 中间件模式
http.Handle("/", w.Middleware(handler))

// 获取统计
stats := w.Stats()
```

### ML 服务 API

```go
// 创建服务
svc, _ := ml.NewMLService(config)

// 推理
result := svc.Infer(profile, nil)

// 验证
ok := svc.Validate(profile)

// 反馈学习
svc.Feedback(features, reward)
```

---

## 监控指标

### 爬虫指标

```
crawler_requests_total
  - label: status (success/blocked/error)
  - label: fingerprint

crawler_block_rate
  - gauge: 0.0 - 1.0

crawler_success_rate
  - gauge: 0.0 - 1.0
```

### WAF 指标

```
waf_requests_total
  - label: action (allow/challenge/block/throttle)
  - label: layer

detection_layer_hits_total
  - label: layer (network/tls/http/behavior/device)
  - label: result (hit/miss)

blocklist_size
  - gauge
```

---

## 最佳实践

1. **渐进式部署**
   - 从 Learning 模式开始收集基线
   - 逐步切换到 Detection 模式
   - 最后启用 Protection 模式

2. **指纹管理**
   - 定期更新 Profile 池
   - 监控各指纹成功率
   - 移除被识别的指纹

3. **数据质量**
   - 确保正负样本平衡
   - 定期清洗训练数据
   - 标注边界案例

4. **性能优化**
   - 使用缓存避免重复分析
   - 异步处理数据回流
   - 合理设置阈值平衡误杀

5. **安全审计**
   - 记录所有拦截决策
   - 定期审查代理策略
   - 监控异常访问模式