<!-- markdownlint-disable MD009 MD034 MD037 MD041 -->

// Package fingerprint 代码注释完整模板和示例
//
// 本文档提供了项目中各种代码注释的完整模板和实际示例
// 所有开发者必须参照这些模板编写代码注释
//
// 创建日期: 2026-02-28
// 版本: 1.0

package fingerprint

// ============================================================================
// 1. 包文档注释模板
// ============================================================================
//
// 每个包的第一个源文件必须包含包文档注释
// 格式如下（这是示例，实际项目已经有）:
//
/*
// Package fingerprint 提供高性能的浏览器 TLS 指纹识别和模拟功能
//
// 本包实现了完整的浏览器指纹识别系统，包括:
//
// 1. 指纹配置管理 - 维护 70+ 个真实浏览器的 TLS 配置
// 2. User-Agent 生成 - 生成与浏览器版本和操作系统匹配的 User-Agent
// 3. HTTP Headers 生成 - 生成符合浏览器特性的 HTTP Headers
// 4. 指纹计算 - 支持 JA3 和 JA4 指纹算法
// 5. 异常检测 - 检测无头浏览器和自动化工具
// 6. 矛盾检测 - 验证指纹属性的逻辑一致性
// 7. 被动识别 - 从 HTTP Headers 识别浏览器信息
//
// 核心类型:
//
//   - ClientProfile: 代表一个浏览器指纹配置
//   - FingerprintResult: 包含完整指纹、User-Agent 和 Headers
//   - HTTPHeaders: 标准化的 HTTP 请求头
//   - AnomalyDetector: 异常指纹检测器
//   - ContradictionDetector: 指纹矛盾检测器
//   - PassiveRecognizer: 被动浏览器识别器
//
// 主要函数:
//
//   GetRandomFingerprint() - 获取随机指纹
//   GetRandomFingerprintByBrowser() - 按浏览器类型获取指纹
//   ComputeJA3ByProfileName() - 计算 JA3 指纹
//   ComputeJA4ByProfileName() - 计算 JA4 指纹
//   RecognizeFromHeaders() - 从 Headers 识别浏览器
//
// 使用示例:
//
//   // 获取随机指纹
//   result, err := fingerprint.GetRandomFingerprint()
//   if err != nil {
//       log.Fatal(err)
//   }
//   
//   // 使用指纹进行 HTTP 请求
//   req, _ := http.NewRequest("GET", "https://example.com", nil)
//   
//   // 设置 Headers
//   for key, value := range result.Headers.ToMap() {
//       req.Header.Set(key, value)
//   }
//
// 性能特性:
//
//   - GetRandomFingerprint: ~7.4 微秒，分配 1.8 KB
//   - 支持并发调用，性能提升 5-6 倍
//   - 零分配的随机语言和操作系统选择
//
// 线程安全性:
//
//   所有公开函数都是线程安全的
//   可以安全地在多个 goroutine 中并发调用
//
package fingerprint
*/

// ============================================================================
// 2. 常量注释模板
// ============================================================================
//
// 模板:
/*
const (
    // ConstantName 常量的简短说明
    // 详细说明（如需要）
    ConstantName = value
    
    // AnotherConstant 另一个常量的说明
    AnotherConstant = value
)
*/
//
// 实际示例:
/*
const (
    // DefaultTimeout 默认请求超时时间，单位为秒
    // 此值应该足够完成大多数 HTTP 请求
    DefaultTimeout = 30
    
    // MaxRetries 最大重试次数
    // 超过此次数的请求将被放弃
    MaxRetries = 3
    
    // BrowserChrome Chrome 浏览器类型标识符
    // 用于在指纹映射中查询 Chrome 相关配置
    BrowserChrome = "chrome"
)
*/

