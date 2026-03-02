<!-- markdownlint-disable MD011 MD041 -->

// Fingerprint 项目代码审查规范
//
// 本文档是对通用 Go 开发规范的扩展，针对 Fingerprint 项目的具体要求
//
// 创建日期: 2026-02-28
// 版本: 1.0

package fingerprint

// ============================================================================
// Fingerprint 项目特定规范
// ============================================================================

// ============================================================================
// 1. 包组织规范
// ============================================================================
//
// 项目目录结构:
//
// fingerprint/
// ├── 公开接口 (*.go)
// │   ├── types.go          - 类型定义和常量
// │   ├── random.go         - 随机指纹生成
// │   ├── headers.go        - HTTP Headers 生成
// │   ├── useragent.go      - User-Agent 生成
// │   ├── ja3.go            - JA3 指纹计算
// │   ├── ja4.go            - JA4 指纹计算
// │   ├── defense.go        - 防护检测（异常、矛盾）
// │   ├── noise.go          - 噪声注入
// │   ├── profiles.go       - 指纹映射导出
// │   └── automation.go     - 自动化工具检测
// │
// ├── internal/
// │   └── utils/            - 内部工具函数
// │       ├── rand.go       - 随机数生成
// │       ├── strings.go    - 字符串处理
// │       └── useragent.go  - User-Agent 工具
// │
// ├── profiles/             - 指纹配置包
// │   ├── profiles.go       - 指纹类定义和映射
// │   ├── internal_browser_profiles.go
// │   ├── contributed_browser_profiles.go
// │   ├── edge_profiles.go
// │   └── ...其他配置
// │
// ├── test/                 - 测试文件
// │   ├── fingerprint_test.go
// │   ├── benchmark_test.go
// │   └── integration_test.go
// │
// └── examples/             - 示例代码
//

// ============================================================================
// 2. 指纹相关代码规范
// ============================================================================
//
// 2.1 ClientHello 处理规范
// 所有涉及 ClientHello 的代码都必须遵循以下规范:
//
// ✅ 必须:
//   - 使用 utls 库的标准类型
//   - 清楚地说明支持的 TLS 版本
//   - 文档化 GREASE 值的过滤
//   - 验证 ClientHello 的有效性
//
// ✅ 示例:
// // GetClientHelloSpec 获取 TLS ClientHello 规范
// //
// // 返回经过验证的 ClientHello 规范，用于 TLS 握手
// // 所有 GREASE 值已被过滤
// //
// // 支持的 TLS 版本:
// //   - TLS 1.2
// //   - TLS 1.3
// //
// func (cp *ClientProfile) GetClientHelloSpec() (*tls.ClientHelloSpec, error) {
//     if cp.clientHelloId == "" {
//         return nil, fmt.Errorf("empty ClientHello ID")
//     }
//     // 实现
// }
//
// 2.2 指纹验证规范
// 每个指纹必须包含以下验证:
//
// ✅ 验证内容:
//   - ClientHello 不为空
//   - ClientHello 格式有效
//   - TLS 版本支持
//   - 密码套件合法
//   - 扩展列表有效
//
// 2.3 指纹映射规范
// MappedTLSClients 映射必须满足:
//
// ✅ 必须:
//   - 每个映射都有注释说明
//   - 键使用小写英文和下划线 (chrome_133)
//   - 值是有效的 ClientProfile
//   - 包含默认指纹 (DefaultClientProfile)
//
// ✅ 示例:
// var MappedTLSClients = map[string]ClientProfile{
//     // Chrome 系列
//     "chrome_133": Chrome_133,  // 最新 Chrome
//     "chrome_132": Chrome_132,  // 前一版本
//     "chrome_131": Chrome_131,  // 更早版本
//
//     // Firefox 系列
//     "firefox_135": Firefox_135,
//     "firefox_134": Firefox_134,
//
//     // Safari 系列
//     "safari_18_5": Safari_18_5,
//     "safari_18_0": Safari_18_0,
//
//     // 移动端
//     "okhttp4_android_13": Okhttp4Android13,
//     "okhttp4_android_12": Okhttp4Android12,
// }

// ============================================================================
// 3. User-Agent 生成规范
// ============================================================================
//
// 3.1 User-Agent 合法性
// 生成的 User-Agent 必须满足:
//
// ✅ 要求:
//   - 格式符合 RFC 标准
//   - 与浏览器版本一致
//   - 与操作系统一致
//   - 避免已知的自动化工具标记
//
// ✅ 示例:
// // 正确的 Chrome User-Agent
// Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 \
// (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36
//
// // 不正确的 User-Agent (包含 HeadlessChrome)
// Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 \
// (KHTML, like Gecko) HeadlessChrome/133.0.0.0 Safari/537.36
//
// 3.2 User-Agent 缓存
// User-Agent 缓存必须明确标注
//
// ✅ 正确:
// // userAgentCache 缓存已生成的 User-Agent 字符串
// // 键格式: "{browserType}_{version}_{os}"
// var userAgentCache = make(map[string]string)
//
// 3.3 User-Agent 更新
// 当浏览器版本更新时:
//
// ✅ 必须:
//   - 添加新版本模板
//   - 更新对应的测试
//   - 验证与服务器兼容性
//   - 更新文档

