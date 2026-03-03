# API 文档

本文档描述 fingerprint 库的公共 API 使用方式。

## 快速开始

### 安装

```bash
go get github.com/vistone/fingerprint
```

### 基本使用

```go
package main

import (
    "github.com/vistone/fingerprint/profiles"
    "github.com/vistone/fingerprint/tls/ja3"
)

func main() {
    // 选择 Profile
    profile := profiles.Chrome_120
    
    // 获取 TLS 配置
    spec := profile.GetClientHelloSpec()
    
    // 使用 utls 建立连接
    // ...
}
```

## Profile API

### ClientProfile

完整的客户端指纹配置。

```go
type ClientProfile struct {
    ClientHelloSpec     tls.ClientHelloSpec
    HTTP2Settings       map[http2.SettingID]uint32
    HTTP2PseudoHeaders  []string
    ConnectionFlow      uint32
    PriorityFrames      []http2.Priority
}
```

#### 方法

**GetClientHelloSpec**

```go
func (p ClientProfile) GetClientHelloSpec() (tls.ClientHelloSpec, error)
```

返回 TLS ClientHello 配置。

```go
profile := profiles.Chrome_120
spec, err := profile.GetClientHelloSpec()
if err != nil {
    log.Fatal(err)
}

// 使用 utls
cfg := &tls.Config{ServerName: "example.com"}
conn := tls.UClient(conn, cfg, tls.HelloCustom)
conn.ApplyPreset(spec)
```

**GetSettings**

```go
func (p ClientProfile) GetSettings() map[http2.SettingID]uint32
```

返回 HTTP/2 SETTINGS 帧参数。

```go
settings := profile.GetSettings()
for id, value := range settings {
    fmt.Printf("Setting %d = %d\n", id, value)
}
```

**GetPseudoHeaderOrder**

```go
func (p ClientProfile) GetPseudoHeaderOrder() []string
```

返回 HTTP/2 伪头部顺序。

```go
order := profile.GetPseudoHeaderOrder()
// 返回: []string{":method", ":authority", ":scheme", ":path"}
```

### Profile 选择

```go
// 内置 Profile
profile := profiles.Chrome_120
profile := profiles.Firefox_121
profile := profiles.Safari_17

// 通过名称获取
profile, ok := profiles.Get("chrome_120")
if !ok {
    log.Fatal("profile not found")
}

// 通过 User-Agent 自动匹配
profile := profiles.FromUserAgent(ua)
```

## TLS API

### JA3

#### Parse

```go
func Parse(ja3 string) (*JA3, error)
```

解析 JA3 字符串。

```go
ja3 := "769,47-53-5-10,0-10-11,23-24-25,0"
parsed, err := ja3.Parse(ja3)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Version: %d\n", parsed.Version)
fmt.Printf("Ciphers: %v\n", parsed.CipherSuites)
```

#### Calculate

```go
func Calculate(data []byte) string
```

从 ClientHello 数据计算 JA3 指纹。

```go
clientHello := []byte{...} // TLS ClientHello 消息
fingerprint := ja3.Calculate(clientHello)
fmt.Printf("JA3: %s\n", fingerprint)
```

#### String

```go
func (j *JA3) String() string
```

将 JA3 结构序列化为字符串。

```go
ja3Obj := &ja3.JA3{
    Version:     769,
    CipherSuites: []uint16{47, 53, 5, 10},
    Extensions:   []uint16{0, 10, 11},
    EllipticCurves: []uint16{23, 24, 25},
    EllipticCurvePointFormats: []uint8{0},
}
fmt.Println(ja3Obj.String())
```

### JA4

```go
func CalculateJA4(clientHello []byte) (string, error)
```

计算 JA4 指纹。

```go
fp, err := ja4.CalculateJA4(clientHello)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("JA4: %s\n", fp)
// 输出: t13d1516h2_8daaf6152771_b1ff8e5f1d09
```