// ============================================================================
// 3. 类型定义注释模板
// ============================================================================
//
// 模板:
/*
// TypeName 类型的简短说明
//
// 详细描述，包括:
// - 用途说明
// - 使用场景
// - 创建方式
//
// 字段说明应该在字段旁边
type TypeName struct {
    // FieldName 字段的说明
    // 包括字段含义、类型限制、有效范围等
    FieldName string
    
    // AnotherField 另一个字段的说明
    AnotherField int
}
*/
//
// 实际示例 - defense.go 中的 AnomalyDetector:
/*
// AnomalyDetector 异常指纹检测器
//
// 该结构体用于分析指纹数据中的可疑模式，判断指纹是否为
// 机器人或伪造指纹。它通过多个启发式算法检测异常，包括:
//
// 1. 熵值检测 - 检查数据的熵值是否过低或过高
// 2. 自动化工具标记检测 - 扫描已知的自动化工具特征
// 3. User-Agent 分析 - 识别无头浏览器和虚假浏览器
//
// 零值 AnomalyDetector 直接可用，无需初始化。
//
// 示例:
//   detector := &AnomalyDetector{}
//   if detector.DetectAnomalies(data) {
//       log.Warn("检测到异常指纹")
//   }
//
type AnomalyDetector struct{}
*/

// ============================================================================
// 4. 接口定义注释模板
// ============================================================================
//
// 模板:
/*
// InterfaceName 接口的简短说明
//
// 详细描述接口的用途和使用场景
//
// 实现此接口的类型必须满足以下要求:
// - 说明 1
// - 说明 2
//
type InterfaceName interface {
    // MethodName 方法的说明
    // 参数和返回值的详细说明
    MethodName(ctx context.Context, param string) (result interface{}, err error)
}
*/
//
// 实际示例 - 如果项目中有的话:
/*
// FingerprintProvider 指纹提供者接口
//
// 定义了获取指纹和列出可用指纹的方法
// 任何实现此接口的类型都可以作为指纹数据源
//
type FingerprintProvider interface {
    // GetFingerprint 获取指定名称的指纹配置
    // 
    // 参数:
    //   name: 指纹名称，如 "chrome_133"
    //
    // 返回值:
    //   *Fingerprint: 指纹配置，如果不存在返回 nil
    //   error: 获取过程中的错误
    //
    // 如果指纹不存在，返回 (nil, nil) 或相应错误
    GetFingerprint(name string) (*Fingerprint, error)
    
    // ListFingerprints 列出所有可用的指纹名称
    //
    // 返回值:
    //   []string: 指纹名称列表
    //   error: 列举过程中的错误
    //
    ListFingerprints() ([]string, error)
}
*/

// ============================================================================
// 5. 函数和方法注释完整模板
// ============================================================================
//
// 模板结构:
/*
// FunctionName 函数的简短说明 (一句话)
//
// 详细描述 (可选):
// 这里是对函数功能的详细说明，包括算法原理、重要细节等
// 如果函数比较复杂，应该有详细的描述
//
// 参数说明 (如有参数):
//   param1: 参数1的详细说明，包括类型含义和有效范围
//   param2: 参数2的详细说明
//
// 返回值说明 (如有返回值):
//   resultType: 返回值的说明，包括含义和使用方式
//   error: 可能返回的错误类型说明
//
// 错误处理 (如会返回错误):
//   如果发生 X 情况，返回 ErrX
//   如果发生 Y 情况，返回 ErrY
//
// 示例 (强烈推荐):
//   param1 := "example"
//   result, err := FunctionName(param1)
//   if err != nil {
//       log.Fatal(err)
//   }
//   // 使用 result
//
// 线程安全性 (如适用):
//   是 / 否 / 仅在 X 条件下安全
//
// 性能 (如重要):
//   平均耗时: X 微秒
//   内存分配: X 字节
//   时间复杂度: O(n)
//
// 弃用说明 (如适用):
//   [Deprecated] 此函数已弃用，请使用 NewFunctionName 代替
//
// 注意事项 (如有特殊说明):
//   - 说明 1
//   - 说明 2
//
// 相关函数:
//   - RelatedFunction1
//   - RelatedFunction2
//
func FunctionName(param1, param2 string) (result string, err error) {
    // 实现
    return "", nil
}
*/
//
// 实际示例 - defense.go 中的方法:
/*
// DetectAnomalies 检测指纹数据中的异常
//
// 该方法通过多个启发式算法检测指纹数据中的可疑模式。
// 它会按顺序检查:
// 1. 数据熵值是否过低（表示数据重复性过高）
// 2. 数据熵值是否过高（表示数据过于随机）
// 3. 已知自动化工具的特征标记
//
// 参数:
//   data: 指纹原始字节数据
//
// 返回值:
//   bool: 如果检测到异常返回 true，否则返回 false
//
// 示例:
//   detector := NewAnomalyDetector()
//   data := []byte("HeadlessChrome/120.0.0.0")
//   if detector.DetectAnomalies(data) {
//       fmt.Println("检测到异常")
//   }
//
// 线程安全性: 是
//
// 注意事项:
//   - 空数据切片返回 false
//   - 检测使用启发式算法，可能有误报
//   - 性能复杂度为 O(n)，其中 n 为数据长度
//
func (d *AnomalyDetector) DetectAnomalies(data []byte) bool {
    // 实现
    return false
}
*/