// ============================================================================
// 4. HTTP Headers 生成规范
// ============================================================================
//
// 4.1 Headers 结构
// HTTPHeaders 必须包含所有必需字段:
//
// ✅ 必需字段:
//   - Accept          - 必须存在且有效
//   - AcceptLanguage  - 必须存在，支持多语言
//   - AcceptEncoding  - 应当存在
//   - UserAgent       - 必须与指纹匹配
//   - SecFetchSite    - Chrome 系必须
//   - SecFetchMode    - Chrome 系必须
//   - Custom          - 用户自定义 headers
//
// 4.2 Headers 一致性
// Headers 与指纹必须保持一致:
//
// ✅ 规则:
//   - Chrome headers 中应有 Sec-CH-UA
//   - Firefox headers 不应有 Sec-CH-UA
//   - 版本号与 User-Agent 一致
//   - 移动设备有对应标记
//
// 4.3 自定义 Headers
// Custom 字段用于用户自定义:
//
// ✅ 示例:
// headers.Custom = map[string]string{
//     "Authorization": "Bearer token",
//     "X-Custom-Header": "value",
//     "Cookie": "session=xyz",
// }
//
// ❌ 不允许覆盖标准 headers

// ============================================================================
// 5. 浏览器检测规范
// ============================================================================
//
// 5.1 浏览器类型定义
// 必须使用预定义的常量:
//
// ✅ 正确:
// const (
//     BrowserChrome  BrowserType = "chrome"
//     BrowserFirefox BrowserType = "firefox"
//     BrowserSafari  BrowserType = "safari"
//     BrowserOpera   BrowserType = "opera"
//     BrowserEdge    BrowserType = "edge"
// )
//
// ❌ 避免:
// if browserName == "Chrome" {  // 应该使用常量
// }
//
// 5.2 浏览器识别算法
// 识别算法必须包括:
//
// ✅ 步骤:
//   1. 检查特殊标记 (Edge, Opera 在 Chrome 之前)
//   2. 通过 User-Agent 签名识别
//   3. 通过 HTTP Headers 验证
//   4. 返回置信度分数
//
// 5.3 识别结果格式
// RecognitionResult 必须完整:
//
// ✅ 必需字段:
//   - Browser: 识别的浏览器类型
//   - OS: 识别的操作系统
//   - BrowserVersion: 浏览器版本字符串
//   - Confidence: 置信度 (0.0-1.0)
//   - IsMobile: 是否移动设备
//   - IsBot: 是否机器人

// ============================================================================
// 6. 异常检测规范
// ============================================================================
//
// 6.1 AnomalyDetector 检测项
// 异常检测器必须检测以下项目:
//
// ✅ 检测内容:
//   - 无头浏览器标记 (HeadlessChrome, Phantom JS 等)
//   - 自动化工具标记 (Selenium, Puppeteer 等)
//   - 熵值异常 (过低或过高)
//   - User-Agent 与 headers 矛盾
//
// 6.2 检测结果
// 返回布尔值表示是否检测到异常:
//
// ✅ 示例:
// detector := NewAnomalyDetector()
// isAnomaly := detector.DetectAnomalies(data)
// if isAnomaly {
//     log.Warn("异常指纹检测")
// }
//
// 6.3 无头浏览器检测
// DetectHeadlessBrowser 必须检测:
//
// ✅ 标记:
//   - "HeadlessChrome"
//   - "PhantomJS"
//   - "Selenium"
//   - "WebDriver"
//   - "Puppeteer"
//   - "Playwright"
//   - "Cypress"
//
// ✅ 规则:
//   - 不区分大小写
//   - 检查整个 User-Agent 字符串
//   - 返回第一次匹配结果