### JA4S

```go
func CalculateJA4S(serverHello []byte) (string, error)
```

计算 JA4S（Server Hello）指纹。

```go
fp, err := ja4s.CalculateJA4S(serverHello)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("JA4S: %s\n", fp)
```

## HTTP API

### HTTP/2 分析

#### AnalyzeFrame

```go
func AnalyzeFrame(data []byte) (*FrameInfo, error)
```

分析 HTTP/2 帧。

```go
frameData := []byte{0x00, 0x00, 0x00, 0x04, 0x01, ...}
info, err := http2.AnalyzeFrame(frameData)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Type: %s\n", info.Type)
fmt.Printf("Flags: %d\n", info.Flags)
fmt.Printf("StreamID: %d\n", info.StreamID)
```

#### ParseSettings

```go
func ParseSettings(data []byte) (map[SettingID]uint32, error)
```

解析 SETTINGS 帧负载。

```go
settings, err := http2.ParseSettings(frameData)
if err != nil {
    log.Fatal(err)
}

for id, value := range settings {
    fmt.Printf("%s = %d\n", id.String(), value)
}
```

### JA4H

```go
func CalculateJA4H(req *http.Request) (string, error)
```

计算 HTTP 请求指纹。

```go
req, _ := http.NewRequest("GET", "https://example.com/", nil)
req.Header.Set("User-Agent", "Mozilla/5.0...")
req.Header.Set("Accept", "text/html")

fp, err := ja4h.CalculateJA4H(req)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("JA4H: %s\n", fp)
```

### Client Hints

```go
func ParseClientHints(headers http.Header) (*ClientHints, error)
```

解析 Client Hints 头。

```go
hints, err := clienthints.ParseClientHints(req.Header)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Platform: %s\n", hints.Platform)
fmt.Printf("Mobile: %v\n", hints.Mobile)
```

## Config API

### 全局配置

```go
import "github.com/vistone/fingerprint/internal/config"

// 获取当前配置
cfg := config.Get()

// 获取克隆（线程安全）
localCfg := cfg.Clone()
```

### 动态配置

```go
// 监听配置变更
ch := make(chan config.Change)
config.Watch(ch)

go func() {
    for change := range ch {
        fmt.Printf("Config changed: %s\n", change.Path)
    }
}()

// 手动重载
if err := config.Reload(); err != nil {
    log.Fatal(err)
}
```

### Feature Extraction 配置

```go
featCfg := config.GetFeatureExtractionConfig()

if featCfg.TLS.JA3.Enabled {
    // 启用 JA3 提取
}

if featCfg.HTTP.Headers.Enabled {
    // 启用 HTTP 头部提取
}
```

## TCP/IP API

### 分析网络行为

```go
import "github.com/vistone/fingerprint/internal/tcpip"

// 分析数据包
behavior, err := tcpip.AnalyzeNetworkBehavior(packet)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("OS: %s\n", behavior.OS)
fmt.Printf("TTL: %d\n", behavior.TTL)
fmt.Printf("Window Size: %d\n", behavior.WindowSize)
```

### TCP 签名匹配

```go
signature := tcpip.TCPSignature{
    WindowSize: 65535,
    Options:    []byte{0x02, 0x04, 0x05, 0xb4, 0x01, 0x03, 0x03, 0x08},
    TTL:        64,
}

os, confidence := tcpip.MatchOSSignature(signature)
fmt.Printf("Detected OS: %s (confidence: %.2f)\n", os, confidence)
```

## Metrics API

### 默认服务器

```go
import "github.com/vistone/fingerprint/internal/metrics"

// 启动指标服务器
metrics.InitDefaultServer(":9090")
defer metrics.StopDefaultServer()
```

### 自定义指标

