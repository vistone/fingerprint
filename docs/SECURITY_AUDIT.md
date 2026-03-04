# 安全审计报告

本文档提供 fingerprint 库的安全评估结果和加固建议。

## 审计范围

- **代码库**: github.com/vistone/fingerprint
- **审计日期**: 2026-03-02
- **审计版本**: main 分支
- **审计范围**: TLS/HTTP/Profile/Config/Internal 模块

## 执行摘要

| 风险等级 | 数量 | 描述 |
| ---------- | ------ | ------ |
| 严重 (Critical) | 0 | 无 |
| 高 (High) | 2 | 输入验证不足、配置注入风险 |
| 中 (Medium) | 4 | 信息泄露、DoS 风险、日志敏感数据 |
| 低 (Low) | 3 | 文档不足、测试覆盖不全 |

## 详细发现

### HIGH-1: JA3 输入验证不足

**风险描述**: JA3 解析器对输入字符串的验证不够严格，可能导致 panic 或资源耗尽。

**位置**: `tls/ja3/parse.go`

**代码片段**:
```go
// 当前实现
parts := strings.Split(ja3, ",")
version, _ := strconv.Atoi(parts[0])  // 可能 panic
```plaintext

**风险**:
- 超长输入导致内存分配过多
- 格式错误的输入导致 panic
- 负数或大数导致后续逻辑异常

**修复建议**:
```go
func Parse(ja3 string) (*JA3, error) {
    // 长度限制
    if len(ja3) > maxJA3Length {
        return nil, fmt.Errorf("JA3 string too long: %d", len(ja3))
    }
    
    parts := strings.Split(ja3, ",")
    if len(parts) != 5 {
        return nil, fmt.Errorf("invalid JA3 format: expected 5 parts, got %d", len(parts))
    }
    
    // 安全的数值解析
    version, err := strconv.ParseUint(parts[0], 10, 16)
    if err != nil {
        return nil, fmt.Errorf("invalid version: %w", err)
    }
    
    // ...
}
```plaintext

**优先级**: High
**状态**: 待修复

---

### HIGH-2: Profile 配置注入风险

**风险描述**: YAML Profile 解析可能允许代码执行或文件系统访问。

**位置**: `cmd/profilegen/parser.go`

**代码片段**:
```go
// 当前实现直接解析 YAML
data, _ := os.ReadFile(path)
yaml.Unmarshal(data, &profile)
```plaintext

**风险**:
- YAML 反序列化可能执行任意代码
- 路径遍历允许读取任意文件
- 嵌入的 Go 代码可能被执行

**修复建议**:
```go
func LoadProfile(path string) (*Profile, error) {
    // 路径规范化
    absPath, err := filepath.Abs(path)
    if err != nil {
        return nil, err
    }
    
    // 限制在允许目录内
    if !strings.HasPrefix(absPath, allowedBasePath) {
        return nil, fmt.Errorf("path outside allowed directory: %s", path)
    }
    
    data, err := os.ReadFile(absPath)
    if err != nil {
        return nil, err
    }
    
    // 大小限制
    if len(data) > maxProfileSize {
        return nil, fmt.Errorf("profile too large: %d bytes", len(data))
    }
    
    // 使用安全的 YAML 解析器
    decoder := yaml.NewDecoder(bytes.NewReader(data))
    decoder.SetStrict(true)  // 拒绝未知字段
    
    var profile Profile
    if err := decoder.Decode(&profile); err != nil {
        return nil, err
    }
    
    // 验证
    if err := validateProfile(&profile); err != nil {
        return nil, err
    }
    
    return &profile, nil
}
```plaintext

**优先级**: High
**状态**: 待修复

---

### MEDIUM-1: 敏感信息日志泄露

**风险描述**: 调试日志可能包含敏感信息（TLS 扩展中的域名、Cookie 等）。

**位置**: 多处调试日志

**代码片段**:
```go
// 可能泄露 SNI
logger.Debug("parsed extensions", "extensions", spew.Sdump(extensions))
```plaintext

**风险**:
- 日志文件泄露用户隐私
- 调试信息包含敏感域名
- 生产环境启用调试可能导致泄露

**修复建议**:
```go
// 创建安全的日志辅助函数
func LogExtensionsSafe(logger Logger, extensions []Extension) {
    // 过滤敏感字段
    safeExts := make([]Extension, len(extensions))
    for i, ext := range extensions {
        safeExts[i] = ext.Sanitize()  // 移除敏感数据
    }
    logger.Debug("parsed extensions", "count", len(safeExts), "types", getExtensionTypes(safeExts))
}

// Extension 接口添加 Sanitize 方法
type Extension interface {
    // ...
    Sanitize() Extension  // 返回副本，移除敏感字段
}
```plaintext