// ============================================================================
// 7. 矛盾检测规范
// ============================================================================
//
// 7.1 ContradictionDetector 检测项
// 矛盾检测器必须检测以下矛盾:
//
// ✅ 检测内容:
//   - User-Agent 与操作系统不匹配
//   - User-Agent 与特性矛盾
//   - 移动设备与屏幕尺寸矛盾
//   - OS 与 Platform 字段矛盾
//
// 7.2 检测示例
//
// ✅ 矛盾示例 1 - OS 与 Platform 不匹配:
// attributes := map[string]string{
//     "os": "Windows NT 10.0",
//     "platform": "MacIntel",  // 矛盾！
// }
// if detector.CheckContradictions(attributes) {
//     log.Warn("检测到矛盾: Windows OS 但 Mac platform")
// }
//
// ✅ 矛盾示例 2 - User-Agent 与 OS 不匹配:
// attributes := map[string]string{
//     "user_agent": "Mozilla/5.0 (Windows NT 10.0...)",
//     "os": "Mac OS",  // 矛盾！
// }
//
// ✅ 矛盾示例 3 - 移动设备与屏幕尺寸:
// attributes := map[string]string{
//     "is_mobile": "true",
//     "screen_width": "3840",  // 矛盾！移动设备不会有这么大屏幕
// }

// ============================================================================
// 8. JA3/JA4 指纹规范
// ============================================================================
//
// 8.1 JA3 计算规范
// JA3 哈希计算必须遵循标准:
//
// ✅ 步骤:
//   1. 提取 TLS 版本
//   2. 提取密码套件列表
//   3. 提取扩展列表
//   4. 提取椭圆曲线列表
//   5. 提取点格式列表
//   6. 过滤 GREASE 值
//   7. 拼接为字符串
//   8. 计算 MD5 哈希
//
// ✅ JA3 字符串格式:
// TLSVersion,Ciphers,Extensions,EllipticCurves,EllipticCurvePointFormats
//
// 示例:
// 771,4865-4866-4867,0-23-65281,23-24-25,0
//
// 8.2 JA4 计算规范
// JA4 哈希计算包括两部分:
//
// ✅ JA4_a (签名):
// SSLVersion,NumberOfCiphers,NumberOfExtensions,SNI,AlpnInfo
//
// ✅ JA4_b (密码套件):
// SHA256(sorted_ciphers)[:12]
//
// ✅ JA4_c (扩展 + 签名算法):
// SHA256(sorted_extensions_and_signatures)[:12]
//
// 8.3 GREASE 值过滤
// 必须正确识别和过滤 GREASE 值:
//
// ✅ GREASE 值识别:
// // RFC 8701 定义 GREASE 值为 0xXAXA 格式
// func isGREASEValue(v uint16) bool {
//     return v&0x0f0f == 0x0a0a && v&0xf0f0 == v&0xf0f0
// }
//
// ✅ 必须过滤的位置:
//   - 密码套件列表
//   - 扩展列表
//   - 椭圆曲线列表
//   - 版本列表

// ============================================================================
// 9. 噪声注入规范
// ============================================================================
//
// 9.1 NoiseConfig 配置
// 噪声配置必须包含:
//
// ✅ 必需字段:
//   - Intensity (0.0-1.0): 噪声强度
//   - EnableCanvas: 是否启用 Canvas 噪声
//   - EnableAudio: 是否启用 Audio 噪声
//   - EnableWebGL: 是否启用 WebGL 噪声
//   - EnableFont: 是否启用 Font 噪声
//   - EnableScreen: 是否启用 Screen 噪声
//
// 9.2 噪声生成规范
// 生成的噪声必须:
//
// ✅ 要求:
//   - 每次生成不同的噪声
//   - 噪声强度与配置一致
//   - 噪声值在合理范围内
//   - 不影响基础指纹有效性
//
// 9.3 噪声应用场景
// 噪声应该在以下场景使用:
//
// ✅ 使用场景:
//   - Canvas 指纹（防止被识别）
//   - WebGL 指纹（防止被识别）
//   - 音频上下文（防止被识别）
//   - 字体列表（轻微变化）

// ============================================================================
// 10. 代码审查检查清单 (Fingerprint 专项)
// ============================================================================
//
// 在提交指纹相关代码前，检查:
//
// 指纹数据:
//   ✅ ClientHello 有效且完整
//   ✅ User-Agent 符合浏览器特性
//   ✅ Headers 与指纹一致
//   ✅ 没有已知的自动化工具标记
//   ✅ 版本号准确无误
//
// 规范性:
//   ✅ 使用预定义的浏览器类型常量
//   ✅ 使用预定义的操作系统常量
//   ✅ 指纹键值使用统一格式 (browser_version)
//   ✅ 每个指纹在 MappedTLSClients 中都有注册
//
// 检测功能:
//   ✅ 异常检测覆盖所有已知的自动化工具
//   ✅ 矛盾检测检查所有相关字段
//   ✅ JA3/JA4 计算结果验证正确
//   ✅ 被动识别的置信度合理
//
// 测试覆盖:
//   ✅ 每个新指纹都有对应的测试
//   ✅ 异常检测有正面和负面测试用例
//   ✅ Headers 生成有完整的测试
//   ✅ 性能测试包含新指纹
//
// 文档:
//   ✅ 新指纹在文档中有说明
//   ✅ 更新了支持的浏览器列表
//   ✅ 更新了性能数据
//   ✅ 提交信息清晰

