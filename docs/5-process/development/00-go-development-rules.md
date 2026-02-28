// Package fingerprint GO 开发规范 (DEVELOPMENT RULES)
//
// 本文档规定了本项目所有 Go 代码必须遵循的开发标准和规范。
// 这是一个强制性规范，所有代码提交必须符合这些要求。
//
// 创建日期: 2026-02-28
// 版本: 1.0
// 维护者: vistone
// 更新周期: 按需

package fingerprint

// ============================================================================
// GO 语言开发规范 - 完整指南
// ============================================================================

// 1. 包声明和导入规范
// ============================================================================
// 
// 包声明必须在文件的第一行（注释之后）
// 示例: package fingerprint
//
// 导入规则:
//   1. 标准库导入分组放在最前面
//   2. 第三方库导入分组放在中间
//   3. 本项目导入分组放在最后
//   4. 每个分组之间用空行分隔
//
// ✅ 正确示例:
// import (
//     "context"        // 标准库
//     "encoding/json"  // 标准库
//     ""               // 空行分隔
//     "github.com/xxx" // 第三方库
//     ""               // 空行分隔
//     "github.com/vistone/xxx" // 本项目库
// )
//
// ❌ 错误示例:
// import (
//     "github.com/xxx"
//     "context"
//     "encoding/json"
// )

// ============================================================================
// 2. 文件结构规范
// ============================================================================
//
// 每个 Go 文件必须有以下结构:
//
// 1. 文件头注释 (Package Documentation)
//    - Package 名称说明
//    - 文件目的
//    - 主要功能描述
//
// 2. Package 声明
//
// 3. Import 分组
//
// 4. 常量定义 (const)
//
// 5. 变量定义 (var)
//
// 6. 类型定义 (type)
//
// 7. 接口定义 (interface)
//
// 8. 函数实现
//    - 公开函数 (大写开头)
//    - 私有函数 (小写开头)
//
// 9. 方法实现
//
// ✅ 标准结构示例:
/*
package fingerprint

import (
    "fmt"
    "strings"
)

// ============================================================================
// 常量定义
// ============================================================================

const (
    // DefaultTimeout 默认超时时间（单位：秒）
    DefaultTimeout = 30
    
    // MaxRetries 最大重试次数
    MaxRetries = 3
)

// ============================================================================
// 类型定义
// ============================================================================

// MyStruct 我的结构体描述
type MyStruct struct {
    // Field1 字段1的说明
    Field1 string
}

// ============================================================================
// 接口定义
// ============================================================================

// MyInterface 我的接口描述
type MyInterface interface {
    // Method1 方法1的说明
    Method1(ctx context.Context) error
}

// ============================================================================
// 公开函数
// ============================================================================

// PublicFunction 公开函数的说明
func PublicFunction(param string) (result string, err error) {
    // 实现
    return "", nil
}

// ============================================================================
// 私有函数
// ============================================================================

// privateFunction 私有函数的说明
func privateFunction(param string) string {
    // 实现
    return param
}
*/

// ============================================================================
// 3. 注释规范
// ============================================================================
//
// 3.1 文件头注释规范
// 每个 .go 文件开头必须有包文档注释
//
// ✅ 正确示例:
// Package fingerprint 提供浏览器 TLS 指纹识别功能
//
// 本包实现了...
// 主要功能包括:
// - 功能1
// - 功能2
//
// 示例:
//   result, err := GetRandomFingerprint()
//
// package fingerprint
//
// ❌ 错误示例:
// package fingerprint  // 没有包文档

// 3.2 常量和变量注释规范
// 每个 const 和 var 声明都必须有注释
//
// ✅ 正确示例:
// const (
//     // BrowserChrome Chrome 浏览器标识
//     BrowserChrome BrowserType = "chrome"
//     
//     // BrowserFirefox Firefox 浏览器标识
//     BrowserFirefox BrowserType = "firefox"
// )
//
// ❌ 错误示例:
// const (
//     BrowserChrome BrowserType = "chrome"
//     BrowserFirefox BrowserType = "firefox"
// )