```go
// 创建计数器
counter := metrics.NewCounter("my_custom_total", "Description")
counter.Inc()

// 创建直方图
histogram := metrics.NewHistogram("my_duration_ms", "Description", 
    []float64{10, 50, 100, 500, 1000})
histogram.Observe(duration)
```

### 记录指标

```go
// 记录指纹生成
metrics.RecordFingerprintGeneration("Chrome", "Windows", nil)

// 记录错误
metrics.RecordGenerationError("Chrome")

// 记录缓存命中/未命中
metrics.RecordCacheHit()
metrics.RecordCacheMiss()
```

## 错误处理

### 预定义错误

```go
var (
    ErrInvalidProfile = errors.New("invalid client profile")
    ErrNotFound       = errors.New("profile not found")
    ErrParseError     = errors.New("parse error")
)
```

### 错误检查

```go
profile, err := profiles.Get(name)
if err != nil {
    if errors.Is(err, profiles.ErrNotFound) {
        // 使用默认 Profile
        profile = profiles.Chrome_120
    } else {
        log.Fatal(err)
    }
}
```

## 完整示例

### 指纹代理服务器

```go
package main

import (
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    
    "github.com/vistone/fingerprint/internal/metrics"
    "github.com/vistone/fingerprint/profiles"
)

func main() {
    // 启动指标服务器
    metrics.InitDefaultServer(":9090")
    defer metrics.StopDefaultServer()
    
    // 目标服务器
    target, _ := url.Parse("http://localhost:8080")
    proxy := httputil.NewSingleHostReverseProxy(target)
    
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // 记录请求
        start := time.Now()
        defer func() {
            metrics.RecordFingerprintGeneration(
                profiles.DetectBrowser(r),
                profiles.DetectOS(r),
                nil,
            )
            metrics.ObserveGenerationDuration(time.Since(start))
        }()
        
        // 代理请求
        proxy.ServeHTTP(w, r)
    })
    
    log.Println("Server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 自定义 Profile 加载器

```go
type FileProfileSource struct {
    Dir string
}

func (s *FileProfileSource) Load() ([]profiles.ClientProfile, error) {
    entries, err := os.ReadDir(s.Dir)
    if err != nil {
        return nil, err
    }
    
    var profiles []profiles.ClientProfile
    for _, entry := range entries {
        if filepath.Ext(entry.Name()) != ".yaml" {
            continue
        }
        
        data, err := os.ReadFile(filepath.Join(s.Dir, entry.Name()))
        if err != nil {
            return nil, err
        }
        
        var profile profiles.ClientProfile
        if err := yaml.Unmarshal(data, &profile); err != nil {
            return nil, err
        }
        
        profiles = append(profiles, profile)
    }
    
    return profiles, nil
}

// 使用
source := &FileProfileSource{Dir: "./custom-profiles"}
profiles, err := source.Load()
```

## 最佳实践

### 1. Profile 缓存

```go
var profileCache = sync.Map{}

func GetCachedProfile(name string) (profiles.ClientProfile, bool) {
    if v, ok := profileCache.Load(name); ok {
        return v.(profiles.ClientProfile), true
    }
    
    profile, ok := profiles.Get(name)
    if ok {
        profileCache.Store(name, profile)
    }
    return profile, ok
}
```

### 2. 连接池

```go
var dialer = &utls.Dialer{
    Config: &tls.Config{
        InsecureSkipVerify: true,
    },
}

func GetConnection(ctx context.Context, addr string, spec tls.ClientHelloSpec) (*tls.Conn, error) {
    conn, err := dialer.DialContext(ctx, "tcp", addr)
    if err != nil {
        return nil, err
    }
    
    uconn := tls.UClient(conn, nil, tls.HelloCustom)
    if err := uconn.ApplyPreset(spec); err != nil {
        conn.Close()
        return nil, err
    }
    
    return uconn, nil
}
```

### 3. 超时控制

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

conn, err := GetConnection(ctx, "example.com:443", spec)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        // 处理超时
    }
    return err
}
```
