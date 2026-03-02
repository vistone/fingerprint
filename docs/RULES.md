# 📋 文档组织规则 (Documentation Organization Rules)

<!-- markdownlint-disable MD022 MD031 MD032 MD040 -->

**版本**: 1.0  
**生效日期**: 2026-02-28  
**维护者**: vistone  
**更新周期**: 按需

---

## 🎯 核心原则

### 1. 唯一真理源 (Single Source of Truth)
- ✅ **代码是唯一真理** - 文档必须与代码保持一致
- ✅ **代码优先** - 代码变更时，文档必须同步更新
- ✅ **可验证性** - 文档中的所有示例代码必须可测试

### 2. 严格的目录结构
```
docs/
├── 1-analysis/          (项目分析文档 - 由分析工具生成或维护)
├── 2-guides/            (用户指南 - 针对特定用途的教程)
├── 3-references/        (参考文档 - API文档、配置说明)
├── 4-templates/         (模板文件 - CI/CD、问题模板等)
├── 5-process/           (流程文档 - 开发流程、发布流程)
├── README.md            (文档导航首页)
└── RULES.md             (本文件 - 文档规则)
```

### 3. 文件命名规则
```
✅ 允许的格式:
- 英文名称，单词用连字符分隔: my-document.md
- 数字前缀用于排序: 01-getting-started.md
- 版本号: my-document-v1.0.md

❌ 不允许:
- 中文文件名: 快速开始.md
- 下划线: my_document.md
- 大写字母: MyDocument.md
- 特殊字符: my@document.md
```

### 4. 文档生命周期
```
创建 → 审查 → 验证 → 发布 → 维护 → 归档
  ↑                                    ↓
  └────────────────── 更新循环 ────────┘
```

---

## 📁 目录结构详解

### 1-analysis/ (分析文档)
**目的**: 存储项目分析结果、质量评估报告

**文件规则**:
- 由自动化工具生成或代码审查生成
- 包含指标数据、性能基准、质量评分
- 自动生成的文件需要标注生成日期和工具名

**包含的文件**:
```
1-analysis/
├── project-health.md          # 项目健康报告
├── code-quality.md            # 代码质量评估
├── performance-benchmarks.md  # 性能基准数据
├── dependency-analysis.md     # 依赖关系分析
├── security-audit.md          # 安全审计报告
└── improvement-recommendations.md  # 改进建议
```

**示例文件头**:
```markdown
# 项目健康报告

**生成日期**: 2026-02-28  
**分析工具**: go vet, gofmt, go test -bench  
**版本**: v2.0.0  
**下次更新**: 2026-03-31  

[内容...]
```

---

### 2-guides/ (用户指南)
**目的**: 为不同用户提供步骤化的操作指南

**文件规则**:
- 针对特定任务或用户场景
- 包含详细的步骤和示例
- 必须包含"前置条件"和"预期结果"

**包含的文件**:
```
2-guides/
├── 01-quick-start.md              # 快速开始指南
├── 02-installation.md              # 安装指南
├── 03-basic-usage.md               # 基础用法
├── 04-advanced-usage.md            # 高级用法
├── 05-deployment.md                # 部署指南
├── 06-troubleshooting.md           # 故障排除
└── 07-best-practices.md            # 最佳实践

developer/
├── 01-setup-dev-environment.md     # 开发环境配置
├── 02-build-and-test.md            # 编译和测试
├── 03-code-style.md                # 代码风格指南
├── 04-testing-guide.md             # 测试指南
└── 05-debugging.md                 # 调试指南

contributor/
├── 01-contribution-guide.md        # 贡献指南
├── 02-pull-request-process.md      # PR 流程
├── 03-code-review.md               # 代码审查
└── 04-commit-message.md            # 提交信息规范
```

**示例文件头**:
```markdown
# 快速开始指南

## 前置条件
- Go 1.25.4 或更高版本
- Git 已安装
- 基本的 Go 编程知识

## 预期结果
完成本指南后，你将能够：
- 安装项目
- 运行第一个程序
- 理解基本概念

## 步骤
1. [步骤...]
2. [步骤...]

## 常见问题
[FAQ...]
```

---

### 3-references/ (参考文档)
**目的**: 存储 API 文档、配置说明、数据结构定义等

**文件规则**:
- 必须与代码同步更新
- API 文档自动生成 (go doc)
- 包含完整的参数说明和返回值

