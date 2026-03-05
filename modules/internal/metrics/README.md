# Prometheus 指标

该包提供了基于 Prometheus 的指标采集功能。

## 可用指标

### 指纹生成指标

| 指标名 | 类型 | 描述 |
| -------- | ------ | ------ |
| `fingerprint_generation_total` | Counter | 指纹生成总次数，标签: browser, os |
| `fingerprint_generation_duration_ms` | Histogram | 指纹生成耗时（毫秒），标签: browser |

### 缓存指标

| 指标名 | 类型 | 描述 |
| -------- | ------ | ------ |
| `fingerprint_profile_cache_hit_total` | Counter | Profile 缓存命中次数 |
| `fingerprint_profile_cache_miss_total` | Counter | Profile 缓存未命中次数 |

### 连接指标

| 指标名 | 类型 | 描述 |
| -------- | ------ | ------ |
| `fingerprint_active_connections` | Gauge | 当前活跃连接数 |

### 行为分析指标

| 指标名 | 类型 | 描述 |
| -------- | ------ | ------ |
| `fingerprint_behavior_signals_total` | Counter | 行为信号检测次数，标签: risk_level |

### HTTP/2 分析指标

| 指标名 | 类型 | 描述 |
| -------- | ------ | ------ |
| `fingerprint_http2_analysis_duration_ms` | Histogram | HTTP/2 签名分析耗时 |

## 使用方法

### 在代码中记录指标

```go
import "github.com/vistone/fingerprint/internal/metrics"

func SomeFunction() {
    // 记录指纹生成
    metrics.RecordFingerprintGeneration("chrome", "windows", 5.2)
    
    // 记录缓存访问
    metrics.RecordProfileCacheAccess(true)  // 命中
    metrics.RecordProfileCacheAccess(false) // 未命中
    
    // 记录行为信号
    metrics.RecordBehaviorSignal("high")
}
```plaintext

### 暴露指标端点

在主程序中添加 HTTP 端点：

```go
import (
    "net/http"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":8080", nil)
}
```plaintext

### Prometheus 配置

```yaml
scrape_configs:
  - job_name: 'fingerprint'
    static_configs:
      - targets: ['localhost:8080']
```plaintext

## Grafana 仪表板

建议创建的仪表板：

1. **指纹生成概览**
   - 每秒生成数（按浏览器分组）
   - 平均生成耗时
   - 成功率

2. **缓存性能**
   - 缓存命中率
   - 命中/未命中趋势

3. **行为分析**
   - 各风险等级信号数
   - 信号趋势