// 3.3 类型定义注释规范
// 每个类型定义必须有说明性注释
//
// ✅ 正确示例:
// // FingerprintResult 指纹检测结果
// // 包含 TLS 指纹配置、User-Agent 和 HTTP Headers
// type FingerprintResult struct {
//     // Profile 客户端指纹配置
//     Profile ClientProfile
//     
//     // UserAgent 对应的 User-Agent 字符串
//     UserAgent string
// }
//
// ❌ 错误示例:
// type FingerprintResult struct {
//     Profile ClientProfile
//     UserAgent string
// }

// 3.4 函数和方法注释规范
// 每个公开函数和方法必须有完整的 godoc 注释
//
// 格式:
// // FunctionName 简短说明
// //
// // 详细描述（如需要）
// //
// // 参数:
// //   - param1: 参数1说明
// //   - param2: 参数2说明
// //
// // 返回值:
// //   - 返回值1: 说明
// //   - 返回值2: 说明
// //
// // 示例:
// //   result, err := FunctionName("param")
// //
// // 注意事项:
// //   - 说明1
// //   - 说明2
// //
// func FunctionName(param1, param2 string) (result string, err error) {
//     // 实现
// }
//
// ✅ 完整示例:
// // GetRandomFingerprint 获取随机浏览器指纹
// //
// // 该函数从所有可用的浏览器指纹中随机选择一个，
// // 并返回完整的指纹配置、User-Agent 和 HTTP Headers。
// //
// // 返回值:
// //   - *FingerprintResult: 指纹结果结构体
// //   - error: 错误信息（如指纹库为空）
// //
// // 示例:
// //   result, err := GetRandomFingerprint()
// //   if err != nil {
// //       log.Fatal(err)
// //   }
// //   println(result.UserAgent)
// //
// // 线程安全: 是
// //
// func GetRandomFingerprint() (*FingerprintResult, error) {
//     // 实现
// }

// ============================================================================
// 4. 命名规范
// ============================================================================
//
// 4.1 包名规范
// - 全小写，单个单词
// - 避免下划线和连字符
// - 避免使用通用名字（如 common, util, utils）
//
// ✅ 正确: fingerprint, profiles, internal/utils
// ❌ 错误: finger_print, FingePrint, FingerPrints
//
// 4.2 常量命名规范
// - 采用 PascalCase（帕斯卡命名法）
// - 完整单词，避免缩写
// - 公开常量大写开头
// - 私有常量小写开头
//
// ✅ 正确:
// const (
//     BrowserChrome = "chrome"
//     browserChrome = "chrome"  // 私有
//     MaxRetries = 3
// )
//
// ❌ 错误:
// const (
//     BROWSER_CHROME = "chrome"
//     browser_chrome = "chrome"
//     max_retries = 3
// )
//
// 4.3 变量命名规范
// - 采用 camelCase（驼峰命名法）
// - 完整单词或明确的缩写
// - 避免单字母变量（除了循环计数器 i, j, k）
// - 公开变量大写开头
// - 私有变量小写开头
//
// ✅ 正确:
// var (
//     GlobalConfig Configuration      // 公开全局变量
//     globalConfig Configuration      // 私有全局变量
//     localCounter int                // 本地变量
// )
//
// ❌ 错误:
// var (
//     global_config Configuration
//     GLOBAL_CONFIG Configuration
//     gc Configuration
// )
//
// 4.4 函数命名规范
// - 采用 PascalCase（公开函数）或 camelCase（私有函数）
// - 使用动词开头表示动作
// - 返回布尔值的函数用 Is, Has, Can 开头
// - 返回错误的函数应明确说明
//
// ✅ 正确函数名:
// func GetRandomFingerprint() {...}         // 获取
// func GetUserAgent(name string) {...}      // 获取
// func CreateFingerprint() {...}            // 创建
// func UpdateProfile(p Profile) {...}       // 更新
// func DeleteCache() {...}                  // 删除
// func HasFingerprint(name string) bool {...} // 检查存在
// func IsHeadlessBrowser(ua string) bool {...} // 检查条件
// func ValidateFingerprint(f Fingerprint) error {...} // 验证
//
// ❌ 错误函数名:
// func RandomFingerprint() {...}            // 缺少动词
// func getFP() {...}                        // 缩写不明确
// func get_random_fingerprint() {...}       // 下划线
// func GET_RANDOM_FINGERPRINT() {...}       // 全大写
//
// 4.5 方法命名规范
// - 与函数命名相同规则
// - 接收者建议使用 1-2 个字母缩写
//
// ✅ 正确:
// func (d *AnomalyDetector) DetectAnomalies(data []byte) bool {...}
// func (r *PassiveRecognizer) RecognizeFromHeaders(headers map[string]string) {...}
//
// ❌ 错误:
// func (anomaly_detector *AnomalyDetector) detect_anomalies(data []byte) bool {...}
//
// 4.6 接口命名规范
// - 单一方法接口用 er 后缀
// - 多个方法接口用名词形式
//
// ✅ 正确:
// type Reader interface {
//     Read(p []byte) (n int, err error)
// }
//
// type Writer interface {
//     Write(p []byte) (n int, err error)
// }
//
// type FingerprintProvider interface {
//     GetFingerprint(name string) (*Fingerprint, error)
//     ListFingerprints() ([]string, error)
// }
//
// ❌ 错误:
// type ReaderInterface interface {...}
// type Reader_Interface interface {...}
//
// 4.7 结构体字段命名规范
// - 采用 PascalCase 或 camelCase
// - 字段含义明确，避免缩写
// - 每个字段都要有注释说明
// - 同类型字段可以合并声明
//
// ✅ 正确:
// type HTTPHeaders struct {
//     // UserAgent User-Agent 请求头
//     UserAgent string
//     
//     // AcceptLanguage Accept-Language 请求头（支持多语言）
//     AcceptLanguage string
//     
//     // Custom 用户自定义的 headers
//     Custom map[string]string
// }
//
// ❌ 错误:
// type HTTPHeaders struct {
//     UA string
//     AL string
//     custom map[string]string  // 应该大写
// }