**包含的文件**:
```
3-references/
├── api/
│   ├── public-api.md           # 公开 API 列表
│   ├── types.md                # 类型定义
│   ├── functions.md            # 函数文档
│   └── interfaces.md           # 接口文档
├── cli/
│   └── commands.md             # 命令行参数
├── config/
│   ├── configuration.md        # 配置说明
│   ├── environment-variables.md # 环境变量
│   └── config-examples.md      # 配置示例
└── data-structures/
    ├── fingerprint-types.md    # 指纹类型
    ├── http-headers.md         # HTTP Headers
    └── browser-profiles.md     # 浏览器配置
```

**示例文件头**:
```markdown
# API 参考 - 函数文档

**最后更新**: 2026-02-28  
**对应代码版本**: v2.0.0  
**自动生成**: 否 (手工维护)  

## GetRandomFingerprint

### 签名
\`\`\`go
func GetRandomFingerprint() (*FingerprintResult, error)
\`\`\`

### 参数
无

### 返回值
- `*FingerprintResult`: 指纹结果
- `error`: 错误信息

[完整文档...]
```

---

### 4-templates/ (模板文件)
**目的**: 存储 CI/CD、Issue、PR 等模板

**文件规则**:
- 不需要中文，保持英文
- 版本号标注
- 包含填写说明

**包含的文件**:
```
4-templates/
├── github/
│   ├── issue-template.md
│   ├── pull-request-template.md
│   ├── bug-report.md
│   ├── feature-request.md
│   └── security-issue.md
├── ci-cd/
│   ├── github-actions-test.yml
│   ├── github-actions-lint.yml
│   └── github-actions-release.yml
├── code/
│   ├── new-file-template.go
│   └── function-template.go
└── project/
    ├── changelog-template.md
    └── release-notes-template.md
```

---

### 5-process/ (流程文档)
**目的**: 记录开发流程、发布流程、管理流程

**文件规则**:
- 必须明确责任人和时间线
- 包含决策树和检查清单
- 定期审查和更新

**包含的文件**:
```
5-process/
├── development/
│   ├── development-workflow.md   # 开发工作流
│   ├── feature-development.md    # 功能开发流程
│   ├── bug-fix-process.md        # Bug 修复流程
│   └── code-review-process.md    # 代码审查流程
├── release/
│   ├── release-process.md        # 发布流程
│   ├── version-strategy.md       # 版本策略
│   ├── changelog-management.md   # 变更日志管理
│   └── deployment-process.md     # 部署流程
├── maintenance/
│   ├── dependency-management.md  # 依赖管理
│   ├── security-updates.md       # 安全更新
│   ├── performance-monitoring.md # 性能监控
│   └── documentation-updates.md  # 文档更新
└── governance/
    ├── decision-making.md        # 决策流程
    ├── community-management.md   # 社区管理
    └── roadmap.md                # 项目路线图
```

---

## 📏 文档内容规范

### 每份文档必须包含

#### Header (文件头)
```markdown
# 标题 (一级标题)

**文档类型**: [分析|指南|参考|流程]  
**版本**: 1.0  
**最后更新**: 2026-02-28  
**维护者**: [name]  
**对应代码版本**: v2.0.0  [仅参考文档需要]  
**下次审查**: 2026-03-31  
```

#### Content (内容)
```markdown
## 概述 (Optional)
简要说明文档目的

## 前置条件 (如需要)
- 条件 1
- 条件 2

## 内容主体
[按需分节]

## 常见问题 (如需要)
### Q: 问题?
A: 答案

## 相关文档
- [改进计划](./2-guides/02-improvement-plan.md)
- [开发规范](./5-process/development/README.md)

## 变更历史
- v1.0 (2026-02-28): 初始版本
```

### 文件链接规则
```markdown
✅ 推荐 (相对路径):
[改进计划](./2-guides/02-improvement-plan.md)
[快速参考](./3-references/00-quick-reference.md)

❌ 避免 (绝对路径):
文档: /media/stone/data/dev/docs/guides/02-installation.md
```

---

## 🔄 文档与代码对齐流程

### 1. 代码变更时
```
代码变更 (Commit)
    ↓
检查是否涉及公开 API
    ↓ 是
    需要更新对应的参考文档
    ↓
检查是否改变使用方式
    ↓ 是
    需要更新对应的指南
    ↓
在同一 PR 中提交代码和文档更新
    ↓
代码审查时同时审查文档
```

### 2. 文档变更时
```
文档变更 (Commit)
    ↓
检查是否与当前代码一致
    ↓ 否
    更新代码或修正文档
    ↓
验证所有示例代码可执行
    ↓
更新"最后更新日期"
```

