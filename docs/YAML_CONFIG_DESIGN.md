# YAML 配置设计

本文档定义 fingerprint 库的 YAML 配置系统规范。

## 设计目标

1. **人类可读**: YAML 格式易于阅读和编辑
2. **类型安全**: 严格的模式验证
3. **版本控制**: 支持配置版本管理
4. **热重载**: 无需重启即可更新配置
5. **可扩展**: 支持未来扩展

## 配置文件结构

```plaintext
config/
├── config.yaml              # 主配置文件
├── profiles/                # Profile 配置
│   ├── chrome/
│   │   ├── chrome_120.yaml
│   │   └── chrome_121.yaml
│   ├── firefox/
│   └── safari/
├── rules/                   # 规则配置
│   ├── detection.yaml       # 检测规则
│   └── matching.yaml        # 匹配规则
└── features/                # 功能开关
    ├── ja3.yaml
    ├── ja4.yaml
    └── http2.yaml
```plaintext

## Profile YAML 规范

### 完整示例

```yaml
# Profile 元数据
metadata:
  name: "chrome_120"                    # 唯一标识
  display_name: "Chrome 120"            # 显示名称
  version: "120.0.0"                    # 浏览器版本
  category: "desktop"                   # 类别: desktop, mobile, tablet
  
# 浏览器信息
browser:
  name: "Chrome"                        # 浏览器名称
  version: "120.0.0"                    # 版本号
  engine: "Blink"                       # 渲染引擎
  engine_version: "120.0.0"

# 操作系统信息
os:
  name: "Windows"                       # OS 名称
  version: "10"                         # OS 版本
  platform: "x86_64"                    # 平台架构

# TLS 配置
tls:
  version: "1.3"                        # 最高支持版本
  record_version: "1.0"                 # 记录层版本
  handshake_version: "1.2"              # 握手版本
  
  # Cipher Suites（按优先级排序）
  cipher_suites:
    - TLS_AES_128_GCM_SHA256
    - TLS_AES_256_GCM_SHA384
    - TLS_CHACHA20_POLY1305_SHA256
    - TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
    - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
    - TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
    - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
    - TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256
    - TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256
    - TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA
    - TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA
    - TLS_RSA_WITH_AES_128_GCM_SHA256
    - TLS_RSA_WITH_AES_256_GCM_SHA384
    - TLS_RSA_WITH_AES_128_CBC_SHA
    - TLS_RSA_WITH_AES_256_CBC_SHA
  
  # 压缩方法
  compression_methods:
    - null
  
  # 扩展列表（按发送顺序）
  extensions:
    # Server Name Indication
    - type: server_name
      id: 0
      
    # Extended Master Secret
    - type: extended_master_secret
      id: 23
      
    # Renegotiation Info
    - type: renegotiation_info
      id: 65281
      
    # Supported Groups (Elliptic Curves)
    - type: supported_groups
      id: 10
      data:
        curves:
          - X25519
          - SECP256R1
          - SECP384R1
          - SECP521R1
          - X448
          - FFDHE2048
          - FFDHE3072
          - FFDHE4096
          - FFDHE6144
          - FFDHE8192
          
    # EC Point Formats
    - type: ec_point_formats
      id: 11
      data:
        formats:
          - uncompressed
          
    # Session Ticket
    - type: session_ticket
      id: 35
      
    # Application Layer Protocol Negotiation
    - type: application_layer_protocol_negotiation
      id: 16
      data:
        protocols:
          - h2
          - http/1.1
          
    # Status Request (OCSP Stapling)
    - type: status_request
      id: 5
      data:
        responder_id_list: []
        extensions: []
        
    # Signature Algorithms
    - type: signature_algorithms
      id: 13
      data:
        algorithms:
          - ECDSA_SECP256R1_SHA256
          - RSA_PSS_RSAE_SHA256
          - RSA_PKCS1_SHA256
          - ECDSA_SECP384R1_SHA384
          - RSA_PSS_RSAE_SHA384
          - RSA_PKCS1_SHA384
          - RSA_PSS_RSAE_SHA512
          - RSA_PKCS1_SHA512
          
    # Signed Certificate Timestamps
    - type: signed_certificate_timestamp
      id: 18
      
    # Key Share
    - type: key_share
      id: 51
      data:
        shares:
          - group: X25519
            length: 32
          - group: SECP256R1
            length: 65
            
    # Supported Versions
    - type: supported_versions
      id: 43
      data:
        versions:
          - "1.3"
          - "1.2"
          - "1.1"
          - "1.0"
          
    # Cookie
    - type: cookie
      id: 44
      
    # PSK Key Exchange Modes
    - type: psk_key_exchange_modes
      id: 45
      data:
        modes:
          - psk_dhe_ke
          
    # Certificate Authorities
    - type: certificate_authorities
      id: 47
      
    # GREASE
    - type: grease
      id: "0x9a9a"
      
    # Padding
    - type: padding
      id: 21
      data:
        length: 517

# HTTP/2 配置
http2:
  # SETTINGS 帧参数
  settings:
    header_table_size: 65536
    enable_push: 0
    max_concurrent_streams: 1000
    initial_window_size: 6291456
    max_frame_size: 16384
    max_header_list_size: 262144
    
  # 初始流窗口大小
  connection_flow: 15663105
  
  # 伪头部顺序
  pseudo_header_order:
    - ":method"
    - ":authority"
    - ":scheme"
    - ":path"
    
  # 优先级帧（可选）
  priority_frames:
    - stream_id: 0
      weight: 256
      depends_on: 0
      exclusive: false
      
  # 自定义帧处理
  frame_handlers:
    - type: settings
      ack_required: true
    - type: headers
      priority: true

# HTTP/1.x 配置
http1:
  # 头部顺序
  header_order:
    - host
    - connection
    - upgrade-insecure-requests
    - user-agent
    - accept
    - sec-fetch-site
    - sec-fetch-mode
    - sec-fetch-user
    - sec-fetch-dest
    - accept-encoding
    - accept-language
    - cookie
    
  # 头部值（模板）
  headers:
    user_agent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
    accept: "text/html,application/xhtml+xml,application/xml;q=0.9"
    accept_language: "en-US,en;q=0.9"
    accept_encoding: "gzip, deflate, br"
    
  # 请求配置
  request:
    http_version: "1.1"
    keep_alive: true
    pipeline: false
    chunked: true

# 行为特征
behavior:
  # 连接行为
  connection:
    max_connections: 6
    max_persistent: 4
    pipeline: false
    
  # 超时设置
  timeout:
    connection: 30s
    handshake: 10s
    request: 30s
    keep_alive: 120s
    
  # 重试策略
  retry:
    max_retries: 3
    backoff: exponential
    
# 匹配规则
matching:
  # User-Agent 模式
  user_agent_patterns:
    - "Chrome/120"
    - "Chrome/121"
    
  # Client Hints
  client_hints:
    sec_ch_ua: '"Not_A Brand";v="8", "Chromium";v="120"'
    sec_ch_ua_mobile: "?0"
    sec_ch_ua_platform: '"Windows"'
    
  # TLS 特征匹配
  tls_fingerprints:
    ja3_hash: "cd08e31494f9531f560d64c695473da9"
```plaintext

