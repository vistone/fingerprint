# 测试数据规范文档

## 核心原则

**所有测试必须使用真实数据，禁止使用模拟/虚拟数据。**

## 什么是真实数据

### ✅ 允许使用的真实数据

1. **真实的浏览器指纹配置**
   - `profiles.MappedTLSClients` 中的实际指纹配置
   - 示例：`profiles.MappedTLSClients["chrome_133"]`

2. **真实的 TLS ClientHelloSpec**
   - 通过 `profile.GetClientHelloSpec()` 获取的实际 TLS 配置
   - 包含真实的密码套件、扩展列表等

3. **真实的 HTTP Headers**
   - 通过 `random.GetRandomFingerprint()` 获取的完整 Headers
   - 包含真实的 User-Agent、Accept-Language 等

4. **真实的 User-Agent 字符串**
   - 通过 `useragent.GetUserAgentByProfileName()` 生成的 UA
   - 基于真实浏览器模板生成

5. **真实的网络协议数据**
   - 从实际网络抓包获取的数据包
   - 与真实服务器交互获得的响应数据

### ❌ 禁止使用的模拟数据

1. **手动的字节数组构造**
   ```go
   // ❌ 错误：手动构造数据
   data := []byte{0xfe, 0x0d, 0x01, 0x00}
   ```

2. **硬编码的测试字符串**
   ```go
   // ❌ 错误：硬编码 UA
   ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/133"
   ```

3. **手动的 TLS 配置**
   ```go
   // ❌ 错误：手动构造 spec
   spec := tls.ClientHelloSpec{
       TLSVersMax: tls.VersionTLS13,
       CipherSuites: []uint16{
           tls.TLS_AES_128_GCM_SHA256,
       },
   }
   ```

4. **虚构的网络参数**
   ```go
   // ❌ 错误：虚构的 TTL/MSS
   ttl := 64
   mss := 1460
   ```

## 正确的测试示例

### 示例 1: 测试 JA3 指纹计算

```go
// ✅ 正确：使用真实指纹数据
func TestComputeJA3FromRealProfiles(t *testing.T) {
    // 使用真实的 Chrome 133 指纹
    chromeProfile, ok := profiles.MappedTLSClients["chrome_133"]
    if !ok {
        t.Fatal("chrome_133 profile not found")
    }

    spec, err := chromeProfile.GetClientHelloSpec()
    if err != nil {
        t.Skipf("chrome_133 does not support spec export: %v", err)
        return
    }

    result, err := ComputeJA3FromSpec(spec)
    if err != nil {
        t.Fatalf("ComputeJA3FromSpec failed: %v", err)
    }

    if result.Hash == "" {
        t.Error("Expected non-empty hash for chrome_133")
    }
}
```

### 示例 2: 测试异常检测

```go
// ✅ 正确：使用真实指纹的 User-Agent
func TestAnomalyDetector(t *testing.T) {
    detector := defense.NewAnomalyDetector()

    // 获取真实 Chrome 133 指纹的 User-Agent
    result, err := random.GetRandomFingerprintByBrowser("chrome")
    if err != nil {
        t.Fatalf("获取 Chrome 指纹失败: %v", err)
    }
    normalUA := result.UserAgent

    // 使用真实 UA 测试
    if detector.DetectHeadlessBrowser(normalUA) {
        t.Error("真实 Chrome UA 不应被检测为无头浏览器")
    }
}
```

### 示例 3: 测试 Headers 操作

```go
// ✅ 正确：使用真实指纹的 Headers
func TestHeadersCustomization(t *testing.T) {
    result, err := random.GetRandomFingerprint()
    if err != nil {
        t.Fatalf("获取随机指纹失败: %v", err)
    }

    // 在真实 Headers 上操作
    result.Headers.Set("Cookie", "session_id=test123")
    headers := result.Headers.ToMap()

    if cookie, ok := headers["Cookie"]; !ok || cookie != "session_id=test123" {
        t.Error("Cookie header 设置失败")
    }
}
```

## 例外情况

以下情况可以使用构造数据，但必须添加明确注释：

1. **协议解析器的单元测试**
   - 需要测试特定字节序列的解析逻辑
   - 必须基于真实协议规范构造
   - 示例：ECH 扩展的特定版本格式测试

2. **边界条件测试**
   - 测试空值、极值等边界情况
   - 示例：`[]byte{}` 或 `nil` 输入

3. **错误处理测试**
   - 测试无效数据的错误返回
   - 示例：格式错误的数据包

```go
// ✅ 允许：测试协议解析器的特定格式
func TestParseECHExtension(t *testing.T) {
    // 基于 ECH Draft 13 规范构造的测试数据
    echData := make([]byte, 10)
    echData[0] = 0xfe
    echData[1] = 0x0d // Version Draft 13
    echData[2] = 0x01 // Outer Hello
    // ... 其余字段

    ech, err := ParseECHExtension(ExtensionEncryptedClientHello, echData)
    if err != nil {
        t.Fatalf("ParseECHExtension() error = %v", err)
    }
}
```

## 测试数据检查清单

在提交测试代码前，请确认：

- [ ] 所有指纹数据来自 `profiles.MappedTLSClients`
- [ ] 所有 User-Agent 通过 `useragent.GetUserAgentByProfileName` 获取
- [ ] 所有 Headers 通过 `random.GetRandomFingerprint` 获取
- [ ] 所有 TLS 配置通过 `profile.GetClientHelloSpec()` 获取
- [ ] 没有硬编码的浏览器版本号
- [ ] 没有手动的字节数组构造（除非测试协议解析器）
- [ ] 所有测试数据都是实际项目中会使用的数据

## 推荐的测试模式

### 模式 1: 遍历所有真实指纹

```go
func TestAllProfiles(t *testing.T) {
    for name, profile := range profiles.MappedTLSClients {
        t.Run(name, func(t *testing.T) {
            // 在每个真实指纹上测试
            spec, err := profile.GetClientHelloSpec()
            if err != nil {
                t.Skipf("Profile %s does not support spec export", name)
                return
            }
            // 测试逻辑...
        })
    }
}
```

### 模式 2: 测试特定浏览器系列

```go
func TestChromeProfiles(t *testing.T) {
    chromeProfiles := []string{
        "chrome_133", "chrome_120", "chrome_131",
    }

    for _, name := range chromeProfiles {
        t.Run(name, func(t *testing.T) {
            profile, ok := profiles.MappedTLSClients[name]
            if !ok {
                t.Skipf("Profile %s not found", name)
                return
            }
            // 测试逻辑...
        })
    }
}
```

### 模式 3: 使用多浏览器测试

```go
func TestMultipleBrowsers(t *testing.T) {
    browsers := []string{"chrome", "firefox", "safari", "edge"}

    for _, browser := range browsers {
        t.Run(browser, func(t *testing.T) {
            result, err := random.GetRandomFingerprintByBrowser(browser)
            if err != nil {
                t.Fatalf("获取 %s 指纹失败: %v", browser, err)
            }
            // 测试逻辑...
        })
    }
}
```

## 违规检查

使用以下命令检查测试文件中是否包含模拟数据：

```bash
# 检查硬编码的 User-Agent
grep -r "Mozilla/5.0" --include="*_test.go" .

# 检查手动的字节构造
grep -r "\[\]byte{" --include="*_test.go" .

# 检查手动的 TLS 版本
grep -r "VersionTLS" --include="*_test.go" .
```

## 文档维护

- 本规范随项目更新而更新
- 新的测试必须遵循本规范
- 旧测试逐步重构以符合规范

---

**最后更新**: 2026-03-03
