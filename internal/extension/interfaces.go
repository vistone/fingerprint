package extension

import "context"

// Parser 扩展解析接口
// 负责将原始字节数据解析为结构化的 ExtensionData
type Parser interface {
	// Parse 解析扩展数据
	// data: 原始扩展数据（不包含扩展头）
	// parentContext: 父级上下文（如完整的 ClientHello）
	// 返回: 解析后的 ExtensionData，错误信息
	Parse(data []byte, parentContext context.Context) (ExtensionData, error)

	// GetType 返回此解析器负责的扩展类型
	GetType() ExtensionType

	// GetVersion 返回解析器版本
	GetVersion() string
}

// Analyzer 扩展分析接口
// 负责分析 ExtensionData 并生成 AnalysisResult
type Analyzer interface {
	// Analyze 分析扩展数据
	// data: 已解析的 ExtensionData
	// config: 分析配置（可选）
	// 返回: 分析结果，错误信息
	Analyze(data ExtensionData, config map[string]interface{}) (AnalysisResult, error)

	// GetType 返回此分析器负责的扩展类型
	GetType() ExtensionType

	// GetVersion 返回分析器版本
	GetVersion() string

	// SupportsConfig 返回支持的配置键列表
	SupportsConfig() []string
}

// Handler 扩展处理接口
// 事件驱动的扩展处理，用于流式处理和中间件
type Handler interface {
	// Handle 处理扩展事件
	// event: 扩展事件
	// 返回: 处理结果，错误信息
	Handle(event *ExtensionEvent) (*EventResult, error)

	// GetType 返回此处理器负责的扩展类型
	GetType() ExtensionType

	// GetPriority 返回处理器优先级（高优先级先执行）
	// 0-100，默认 50
	GetPriority() int

	// GetName 返回处理器名称
	GetName() string
}

// Plugin 第三方插件接口
// 允许外部开发者扩展指纹库功能
type Plugin interface {
	// GetInfo 获取插件信息
	GetInfo() *PluginInfo

	// Init 初始化插件
	// config: 插件配置
	// 返回: 错误信息
	Init(config map[string]interface{}) error

	// Register 注册插件的扩展
	// 此方法应调用 RegisterExtension, RegisterParser 等
	Register() error

	// Unload 卸载插件
	Unload() error

	// Validate 验证插件有效性
	Validate() error

	// GetDependencies 返回插件依赖列表
	GetDependencies() []string

	// GetVersion 返回插件版本
	GetVersion() string
}

// ExtensionEvent 扩展事件
type ExtensionEvent struct {
	// 事件类型
	Type string // "parse", "analyze", "transform"

	// 扩展类型
	ExtensionType ExtensionType

	// 事件数据
	Data interface{}

	// 事件发生时的上下文
	Context context.Context

	// 事件元数据
	Metadata map[string]interface{}

	// 时间戳
	Timestamp int64
}

// EventResult 事件处理结果
type EventResult struct {
	// 是否成功处理
	Success bool

	// 错误消息
	Error string

	// 处理结果数据
	Result interface{}

	// 是否继续传递给下一个处理器
	ContinueProcessing bool

	// 处理器名称
	HandlerName string
}

// PluginInfo 插件信息
type PluginInfo struct {
	// 插件 ID（唯一标识）
	ID string

	// 插件名称
	Name string

	// 插件描述
	Description string

	// 插件版本
	Version string

	// 作者
	Author string

	// 联系方式
	Contact string

	// License
	License string

	// 主页
	Homepage string

	// 最小 SDK 版本
	MinSDKVersion string

	// 最大 SDK 版本
	MaxSDKVersion string
}

// Transform 数据转换接口
// 用于扩展的链式转换
type Transform interface {
	// Transform 进行转换
	// input: 输入数据
	// 返回: 转换后的数据，错误信息
	Transform(input interface{}) (interface{}, error)

	// GetName 返回转换名称
	GetName() string

	// GetInputType 返回输入类型名称
	GetInputType() string

	// GetOutputType 返回输出类型名称
	GetOutputType() string
}

// Validator 数据验证接口
// 用于扩展数据的校验
type Validator interface {
	// Validate 验证数据
	// value: 要验证的值
	// 返回: 是否有效，错误信息
	Validate(value interface{}) (bool, error)

	// GetName 返回验证器名称
	GetName() string
}

// ComparerComparable protocol for extension data comparison
// 用于扩展间的比较和差异检测
type Comparer interface {
	// Compare 比较两个数据
	// data1, data2: 要比较的数据
	// 返回: 相似度（0.0-1.0），差异列表
	Compare(data1, data2 interface{}) (float64, []string, error)

	// GetName 返回比较器名称
	GetName() string
}
