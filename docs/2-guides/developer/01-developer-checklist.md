# ✅ 开发者检查清单

<!-- markdownlint-disable MD009 MD022 MD031 MD032 MD036 -->

本文档提供了可操作的、分步骤的改进检查清单。

---

## 🚀 快速启动 (第 1 天)

### 第 1 小时: 代码格式化

- [ ] 在项目根目录打开终端
- [ ] 运行格式化命令
  ```bash
  cd /media/stone/data/dev
  gofmt -w .
  ```
- [ ] 验证无文件需要格式化
  ```bash
  gofmt -l .  # 应输出: 空
  ```
- [ ] 提交修改
  ```bash
  git add -A
  git commit -m "style: fix code formatting with gofmt"
  ```
- [ ] 推送到远程
  ```bash
  git push origin main
  ```

**✨ 第一个改进完成! 预计耗时: 5 分钟**

---

### 第 2-5 小时: 修复 Go Vet 警告

- [ ] 运行 Go vet 查看所有警告
  ```bash
  go vet ./... 2>&1 | tee vet_warnings.txt
  ```
- [ ] 分析警告 - 大部分是结构体初始化问题
- [ ] 打开第一个问题文件: `profiles/internal_browser_profiles.go`
- [ ] 逐一修改结构体初始化
  - [ ] 查找 `SupportedCurvesExtension{` 
  - [ ] 改为 `SupportedCurvesExtension{Curves: ...`
  - [ ] 查找 `KeyShareExtension{`
  - [ ] 改为 `KeyShareExtension{KeyShares: ...`
  - 🔄 对所有扩展类型重复此过程

- [ ] 打开第二个问题文件: `profiles/contributed_browser_profiles.go`
- [ ] 使用相同步骤修复所有警告

- [ ] 重新运行 vet 验证修复
  ```bash
  go vet ./...  # 应输出: 无错误
  ```

- [ ] 运行测试确保功能完整
  ```bash
  go test ./... -v  # 应通过所有测试
  ```

- [ ] 提交修改
  ```bash
  git add -A
  git commit -m "fix: resolve go vet warnings with named struct fields"
  git push origin main
  ```

**✨ 第二个改进完成! 预计耗时: 3-4 小时**

---

### 提交 go.sum 修改

- [ ] 查看 go.sum 修改状态
  ```bash
  git status go.sum
  ```
- [ ] 添加并提交
  ```bash
  git add go.sum
  git commit -m "chore: update go.sum after go mod tidy"
  git push origin main
  ```

**✨ 第三个改进完成! 预计耗时: 10 分钟**

---

## 📦 CI/CD 建立 (第 2-3 周)

### 创建 GitHub Actions 工作流

- [ ] 创建目录结构
  ```bash
  mkdir -p .github/workflows
  ```

- [ ] 创建测试工作流文件 (`.github/workflows/test.yml`)
  ```bash
  # 复制以下内容到 .github/workflows/test.yml
  # 完整内容见 ../02-improvement-plan.md 任务 2.1
  touch .github/workflows/test.yml
  ```

- [ ] 编辑文件并填入测试工作流内容
  - [ ] 配置 Go 版本矩阵
  - [ ] 配置缓存策略
  - [ ] 配置测试命令
  - [ ] 配置基准测试
  - [ ] 配置覆盖率上传

- [ ] 创建 linting 工作流文件 (`.github/workflows/lint.yml`)
  ```bash
  touch .github/workflows/lint.yml
  ```

- [ ] 编辑文件并填入 linting 工作流内容
  - [ ] 配置 golangci-lint
  - [ ] 配置 go vet
  - [ ] 配置 gofmt 检查

- [ ] 提交工作流文件
  ```bash
  git add .github/workflows/
  git commit -m "ci: add GitHub Actions workflows for testing and linting"
  git push origin main
  ```