**优先级**: Medium
**状态**: 待修复

---

### MEDIUM-2: 正则表达式 DoS (ReDoS)

**风险描述**: User-Agent 解析使用的正则表达式可能存在回溯问题。

**位置**: `http/useragent/parse.go`

**代码片段**:
```go
// 可能存在回溯问题
uaPattern := regexp.MustCompile(`(?i)chrome/([\d\.]+)`)
matches := uaPattern.FindStringSubmatch(ua)  // 在特定输入下可能很慢
```plaintext

**风险**:
- 恶意构造的 User-Agent 导致 CPU 耗尽
- 影响服务可用性

**修复建议**:
```go
func ParseUserAgent(ua string) (*UserAgentInfo, error) {
    // 长度限制
    if len(ua) > maxUALength {
        return nil, fmt.Errorf("User-Agent too long")
    }
    
    // 使用超时
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    
    result := make(chan *UserAgentInfo, 1)
    go func() {
        result <- parseUserAgentInternal(ua)
    }()
    
    select {
    case info := <-result:
        return info, nil
    case <-ctx.Done():
        return nil, fmt.Errorf("User-Agent parsing timeout")
    }
}
```plaintext

**优先级**: Medium
**状态**: 待修复

---

### MEDIUM-3: 内存分配不受限制

**风险描述**: 解析大型输入时没有内存限制，可能导致 OOM。

**位置**: 多个解析函数

**风险**:
- 大型 TLS 握手消息导致内存耗尽
- HTTP/2 帧解析无限分配内存

**修复建议**:
```go
// 设置全局解析限制
const (
    maxTLSHandshakeSize = 64 * 1024      // 64KB
    maxHTTP2FrameSize   = 16 * 1024      // 16KB (标准限制)
    maxExtensionsSize   = 8 * 1024       // 8KB
)

func ParseClientHello(data []byte) (*ClientHello, error) {
    if len(data) > maxTLSHandshakeSize {
        return nil, fmt.Errorf("handshake too large: %d bytes", len(data))
    }
    // ...
}
```plaintext

**优先级**: Medium
**状态**: 待修复

---

### MEDIUM-4: 版本信息泄露

**风险描述**: 错误响应或日志可能泄露库版本信息，帮助攻击者。

**位置**: 错误处理、HTTP 响应头

**修复建议**:
```go
// 不要在错误中暴露版本
var ErrInternal = errors.New("internal error")  // 通用错误

// 只在 Debug 级别记录详细版本
if logger.Level() == DebugLevel {
    logger.Debug("library version", "version", Version)
}
```plaintext

**优先级**: Medium
**状态**: 待修复

---

### LOW-1: 测试覆盖率不足

**风险描述**: 安全相关代码缺乏充分的测试覆盖。

**数据**:
- `tls/ja3`: 覆盖率 75%
- `http/http2`: 覆盖率 65%
- `internal/security`: 覆盖率 50%

**建议**:
- 添加模糊测试 (Fuzzing)
- 添加边界条件测试
- 添加安全相关测试用例

**优先级**: Low
**状态**: 待处理

---

### LOW-2: 依赖安全检查缺失

**风险描述**: 缺少自动化依赖漏洞扫描。

**建议**:
```yaml
# .github/workflows/security.yml
name: Security Scan
on:
  schedule:
    - cron: '0 0 * * *'  # 每日运行
  push:
    branches: [main]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: securego/gosec@master
      - uses: snyk/actions/golang@master
        with:
          args: --severity-threshold=high
```plaintext

**优先级**: Low
**状态**: 待处理

---

### LOW-3: 安全文档不完整

**风险描述**: 缺乏安全最佳实践文档。

**建议**: 创建 `SECURITY.md` 包含：
- 安全更新策略
- 漏洞报告流程
- 安全加固指南

**优先级**: Low
**状态**: 待处理

## 安全加固建议

### 1. 输入验证

```go
package security

import "golang.org/x/text/unicode/norm"

// Validator 输入验证器
type Validator struct {
    maxLength int
    patterns  map[string]*regexp.Regexp
}

func (v *Validator) ValidateString(s string, rules ...ValidationRule) error {
    // 规范化
    s = norm.NFC.String(s)
    
    // 长度检查
    if len(s) > v.maxLength {
        return ErrTooLong
    }
    
    // 字符白名单
    if !v.isValidCharacters(s) {
        return ErrInvalidCharacters
    }
    
    return nil
}
```plaintext

