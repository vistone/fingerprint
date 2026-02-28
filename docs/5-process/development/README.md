# 开发规范文档
所有 Go 开发者必须遵循本目录中的规范。
## 📖 文档导航
### 🚀 快速开始
1. **首先阅读**: [03-quick-reference.go](./03-quick-reference.go)
   - 5 分钟了解核心要点
   - 核心检查清单
### 📚 详细规范
2. **00-go-development-rules.md** - 完整的 Go 语言开发规范
   - 适合: 所有 Go 开发者
   - 内容: 100+ 个详细规范和示例
   - 时间: 首次阅读 30-60 分钟
3. **01-fingerprint-project-rules.md** - 项目特定规范
   - 适合: 本项目开发者
   - 内容: 指纹相关的特定规范
   - 时间: 首次阅读 20-30 分钟
4. **02-code-comment-templates.md** - 注释模板
   - 适合: 编写代码注释时参考
   - 内容: 30+ 个注释模板和示例
   - 时间: 参考即可
## 🎯 规范执行
### 必须遵循
- ✅ 所有代码必须符合命名规范
- ✅ 所有公开 API 必须有注释
- ✅ 所有函数必须正确处理错误
- ✅ 所有代码必须通过 `gofmt` 和 `go vet`
### 代码审查
提交 PR 前检查:
- [ ] 阅读 03-quick-reference.go 的检查清单
- [ ] 确保代码符合所有项目
- [ ] 测试覆盖主要场景
- [ ] 文档已更新
### 强制执行
不符合规范的代码提交将被拒绝。
## 📍 查找方法
**我需要...**
- 快速了解规范 → 03-quick-reference.go
- 学习命名规范 → 00-go-development-rules.md 第 4 节
- 学习注释规范 → 02-code-comment-templates.md
- 了解 User-Agent 规范 → 01-fingerprint-project-rules.md 第 3 节
- 了解错误处理 → 00-go-development-rules.md 第 5 节
- 学习性能优化 → 00-go-development-rules.md 第 10 节
- 了解常见错误 → 00-go-development-rules.md 第 14 节 或 03-quick-reference.go 第 10 节
## ✨ 核心要点
### 命名规范
- 常量: `PascalCase` (BrowserChrome)
- 变量: `camelCase` (localCounter)
- 函数: `PascalCase + 动词` (GetRandomFingerprint)
### 注释规范
- 公开 API: 必须有 godoc 注释
- 复杂逻辑: 必须有行内注释
- 格式: `TODO(name)`, `FIXME(name)`
### 代码结构
- 导入: 标准库 → 第三方库 → 本项目
- 定义: const → var → type → interface → func → method
- 错误: 始终检查并包装
## 📊 规范覆盖范围
✅ Go 语言所有方面
- 包组织、导入、文件结构
- 常量、变量、类型、接口、函数
- 注释、命名、错误处理
- 性能、并发、测试
✅ 项目特定
- 指纹管理、User-Agent、Headers
- 异常检测、矛盾检测
- JA3/JA4 算法
## 🚀 立即开始
1. 打开 [03-quick-reference.go](./03-quick-reference.go) 了解核心要点
2. 在编码时参考相应规范
3. 代码审查时检查清单
## 📞 反馈和更新
规范如有问题或需要改进，请:
1. 提交 Issue
2. 在 PR 中讨论
3. 更新相应规范文件
---
**最后更新**: 2026-02-28
**版本**: 1.0
**维护者**: vistone
