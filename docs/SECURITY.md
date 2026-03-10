# 安全政策

## 报告安全漏洞

### ⚠️ 不要公开报告

如果发现安全漏洞，**请不要在 GitHub Issues 中公开讨论**。这可能会给所有用户带来风险。

### 如何报告

请发送电子邮件至安全联系人，详细说明：

1. **漏洞描述** - 清晰详细的技术说明
2. **影响范围** - 哪些组件/功能受影响
3. **复现步骤** - 如何重现该漏洞
4. **建议修复** - 如果您有解决方案

### 报告处理流程

1. **确认收到** - 我们将在 48 小时内确认收到报告
2. **评估风险** - 评估漏洞的严重性和影响
3. **制定修复** - 开发补丁并进行测试
4. **发布更新** - 发布安全更新版本
5. **公开披露** - 漏洞修复后可以公开讨论

**预期响应时间：**
- 严重漏洞：1-2 周内修复和发布
- 中等漏洞：2-4 周内修复和发布
- 低等漏洞：下一个常规更新中修复

## 安全最佳实践

### 对于库用户

1. **定期更新** - 及时更新到最新版本
   ```bash
   go get -u github.com/vistone/fingerprint/modules/fingerprint
   ```

2. **订阅发布** - 关注 GitHub Release 通知
   - 访问 https://github.com/vistone/fingerprint/releases
   - 点击 "Watch" 并选择 "Releases only"

3. **验证版本** - 确保使用的版本在 go.mod 中明确声明
   ```go
   require github.com/vistone/fingerprint/modules/fingerprint v1.0.6
   ```

4. **依赖审计** - 定期检查依赖中的已知漏洞
   ```bash
   go list -u -m all
   ```

### 对于贡献者

1. **代码审查** - 所有代码变更需要审查
2. **测试覆盖** - 新代码必须有单元测试
3. **不要硬编码密钥** - 敏感信息不应出现在代码中
4. **清理日志** - 移除调试日志和敏感信息

## 已知安全问题

当前版本（v1.0.6）中没有已知的未修复的安全问题。

### 历史安全修复

详见 [CHANGELOG.md](./CHANGELOG.md) 的 "Fixed" 部分。

## 依赖安全

本项目使用以下主要依赖：

| 依赖包 | 用途 | 已知问题 |
|------|------|--------|
| Go 标准库 | 核心功能 | 无 |
| 内部模块 | 核心 + profiles | 无 |

### 检查依赖漏洞

```bash
# 使用 go list 查看所有依赖
go list -m all

# 使用 govulncheck 检查已知漏洞（需要安装）
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## 版本支持

| 版本 | 状态 | 安全更新 |
|------|------|--------|
| v1.0.6+ | 当前 | ✅ 有效 |
| v1.0.5 | 维护中 | ✅ 关键修复 |
| v1.0.4 及之前 | 旧版 | ❌ 无 |

## TLS/密码学相关

### TLS 指纹库的安全特性

本库用于**生成和分析** TLS 指纹，而不是执行密码学操作。安全相关注意事项：

1. **不进行密钥操作** - 不处理/存储私钥
2. **不实现密码算法** - 仅进行指纹分析
3. **读取公开信息** - 只处理 TLS 握手中的公开值

### 使用注意

- ✅ 用于浏览器识别和流量分析
- ✅ 用于安全监控和异常检测
- ❌ 不能用于规避安全防护
- ❌ 不能用于恶意目的

## 合规性

本项目遵守以下标准：

- **Go 编码规范** - 遵循官方 Go 最佳实践
- **License** - BSD 3-Clause License（见 LICENSE 文件）
- **数据隐私** - 不收集或存储用户数据

## 安全审计

本项目未进行第三方安全审计。如果您使用此库处理敏感应用，**建议进行您自己的安全评估**。

## 安全相关资源

- [OWASP - Transport Layer Protection](https://owasp.org/www-community/attacks/Manipulator-in-the-middle_attack)
- [RFC 5246 - TLS 1.2](https://tools.ietf.org/html/rfc5246)
- [RFC 8446 - TLS 1.3](https://tools.ietf.org/html/rfc8446)
- [IANA - TLS Parameters](https://www.iana.org/assignments/tls-parameters)

## 联系信息

**安全联系电子邮件：** [security@example.com](mailto:security@example.com)

（如果您发现此项目中的具体安全问题，请通过上述渠道报告）

## 更新日志

### 2026-03-10
- 发布初始安全政策
- 建立安全报告流程
- 记录已知问题和依赖审计指导

---

**最后更新：2026 年 3 月 10 日**

**当前版本：v1.0.6** ✅ 无已知安全问题