// ============================================================================
// 5. 错误处理规范
// ============================================================================
//
// 5.1 错误检查
// 必须检查所有可能返回错误的函数调用
//
// ✅ 正确:
// result, err := GetRandomFingerprint()
// if err != nil {
//     return fmt.Errorf("failed to get fingerprint: %w", err)
// }
// // 使用 result
//
// ❌ 错误:
// result, _ := GetRandomFingerprint()  // 忽略错误
// // 直接使用 result
//
// ❌ 错误:
// result, err := GetRandomFingerprint()  // 检查错误后继续使用
// // 使用 result
//
// 5.2 错误创建规范
// - 使用 fmt.Errorf 创建带上下文的错误
// - 使用 %w 包装错误以支持 errors.Is/As
// - 避免创建同一错误多次
//
// ✅ 正确:
// if len(profiles) == 0 {
//     return nil, fmt.Errorf("no profiles available: %w", ErrNoProfiles)
// }
//
// ❌ 错误:
// if len(profiles) == 0 {
//     return nil, errors.New("no profiles")
// }
//
// 5.3 错误信息规范
// - 错误信息小写开头（除非以专有名词开头）
// - 提供足够的上下文信息
// - 避免在错误中使用特殊格式字符
//
// ✅ 正确:
// return fmt.Errorf("failed to parse TLS version: %w", err)
// return fmt.Errorf("invalid fingerprint %q: %w", name, err)
//
// ❌ 错误:
// return fmt.Errorf("Failed to parse TLS version: %w", err)
// return fmt.Errorf("ERROR: Invalid fingerprint!")
//
// 5.4 哨兵错误规范
// 定义项目级的哨兵错误
//
// ✅ 正确:
// var (
//     // ErrNoProfiles 表示没有可用的指纹配置
//     ErrNoProfiles = errors.New("no profiles available")
//     
//     // ErrInvalidFingerprint 表示指纹格式无效
//     ErrInvalidFingerprint = errors.New("invalid fingerprint format")
// )
//
// 使用方式:
// if err != nil {
//     if errors.Is(err, ErrNoProfiles) {
//         // 处理没有配置的情况
//     }
// }