// ============================================================================
// 6. 复杂逻辑和算法的注释
// ============================================================================
//
// 模板:
/*
// 算法或流程的总体描述
func ComplexFunction() {
    // 第一步：说明第一步在做什么
    // 为什么要做这一步
    // ...代码...
    
    // 第二步：说明第二步在做什么
    // 为什么要做这一步
    // ...代码...
    
    // 关键点：需要特别注意的地方
    // ...代码...
}
*/
//
// 实际示例 - ja3.go 中的 GREASE 过滤:
/*
// isGREASEValue 检查是否为 GREASE 值 (RFC 8701)
//
// GREASE (GRE ASE) 是随机扩展和保留值，用于确保
// 中间设备安全地处理未来的 TLS 扩展。
//
// GREASE 值格式: 0xXAXA (十六进制)
// 其中 X 是任意十六进制数字
//
// 算法原理:
// 1. 检查低 4 位和高 4 位是否都是 0x0a
// 2. 检查中间 4 位是否是 0x0f
//
// 示例值:
//   0x0a0a (10, 10) - 是 GREASE
//   0x1a1a (26, 26) - 是 GREASE
//   0x0a1a (10, 26) - 不是 GREASE
//
func isGREASEValue(v uint16) bool {
    // 检查是否为 GREASE 值
    // 0xXAXA 格式的检查
    return v&0x0f0f == 0x0a0a && v&0xf0f0 == v&0xf0f0
}
*/

// ============================================================================
// 7. 全局变量和常量的分组注释
// ============================================================================
//
// 模板:
/*
// ============================================================================
// 浏览器类型定义
// ============================================================================

const (
    // BrowserChrome Chrome 浏览器标识
    BrowserChrome BrowserType = "chrome"
    
    // BrowserFirefox Firefox 浏览器标识
    BrowserFirefox BrowserType = "firefox"
)

// ============================================================================
// 操作系统定义
// ============================================================================

const (
    // OSWindows10 Windows 10 操作系统
    OSWindows10 OperatingSystem = "Windows NT 10.0; Win64; x64"
    
    // OSMacOS14 macOS 14 操作系统
    OSMacOS14 OperatingSystem = "Macintosh; Intel Mac OS X 14_0_0"
)

// ============================================================================
// 全局配置
// ============================================================================

var (
    // DefaultConfig 默认配置实例
    // 包含所有字段的默认值
    DefaultConfig = Configuration{
        Timeout: 30,
    }
    
    // cachedFingerprints 缓存已加载的指纹
    // 键: 指纹名称，值: 指纹配置
    cachedFingerprints = make(map[string]*Fingerprint)
)
*/

