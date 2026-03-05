package extension

// Component 所有可治理组件的基础接口
//
// 遵循单一职责原则，每个组件只定义最小必要的行为
// 具体组件可通过组合来扩展功能
//
// 使用示例：
//
//	var comp Component = myParser
//	info := comp.GetInfo()
type Component interface {
	// GetInfo 获取组件信息
	GetInfo() ComponentInfo
}

// ComponentInfo 组件信息
type ComponentInfo struct {
	// 组件名称
	Name string

	// 组件版本
	Version string

	// 组件描述
	Description string

	// 组件作者
	Author string
}

// Auditable 可审计的组件接口
//
// 遵循 Go 的隐式接口设计，组件无需显式声明实现
type Auditable interface {
	// RecordEvent 记录事件
	// 返回: 错误信息
	RecordEvent(eventType, severity, message string, details map[string]interface{}) error
}

// Closeable 支持资源清理的组件接口
//
// 任何需要释放资源的组件都可实现此接口
// 与 io.Closer 一致，便于使用
type Closeable interface {
	// Close 关闭组件并释放资源
	Close() error
}

// Initializable 支持初始化的组件接口
//
// 用于延迟初始化和复杂的设置过程
type Initializable interface {
	// Initialize 初始化组件
	// config: 初始化配置
	// 返回: 错误信息
	Initialize(config map[string]interface{}) error

	// IsInitialized 检查是否已初始化
	IsInitialized() bool
}

// Identifiable 可识别的组件接口
//
// 任何有唯一标识符的对象都应实现此接口
type Identifiable interface {
	// GetID 获取唯一标识符
	GetID() string
}
