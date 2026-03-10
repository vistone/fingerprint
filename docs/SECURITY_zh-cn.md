# 安全政策

## 报告漏洞

**不要** 在 GitHub Issues 中公开报告安全漏洞。

请发送电子邮件至 security@example.com，详细说明：

1. 漏洞描述
2. 影响范围
3. 复现步骤
4. 建议修复

### 响应时间

- 严重漏洞：1-2 周内修复
- 中等漏洞：2-4 周内修复
- 低等漏洞：下一个版本中修复

## 最佳实践

### 用户

- 定期更新：`go get -u github.com/vistone/fingerprint/modules/fingerprint`
- 订阅release 通知：https://github.com/vistone/fingerprint/releases
- 明确声明版本：`require ... v1.0.7`
- 审计依赖：`go list -u -m all`

### 贡献者

- 所有代码变更需代码审查
- 新代码必须有单元测试
- 不硬编码密钥和敏感信息
- 移除调试日志

## 已知问题

当前版本（v1.0.7）无已知未修复的安全问题。

## 依赖

主要依赖：

| 包 | 用途 | 状态 |
|---|------|------|
| Go 标准库 | 核心 | ✅ |
| 内部模块 | 核心 + profiles | ✅ |

### 检查漏洞

```bash
# 列出所有依赖
go list -m all

# 检查已知漏洞（需安装）
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## 版本支持

| 版本 | 状态 | 安全更新 |
|------|------|--------|
| v1.0.7+ | 当前 | ✅ |
| v1.0.6 | 维护中 | ✅ |
| v1.0.5 及之前 | 旧版 | ❌ |

## TLS 指纹库

本库用于 **生成和分析** TLS 指纹，不执行密码学操作。

### 特性

- ✅ 不处理/存储私钥
- ✅ 不实现密码算法
- ✅ 仅读取 TLS 握手中的公开值

### 使用

- ✅ 浏览器识别和流量分析
- ✅ 安全监控和异常检测
- ❌ 规避安全防护
- ❌ 恶意目的

## 合规性

- Go 编码规范
- BSD 3-Clause License
- 不收集或存储用户数据

## 审计

本项目未进行第三方安全审计。处理敏感应用请自行评估。

## 资源

- [OWASP - Transport Layer Protection](https://owasp.org/www-community/attacks/Manipulator-in-the-middle_attack)
- [RFC 5246 - TLS 1.2](https://tools.ietf.org/html/rfc5246)
- [RFC 8446 - TLS 1.3](https://tools.ietf.org/html/rfc8446)

---

**最后更新：2026-03-10**  
**当前版本：v1.0.7** ✅
