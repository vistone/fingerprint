// Fingerprint 项目 Go 开发规范总结
//
// 快速参考指南 - 所有开发者必读
//
// 创建日期: 2026-02-28
// 版本: 1.0

package fingerprint

// ============================================================================
// 开发规范核心要点 - 快速检查清单
// ============================================================================
//
// 在提交代码前，必须检查以下项目:

// ============================================================================
// 1. 代码结构 (Code Structure)
// ============================================================================
//
// ✅ 必须:
//   - 每个包的第一个文件有包文档注释
//   - 导入按照标准库、第三方库、本项目的顺序分组
//   - 文件内按照 const → var → type → interface → func → method 排列
//   - 相同类型的定义使用分组 (const (...), var (...))
//
// ❌ 避免:
//   - 混乱的导入顺序
//   - 定义和使用的顺序错乱
//   - 单个文件超过 1000 行

// ============================================================================
// 2. 命名规范 (Naming Convention)
// ============================================================================
//
// ✅ 常量 (Constants):
//   - PascalCase: BrowserChrome, DefaultTimeout
//   - 公开常量大写开头，私有常量小写开头
//   - 使用完整单词，避免缩写
//
// ✅ 变量 (Variables):
//   - camelCase: localCounter, globalConfig
//   - 公开变量大写开头，私有变量小写开头
//   - 避免单字母变量（除了循环计数器 i, j, k）
//
// ✅ 函数 (Functions):
//   - PascalCase (公开): GetRandomFingerprint()
//   - camelCase (私有): randomChoice()
//   - 使用动词开头: Get, Create, Update, Delete, Check, Is, Has
//   - 返回 bool 的函数: IsValid(), HasError(), CanProcess()
//
// ✅ 接口 (Interfaces):
//   - 单方法接口用 er 后缀: Reader, Writer, Closer
//   - 多方法接口用名词形式: FingerprintProvider
//
// ✅ 方法接收者 (Receivers):
//   - 1-2 个字母缩写: (d *AnomalyDetector), (c *Cache)
//   - 不使用 self 或 this
//
// ❌ 避免:
//   - 下划线: my_variable, get_fingerprint
//   - 缩写混乱: gfp, ud, chd
//   - 全大写: MY_CONSTANT (除非特殊情况)
//   - 含混两可的名称: data, info, thing

// ============================================================================
// 3. 注释规范 (Comment Convention)
// ============================================================================
//
// ✅ 公开 API 必须有 godoc 注释:
/*
// GetRandomFingerprint 获取随机浏览器指纹
//
// 详细描述（可选）
//
// 返回值:
//   - *FingerprintResult: 指纹结果
//   - error: 错误信息
//
// 示例:
//   result, err := GetRandomFingerprint()
//
func GetRandomFingerprint() (*FingerprintResult, error) {
}
*/
//
// ✅ 所有公开的 type, const, var 都必须有注释
// ✅ 复杂逻辑必须有行内注释解释为什么这样做
// ✅ 使用标准的 TODO/FIXME 格式: // TODO(name): description
//
// ❌ 避免:
//   - 无意义的注释: // i++  or  // 设置值
//   - 过时的注释
//   - 注释代码片段（应该删除）

// ============================================================================
// 4. 错误处理 (Error Handling)
// ============================================================================
//
// ✅ 必须检查所有可能返回错误的调用:
//   result, err := GetRandomFingerprint()
//   if err != nil {
//       return fmt.Errorf("failed: %w", err)
//   }
//
// ✅ 使用 fmt.Errorf 包装错误，使用 %w 格式动词
// ✅ 定义哨兵错误并使用 errors.Is() 检查
// ✅ 使用 defer 确保资源释放
//
// ❌ 避免:
//   - 忽略错误: _, _ := function()
//   - 创建通用错误: errors.New("error")
//   - 忘记检查 nil

// ============================================================================
// 5. 性能优化 (Performance)
// ============================================================================
//
// ✅ 检查项:
//   - 已知容量时预分配切片: make([]T, 0, capacity)
//   - 使用 strings.Builder 进行字符串拼接
//   - 避免频繁的内存分配
//   - 使用适当的数据结构（map vs slice）
//   - 缓存重复计算结果
//
// ✅ 关键函数的性能目标:
//   - GetRandomFingerprint: < 10 微秒
//   - GenerateHeaders: < 2 微秒
//   - ComputeJA3: < 1 微秒
//
// ❌ 避免:
//   - 循环中的 string + 拼接
//   - 创建不必要的中间对象
//   - 过度的 goroutine 创建

// ============================================================================
// 6. 指纹特定规范 (Fingerprint-Specific)
// ============================================================================
//
// ✅ 指纹键值格式: browser_version (e.g., chrome_133, firefox_135)
// ✅ 所有指纹必须在 MappedTLSClients 中注册
// ✅ User-Agent 必须与浏览器版本和操作系统一致
// ✅ Headers 必须与指纹和 User-Agent 匹配
// ✅ 异常检测必须覆盖所有已知的自动化工具
//
// ❌ 避免:
//   - 使用已弃用的指纹名称
//   - 不一致的版本号
//   - 遗漏的指纹注册

