# 🚫 文档放置规范（强制执行）

**生效日期**: 2026-02-28  
**执行等级**: 强制执行  
**违反后果**: PR 拒绝  

---

## ⚠️ 重要警告

**从本日期起，任何违反此规范的文档提交将被拒绝！**

违反文档放置规范的 PR 将被自动拒绝，不进行代码审查。

---

## 📍 文档放置规则

### 根目录 (/) - 严格限制

#### ✅ 允许的文件

```
✅ 允许
/README.md              # 主项目文档（1 个）
/go.mod                 # Go 模块定义
/go.sum                 # Go 依赖锁定
/Makefile               # 构建工具
/LICENSE                # 许可证
/.gitignore             # Git 忽略规则
/.editorconfig          # 编辑器配置
/.golangci.yml          # 代码检查配置
/.github/               # GitHub 配置
/*.go                   # Go 源代码
/examples/              # 示例代码
/internal/              # 内部包
/profiles/              # 指纹配置包
/test/                  # 测试文件
/docs/                  # 文档目录
```

#### ❌ 禁止的文件

```
❌ 禁止 - 这些必须放在 docs/
CHANGELOG.md            ❌ 应在 docs/CHANGELOG.md
CONTRIBUTING.md         ❌ 应在 docs/CONTRIBUTING.md
SECURITY.md             ❌ 应在 docs/SECURITY.md
开发规范.md             ❌ 应在 docs/5-process/
快速开始.md             ❌ 应在 docs/2-guides/
API文档.md              ❌ 应在 docs/3-references/
任何其他 *.md           ❌ 应在 docs/
任何报告.md             ❌ 应在 docs/
任何指南.md             ❌ 应在 docs/
```

### docs 目录 (docs/) - 5 层分类

所有文档必须严格按照 5 层分类系统放置：

```
docs/
│
├── 根层文档 (6 个)
│   ├── README.md                                ✅ 文档导航
│   ├── RULES.md                                 ✅ 文档组织规则
│   ├── CHANGELOG.md                             ✅ 版本历史
│   ├── CONTRIBUTING.md                          ✅ 贡献指南
│   ├── SECURITY.md                              ✅ 安全政策
│   └── DOCUMENTATION-ORGANIZATION-GUIDE.md      ✅ 详细指南
│
├── 1-analysis/                                  ✅ 分析报告
│   ├── 项目分析文档
│   ├── 健康检查报告
│   └── ...
│
├── 2-guides/                                    ✅ 指南文档
│   ├── 用户指南
│   ├── 快速开始
│   ├── 开发者/
│   │   ├── 开发检查清单
│   │   ├── 错误修复总结
│   │   └── ...
│   ├── 贡献者/
│   └── ...
│
├── 3-references/                                ✅ 参考文档
│   ├── API 文档
│   ├── 配置参考
│   ├── 数据结构
│   └── ...
│
├── 4-templates/                                 ✅ 模板库
│   ├── CI/CD 模板
│   ├── 代码模板
│   ├── GitHub 模板
│   └── 项目模板
│
└── 5-process/                                   ✅ 流程规范
    ├── development/
    │   ├── Go 开发规范
    │   ├── 项目规范
    │   ├── 注释模板
    │   └── 快速参考
    ├── governance/
    ├── maintenance/
    └── release/
```

---

## 🔍 检查清单

提交文档时必须检查：

### 提交前 ✅

- [ ] 此文档是否应该在根目录？
  - ❌ 如果答案是"不确定"或"也许"，那么它不应该在根目录
  - ✅ 只有 README.md 可以在根目录

- [ ] 此文档属于哪个分类？
  - [ ] 分析报告 → docs/1-analysis/
  - [ ] 指南文档 → docs/2-guides/
  - [ ] 参考文档 → docs/3-references/
  - [ ] 模板文件 → docs/4-templates/
  - [ ] 流程规范 → docs/5-process/
  - [ ] 顶层文档 → docs/

- [ ] 是否存在重复的文档？
  - [ ] 检查 docs/ 中是否已存在相似文档
  - [ ] 删除旧版本文档

- [ ] 文档是否有明确的用途？
  - [ ] 不清楚的文档不应该被添加
  - [ ] 等等改进后再添加

