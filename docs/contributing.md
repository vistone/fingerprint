# 贡献指南

感谢您对 Fingerprint 项目的贡献兴趣！本文档提供贡献指南和说明。

## 行为准则

参与此项目即表示您同意遵守我们的行为准则。请在所有互动中保持尊重和建设性。

## 准备工作

### 前置条件

- Go 1.24 或更高版本（我们测试 1.24 和 1.25）
- Git
- Make（可选，但推荐）

### 开发环境设置

1. **克隆仓库**
   ```bash
   git clone https://github.com/vistone/fingerprint.git
   cd fingerprint
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **安装开发工具**（可选）
   ```bash
   make install-tools
   ```

4. **验证设置**
   ```bash
   go test ./...
   ```

## 开发工作流

### 1. 创建功能分支

```bash
git checkout -b feature/your-feature-name
# 或 bug 修复
git checkout -b fix/your-bug-fix
```

分支命名约定：
- `feature/` 新功能
- `fix/` 错误修复
- `docs/` 文档更改
- `refactor/` 代码重构
- `perf/` 性能改进

### 2. 进行更改

遵循我们的开发标准：

1. **代码质量**
   - 运行 `make format` 格式化代码
   - 运行 `make lint` 检查代码质量
   - 运行 `make test` 验证测试通过
   - 确保 `go vet ./...` 通过

2. **文档**
   - 为公开 API 添加/更新注释
   - 使用您的更改更新 changelog.md
   - 更新相关文档文件

3. **测试**
   - 为新功能编写测试
   - 确保所有测试通过
   - 目标是 80%+ 代码覆盖率

### 3. 提交更改

遵循以下格式编写清晰的提交消息：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型**（必需）：
- `feat`: 新功能
- `fix`: 错误修复
- `docs`: 文档更改
- `style`: 代码风格更改
- `refactor`: 代码重构
- `perf`: 性能改进
- `test`: 添加或更新测试
- `chore`: 其他不影响代码的更改

**范围**（可选）：
- `profiles`: 浏览器配置变更
- `ja3`: JA3 指纹变更
- `ja4`: JA4 指纹变更
- `defense`: 异常/矛盾检测变更
- `api`: 公开 API 变更

**主题**（必需）：
- 使用命令式语气（"add" 而不是 "added"）
- 不要大写首字母
- 行末无句号
- 限制在 50 个字符以内

**示例**：
```
feat(profiles): 添加 Safari 18 浏览器配置
fix(ja3): 正确处理 GREASE 值
docs: 更新 API 文档
test(defense): 添加异常检测测试
```

### 4. 推送和创建拉取请求

```bash
git push origin feature/your-feature-name
```

然后在 GitHub 上创建拉取请求，包含：

1. **清晰的标题** 遵循提交消息格式
2. **描述** 说明：
   - 进行了哪些更改
   - 为什么进行更改
   - 任何相关问题（使用 `Closes #123`）
3. **截图/示例** 如适用
4. **检查清单** （PR 模板自动提供）

## 代码审查流程

1. **自动化检查**
   - GitHub Actions 将运行测试、检查和覆盖率
   - 所有检查必须通过

2. **手动审查**
   - 至少一名维护者将审查代码
   - 可能会请求更改或提出问题
   - 请对反馈做出响应

3. **批准和合并**
   - 批准后且检查通过，PR 将被合并
   - 维护者将处理合并

## 质量标准

### 代码质量要求

- ✅ 所有测试通过 (`go test ./...`)
- ✅ 没有 go vet 警告 (`go vet ./...`)
- ✅ 代码已格式化 (`gofmt -s -w .`)
- ✅ 检查通过 (`make lint`)
- ✅ 没有安全问题 (`gosec ./...`)
- ✅ 测试覆盖率维持或改进

### 文档要求

- ✅ 公开函数/类型有 godoc 注释
- ✅ 复杂逻辑有行内注释
- ✅ changelog.md 已更新
- ✅ readme.md 已更新（如需要）
- ✅ docs/examples 中的示例有效

### 性能要求

- ✅ 没有性能回归
- ✅ 基准测试通过 (`make benchmark`)
- ✅ 内存使用可接受
- ✅ 并发性能维持

## 测试指南

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行详细输出
go test ./... -v

# 运行特定测试
go test ./test -run TestSpecificName

# 运行并显示覆盖率
go test ./... -cover

# 运行基准测试
go test ./test -bench=. -benchmem
```

### 编写测试

1. **测试文件命名**: `*_test.go`
2. **测试函数命名**: `TestFunctionName`
3. **表驱动测试**: 用于多个场景
4. **测试组织**: 分组相关测试

示例:
```go
func TestGetRandomFingerprint(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"valid", false},
		{"no fingerprints", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetRandomFingerprint()
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error")
			}
			if result == nil && !tt.wantErr {
				t.Errorf("unexpected nil result")
			}
		})
	}
}
```

## 文档

### 文档类型

1. **API 文档**: 代码中的 Godoc 注释
2. **用户指南**: `docs/2-guides/`
3. **参考文档**: `docs/3-references/`
4. **开发指南**: `docs/5-process/development/`

### 编写文档

1. 使用清晰简洁的中文
2. 包含代码示例
3. 解释为什么，不仅仅是怎样
4. 与代码保持最新
5. 链接相关文档

## 报告问题

### 错误报告

包含：
- Go 版本
- 操作系统
- 复现步骤
- 预期行为
- 实际行为
- 代码示例（如适用）

### 功能请求

包含：
- 用例描述
- 为什么需要
- 建议的 API/接口
- 类似功能的示例

## 安全

### 报告安全问题

⚠️ **不要公开提交安全漏洞问题**

请私密报告安全问题至：security@example.com

包含：
- 漏洞类型
- 代码位置
- 潜在影响
- 建议修复（如有）

详见 [security.md](./security.md)

## 其他资源

- 📖 [开发指南](./5-process/development/readme.md)
- 🚀 [快速开始](./2-guides/)
- 📋 [开发检查清单](./2-guides/developer/)
- 🔍 [Go 开发规范](./5-process/development/)

## 有问题？

- 📧 邮箱: support@example.com
- 💬 讨论: GitHub Discussions
- 📝 文档: 见 docs/readme.md

---

## 许可证

对本项目的贡献即表示您同意您的贡献采用与本项目相同的许可证。

感谢您的贡献！🎉
