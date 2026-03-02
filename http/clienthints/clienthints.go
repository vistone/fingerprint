package clienthints

import fp "github.com/vistone/fingerprint"

// ClientHintsPolicy Client Hints 策略配置。
type ClientHintsPolicy = fp.ClientHintsPolicy

// ClientHintsData 完整的 Client Hints 数据。
type ClientHintsData = fp.ClientHintsData

// ClientProfile 客户端指纹配置。
type ClientProfile = fp.ClientProfile

// BrowserType 浏览器类型。
type BrowserType = fp.BrowserType

// NewPolicy 创建默认策略。
func NewPolicy(browserType BrowserType) *ClientHintsPolicy {
	return fp.NewClientHintsPolicy(browserType)
}

// GenerateFromProfile 从 profile 生成 Client Hints。
func GenerateFromProfile(profile *ClientProfile, policy *ClientHintsPolicy) *ClientHintsData {
	return fp.GenerateClientHintsFromProfile(profile, policy)
}