### 3. 新版本发布时
```
版本标签创建 (v2.0.1)
    ↓
审查所有文档
    ↓
更新对应代码版本号
    ↓
归档上个版本的文档 (可选)
    ↓
发布版本
```

---

## 📊 文档维护清单

### 每周
- [ ] 检查是否有更新的 PR 需要文档同步
- [ ] 验证文档中的代码示例仍然有效

### 每月
- [ ] 审查所有文档的"最后更新"日期
- [ ] 检查是否有过期的文档需要更新

### 每季度
- [ ] 全面审查文档组织结构
- [ ] 评估是否需要新增或移除文档
- [ ] 更新依赖关系文档

### 每年
- [ ] 大版本更新时的文档重构
- [ ] 文档风格一致性检查
- [ ] 归档旧版本文档

---

## ✅ 文档质量检查清单

### 提交任何文档前，检查以下事项

- [ ] 文件名符合规范 (英文、连字符分隔)
- [ ] 文件放在正确的目录
- [ ] 包含完整的文件头 (标题、版本、日期、维护者)
- [ ] 所有示例代码都经过测试
- [ ] 内部链接使用相对路径
- [ ] 没有过时的信息
- [ ] 与当前代码版本一致
- [ ] 包含相关文档的链接

---

## 🚫 常见错误

### ❌ 不要做

1. **乱写文档**
   - ❌ 在根目录创建文档
   - ✅ 放在 `docs/` 的对应子目录

2. **不与代码同步**
   - ❌ 文档中的示例代码无法运行
   - ✅ 所有示例都经过测试

3. **混用中英文**
   - ❌ 文件名: `快速开始.md`
   - ✅ 文件名: `quick-start.md`

4. **不使用相对路径**
    - ❌ `文档: /media/stone/data/dev/docs/ref.md`
    - ✅ `[文档](./3-references/00-quick-reference.md)`

5. **不更新版本信息**
   - ❌ 文档说的是 v1.0，但代码已是 v2.0
   - ✅ 版本号与代码同步更新

6. **孤立的文档**
   - ❌ 文档没有任何链接指向它
   - ✅ 在导航或索引文件中引用

---

## 🔗 文档发现机制

### 导航首页
`docs/README.md` - 所有文档的入口点

### 自动生成的索引
定期扫描 `docs/` 目录生成索引 (可选)

### 代码中的文档引用
在代码注释中引用相关文档:
```go
// 详细使用说明见: docs/2-guides/03-basic-usage.md
func GetRandomFingerprint() (*FingerprintResult, error) {
    // ...
}
```

---

## 📝 示例文档结构

### 参考文档示例
```markdown
# GetRandomFingerprint API

**最后更新**: 2026-02-28  
**对应代码版本**: v2.0.0  

## 概述
随机获取一个指纹配置...

## 函数签名
\`\`\`go
func GetRandomFingerprint() (*FingerprintResult, error)
\`\`\`

## 参数
无

## 返回值
...

## 示例
\`\`\`go
result, err := fingerprint.GetRandomFingerprint()
\`\`\`

## 相关文档
- 改进计划: ../2-guides/02-improvement-plan.md
- 快速参考: ../3-references/00-quick-reference.md
```

### 指南文档示例
```markdown
# 快速开始

**版本**: 1.0  
**最后更新**: 2026-02-28  

## 前置条件
- Go 1.25.4+
- Git

## 步骤

### 1. 安装
\`\`\`bash
go get github.com/vistone/fingerprint
\`\`\`

### 2. 第一个程序
\`\`\`go
package main
import "github.com/vistone/fingerprint"

func main() {
    fp, _ := fingerprint.GetRandomFingerprint()
    println(fp.UserAgent)
}
\`\`\`

## 下一步
改进计划: ./02-improvement-plan.md
```

---

## 🎯 遵循规则的好处

✅ **一致性** - 所有文档格式统一  
✅ **可维护性** - 易于查找和更新  
✅ **可发现性** - 用户能快速找到需要的文档  
✅ **质量保证** - 文档与代码同步更新  
✅ **扩展性** - 新文档可以无缝集成  

---

## 📞 文档维护联系

**文档负责人**: [待指定]  
**审查周期**: 按需 / 每月一次  
**更新方式**: 通过 PR 提交  

---

**最后更新**: 2026-02-28  
**规则版本**: 1.0  
**状态**: ✅ 生效中