// ============================================================================
// 6. 变量和指针规范
// ============================================================================
//
// 6.1 变量初始化
// 使用简短声明或 var 块初始化变量
//
// ✅ 正确:
// var (
//     DefaultTimeout = 30
//     MaxRetries = 3
// )
//
// ✅ 正确:
// timeout := 30
// maxRetries := 3
//
// ❌ 错误:
// var DefaultTimeout = 30
// var MaxRetries = 3
//
// 6.2 指针使用规范
// - 只在必要时使用指针（值修改、避免复制大对象）
// - 接收者通常使用指针
// - 参数优先考虑值类型
//
// ✅ 正确:
// func (d *AnomalyDetector) DetectAnomalies(data []byte) bool {...}
// // data 是 []byte，不需要指针（切片已是引用）
//
// ✅ 正确:
// func UpdateProfile(profile *Profile) error {
//     // 需要修改 profile，使用指针
//     profile.Updated = time.Now()
//     return nil
// }
//
// ❌ 错误:
// func (d AnomalyDetector) DetectAnomalies(data *[]byte) bool {...}
// // 接收者应该是指针，data 是切片不需要指针
//
// 6.3 nil 检查规范
// 在使用指针前必须检查 nil
//
// ✅ 正确:
// if profile == nil {
//     return fmt.Errorf("profile cannot be nil")
// }
//
// ❌ 错误:
// // 直接使用 profile.Field，可能 panic
// value := profile.Field

// ============================================================================
// 7. 循环和条件规范
// ============================================================================
//
// 7.1 条件语句
// 避免深度嵌套，使用卫语句（guard clause）
//
// ✅ 正确:
// if err != nil {
//     return err
// }
// if len(data) == 0 {
//     return fmt.Errorf("data is empty")
// }
// // 主要逻辑
// return process(data)
//
// ❌ 错误:
// if err == nil {
//     if len(data) > 0 {
//         // 主要逻辑深度嵌套
//         return process(data)
//     }
// }
// return err
//
// 7.2 循环规范
// 使用明确的循环变量
//
// ✅ 正确:
// for i := 0; i < len(items); i++ {
//     // 处理
// }
//
// ✅ 正确:
// for _, item := range items {
//     // 处理
// }
//
// ✅ 正确:
// for key, value := range mapping {
//     // 处理
// }
//
// ❌ 错误:
// for {  // 无限循环，容易出错
//     if !hasMore {
//         break
//     }
// }

// ============================================================================
// 8. 类型转换规范
// ============================================================================
//
// 8.1 类型转换
// 显式进行类型转换，不依赖隐式转换
//
// ✅ 正确:
// var count int64 = 100
// value := int(count)
//
// ✅ 正确:
// str := "hello"
// bytes := []byte(str)
//
// ❌ 错误:
// count := int64(100)
// value := count  // 不同类型，Go 不允许隐式转换
//
// 8.2 类型断言
// 使用 ok 模式进行安全的类型断言
//
// ✅ 正确:
// if value, ok := data.(string); ok {
//     // 使用 value
// } else {
//     // 处理错误情况
// }
//
// ❌ 错误:
// value := data.(string)  // 可能 panic
//
// ✅ 正确:
// switch value := data.(type) {
// case string:
//     // 处理 string
// case int:
//     // 处理 int
// default:
//     // 处理未知类型
// }

// ============================================================================
// 9. 并发和锁规范
// ============================================================================
//
// 9.1 Goroutine 使用
// - 在使用 goroutine 前明确知道它何时结束
// - 使用 context.Context 控制取消和超时
// - 用 sync.WaitGroup 等待 goroutine 完成
//
// ✅ 正确:
// var wg sync.WaitGroup
// wg.Add(1)
// go func() {
//     defer wg.Done()
//     // 工作
// }()
// wg.Wait()
//
// ❌ 错误:
// go func() {
//     // 工作
// }()
// // 没有等待 goroutine 完成
//
// 9.2 互斥锁规范
// - 使用 defer 解锁，确保锁被释放
// - 锁的范围要尽可能小
// - 避免在锁内进行耗时操作
//
// ✅ 正确:
// func (c *Cache) Get(key string) (string, bool) {
//     c.mu.RLock()
//     defer c.mu.RUnlock()
//     value, ok := c.data[key]
//     return value, ok
// }
//
// ❌ 错误:
// func (c *Cache) Get(key string) (string, bool) {
//     c.mu.RLock()
//     value, ok := c.data[key]
//     c.mu.RUnlock()  // 可能被 panic 跳过
//     return value, ok
// }

