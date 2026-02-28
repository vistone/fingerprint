# 📋 文档整理清单与规范指南

**版本**: 1.0  
**日期**: 2026-02-28  
**状态**: ✅ 规范已制定

---

## 🎯 文档组织目标

✅ **杜绝根目录文件混乱** - 所有文档必须放在 `docs/` 目录  
✅ **严格的结构化组织** - 按照类型分类到 5 个主目录  
✅ **文档与代码对齐** - 文档必须与当前代码版本一致  
✅ **明确的治理规则** - 有清晰的维护和更新流程  

---

## 📁 正确的目录结构

```
项目根目录/
├── .git/
├── .github/
├── examples/
├── internal/
├── profiles/
├── test/
├── docs/                          ← 所有文档放这里
│   ├── README.md                  ← 文档导航首页
│   ├── RULES.md                   ← 文档规则(本文件)
│   ├── 00-documentation-index.md  ← 文档索引
│   │
│   ├── 1-analysis/                ← 分析报告
│   │   ├── 00-comprehensive-analysis.md
│   │   ├── 01-project-health-dashboard.md
│   │   ├── 02-code-quality.md
│   │   ├── 03-performance-benchmarks.md
│   │   ├── 04-dependency-analysis.md
│   │   ├── 05-security-audit.md
│   │   └── 06-improvement-recommendations.md
│   │
│   ├── 2-guides/                  ← 用户指南
│   │   ├── 01-improvement-plan.md
│   │   ├── 02-quick-reference.md
│   │   ├── 03-best-practices.md
│   │   ├── developer/             ← 开发者指南
│   │   │   └── 01-developer-checklist.md
│   │   └── contributor/           ← 贡献者指南
│   │       └── (贡献相关文档)
│   │
│   ├── 3-references/              ← 参考文档
│   │   ├── api/
│   │   │   ├── public-api.md
│   │   │   ├── types.md
│   │   │   └── functions.md
│   │   ├── config/
│   │   │   └── configuration.md
│   │   └── data-structures/
│   │       └── fingerprint-types.md
│   │
│   ├── 4-templates/               ← 模板文件
│   │   ├── github/
│   │   │   ├── issue-template.md
│   │   │   └── pull-request-template.md
│   │   ├── ci-cd/
│   │   │   └── github-actions.yml
│   │   └── project/
│   │       └── changelog-template.md
│   │
│   └── 5-process/                 ← 流程文档
│       ├── development/
│       │   └── development-workflow.md
│       ├── release/
│       │   └── release-process.md
│       ├── maintenance/
│       │   └── maintenance-workflow.md
│       └── governance/
│           └── decision-making.md
│
├── go.mod
├── go.sum
├── README.md                      ← 项目主 README(不在 docs)
├── LICENSE
└── (源代码文件)
```

---

## 🗂️ 文档归类规则

### 1-analysis/ (分析报告)
**何时创建**: 项目分析、质量评估时  
**谁来维护**: 项目维护者、技术负责人  
**更新频率**: 按需、每季度一次  
**版本同步**: ✅ 必须与代码版本对齐

**应包含的文档**:
```
✅ comprehensive-analysis.md        - 完整项目分析报告
✅ project-health-dashboard.md      - 项目健康指数
✅ code-quality.md                  - 代码质量评分
✅ performance-benchmarks.md        - 性能基准数据
✅ dependency-analysis.md           - 依赖关系分析
✅ security-audit.md                - 安全审计报告
✅ improvement-recommendations.md   - 改进建议
```

**文件命名规则**:
- 前缀数字: 00-09 (用于排序)
- 英文名称，连字符分隔
- 示例: `00-comprehensive-analysis.md`

**文件头要求**:
```markdown
# 标题

**生成日期**: 2026-02-28  
**分析工具**: go vet, gofmt, go test -bench  
**代码版本**: v2.0.0  
**更新周期**: 每季度 / 按需  

[内容...]
```

---

### 2-guides/ (用户指南)
**何时创建**: 用户需要学习如何使用项目时  
**谁来维护**: 开发者、文档维护者  
**更新频率**: 代码变更时  
**版本同步**: ✅ 必须与代码版本对齐

**根目录应包含的文档**:
```
✅ 01-improvement-plan.md           - 改进计划
✅ 02-quick-reference.md            - 快速参考
✅ 03-best-practices.md             - 最佳实践
✅ 04-installation.md               - 安装指南
✅ 05-getting-started.md            - 快速开始
```