// ============================================================================
// 8. 错误定义注释
// ============================================================================
//
// 模板:
/*
var (
    // ErrInvalidInput 表示输入数据无效
    // 常见原因: 空值、格式错误、超出范围
    ErrInvalidInput = errors.New("invalid input")
    
    // ErrNotFound 表示请求的资源不存在
    // 应该检查资源名称是否正确拼写
    ErrNotFound = errors.New("resource not found")
    
    // ErrUnauthorized 表示没有权限执行操作
    // 需要提供有效的认证凭证
    ErrUnauthorized = errors.New("unauthorized")
)
*/

// ============================================================================
// 9. TODO 和 FIXME 注释规范
// ============================================================================
//
// 正确格式:
/*
// TODO(developer_name): 简明扼要的任务描述，最好包含原因或上下文
// 例如: TODO(john): 优化这个算法的时间复杂度，当前为 O(n^2)

// FIXME(developer_name): 说明什么需要修复以及为什么
// 例如: FIXME(jane): 这里有内存泄漏，需要在函数结束时释放资源

// NOTE: 这是一个重要提示
// 例如: NOTE: 此操作非线程安全，需要外部同步机制

// BUG: 这是一个已知的 bug
// 例如: BUG: 当输入超过 1000 字符时会崩溃
*/
//
// 示例:
/*
// TODO(maintainer): 考虑为这个方法添加缓存，以提高性能
func (d *AnomalyDetector) DetectHeadlessBrowser(userAgent string) bool {
    // ...
}

// FIXME(developer): 当 data 为空时应该返回错误而不是 false
func (d *AnomalyDetector) DetectAnomalies(data []byte) bool {
    if len(data) == 0 {
        return false  // FIXME: 应该返回 error
    }
    // ...
}

// NOTE: 这个函数不是线程安全的
// 调用者需要确保外部的同步
func UnsafeOperation() {
    // ...
}

// BUG: 当处理非常大的数据集时，程序会OOM
func ProcessLargeDataset(data []byte) error {
    // ...
}
*/

// ============================================================================
// 10. 弃用函数的注释
// ============================================================================
//
// 模板:
/*
// Deprecated: GetOldFunction 已弃用
// 请使用 GetNewFunction 代替
//
// 旧实现有性能问题，新实现快 10 倍
// 迁移指南: 见 docs/migration-guide.md
//
// 移除计划: 将在 v3.0.0 版本中移除
//
func GetOldFunction() string {
    // 实现保持不变用于兼容性
}

// GetNewFunction 新的实现，性能更好
func GetNewFunction() string {
    // 新实现
}
*/

// ============================================================================
// 11. 性能关键代码的注释
// ============================================================================
//
// 例子:
/*
// 这个函数是性能关键路径，请谨慎修改
// 性能指标:
// - 目标耗时: < 1 微秒
// - 内存分配: 0 字节
// - 基准数据: BenchmarkRandomChoice
//
// 优化历史:
// - v1.0: 初始实现，5 微秒
// - v1.1: 使用预分配切片，1 微秒
// - v1.2: 使用 math/rand 而非 crypto/rand，0.5 微秒
//
func RandomChoice(items []string) string {
    // 实现
    return items[0]
}
*/

// ============================================================================
// 12. 文档和行内注释的平衡
// ============================================================================
//
// ✅ 好的注释:
/*
// 过滤 GREASE 值，这些是 TLS 规范中的占位符值
// 不应该被包含在指纹中
ciphers := filterGREASEUint16(spec.CipherSuites)

// 验证 ClientHello 格式
if err := validateClientHello(data); err != nil {
    return fmt.Errorf("invalid ClientHello: %w", err)
}
*/
//
// ❌ 不好的注释:
/*
// 过滤 GREASE
ciphers := filterGREASEUint16(spec.CipherSuites)

// c = c + 1
c = c + 1

// 检查错误
if err != nil {
    // 返回错误
    return err
}
*/

// ============================================================================
// END OF COMMENT TEMPLATE
// ============================================================================