// ============================================================================
// 11. 版本管理规范
// ============================================================================
//
// 11.1 版本号格式
// 使用 Semantic Versioning (SemVer):
//
// ✅ 格式: MAJOR.MINOR.PATCH
// 例如: v2.0.1
//
// - MAJOR: 不兼容的 API 变更
// - MINOR: 向后兼容的功能增加
// - PATCH: 向后兼容的 bug 修复
//
// 11.2 Changelog 格式
// 每个版本必须有清晰的 Changelog:
//
// ✅ 格式:
// ## [2.0.1] - 2026-02-28
//
// ### Added
// - 新增 Edge 浏览器指纹支持
// - 新增 JA4 指纹计算
//
// ### Changed
// - 优化了指纹匹配性能
// - 改进了异常检测算法
//
// ### Fixed
// - 修复了 User-Agent 生成的 bug
// - 修复了矛盾检测的误报
//
// 11.3 发布检查清单
// 发布新版本前检查:
//
// ✅ 检查项:
//   - 所有测试通过
//   - 性能基准满足要求
//   - 文档已更新
//   - Changelog 已填写
//   - 版本号已更新
//   - 没有已知 bug
//   - 代码审查已通过

// ============================================================================
// 12. 文档管理规范
// ============================================================================
//
// 12.1 CHANGELOG 变更记录
// 所有代码变更必须在 CHANGELOG.md 中记录:
//
// ✅ 必须:
//   - 每次提交功能/修复时更新 CHANGELOG.md
//   - 使用标准分类: Added/Changed/Fixed/Deprecated/Removed/Security
//   - 描述简洁明确，说明"是什么"而非"为什么"
//   - 按时间倒序排列（最新的在最上面）
//
// ❌ 禁止:
//   - 创建单独的"完成报告"、"实施总结"、"阶段报告"等文档
//   - 为每个功能/模块创建单独的变更说明文档
//   - 使用项目阶段命名（如 P0/P1/Week1）作为文档名
//   - 在多个地方重复记录同一变更
//
// 12.2 技术文档创建规范
// 只在以下情况创建文档:
//
// ✅ 允许创建:
//   - API 参考文档（描述公开接口用法）
//   - 架构设计文档（说明系统设计决策）
//   - 开发规范文档（团队协作规则，如本文件）
//   - 快速参考/备忘录（常用代码模式）
//
// ❌ 禁止创建:
//   - 实施计划文档（用 TODO/Issue 管理）
//   - 完成报告文档（用 CHANGELOG 记录）
//   - 总结报告文档（用 Git Commit Message）
//   - 阶段性进度文档（用项目管理工具）
//
// 12.3 文档命名规范
// 技术文档必须使用业务化命名:
//
// ✅ 正确命名:
//   - `ja4s-api-reference.md` - 描述 JA4S 指纹 API
//   - `architecture-overview.md` - 系统架构说明
//   - `fingerprint-project-rules.md` - 项目规范
//
// ❌ 错误命名:
//   - `p0-implementation-plan.md` - 项目阶段命名
//   - `week1-completion-report.md` - 时间阶段命名
//   - `feature-summary.md` - 通用总结文档
//
// 12.4 CHANGELOG 示例
// 正确的变更记录格式:
//
// ✅ 示例:
// ## [未发布]
//
// ### 新增
// - JA4S 指纹分析：TLS ServerHello 指纹计算与异常检测
// - HTTP/2 帧签名分析：帧序列特征识别与客户端匹配
// - JA4H 指纹分析：HTTP 请求头特征与浏览器行为识别
//
// ### 变更
// - 优化指纹匹配算法性能提升 30%
//
// ### 修复
// - 修复 TLS 1.3 握手中的扩展顺序问题

// ============================================================================
// 13. 性能基准规范
// ============================================================================
//
// 13.1 基准测试
// 关键操作必须有基准测试:
//
// ✅ 必测项:
//   - GetRandomFingerprint()
//   - GetUserAgentByProfileName()
//   - GenerateHeaders()
//   - DetectAnomalies()
//   - ComputeJA3ByProfileName()
//   - ComputeJA4ByProfileName()
//
// 13.2 性能目标
// 设定的性能目标:
//
// ✅ 目标值:
//   - GetRandomFingerprint: < 10 μs
//   - GenerateHeaders: < 2 μs
//   - ComputeJA3: < 1 μs
//   - 内存分配: 最小化
//
// 13.3 性能回归测试
// PR 提交时必须验证:
//
// ✅ 检查:
//   - 没有性能回归
//   - 内存使用未增加
//   - 并发性能未下降

// ============================================================================
// END OF FINGERPRINT PROJECT RULES
// ============================================================================