// ============================================================================
// 7. 测试规范 (Testing)
// ============================================================================
//
// ✅ 测试函数命名: TestFunctionName, BenchmarkFunctionName
// ✅ 使用 table-driven 测试多个场景
// ✅ 测试正常情况和错误情况
// ✅ 性能关键函数需要 benchmark
// ✅ 覆盖率目标 > 80%
//
// ❌ 避免:
//   - 只测试正常情况
//   - 使用相对导入
//   - 硬编码的测试数据

// ============================================================================
// 8. 代码审查检查清单 (Code Review Checklist)
// ============================================================================
//
// 提交 PR 前，确保通过以下检查:
//
// 格式和风格:
//   ✅ 运行了 gofmt 格式化
//   ✅ 运行了 go vet 检查
//   ✅ 没有 golint 警告
//   ✅ 代码缩进正确（tab，不是空格）
//
// 命名和注释:
//   ✅ 所有公开 API 有 godoc 注释
//   ✅ 复杂逻辑有行内注释
//   ✅ 命名遵循规范（大小写、完整单词）
//   ✅ 没有拼写错误
//
// 逻辑和正确性:
//   ✅ 所有错误都被检查
//   ✅ 使用卫语句避免深度嵌套
//   ✅ 没有明显的 bug
//   ✅ 算法逻辑正确
//
// 性能:
//   ✅ 没有明显的性能问题
//   ✅ 大对象使用指针传递
//   ✅ 避免循环中的频繁分配
//   ✅ 性能关键路径已优化
//
// 测试:
//   ✅ 新代码有对应的测试
//   ✅ 测试覆盖主要场景和错误情况
//   ✅ 所有测试都通过
//   ✅ 性能测试包含 benchmark
//
// 文档:
//   ✅ 更新了相关文档
//   ✅ 提交信息清晰完整
//   ✅ 包含了示例代码

// ============================================================================
// 9. 快速参考：常见模式
// ============================================================================
//
// 模式 1: 错误检查和返回
/*
result, err := SomeFunction()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
// 使用 result
*/
//
// 模式 2: 资源清理
/*
file, err := os.Open("filename")
if err != nil {
    return err
}
defer file.Close()
// 使用 file
*/
//
// 模式 3: 类型断言安全检查
/*
if value, ok := data.(string); ok {
    // 使用 value
} else {
    // 处理错误情况
}
*/
//
// 模式 4: 卫语句
/*
func Process(data []byte) error {
    if data == nil {
        return ErrInvalidInput
    }
    if len(data) == 0 {
        return ErrEmptyInput
    }
    // 主要逻辑
    return nil
}
*/
//
// 模式 5: Table-driven 测试
/*
func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid", "valid", false},
        {"empty", "", true},
        {"invalid", "!@#", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate(%q) error = %v", tt.input, err)
            }
        })
    }
}
*/
//
// 模式 6: 字符串构建
/*
var builder strings.Builder
for _, item := range items {
    builder.WriteString(item)
}
result := builder.String()
*/

// ============================================================================
// 10. 常见错误和修复
// ============================================================================
//
// 错误 1: 忽略错误
// ❌ file, _ := os.Open("file.txt")
// ✅ file, err := os.Open("file.txt")
//    if err != nil { return err }
//
// 错误 2: 变量名不清楚
// ❌ d := getData()
// ✅ fingerprints := GetFingerprintList()
//
// 错误 3: 注释不清楚
// ❌ // 检查数据
// ✅ // 验证输入数据的格式和范围
//
// 错误 4: 没有文档注释
// ❌ func process() { }
// ✅ // process 处理输入数据
//    func process() { }
//
// 错误 5: 深度嵌套
// ❌ if a { if b { if c { doSomething() } } }
// ✅ if !a || !b || !c { return }
//    doSomething()
//
// 错误 6: 硬编码的魔数
// ❌ if timeout > 30 { }
// ✅ const DefaultTimeout = 30
//    if timeout > DefaultTimeout { }

// ============================================================================
// 11. 关键文档链接
// ============================================================================
//
// 完整规范文档位置:
//   docs/5-process/development/00-go-development-rules.md
//   docs/5-process/development/01-fingerprint-project-rules.md
//   docs/5-process/development/02-code-comment-templates.md
//
// 官方参考:
//   Effective Go: https://golang.org/doc/effective_go
//   Go Code Review Comments: https://github.com/golang/go/wiki/CodeReviewComments
//   Go Package Documentation: https://golang.org/doc/

// ============================================================================
// 12. 版本更新
// ============================================================================
//
// v1.0 (2026-02-28): 初始规范制定
//   - 基于 Go 1.25.4 标准
//   - 包含 Fingerprint 项目特定规范
//   - 完整的注释模板和示例

// ============================================================================
// 规范的强制执行
// ============================================================================
//
// 这些规范是强制性的。所有代码提交必须符合这些要求。
//
// 代码审查时会检查:
//   1. gofmt 格式化检查 (自动)
//   2. go vet 静态分析 (自动)
//   3. 命名规范检查 (手工)
//   4. 注释完整性检查 (手工)
//   5. 错误处理检查 (手工)
//   6. 测试覆盖检查 (自动)
//
// 不符合规范的 PR 将被拒绝。

// ============================================================================
// END OF QUICK REFERENCE
// ============================================================================
//
// 记住: 好的代码不仅是能够工作，而是易于理解、维护和修改。
// 遵循这些规范，你的代码将更容易被团队理解和维护。
