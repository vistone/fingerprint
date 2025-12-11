package fingerprint

import (
	"github.com/vistone/fingerprint/profiles"
)

// ClientProfile 是 profiles.ClientProfile 的类型别名
// 提供主包的统一接口
type ClientProfile = profiles.ClientProfile

// DefaultClientProfile 默认客户端指纹配置（Chrome 133）
var DefaultClientProfile = profiles.DefaultClientProfile

// MappedTLSClients 所有可用的 TLS 客户端指纹映射表
// 可以通过字符串名称快速获取对应的指纹配置
var MappedTLSClients = profiles.MappedTLSClients

// NewClientProfile 创建一个新的客户端指纹配置
// 这是 profiles.NewClientProfile 的重新导出
var NewClientProfile = profiles.NewClientProfile

