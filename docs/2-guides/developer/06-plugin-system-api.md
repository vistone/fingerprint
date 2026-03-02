// 插件系统 API 参考文档
//
// 本文档描述了 fingerprint 项目的插件系统外部接口和使用方法。
// 插件系统支持第三方指纹模块动态加载，提供完整的贡献者 SDK。
//
// 目录：
// 1. 核心概念
// 2. SDK 快速开始
// 3. 指纹构建 API
// 4. 插件加载 API
// 5. 插件注册 API
// 6. 格式规范
// 7. 验证规则
// 8. 示例代码
// 9. 常见问题
//
// ============================================================================
// 1. 核心概念
// ============================================================================
//
// 插件系统由以下核心组件组成：
//
// - Plugin 接口：插件的基础接口，所有插件都必须实现此接口
// - Registry：插件注册表，管理所有已注册的插件
// - Loader：插件加载器，从文件系统加载插件
// - SDK：贡献者开发工具包，简化指纹开发流程
//
// 指纹分为以下几种分类：
//
// - CategoryBrowser：浏览器指纹（Chrome、Firefox 等）
// - CategoryMobile：移动设备指纹（iOS、Android 等）
// - CategoryBot：机器人指纹（爬虫、脚本等）
// - CategoryCustom：自定义指纹
// - CategoryCommunity：社区贡献指纹
//
// ============================================================================
// 2. SDK 快速开始
// ============================================================================
//
// 创建 SDK 实例：
//
//     import "github.com/vistone/fingerprint/internal/contrib"
//
//     sdk, err := contrib.NewSDK(nil) // 使用默认配置
//     if err != nil {
//         log.Fatal(err)
//     }
//
// 使用默认配置时，SDK 会：
// - 从 $HOME/.fingerprint/plugins 加载插件
// - 启用缓存以提高性能
// - 自动验证插件格式
// - 从环境变量 FINGERPRINT_PLUGIN_DIR 读取自定义目录
//
// 自定义 SDK 配置：
//
//     config := &contrib.SDKConfig{
//         PluginDir:     "./my_plugins",
//         EnableCache:   true,
//         AutoLoad:      true,
//         VerifyPlugins: true,
//         SearchPaths: []string{
//             "./my_plugins",
//             "./community_plugins",
//         },
//         AllowRemote:   false,
//         MaxPluginSize: 1024 * 1024, // 1MB
//     }
//
//     sdk, _ := contrib.NewSDK(config)
//
// ============================================================================
// 3. 指纹构建 API
// ============================================================================
//
// 使用 Builder 创建指纹：
//
//     builder := contrib.NewBuilder("my_chrome_fingerprint")
//     
//     plugin, err := builder.
//         WithDisplayName("Chrome 133").
//         WithCategory(plugins.CategoryBrowser).
//         WithBrowser("Chrome", "133", "Windows NT 10.0").
//         WithUserAgent("Mozilla/5.0 (Windows NT 10.0...)").
//         WithTLSVersion(0x0304). // TLS 1.3
//         WithCipherSuites([]uint16{0x1301, 0x1302}).
//         WithExtensions([]uint16{0x0000, 0x000d, 0x0033}).
//         WithEllipticCurves([]uint16{0x001d, 0x0017}).
//         WithAuthor("Your Name", "your@email.com").
//         WithLicense("BSD 3-Clause").
//         WithTags([]string{"browser", "chrome"}).
//         Build()
//     
//     if err != nil {
//         log.Fatal(err)
//     }
//
// Builder 方法列表：
//
// - WithDisplayName(string)            # 显示名称
// - WithDescription(string)            # 描述
// - WithCategory(Category)             # 分类
// - WithBrowser(name, version, os)     # 浏览器信息
// - WithVersion(string)                # 版本
// - WithUserAgent(string)              # User-Agent
// - WithTLSVersion(uint16)             # TLS 版本
// - WithCipherSuites([]uint16)         # 密码套件列表
// - WithExtensions([]uint16)           # TLS 扩展列表
// - WithEllipticCurves([]uint16)       # 椭圆曲线
// - WithSignatureAlgorithms([]uint16)  # 签名算法
// - WithSupportedVersions([]uint16)    # 支持的 TLS 版本
// - WithKeyShareCurves([]uint16)       # 密钥共享曲线
// - WithAuthor(name, email)            # 作者信息
// - WithLicense(string)                # 许可证
// - WithRepository(string)             # 源代码仓库
// - WithTags([]string)                 # 标签
// - WithKeywords([]string)             # 关键词
// - WithMobile(bool)                   # 是否为移动设备
// - WithExtension(key, value)          # 自定义扩展数据
// - Build()                            # 构建插件
// - SaveToFile(path)                   # 保存为 JSON 文件
// - SaveAndRegister(path)              # 保存并注册
//
// ============================================================================
// 4. 插件加载 API
// ============================================================================
//
// 加载插件：
//
//     // 从单个文件加载
//     sdk.LoadFromFile("path/to/fingerprint.json")
//
//     // 从目录加载
//     sdk.LoadFromDirectory("path/to/plugins")
//
//     // 按名称获取插件
//     plugin, err := sdk.GetPlugin("chrome_133")
//
//     // 搜索插件
//     results := sdk.SearchPlugin("chrome")
//
// SDK 方法列表：
//
// - LoadPlugins()                      # 从配置目录加载所有插件
// - LoadFromDirectory(path)            # 从指定目录加载
// - LoadFromFile(path)                 # 从文件加载
// - GetPlugin(id)                      # 获取插件
// - RegisterPlugin(id, plugin)         # 注册插件
// - UnregisterPlugin(id)               # 注销插件
// - ListPlugins()                      # 列出所有插件
// - ListByCategory(category)           # 按分类列出
// - SearchPlugin(query)                # 搜索插件
// - ValidatePlugin(path)               # 验证插件文件
// - GetStats()                         # 获取统计信息
// - SetDefault(id)                     # 设置默认插件
// - GetDefault()                       # 获取默认插件
// - SetAlias(alias, target)            # 设置别名
// - RemoveAlias(alias)                 # 移除别名
//
// ============================================================================
// 5. 插件注册 API
// ============================================================================
//
// 在全局注册表注册插件：
//
//     // 方式 1: 通过 SDK
//     sdk.RegisterPlugin("my_fingerprint", plugin)
//
//     // 方式 2: 通过全局函数
//     plugins.RegisterPlugin("my_fingerprint", plugin, plugins.SourceCommunity)
//
// PluginSource 类型定义插件来源：
//
// - SourceBuiltin：内置插件
// - SourceLocal：本地加载的插件
// - SourceRegistry：从远程注册表加载的插件
// - SourceCommunity：社区贡献的插件
//
// ============================================================================
// 6. 格式规范
// ============================================================================
//
// 指纹文件采用 JSON 格式，结构如下：
//
// {
//   "metadata": {
//     "name": "chrome_133",                    // 指纹标识符（必需）
//     "display_name": "Chrome 133",            // 显示名称
//     "description": "Chrome 133 TLS fingerprint",
//     "category": "browser",                   // 分类（必需）
//     "version": "133.0.0.0",                  // 版本
//     "format": "2.0",                         // 格式版本
//     "browser": "Chrome",                     // 浏览器名称
//     "browser_version": "133.0.0.0",          // 浏览器版本
//     "os": "Windows NT 10.0",                 // 操作系统
//     "is_mobile": false,                      // 是否移动设备
//     "author": "John Doe",                    // 作者
//     "email": "john@example.com",             // 邮箱
//     "license": "BSD 3-Clause",               // 许可证
//     "repository": "https://github.com/...", // 源代码仓库
//     "created_at": "2026-02-28T...",         // 创建时间 (RFC3339)
//     "updated_at": "2026-02-28T...",         // 更新时间 (RFC3339)
//     "tags": ["browser", "chrome", "tls"],    // 标签
//     "keywords": ["chrome", "fingerprint"]    // 关键词
//   },
//   "client_hello": {                          // TLS ClientHello 规范（必需）
//     "tls_version": 772,                      // 0x0304 = TLS 1.3
//     "cipher_suites": [4866, 4865, ...],      // 密码套件
//     "extensions": [0, 23, ...],              // TLS 扩展
//     "elliptic_curves": [29, 23, 24],         // 椭圆曲线
//     "ec_point_formats": [0],                 // EC 点格式
//     "signature_algorithms": [2052, ...],     // 签名算法
//     "supported_versions": [772, 771],        // 支持的 TLS 版本
//     "key_share_curves": [29, 23]             // 密钥共享曲线
//   },
//   "http2": {                                 // HTTP/2 配置（可选）
//     "header_table_size": 65536,
//     "max_concurrent_streams": 1000,
//     "initial_window_size": 65535,
//     "frame_order": ["SETTINGS", "HEADERS", "PRIORITY"]
//   },
//   "user_agent": "Mozilla/5.0...",            // User-Agent（必需）
//   "extensions": {                            // 自定义扩展数据
//     "ja3": "...",
//     "ja4": "...",
//     "custom_field": "value"
//   }
// }
//
// ============================================================================
// 7. 验证规则
// ============================================================================
//
// 插件系统内置以下验证规则：
//
// 1. 元数据验证
//    - name 字段是必需的
//    - category 字段是必需的
//
// 2. ClientHello 验证
//    - client_hello 字段是必需的
//    - tls_version 字段字段是必需的
//    - cipher_suites 列表不能为空
//
// 3. User-Agent 验证
//    - user_agent 字段是必需的
//
// 4. TLS 版本验证
//    - 支持 TLS 1.2, 1.3 及以上
//
// 5. 密码套件验证
//    - 密码套件数量不超过 1000
//
// 自定义验证规则：
//
//     registry := plugins.GetRegistry()
//
//     // 添加自定义验证规则
//     registry.AddValidationRule(func(data *plugins.FingerprintData) error {
//         if data.Metadata.Author == "" {
//             return fmt.Errorf("author is required")
//         }
//         return nil
//     })
//
// ============================================================================
// 8. 示例代码
// ============================================================================
//
// 示例 1: 创建并保存指纹
//
//     import (
//         "github.com/vistone/fingerprint/internal/contrib"
//         "github.com/vistone/fingerprint/internal/plugins"
//     )
//
//     builder := contrib.ExampleChrome133()
//     plugin, _ := builder.Build()
//
//     // 保存到文件
//     builder.SaveToFile("chrome_133.json")
//
//     // 或者保存并注册
//     plugin, _ := builder.SaveAndRegister("chrome_133.json")
//
// 示例 2: 加载和使用插件
//
//     sdk, _ := contrib.NewSDK(nil)
//     plugin, _ := sdk.GetPlugin("chrome_133")
//
//     // 转换为 TLS ClientHelloSpec
//     spec, _ := plugin.ToClientHelloSpec()
//
//     // 使用 User-Agent
//     ua := plugin.GetUserAgent()
//
// 示例 3: 列出和搜索指纹
//
//     sdk, _ := contrib.NewSDK(nil)
//
//     // 列出所有浏览器指纹
//     browsers := sdk.ListByCategory(plugins.CategoryBrowser)
//     for id, info := range browsers {
//         fmt.Printf("%s: %s\n", id, info.Meta.DisplayName)
//     }
//
//     // 搜索指纹
//     results := sdk.SearchPlugin("chrome")
//     for _, result := range results {
//         fmt.Printf("%s: %s\n", result.ID, result.Meta.DisplayName)
//     }
//
// ============================================================================
// 9. 常见问题
// ============================================================================
//
// Q: 如何创建新的指纹？
// A: 使用 contrib.NewBuilder() 创建构建器，链式调用方法设置属性，
//    然后调用 Build() 方法。参见示例 1。
//
// Q: 如何保存指纹到文件？
// A: 调用 Builder.SaveToFile(path) 方法。会生成标准的 JSON 格式文件。
//
// Q: 如何加载已保存的指纹？
// A: 使用 SDK.LoadFromFile(path) 或 SDK.LoadFromDirectory(path)。
//
// Q: 支持哪些 TLS 版本？
// A: 系统支持 TLS 1.0、1.1、1.2、1.3 及以上版本。
//
// Q: 如何验证指纹格式？
// A: 使用 SDK.ValidatePlugin(filePath) 方法验证 JSON 文件格式。
//
// Q: 如何自定义验证규则？
// A: 通过 Registry.AddValidationRule() 添加自定义验证逻辑。
//
// Q: 指纹文件的最大大小是多少？
// A: 默认为 1MB，可通过 SDKConfig.MaxPluginSize 配置。
//
// Q: 支持远程插件吗？
// A: 目前主要支持本地文件加载。远程加载功能计划在后续版本实现。
//
// Q: 如何设置默认指纹？
// A: 使用 SDK.SetDefault(id) 或 Registry.SetDefault(id)。
//
// Q: 如何为指纹设置别名？
// A: 使用 SDK.SetAlias(alias, targetId)。
//
// Q: 环境变量有哪些？
// A: FINGERPRINT_PLUGIN_DIR：自定义插件目录
//
package contrib