// ============================================================================
// 10. 性能规范
// ============================================================================
//
// 10.1 内存分配
// - 避免频繁小对象分配
// - 在已知容量的地方预分配切片
// - 使用 sync.Pool 重用常见对象
//
// ✅ 正确:
// items := make([]string, 0, estimatedCapacity)
// for _, item := range source {
//     items = append(items, item)
// }
//
// ❌ 错误:
// var items []string
// for _, item := range source {
//     items = append(items, item)  // 频繁重新分配
// }
//
// 10.2 字符串操作
// - 在循环中使用 strings.Builder
// - 避免频繁使用 + 拼接字符串
//
// ✅ 正确:
// var builder strings.Builder
// for _, s := range strings {
//     builder.WriteString(s)
// }
// result := builder.String()
//
// ❌ 错误:
// var result string
// for _, s := range strings {
//     result += s  // 每次都创建新字符串
// }
//
// 10.3 避免不必要的复制
// - 优先使用切片而非数组复制
// - 在可能的地方使用引用
//
// ✅ 正确:
// func Process(data []byte) error {
//     // 直接使用切片，不复制
//     return doSomething(data)
// }
//
// ❌ 错误:
// func Process(data []byte) error {
//     // 创建不必要的副本
//     dataCopy := make([]byte, len(data))
//     copy(dataCopy, data)
//     return doSomething(dataCopy)
// }

// ============================================================================
// 11. 文档和代码注释规范
// ============================================================================
//
// 11.1 公开 API 文档注释
// 每个公开的类型、函数、方法必须有说明性注释
// 格式按照 Go 官方 godoc 标准
//
// ✅ 完整示例:
/*
// GetRandomFingerprint 从所有指纹中随机选择一个
//
// 该函数会随机选择一个支持的浏览器指纹，并自动生成对应的
// User-Agent 和 HTTP Headers。操作系统会从所有支持的系统中
// 随机选择。
//
// 返回值:
//   - *FingerprintResult: 包含指纹配置、User-Agent 和 HTTP Headers
//   - error: 错误信息（例如指纹库为空）
//
// 示例:
//   result, err := GetRandomFingerprint()
//   if err != nil {
//       log.Fatal(err)
//   }
//   println(result.UserAgent)
//
// 线程安全性: 是
//
// 性能: 平均耗时 7.4 微秒，分配 1.8 KB 内存
//
func GetRandomFingerprint() (*FingerprintResult, error) {
    // 实现
}
*/
//
// 11.2 私有函数注释
// 私有函数也应该有注释说明其目的
//
// ✅ 正确:
// // randomChoice 从切片中随机选择一个元素
// func randomChoice(items []string) string {
//     // 实现
// }
//
// 11.3 复杂逻辑注释
// 复杂的算法或逻辑应该有详细的行内注释
//
// ✅ 正确:
// // 过滤 GREASE 值（RFC 8701）
// // GREASE 值格式：0xXAXA（十六进制）
// func isGREASEValue(v uint16) bool {
//     return v&0x0f0f == 0x0a0a && v&0xf0f0 == v&0xf0f0
// }
//
// 11.4 TODO 和 FIXME 注释
// 使用标准化的 TODO 格式
//
// ✅ 正确:
// // TODO(maintainer): 优化性能，考虑使用缓存
// // FIXME: 当前实现有内存泄漏，需要修复
// // NOTE: 这个算法时间复杂度为 O(n^2)，需要后续优化
//
// ❌ 错误:
// // TODO 优化
// // fixme: memory leak
// // note 需要优化

// ============================================================================
// 12. 测试规范
// ============================================================================
//
// 12.1 测试文件命名
// 测试文件应该使用 _test.go 后缀
//
// ✅ 正确:
// fingerprint_test.go
// integration_test.go
// benchmark_test.go
//
// ❌ 错误:
// test_fingerprint.go
// fingerprint.test.go
//
// 12.2 测试函数命名
// 测试函数应该以 Test 开头，使用描述性名称
//
// ✅ 正确:
// func TestGetRandomFingerprint(t *testing.T) { ... }
// func TestDetectHeadlessBrowser(t *testing.T) { ... }
// func BenchmarkGetRandomFingerprint(b *testing.B) { ... }
//
// ❌ 错误:
// func TestGRF(t *testing.T) { ... }
// func Test_GetRandomFingerprint(t *testing.T) { ... }
//
// 12.3 测试覆盖
// - 必须测试正常情况和错误情况
// - 使用 table-driven 测试进行多个场景测试
//
// ✅ 正确:
// func TestValidate(t *testing.T) {
//     tests := []struct {
//         name      string
//         input     string
//         wantErr   bool
//     }{
//         {"valid", "valid_input", false},
//         {"empty", "", true},
//         {"invalid", "invalid@#$", true},
//     }
//     
//     for _, tt := range tests {
//         t.Run(tt.name, func(t *testing.T) {
//             err := Validate(tt.input)
//             if (err != nil) != tt.wantErr {
//                 t.Errorf("Validate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
//             }
//         })
//     }
// }