- [ ] 验证 GitHub Actions 工作正常
  - [ ] 打开 GitHub 项目页面
  - [ ] 进入 Actions 标签页
  - [ ] 查看工作流是否正常运行
  - [ ] 确认所有 checks 通过 ✅

**✨ CI/CD 建立完成! 预计耗时: 2-3 小时**

---

### 创建 Makefile

- [ ] 在项目根目录创建 Makefile
  ```bash
  touch Makefile
  ```

- [ ] 编辑 Makefile 添加以下目标
  - [ ] test - 运行所有测试
  - [ ] benchmark - 运行基准测试
  - [ ] lint - 代码检查
  - [ ] format - 代码格式化
  - [ ] clean - 清理构建文件
  - [ ] install-tools - 安装开发工具
  - [ ] help - 显示帮助信息

- [ ] 验证 Makefile 工作正常
  ```bash
  make help
  make test
  make lint
  make benchmark
  ```

- [ ] 提交 Makefile
  ```bash
  git add Makefile
  git commit -m "chore: add Makefile for common development tasks"
  git push origin main
  ```

**✨ 开发工具完成! 预计耗时: 1 小时**

---

## 📚 文档完善 (第 3-4 周)

### 创建 changelog.md

- [ ] 在项目根目录创建文件
  ```bash
  touch changelog.md
  ```

- [ ] 编辑文件，添加以下部分
  - [ ] 项目简介说明
  - [ ] v2.0.0 版本信息
    - [ ] Added (新增功能)
    - [ ] Changed (变更)
    - [ ] Fixed (修复)
  - [ ] v1.0.2 版本信息
  - [ ] v1.0.1 版本信息
  - [ ] v1.0.0 版本信息

- [ ] 验证文件格式正确
  - [ ] 使用 Markdown 预览查看
  - [ ] 检查链接有效性

- [ ] 提交文件
  ```bash
  git add changelog.md
  git commit -m "docs: add changelog.md with version history"
  git push origin main
  ```

**✨ CHANGELOG 完成! 预计耗时: 1 小时**

---

### 创建 contributing.md

- [ ] 创建文件
  ```bash
  touch contributing.md
  ```

- [ ] 编辑文件，添加以下部分
  - [ ] 开发环境设置
  - [ ] 前置要求
  - [ ] 本地开发步骤
  - [ ] 代码质量标准
  - [ ] 必须通过的检查
  - [ ] 推荐实践
  - [ ] 提交流程
  - [ ] 提交信息规范
  - [ ] 许可证说明

- [ ] 验证文件内容
  - [ ] 所有命令可执行
  - [ ] 步骤逻辑清晰
  - [ ] 示例代码正确

- [ ] 提交文件
  ```bash
  git add contributing.md
  git commit -m "docs: add contributing.md with contribution guidelines"
  git push origin main
  ```

**✨ 贡献指南完成! 预计耗时: 1 小时**

---

### 创建 security.md

- [ ] 创建文件
  ```bash
  touch security.md
  ```

- [ ] 编辑文件，添加以下部分
  - [ ] 安全政策说明
  - [ ] 漏洞报告方式
  - [ ] 报告应包含的信息
  - [ ] 安全响应流程
  - [ ] 已知限制
  - [ ] 支持政策

- [ ] 验证文件内容
  - [ ] 联系方式准确
  - [ ] 流程清晰易懂
  - [ ] 时间承诺明确

- [ ] 提交文件
  ```bash
  git add security.md
  git commit -m "docs: add security.md with security policy"
  git push origin main
  ```

**✨ 安全政策完成! 预计耗时: 30 分钟**

---

### 创建 .editorconfig

- [ ] 创建文件
  ```bash
  touch .editorconfig
  ```

- [ ] 编辑文件，添加编辑器配置
  - [ ] 通用配置 (所有文件)
  - [ ] Go 文件配置
  - [ ] Markdown 文件配置
  - [ ] JSON 文件配置
  - [ ] YAML 文件配置