## 配置验证

### 验证命令

```bash
# 验证单个配置文件
go run ./cmd/profilegen validate profiles/specs/chrome_120.yaml

# 批量验证
go run ./cmd/profilegen validate profiles/specs/*.yaml

# 严格模式
go run ./cmd/profilegen validate --strict profiles/specs/chrome_120.yaml
```plaintext

## 配置继承

### 基础模板

```yaml
# profiles/base/desktop_chrome.yaml
template:
  name: "desktop_chrome_base"
  abstract: true
  
tls:
  compression_methods:
    - null
  extensions:
    - type: extended_master_secret
    - type: renegotiation_info
      
http2:
  settings:
    enable_push: 0
    
behavior:
  connection:
    max_connections: 6
```plaintext

### 继承模板

```yaml
# profiles/specs/chrome_120.yaml
metadata:
  name: "chrome_120"
  
extends: "desktop_chrome_base"

browser:
  name: "Chrome"
  version: "120.0.0"
  
tls:
  cipher_suites:
    - TLS_AES_128_GCM_SHA256
  extensions:
    - type: key_share
      data:
        shares:
          - group: X25519
```plaintext

## 环境特定配置

### 配置覆盖

```yaml
# config.production.yaml
overrides:
  - path: "metrics.enabled"
    value: true
  - path: "logging.level"
    value: "warn"
  - path: "cache.size"
    value: 10000
```plaintext

### 环境变量替换

```yaml
metrics:
  enabled: ${METRICS_ENABLED:true}
  port: ${METRICS_PORT:9090}
  
cache:
  size: ${CACHE_SIZE:1000}
  
logging:
  level: ${LOG_LEVEL:info}
```plaintext

## 热重载

```go
// 文件监听
func (w *Watcher) Watch(paths []string) error {
    for _, path := range paths {
        if err := w.watcher.Add(path); err != nil {
            return err
        }
    }
    return nil
}

// 重载回调
config.OnReload(func(oldCfg, newCfg *Config) {
    log.Info("Configuration reloaded")
})
```plaintext

## 配置版本管理

### 版本声明

```yaml
apiVersion: "fingerprint.io/v1"
kind: Profile
metadata:
  name: chrome_120
  version: "1.0.0"
```plaintext

### 迁移工具

```bash
# 检查配置版本
go run ./cmd/profilegen migrate check profiles/

# 自动迁移
go run ./cmd/profilegen migrate run profiles/ --to v2

# 生成报告
go run ./cmd/profilegen migrate report profiles/
```plaintext

## 配置最佳实践

1. **模块化组织**: 使用目录结构组织不同类型配置
2. **命名规范**: 使用小写和下划线命名
3. **文档注释**: 添加 YAML 注释说明
4. **敏感数据**: 不要在配置中硬编码敏感信息