### 2. 资源限制

```go
// ResourceLimiter 资源限制器
type ResourceLimiter struct {
    maxMemory    int64
    maxCPUTime   time.Duration
    maxRequests  int
}

func (l *ResourceLimiter) Wrap(handler Handler) Handler {
    return func(ctx context.Context, req Request) (Response, error) {
        // 内存限制
        if l.maxMemory > 0 {
            ctx = context.WithValue(ctx, maxMemoryKey, l.maxMemory)
        }
        
        // 超时
        if l.maxCPUTime > 0 {
            var cancel context.CancelFunc
            ctx, cancel = context.WithTimeout(ctx, l.maxCPUTime)
            defer cancel()
        }
        
        return handler(ctx, req)
    }
}
```plaintext

### 3. 安全日志

```go
// SecureLogger 安全日志记录器
type SecureLogger struct {
    inner  Logger
    filter FieldFilter
}

func (l *SecureLogger) Debug(msg string, fields ...Field) {
    // 过滤敏感字段
    safeFields := l.filter.Filter(fields)
    l.inner.Debug(msg, safeFields...)
}

func NewSecureLogger(inner Logger) *SecureLogger {
    return &SecureLogger{
        inner: inner,
        filter: NewFieldFilter(
            "password",
            "token",
            "cookie",
            "sni",  // 域名可能敏感
            "authorization",
        ),
    }
}
```plaintext

### 4. 安全头部

```go
// SecurityHeaders 添加安全头部
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 防止 MIME 嗅探
        w.Header().Set("X-Content-Type-Options", "nosniff")
        
        // XSS 保护
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        
        // 点击劫持保护
        w.Header().Set("X-Frame-Options", "DENY")
        
        // CSP
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        
        // 不泄露 Referer
        w.Header().Set("Referrer-Policy", "no-referrer")
        
        next.ServeHTTP(w, r)
    })
}
```plaintext

## 安全配置检查清单

```markdown
## 生产环境部署检查清单

### 网络层
- [ ] 使用 TLS 1.2+
- [ ] 禁用不安全的 Cipher Suites
- [ ] 配置正确的证书
- [ ] 使用防火墙限制访问

### 应用层
- [ ] 禁用调试日志
- [ ] 启用输入验证
- [ ] 配置资源限制
- [ ] 设置请求超时
- [ ] 启用速率限制

### 监控
- [ ] 配置安全告警
- [ ] 审计日志启用
- [ ] 异常检测
- [ ] 性能监控

### 依赖
- [ ] 依赖漏洞扫描
- [ ] 定期更新依赖
- [ ] 最小权限原则
```plaintext

## 漏洞报告流程

```plaintext
1. 发现潜在漏洞
   ↓
2. 私下报告 security@example.com
   - 不要公开披露
   - 提供复现步骤
   - 说明影响范围
   ↓
3. 维护者确认 (48小时内)
   ↓
4. 修复开发
   ↓
5. 安全更新发布
   ↓
6. 公开披露 (修复后 30 天)
```plaintext

## 行动计划

| 任务 | 优先级 | 负责人 | 截止日期 |
| ------ | -------- | -------- | ---------- |
| 修复 HIGH-1: JA3 输入验证 | High | TBD | 1周内 |
| 修复 HIGH-2: Profile 配置注入 | High | TBD | 1周内 |
| 修复 MEDIUM-1: 敏感日志 | Medium | TBD | 2周内 |
| 修复 MEDIUM-2: ReDoS | Medium | TBD | 2周内 |
| 修复 MEDIUM-3: 内存限制 | Medium | TBD | 2周内 |
| 添加安全测试 | Low | TBD | 1个月内 |
| 配置安全扫描 | Low | TBD | 1个月内 |
| 完善安全文档 | Low | TBD | 1个月内 |

## 工具推荐

### 静态分析
- `gosec`: Go 安全扫描
- `staticcheck`: 静态分析
- `govulncheck`: 漏洞检查

### 模糊测试
- `go-fuzz`: 原生模糊测试
- `AFL++`: 高级模糊测试

### 依赖扫描
- `Snyk`: 依赖漏洞扫描
- `Dependabot`: 自动更新
- `OWASP Dependency-Check`

## 参考资源

- [OWASP Go Security](https://owasp.org/www-project-go-security/)
- [Go Security Guidelines](https://github.com/securego/securego.github.io)
- [CWE Top 25](https://cwe.mitre.org/top25/)