**developer/ 子目录**:
```
✅ 01-developer-checklist.md        - 开发者检查清单
✅ 02-setup-environment.md          - 环境配置
✅ 03-build-and-test.md             - 构建和测试
✅ 04-code-style.md                 - 代码风格
✅ 05-testing-guide.md              - 测试指南
```

**contributor/ 子目录**:
```
✅ 01-contribution-guide.md         - 贡献指南
✅ 02-pull-request-process.md       - PR 流程
✅ 03-code-review.md                - 代码审查
✅ 04-commit-message.md             - 提交规范
```

**文件命名规则**:
- 数字前缀: 01-99 (用于排序)
- 英文名称，连字符分隔
- 示例: `02-quick-reference.md`

---

### 3-references/ (参考文档)
**何时创建**: 作为 API 或配置的完整参考  
**谁来维护**: 开发者、架构师  
**更新频率**: **代码变更时必须同步**  
**版本同步**: ✅ 必须与代码版本完全一致

**api/ 子目录**:
```
✅ public-api.md                    - 公开 API 列表
✅ types.md                         - 类型定义
✅ functions.md                     - 函数文档
✅ interfaces.md                    - 接口定义
```

**config/ 子目录**:
```
✅ configuration.md                 - 配置说明
✅ environment-variables.md         - 环境变量
✅ config-examples.md               - 配置示例
```

**data-structures/ 子目录**:
```
✅ fingerprint-types.md             - 指纹类型
✅ http-headers.md                  - HTTP Headers
✅ browser-profiles.md              - 浏览器配置
```

**重要**: 参考文档中所有示例代码必须：
- ✅ 经过测试且可以运行
- ✅ 与当前版本的 API 一致
- ✅ 包含完整的参数说明

---

### 4-templates/ (模板文件)
**何时创建**: 用于自动化流程或标准化工作  
**谁来维护**: 项目管理员、DevOps  
**更新频率**: 流程改变时  
**版本同步**: ⚠️ 可以独立于代码版本

**github/ 子目录**:
```
✅ issue-template.md                - Issue 模板
✅ pull-request-template.md         - PR 模板
✅ bug-report.md                    - Bug 报告模板
✅ feature-request.md               - 功能请求模板
✅ security-issue.md                - 安全问题模板
```

**ci-cd/ 子目录**:
```
✅ github-actions-test.yml          - 测试工作流
✅ github-actions-lint.yml          - Lint 工作流
✅ github-actions-release.yml       - 发布工作流
```

**code/ 子目录**:
```
✅ new-file-template.go             - 新文件模板
✅ function-template.go             - 函数模板
```

**project/ 子目录**:
```
✅ changelog-template.md            - 变更日志模板
✅ release-notes-template.md        - 发布说明模板
```

---

### 5-process/ (流程文档)
**何时创建**: 记录团队的工作流程和决策过程  
**谁来维护**: 项目维护者、流程所有者  
**更新频率**: 流程改变时  
**版本同步**: ⚠️ 可以独立于代码版本

**development/ 子目录**:
```
✅ development-workflow.md          - 开发工作流
✅ feature-development.md           - 功能开发流程
✅ bug-fix-process.md               - Bug 修复流程
✅ code-review-process.md           - 代码审查流程
```

**release/ 子目录**:
```
✅ release-process.md               - 发布流程
✅ version-strategy.md              - 版本策略
✅ changelog-management.md          - 变更管理
✅ deployment-process.md            - 部署流程
```

**maintenance/ 子目录**:
```
✅ dependency-management.md         - 依赖管理
✅ security-updates.md              - 安全更新
✅ performance-monitoring.md        - 性能监控
✅ documentation-updates.md         - 文档更新
```

**governance/ 子目录**:
```
✅ decision-making.md               - 决策流程
✅ community-management.md          - 社区管理
✅ roadmap.md                       - 项目路线图
```

---

## 📝 文档编写规范

### 必须的文件头

```markdown
# [文档标题]

**文档类型**: [analysis|guide|reference|template|process]  
**版本**: [1.0]  
**最后更新**: [2026-02-28]  
**维护者**: [name]  
**代码对应版本**: [v2.0.0] [仅参考文档需要]  
**下次审查**: [2026-03-31]  

## 概述
[文档目的和简要说明]

## [主要内容]

## 相关文档
- [文档名称](path/to/doc.md)

## 变更历史
- v1.0 (2026-02-28): 初始版本
```

### 文件名规范

✅ **允许的格式**:
```
my-document.md          - 英文、连字符分隔
01-getting-started.md   - 数字前缀用于排序
file-v2.0.md            - 可以包含版本号
```