// ============================================================================
// 13. 代码审查检查清单
// ============================================================================
//
// 在提交 PR 前，请检查以下项目:
//
// 代码质量:
//   ✅ 所有公开 API 都有 godoc 注释
//   ✅ 复杂逻辑都有行内注释
//   ✅ 没有明显的性能问题
//   ✅ 错误都被正确处理
//   ✅ 使用了合适的错误类型
//
// 命名规范:
//   ✅ 变量名遵循 camelCase
//   ✅ 常量名遵循 PascalCase
//   ✅ 函数名使用动词开头
//   ✅ 函数名和变量名有明确含义
//   ✅ 接口名以 er 或名词结尾
//
// 代码风格:
//   ✅ 运行了 gofmt 格式化
//   ✅ 运行了 go vet 检查
//   ✅ 没有使用未导入的包
//   ✅ 没有死代码或未使用的变量
//   ✅ 代码可读性好，逻辑清晰
//
// 测试:
//   ✅ 新增功能都有对应测试
//   ✅ 测试覆盖主要场景和错误情况
//   ✅ 测试函数名称清晰
//   ✅ 所有测试都通过
//   ✅ 性能测试包含相关 benchmark
//
// 文档:
//   ✅ 更新了相关文档
//   ✅ 文档与代码一致
//   ✅ 提交信息清晰且描述完整

// ============================================================================
// 14. 常见错误模式和修复
// ============================================================================
//
// 14.1 未检查错误
// ❌ 错误:
// file, _ := os.Open("file.txt")
// scanner := bufio.NewScanner(file)
// data := scanner.Bytes()
//
// ✅ 正确:
// file, err := os.Open("file.txt")
// if err != nil {
//     return fmt.Errorf("failed to open file: %w", err)
// }
// defer file.Close()
// scanner := bufio.NewScanner(file)
// if !scanner.Scan() {
//     if err := scanner.Err(); err != nil {
//         return fmt.Errorf("scan error: %w", err)
//     }
// }
// data := scanner.Bytes()
//
// 14.2 忘记关闭资源
// ❌ 错误:
// file, err := os.Open("file.txt")
// if err != nil {
//     return err
// }
// // 忘记了 defer file.Close()
// return processFile(file)
//
// ✅ 正确:
// file, err := os.Open("file.txt")
// if err != nil {
//     return err
// }
// defer file.Close()
// return processFile(file)
//
// 14.3 goroutine 泄漏
// ❌ 错误:
// for i := 0; i < 100; i++ {
//     go func() {
//         for {
//             doSomething()
//         }
//     }()
// }
// // goroutine 永不结束
//
// ✅ 正确:
// for i := 0; i < 100; i++ {
//     go func(ctx context.Context) {
//         for {
//             select {
//             case <-ctx.Done():
//                 return
//             default:
//                 doSomething()
//             }
//         }
//     }(ctx)
// }
//
// 14.4 错误的字符串比较
// ❌ 错误:
// if browserType == "Chrome" || browserType == "chrome" {
//     // 冗余，应该在源头统一大小写
// }
//
// ✅ 正确:
// browserType = strings.ToLower(browserType)
// if browserType == "chrome" {
//     // 统一比较小写
// }
//
// 或者使用常量:
// if browserType == BrowserChrome {
//     // 使用已定义的常量
// }

// ============================================================================
// 15. 性能优化检查清单
// ============================================================================
//
// 提交前检查这些性能问题:
//
//   ✅ 避免在 for 循环中进行字符串拼接
//   ✅ 预分配已知容量的切片
//   ✅ 避免频繁的内存分配
//   ✅ 使用指针传递大结构体
//   ✅ 避免不必要的类型转换
//   ✅ 合理使用 goroutine，避免过度并发
//   ✅ 缓存重复计算结果
//   ✅ 使用合适的数据结构（map, set 等）
//   ✅ 避免深度嵌套循环
//   ✅ 定期进行性能测试和 profiling

// ============================================================================
// END OF GO DEVELOPMENT RULES
// ============================================================================
