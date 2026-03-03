# Profile 代码生成工具

该工具从 YAML 配置文件生成类型安全的 Go 代码，用于消除 `nolint:composites` 警告。

## 问题背景

当前 `profiles/` 目录下的指纹配置使用无键字段初始化，导致大量 `//nolint:composites` 注释：

```go
//nolint:composites
var Chrome_133 = ClientProfile{
    clientHelloId: tls.ClientHelloID{
        // ...
    },
}
```

## 解决方案

使用 YAML 配置文件 + 代码生成工具，生成类型安全的 Go 代码。

## 使用方法

### 1. 创建 YAML 配置文件

在 `profiles/specs/` 目录下创建 YAML 文件：

```yaml
# profiles/specs/chrome_133.yaml
name: chrome_133
var_name: Chrome_133
display_name: Chrome 133
client: Chrome
version: "133"
random_extension_order: false

cipher_suites:
  - tls.GREASE_PLACEHOLDER
  - tls.TLS_AES_128_GCM_SHA256
  # ...

extensions:
  - type: UtlsGREASEExtension
    params: {}
  - type: SignatureAlgorithmsExtension
    params:
      supported_signature_algorithms:
        - tls.ECDSAWithP256AndSHA256
        - tls.PSSWithSHA256
        # ...

settings:
  SettingHeaderTableSize: 65536
  SettingEnablePush: 0
  # ...
```

### 2. 运行代码生成工具

```bash
go run ./cmd/profilegen -input profiles/specs -output profiles/generated.go
```

### 3. 替换原有代码

将生成的代码复制到 `profiles/internal_browser_profiles.go`，替换原有的手工实现。

## 优势

1. **类型安全**：生成的代码使用命名字段，无 `nolint:composites` 警告
2. **易于维护**：修改 YAML 即可，无需手动编辑 Go 代码
3. **一致风格**：所有配置使用统一的代码风格
4. **版本控制**：YAML 文件更易读、易 diff

## 迁移计划

### 阶段 1：验证（当前）
- [x] 创建代码生成工具
- [x] 创建示例 YAML 配置
- [ ] 验证生成的代码与原有代码等价

### 阶段 2：并行维护
- 新指纹使用 YAML + 代码生成
- 旧指纹保持手工维护

### 阶段 3：逐步替换
- 将旧指纹迁移到 YAML
- 最终移除所有 `nolint:composites` 注释

## 配置文件格式

详见 `profiles/specs/chrome_133.yaml`

### 支持的扩展类型

- `UtlsGREASEExtension`
- `SessionTicketExtension`
- `SignatureAlgorithmsExtension`
- `ApplicationSettingsExtensionNew`
- `KeyShareExtension`
- `SCTExtension`
- `SupportedPointsExtension`
- `SupportedVersionsExtension`
- `StatusRequestExtension`
- `ALPNExtension`
- `SNIExtension`
- `BoringGREASEECH`
- `UtlsCompressCertExtension`
- `SupportedCurvesExtension`
- `PSKKeyExchangeModesExtension`
- `ExtendedMasterSecretExtension`
- `RenegotiationInfoExtension`

## 测试

```bash
# 生成代码
go run ./cmd/profilegen -input profiles/specs -output /tmp/generated.go

# 验证生成的代码可编译
go build /tmp/generated.go
```