❌ **禁止的格式**:
```
my_document.md          - 下划线
MyDocument.md           - 大写字母
my@document.md          - 特殊字符
快速开始.md             - 中文文件名
```

### 内部链接规范

✅ **推荐 (相对路径)**:
```markdown
[快速开始](./01-quick-start.md)
[API 文档](../3-references/api/public-api.md)
[项目 README](../../README.md)
```

❌ **禁止 (绝对路径)**:
```markdown
[文档](/media/stone/data/dev/docs/guides/01-quick-start.md)
[文档](https://github.com/.../docs/guides/01-quick-start.md)
```

---

## 🔄 文档与代码对齐流程

### 何时需要更新文档

```
代码变更
  ↓
检查是否涉及公开 API
  ↓ 是 → 更新 3-references/api/*.md
检查是否改变使用方式
  ↓ 是 → 更新 2-guides/*.md
检查是否影响配置
  ↓ 是 → 更新 3-references/config/*.md
检查是否影响流程
  ↓ 是 → 更新 5-process/*.md
  ↓
同一 PR 中提交代码和文档更新
```

### 代码审查时

- [ ] 代码变更是否有对应文档更新？
- [ ] 文档中的示例代码是否与新代码一致？
- [ ] 是否有过时的文档需要更新？
- [ ] 新增 API 是否有对应文档？

---

## ✅ 文档质量检查清单

提交任何文档前，检查以下事项：

### 文件管理
- [ ] 文件放在正确的目录 (不在根目录)
- [ ] 文件名符合规范 (英文、连字符分隔)
- [ ] 文件头完整 (标题、版本、日期)

### 内容质量
- [ ] 所有示例代码都经过测试
- [ ] 内部链接使用相对路径
- [ ] 没有过时的信息
- [ ] 与当前代码版本一致 [参考文档]

### 版本管理
- [ ] 更新了"最后更新"日期
- [ ] 更新了"版本"号 (如需要)
- [ ] 添加了"变更历史"条目

### 发现性
- [ ] 在导航文档中有引用
- [ ] 与相关文档有交叉链接
- [ ] 出现在 README.md 的导航中

---

## 📊 文档维护清单

### 每周
- [ ] 检查是否有新的 PR 需要文档同步
- [ ] 验证快速参考中的命令仍然有效

### 每月
- [ ] 审查所有文档的"最后更新"日期
- [ ] 检查是否有过期的文档

### 每季度
- [ ] 全面审查所有文档内容
- [ ] 评估是否需要重构或移除文档
- [ ] 更新分析报告

### 每年
- [ ] 大版本发布时的文档审查
- [ ] 文档风格一致性检查
- [ ] 归档旧版本文档

---

## 🚫 常见错误

### ❌ 不要在根目录创建文档

**错误**:
```
project-root/
├── 快速开始.md              ❌ 中文名
├── quick-start.md           ❌ 在根目录
├── README.md
└── docs/
```

**正确**:
```
project-root/
├── README.md
└── docs/
    └── 2-guides/
        └── 01-quick-start.md  ✅ 在正确目录
```

### ❌ 不要让文档与代码不同步

**错误**:
```markdown
# API 参考

func GetRandomFingerprint() (*FingerprintResult, error) {  // ❌ 旧 API
```

**正确**:
```markdown
# API 参考

func GetRandomFingerprint() (*FingerprintResult, error) {  // ✅ 与代码一致
```

### ❌ 不要使用绝对路径

**错误**:
```markdown
[参考](https://github.com/vistone/fingerprint/docs/3-references/api.md)
```

**正确**:
```markdown
[参考](../3-references/api/public-api.md)
```

### ❌ 不要让文档孤立

**错误**:
```
文档存在但无人知晓
- 没有在导航中引用
- 没有其他文档指向它
```

**正确**:
```markdown
# README.md
[API 文档](./3-references/api/public-api.md)

# public-api.md
[相关指南](../../2-guides/03-basic-usage.md)
```

---

## 🎯 本规范的效力

✅ **生效日期**: 2026-02-28  
✅ **适用范围**: 所有项目文档  
✅ **强制执行**: 是 (PR 审查时强制检查)  
✅ **修改流程**: 需要在 PR 中讨论和批准  

---

## 📞 文档管理联系

**文档负责人**: [待指定]  
**问题报告**: 通过 GitHub Issues  
**规范更新**: 通过 PR 到 docs/RULES.md  

---

**规范版本**: 1.0  
**最后更新**: 2026-02-28  
**状态**: ✅ 生效中

