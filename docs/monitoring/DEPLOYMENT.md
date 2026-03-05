# 监控部署指南

本文档介绍如何部署 fingerprint 库的监控系统。

## 架构概览

```plaintext
┌─────────────────────────────────────────────────────────────┐
│                    Application                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Fingerprint  │  │   Behavior   │  │    HTTP/2    │      │
│  │   Generator  │  │   Analyzer   │  │   Analyzer   │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                 │                 │              │
│         └────────────────┬┴────────────────┘               │
│                          │                                  │
│              ┌───────────▼──────────┐                      │
│              │  Metrics Library     │                      │
│              │  (Prometheus client) │                      │
│              └───────────┬──────────┘                      │
└──────────────────────────┼──────────────────────────────────┘
                           │
                           ▼ scrape
┌─────────────────────────────────────────────────────────────┐
│                  Prometheus Server                          │
│  - Scrape metrics from :8080/metrics                       │
│ - Store time-series data                                   │
│ - Evaluate alerting rules                                  │
└──────────────────────────┬──────────────────────────────────┘
                           │
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
    ┌────────────┐  ┌──────────┐  ┌──────────────┐
    │  Grafana   │  │ Alertman-│  │  PagerDuty   │
    │ (Dashboard)│  │   ager   │  │  (On-call)   │
    └────────────┘  └──────────┘  └──────────────┘
```plaintext

## 快速开始

### 1. 启动指标服务器

在应用程序中启动指标服务器：

```go
package main

import (
    "log"
    "github.com/vistone/fingerprint/modules/internal/metrics"
)

func main() {
    // 启动指标服务器（端口 8080）
    metrics.InitDefaultServer(":8080")
    defer metrics.StopDefaultServer()
    
    log.Println("Metrics server started on :8080")
    
    // 你的应用代码...
}
```plaintext

验证指标端点：

```bash
curl http://localhost:8080/metrics
curl http://localhost:8080/health
```plaintext

### 2. 部署 Prometheus

使用 Docker 部署：

```bash
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v $(pwd)/docs/monitoring/prometheus.yaml:/etc/prometheus/prometheus.yaml \
  -v $(pwd)/docs/monitoring/alerting:/etc/prometheus/rules \
  prom/prometheus:latest \
  --config.file=/etc/prometheus/prometheus.yaml
```plaintext

prometheus.yaml 配置：

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'fingerprint'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /metrics
    scrape_interval: 5s

rule_files:
  - /etc/prometheus/rules/prometheus-rules.yaml

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']
```plaintext

### 3. 部署 Alertmanager

```bash
docker run -d \
  --name alertmanager \
  -p 9093:9093 \
  -v $(pwd)/docs/monitoring/alerting/alertmanager-config.yaml:/etc/alertmanager/config.yaml \
  -e SMTP_PASSWORD='your-password' \
  -e SLACK_WEBHOOK_URL='https://hooks.slack.com/...' \
  prom/alertmanager:latest \
  --config.file=/etc/alertmanager/config.yaml
```plaintext

### 4. 部署 Grafana

```bash
docker run -d \
  --name grafana \
  -p 3000:3000 \
  -v $(pwd)/docs/monitoring/grafana-dashboard.json:/var/lib/grafana/dashboards/fingerprint.json \
  -e GF_SECURITY_ADMIN_PASSWORD=admin \
  grafana/grafana:latest
```plaintext

导入仪表板：
1. 访问 http://localhost:3000
2. 登录（admin/admin）
3. Configuration → Data Sources → Add Prometheus
4. 导入 `docs/monitoring/grafana-dashboard.json`

## Kubernetes 部署

### Deployment 配置

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fingerprint-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: fingerprint
  template:
    metadata:
      labels:
        app: fingerprint
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      containers:
        - name: app
          image: your-app:latest
          ports:
            - containerPort: 8080
              name: metrics
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "512Mi"
              cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: fingerprint-metrics
spec:
  selector:
    app: fingerprint
  ports:
    - port: 8080
      targetPort: 8080
      name: metrics
```plaintext

### ServiceMonitor（用于 Prometheus Operator）

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: fingerprint-metrics
  labels:
    release: prometheus
spec:
  selector:
    matchLabels:
      app: fingerprint
  endpoints:
    - port: metrics
      interval: 5s
      path: /metrics
```plaintext

## 告警响应流程

```plaintext
告警触发
    │
    ▼
Prometheus 评估规则
    │
    ▼
Alertmanager 接收
    │
    ├─→ 分组（group_by）
    ├─→ 抑制（inhibit_rules）
    └─→ 路由（routes）
         │
         ├─→ Critical → PagerDuty + Slack #critical
         ├─→ Security → Slack #security
         └─→ Warning  → Slack #platform
    │
    ▼
通知发送
    │
    └─→ On-call 工程师响应
```plaintext

## 性能调优

### Prometheus 内存优化

```yaml
# prometheus.yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: production
    replica: '{{.ExternalURL}}'

storage:
  tsdb:
    retention.time: 30d
    retention.size: 50GB
    min-block-duration: 2h
    max-block-duration: 2h
```plaintext

### 高基数问题处理

如果指标标签值过多（如用户 IP）：

```go
// 避免高基数标签
metrics.FingerprintGenerationTotal.WithLabelValues(
    browser,  // ✓ 有限的枚举值
    os,       // ✓ 有限的枚举值
    userIP,   // ✗ 高基数，不要这样做
).Inc()
```plaintext

## 故障排查

### 问题 1: 指标未采集

```bash
# 检查端点
curl -v http://localhost:8080/metrics

# 检查 Prometheus targets
open http://localhost:9090/targets

# 检查日志
docker logs prometheus
```plaintext

### 问题 2: 告警未触发

```bash
# 手动测试告警表达式
curl -G 'http://localhost:9090/api/v1/query' \
  --data-urlencode 'query=fingerprint_generation_duration_ms > 100'

# 检查 Alertmanager
curl http://localhost:9093/api/v1/alerts
```plaintext

### 问题 3: 内存占用过高

```bash
# 检查高基数指标
curl -s http://localhost:9090/api/v1/label/__name__/values | wc -l

# 查看 TSDB 状态
open http://localhost:9090/tsdb-status
```plaintext

## 安全考虑

1. **网络隔离**
   - 指标端点应限制内部访问
   - 使用防火墙规则限制 Prometheus 访问

2. **敏感数据**
   - 避免在指标中包含 PII（个人身份信息）
   - 使用聚合指标而非原始数据

3. **认证授权**
   ```go
   // 添加基本认证
   http.Handle("/metrics", basicAuth(promhttp.Handler()))
   ```

## 参考链接

- [Prometheus 官方文档](https://prometheus.io/docs/)
- [Grafana Dashboard 最佳实践](https://grafana.com/docs/grafana/latest/best-practices/)
- [Alertmanager 配置指南](https://prometheus.io/docs/alerting/latest/configuration/)