### 代码审查时 🔎

审查者需要检查：

```
PR 审查清单：
□ 是否有新的 *.md 文件放在根目录？
  → 是：要求移除或移动到 docs/
  
□ 新增文档是否放在 docs/ 中？
  → 不是：拒绝 PR
  
□ 文档是否按照 5 层分类放置？
  → 不是：要求重新组织
  
□ 是否有重复的文档？
  → 是：要求删除旧版本
  
□ 文档放置是否合理？
  → 不合理：要求调整分类
```

---

## 📝 常见场景

### 场景 1: 添加新的开发指南

❌ **错误做法**:
```
创建文件: /快速入门.md
```

✅ **正确做法**:
```
创建文件: /docs/2-guides/quick-start.md
或
创建文件: /docs/2-guides/developer/quick-start.md
```

### 场景 2: 添加 API 参考文档

❌ **错误做法**:
```
创建文件: /API.md
```

✅ **正确做法**:
```
创建文件: /docs/3-references/api/README.md
```

### 场景 3: 添加项目模板

❌ **错误做法**:
```
创建文件: /PR-template.md
```

✅ **正确做法**:
```
创建文件: /docs/4-templates/github/pull-request-template.md
或使用: /.github/pull_request_template.md
```

### 场景 4: 添加工作流程文档

❌ **错误做法**:
```
创建文件: /开发流程.md
```

✅ **正确做法**:
```
创建文件: /docs/5-process/development/workflow.md
```

### 场景 5: 添加维护指南

❌ **错误做法**:
```
创建文件: /维护说明.md
```

✅ **正确做法**:
```
创建文件: /docs/5-process/maintenance/guide.md
```

---

## 🚀 快速参考

### 我应该把这个文档放在哪里？

| 文档类型 | 目的 | 位置 |
|---------|------|------|
| 项目主页 | 介绍项目 | `/README.md` |
| 版本历史 | 记录变更 | `docs/CHANGELOG.md` |
| 贡献指南 | 指导贡献者 | `docs/CONTRIBUTING.md` |
| 安全政策 | 报告漏洞 | `docs/SECURITY.md` |
| 项目分析 | 分析项目 | `docs/1-analysis/` |
| 使用指南 | 指导用户 | `docs/2-guides/` |
| API 文档 | API 参考 | `docs/3-references/api/` |
| 代码模板 | 代码示例 | `docs/4-templates/code/` |
| 开发规范 | 开发标准 | `docs/5-process/development/` |

---

## ⚡ 自动化检查（建议）

项目可以使用以下脚本自动检查：

```bash
#!/bin/bash
# check-docs-placement.sh - 检查文档放置

echo "检查根目录文档..."
bad_docs=$(find . -maxdepth 1 -name "*.md" ! -name "README.md" -type f)

if [ -n "$bad_docs" ]; then
  echo "❌ 错误：根目录有不允许的 markdown 文件："
  echo "$bad_docs"
  echo "这些文件应该放在 docs/ 目录中"
  exit 1
fi

echo "✅ 文档放置检查通过"
exit 0
```

---

## 📞 有问题？

**Q: 我的文档不确定放在哪里？**  
A: 放在 docs/ 目录中，再由维护者指导正确的分类位置。

**Q: 可以在根目录创建文件夹吗？**  
A: 不行。除了已经存在的文件夹（examples, internal, profiles, test, docs），不能创建新的根目录文件夹。

**Q: 可以创建新的分类吗？**  
A: 不行。必须严格遵循 5 层分类系统。

---

## 🎯 总结

### 记住这个简单规则：

```
📁 根目录
   ✅ README.md (仅此一个 .md 文件)
   ✅ 代码文件和标准配置
   
📁 docs/
   ✅ 所有其他文档
   ✅ 严格按 5 层分类
```

### 违反规则的后果：

```
❌ PR 拒绝
❌ 不进行代码审查
❌ 要求重新组织文档
✅ 再次提交前必须完全遵守规范
```

---

**这是强制性规范。不遵守的提交将被拒绝。** 🚫

**版本**: 1.0  
**生效日期**: 2026-02-28  
**维护者**: vistone