- [ ] 验证配置正确
  ```bash
  # 在你的编辑器中打开文件
  # 检查缩进、行尾等是否按配置应用
  ```

- [ ] 提交文件
  ```bash
  git add .editorconfig
  git commit -m "chore: add .editorconfig for editor configuration"
  git push origin main
  ```

**✨ 编辑器配置完成! 预计耗时: 15 分钟**

---

## 📖 API 文档 (第 4 周)

### 扩展 Go Doc 注释

- [ ] 打开 `random.go` 文件
- [ ] 为每个公开函数添加文档
  - [ ] GetRandomFingerprint()
  - [ ] GetRandomFingerprintWithOS()
  - [ ] GetRandomFingerprintByBrowser()
  - [ ] GetRandomFingerprintByBrowserWithOS()

- [ ] 打开 `ja3.go` 文件
- [ ] 为每个公开函数添加文档
  - [ ] ComputeJA3FromSpec()
  - [ ] ComputeJA3ByProfileName()
  - [ ] FindProfileByJA3()

- [ ] 打开 `ja4.go` 文件
- [ ] 为每个公开函数添加文档
  - [ ] ComputeJA4ByProfileName()
  - [ ] ComputeJA4FromProfile()

- [ ] 打开 `defense.go` 文件
- [ ] 为每个公开类和方法添加文档
  - [ ] AnomalyDetector
  - [ ] ContradictionDetector
  - [ ] PassiveRecognizer

- [ ] 打开 `headers.go` 文件
- [ ] 为 HTTPHeaders 类型添加文档
- [ ] 为公开方法添加文档

- [ ] 验证文档质量
  ```bash
  go doc ./...
  ```

- [ ] 生成 HTML 文档查看效果
  ```bash
  go doc -html ./... > docs.html
  ```

- [ ] 提交修改
  ```bash
  git add -A
  git commit -m "docs: expand API documentation with Go Doc comments"
  git push origin main
  ```

**✨ API 文档完成! 预计耗时: 2-3 小时**

---

### 创建最佳实践指南

- [ ] 创建文档目录
  ```bash
  mkdir -p docs
  ```

- [ ] 创建文件
  ```bash
  touch docs/best-practices.md
  ```

- [ ] 编辑文件，添加以下部分
  - [ ] 性能优化
    - [ ] 利用零分配函数
    - [ ] 缓存指纹配置
  - [ ] 指纹选择
    - [ ] 按浏览器指定
  - [ ] 安全使用
    - [ ] 验证异常指纹
  - [ ] 噪声注入
    - [ ] 主动防护示例
  - [ ] 指纹识别
    - [ ] 被动识别示例
  - [ ] 性能基准
  - [ ] 常见陷阱

- [ ] 验证所有代码示例
  ```bash
  # 运行示例代码进行测试
  ```

- [ ] 提交文件
  ```bash
  git add docs/
  git commit -m "docs: add best-practices.md with usage guidelines"
  git push origin main
  ```

**✨ 最佳实践指南完成! 预计耗时: 2 小时**

---

## 🎉 最终验证 (完成后)

### 全面检查

- [ ] 代码格式化
  ```bash
  gofmt -l .  # 应输出: 空
  ```

- [ ] 静态分析
  ```bash
  go vet ./...  # 应输出: 无错误
  ```

- [ ] 所有测试通过
  ```bash
  go test ./... -v  # 应全部 PASS
  ```

- [ ] 性能基准正常
  ```bash
  go test ./test -bench=. -benchmem
  ```

- [ ] 代码文档完整
  ```bash
  go doc ./... | wc -l  # 应有大量输出
  ```

- [ ] Git 历史清晰
  ```bash
  git log --oneline | head -10
  ```

- [ ] GitHub Pages (可选)
  - [ ] 配置 GitHub Pages
  - [ ] 生成 API 文档

### 发布版本

