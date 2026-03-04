# 代码生成迁移计划

本文档描述了如何将现有的手工编写指纹配置迁移到 YAML + 代码生成的方案。

## 背景

当前 `profiles/` 目录下包含 90+ 个手工编写的指纹配置，使用了大量的 `//nolint:composites` 注释来压制 go vet 警告。代码生成方案可以：

1. 消除所有 `nolint:composites` 警告
2. 提供类型安全、IDE 友好的代码
3. 简化配置维护（YAML 比 Go 代码更易读）

## 迁移步骤

### Phase 1: 准备（已完成）

- [x] 创建代码生成工具 `cmd/profilegen/`
- [x] 创建示例 YAML 配置 `profiles/specs/chrome_133.yaml`
- [x] 验证代码生成可行性

### Phase 2: 批量生成 YAML 模板

```bash
# 生成所有 profile 的 YAML 模板
go run ./cmd/profilegen/generate_specs.go

# 查看生成的文件
ls -la profiles/specs/
```plaintext

每个生成的文件包含：
- 基本元数据（name, client, version）
- 从现有 profile 提取的 settings, pseudoHeaderOrder
- TODO 标记需要手动填写的字段（cipher_suites, extensions）

### Phase 3: 手动完善 YAML 配置

参考 `chrome_133.yaml` 完成其他配置：

```yaml
# 1. 填写 cipher_suites
cipher_suites:
  - tls.GREASE_PLACEHOLDER
  - tls.TLS_AES_128_GCM_SHA256
  # ... 从原始 Go 代码复制

# 2. 填写 extensions（最复杂的部分）
extensions:
  - type: UtlsGREASEExtension
    params: {}
  - type: SignatureAlgorithmsExtension
    params:
      supported_signature_algorithms:
        - tls.ECDSAWithP256AndSHA256
        # ... 从原始 Go 代码复制

# 3. 验证 YAML 格式
yamllint profiles/specs/firefox_120.yaml
```plaintext

### Phase 4: 生成 Go 代码

```bash
# 生成所有 profile 的 Go 代码
go run ./cmd/profilegen -input profiles/specs -output profiles/generated/profiles.go

# 验证生成的代码
go build ./profiles/generated/...
```plaintext

### Phase 5: 逐步替换

1. **并行运行**：在 `MappedTLSClients` 中同时使用手工和生成的 profile
2. **对比测试**：确保生成的代码与原始代码行为一致
3. **逐步替换**：按浏览器类型分批替换
   - Week 1: Chrome 系列
   - Week 2: Firefox 系列
   - Week 3: Safari 系列
   - Week 4: 其他（Edge, Opera, 移动端）

### Phase 6: 清理

- 删除所有 `//nolint:composites` 注释
- 删除手工编写的 profile 文件（保留备份）
- 更新 `profiles.go` 引用生成的代码

## 工作量估算

| 任务 | 工作量 | 负责人 |
| ------ | -------- | -------- |
| 生成 YAML 模板 | 1 天 | 自动化脚本 |
| 完善 Chrome 配置（~20 个） | 3 天 | 开发团队 |
| 完善 Firefox 配置（~15 个） | 2 天 | 开发团队 |
| 完善 Safari 配置（~10 个） | 2 天 | 开发团队 |
| 完善其他配置（~50 个） | 5 天 | 开发团队 |
| 测试验证 | 3 天 | QA 团队 |
| **总计** | **~16 天** | |

## 风险缓解

1. **配置错误风险**
   - 每个 YAML 配置必须经过严格的单元测试
   - 对比生成的 Go 代码与原始代码的 SpecFactory 输出

2. **迁移中断风险**
   - 采用并行运行策略，随时可以回滚
   - 保留原始 Go 文件备份

3. **维护成本增加**
   - 文档化 YAML 配置格式
   - 提供配置验证工具

## 工具支持

### 1. 批量生成模板

```bash
go run ./cmd/profilegen/generate_specs.go
```plaintext

### 2. 验证 YAML 配置

```bash
# 验证单个配置
go run ./cmd/profilegen/validate.go -config profiles/specs/chrome_133.yaml

# 批量验证
go run ./cmd/profilegen/validate.go -input profiles/specs/
```plaintext

### 3. 对比测试

```bash
# 对比生成的代码与原始代码
go run ./cmd/profilegen/diff.go -profile chrome_133
```plaintext

## 成功标准

1. 所有 90+ 个 profile 都有对应的 YAML 配置
2. 生成的 Go 代码通过所有单元测试
3. 零 `//nolint:composites` 注释
4. 性能指标无退化（生成耗时 < 10μs）

## 下一步行动

1. 运行 `go run ./cmd/profilegen/generate_specs.go` 生成所有 YAML 模板
2. 组建迁移小组，分配浏览器类型任务
3. 建立代码审查流程，确保配置正确性
4. 每周同步进度，调整计划
