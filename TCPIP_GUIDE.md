# TCP/IP 指纹识别系统使用指南

## 概述

TCP/IP 指纹识别系统是一个完整的网络层特征分析工具，用于识别设备、操作系统、检测网络异常和评估安全风险。

## 核心组件

### 1. NetworkBehaviorAnalyzer（网络行为分析器）

用于分析网络流量的行为特征。

```go
import "github.com/vistone/fingerprint/internal/tcpip"

// 创建分析器
analyzer := tcpip.NewNetworkBehaviorAnalyzer()

// 记录数据包和 RTT
analyzer.RecordPacket(packet, 5*time.Millisecond)

// 执行分析
result := analyzer.AnalyzeBehavior()

// 获取结果
fmt.Printf("网络类型: %s\n", result.RTTAnalysis.NetworkType)
fmt.Printf("序列号模式: %s\n", result.SequenceNumberPattern)
```plaintext

### 2. DeviceFingerprintingEngine（设备指纹识别引擎）

用于识别设备类型和操作系统。

```go
// 创建引擎
engine := tcpip.NewDeviceFingerprintingEngine()

// 注册设备轮廓
profile := &tcpip.DeviceProfile{
    Name:        "Windows11_Desktop",
    DeviceType:  "desktop",
    Manufacturer: "Microsoft",
    OS:          "Windows",
    OSVersion:   "11",
}
engine.RegisterDeviceProfile(profile)

// 分析设备
result := engine.AnalyzeDevice(packets, behaviorResult)

// 获取匹配结果
for _, match := range result.DeviceMatches {
    fmt.Printf("%s -> 匹配度: %.2f%%\n", match.DeviceName, match.MatchScore*100)
}
```plaintext

### 3. PacketParser（数据包解析器）

用于解析和验证网络数据包。

```go
// 创建解析器
parser := tcpip.NewPacketParser(rawData)

// 生成签名
signature := tcpip.FormatSignature(64, 1460, 65535, "MSS,WS,SACK,TS")

// 推断初始 TTL
initialTTL := tcpip.InferInitialTTL(currentTTL, hopCount)

// 检查 IP 地址
isPrivate := tcpip.IsPrivateIP("192.168.1.100")
```plaintext

## 检测能力

### 操作系统识别
- Windows (11, 10, Server 2019)
- Linux (Kernel 5.x, 4.x, Ubuntu 22.04)
- macOS (13, 12)
- iOS (16, 15/14)
- Android (13, 12)

### 设备类型
- 台式计算机（Desktop）
- 笔记本电脑（Laptop）
- 移动设备（Phone, Tablet）
- 物联网设备（IoT）
- 服务器（Server）
- 网络设备（Network Device）

### 网络异常检测
- 非标准 TTL 值
- 异常 MSS（最大段大小）
- 小窗口大小
- 异常的 TCP 选项顺序
- 零 ACK 号
- IP 欺骗迹象
- 分片异常

### 设备识别指标
- VPN 检测
- 代理检测
- NAT 检测
- 机器人/自动化流量
- 网络扫描行为
- IP 欺骗

## 使用场景

### 1. 实时威胁检测
```go
// 分析进入的网络流量
for range packets {
    result := analyzer.AnalyzePacket(packet)
    if len(result.Anomalies) > 0 {
        logSecurityAlert(result.Anomalies)
    }
}
```plaintext

### 2. 设备身份识别
```go
// 识别网络中的设备
for _, packet := range devicePackets {
    result := engine.AnalyzeDevice([]*tcpip.TCPPacket{packet}, nil)
    registerNetworkDevice(result.DeviceMatches[0])
}
```plaintext

### 3. 行为分析
```go
// 分析用户的网络行为模式
result := analyzer.AnalyzeBehavior()
userProfile := buildUserProfile(result.BehaviorCharacteristics)
```plaintext

## 配置说明

### 自定义风险评分
```go
indicators := tcpip.RiskIndicators{
    IsBot:      false,
    IsScanner:  true,
    IsVPN:      false,
    IsSpoofed:  false,
    Suspicious: []string{"scanning_behavior"},
}

riskScore := tcpip.CalculateRiskScore(indicators)
```plaintext

### 扩展 OS 数据库
```go
newProfile := &tcpip.OSProfile{
    Name:              "Custom_OS",
    Family:            "Linux",
    Version:           "5.15",
    DefaultTTL:        64,
    DefaultMSS:        1460,
    DefaultWindowSize: 32768,
    TCPOptions:        "MSS,TS,WS",
}
osDatabase["Custom_OS"] = newProfile
```plaintext

## 性能指标

- TCP/IP 分析：< 1ms 每个数据包
- 设备指纹识别：< 5ms 每个分析
- 行为分析：O(n) 其中 n 是数据包数
- 内存使用：< 1MB 用于 1000 个数据包

## 高级用法

### 1. 集成配置系统
```go
config := internal.config.GetConfigCenter()
tcpipConfig := config.Get("tcpip_analysis")
```plaintext

### 2. 使用插件系统扩展
```go
plugin := internal.plugins.GetRegistry().Get("custom_fingerprint")
```plaintext

### 3. 与风险评分集成
```go
risk := calculateOverallRisk(
    deviceFingerprint,
    networkBehavior,
    tlsSignature,
)
```plaintext

## 示例代码

### 运行基础示例
```bash
go run examples/tcpip/main.go
```plaintext

### 运行高级示例
```bash
go run examples/tcpip_advanced/main.go
```plaintext

## 常见问题

**Q: 如何检测 VPN？**
A: 系统会识别非标准 TTL、一致的 IP ID 计数器和异常的 TCP 选项组合作为 VPN 迹象。

**Q: 系统如何确定网络类型？**
A: 基于平均 RTT 时间：本地 LAN (<10ms)、国内 (10-50ms)、区域 (50-150ms)、国际 (>150ms)

**Q: 可以离线使用吗？**
A: 是的，OS 数据库是内置的，不需要网络连接。

## 贡献指南

欢迎通过以下方式贡献：
- 提交新的 OS 特征数据
- 添加设备类型支持
- 改进异常检测算法

## 许可证

由项目许可证管理。