- [ ] 创建版本标签
  ```bash
  git tag -a v2.0.1 -m "Version 2.0.1 - Bug fixes and CI/CD setup"
  git push origin v2.0.1
  ```

- [ ] 在 GitHub 上创建 Release
  - [ ] 标题: v2.0.1 - Bug Fixes and CI/CD Setup
  - [ ] 描述: 列出主要变更
  - [ ] 发布

- [ ] 验证 Actions 工作流
  - [ ] GitHub Actions 自动运行
  - [ ] 所有 checks 通过 ✅

**✨ 项目发布完成!**

---

## 📊 进度追踪

### 第 1 周进度
- [ ] Day 1: 代码格式化 (✅ 5 min)
- [ ] Day 2-3: vet 修复 (⏳ 3-4 hours)
- [ ] Day 4: go.sum 提交 (✅ 10 min)
- [ ] Day 5-7: 代码审查和测试 (2 hours)

**第 1 周总计: 4-5 小时**

### 第 2 周进度
- [ ] Day 1-3: GitHub Actions (2-3 hours)
- [ ] Day 4-5: Makefile (1 hour)
- [ ] Day 5-7: 测试和验证 (1 hour)

**第 2 周总计: 4-5 小时**

### 第 3 周进度
- [ ] Day 1-2: CHANGELOG (1 hour)
- [ ] Day 2-3: CONTRIBUTING (1 hour)
- [ ] Day 3-4: SECURITY (30 min)
- [ ] Day 4-5: .editorconfig (15 min)
- [ ] Day 5-7: 审查和提交 (1.5 hours)

**第 3 周总计: 4-5 小时**

### 第 4 周进度
- [ ] Day 1-3: API 文档 (2-3 hours)
- [ ] Day 3-5: 最佳实践 (2 hours)
- [ ] Day 5-7: 最终验证和发布 (1 hour)

**第 4 周总计: 4-5 小时**

---

## 💡 工作建议

### 时间管理
- 🕐 每天 1-2 小时连续工作效率最高
- 📅 避免间断超过 3 天，保持代码连贯性
- ⏰ 大任务(vet 修复)分块完成

### 协作建议
- 👥 如果团队工作，分配任务避免冲突
- 📝 保持清晰的提交信息
- 🔄 定期 push，避免本地变更过多

### 质量保证
- ✅ 每个任务完成后立即运行测试
- 🔍 修改前后对比验证
- 📊 记录每个任务的完成时间

---

## 🆘 常见问题

### Q1: 修复 vet 警告时找不到所有文件?
**A**: 运行 `go vet ./...` 得到完整列表，然后逐文件处理

### Q2: GitHub Actions 工作流如何本地测试?
**A**: 使用 `act` 工具本地模拟运行，或推送到 GitHub 进行在线测试

### Q3: 如何快速定位需要修改的结构体初始化?
**A**: 使用编辑器搜索功能，搜索 `Extension{` 并逐个替换为带字段名的格式

### Q4: 提交信息有格式要求吗?
**A**: 建议遵循 `<type>(<scope>): <subject>` 格式，详见 contributing.md

### Q5: 修改过程中需要创建分支吗?
**A**: 建议创建 `develop` 分支进行所有修改，完成后再合并到 `main`

---

## 📞 相关资源

- 📄 ../02-improvement-plan.md - 详细改进计划
- 📄 ../../3-references/00-quick-reference.md - 快速参考指南
- 📄 ../../readme.md - 文档导航首页

---

## ✨ 最后的话

感谢你的耐心阅读！这份检查清单将帮助你有条不紊地完成所有改进。

**预期结果**: 完成所有项目后，你将获得：
✅ 无代码警告  
✅ 自动化 CI/CD  
✅ 完整的文档  
✅ 生产级项目标准  

**预计用时**: 16-20 小时  
**难度**: 中等  
**收益**: 显著 🚀

**现在就开始吧!** 🎯

