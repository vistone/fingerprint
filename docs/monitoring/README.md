# 监控与可观测性

本文档介绍 fingerprint 库的监控和可观测性功能。

## Prometheus 指标

### 启用指标收集

```go
package main

import (
    "github.com/vistone/fingerprint/internal/metrics"
)

func main() {
    // 启动指标服务器（默认端口 8080）
    metrics.InitDefaultServer(":8080")
    defer metrics.StopDefaultServer()
    
    // 你的应用代码...
}
```plaintext

### 可用指标

详见 [internal/metrics/README.md](../../internal/metrics/README.md)

### Prometheus 配置

```yaml
scrape_configs:
  - job_name: 'fingerprint'
    static_configs:
      - targets: ['localhost:8080']
    scrape_interval: 15s
```plaintext

## Grafana 仪表板

### 导入仪表板

1. 打开 Grafana → Create → Import
2. 上传 [grafana-dashboard.json](grafana-dashboard.json)
3. 选择 Prometheus 数据源

### 仪表板面板

- **指纹生成速率**: 按浏览器和操作系统分组
- **生成耗时**: p99/p95 延迟
- **缓存命中率**: Profile 缓存性能
- **活跃连接数**: 实时连接监控
- **行为信号**: 按风险等级分布
- **HTTP/2 分析耗时**: 热力图展示

## 日志

建议使用结构化日志：

```go
import "log/slog"

slog.Info("fingerprint generated",
    "browser", result.Profile.GetClientHelloStr(),
    "duration_ms", duration,
)
```plaintext

## 健康检查

指标服务器提供健康检查端点：

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```plaintext

## 性能分析

### 基准测试

```bash
# 运行所有基准测试
go test -bench=. ./...

# 运行特定包的基准测试
go test -bench=BenchmarkGetRandomFingerprint ./generator/random/...
```plaintext

### 性能优化建议

1. **缓存**: Profile 缓存命中率应 > 90%
2. **延迟**: 指纹生成 p99 应 < 10ms
3. **内存**: 使用 `ReportAllocs` 监控分配
