package useragent

import fp "github.com/vistone/fingerprint"

// UserAgentGenerator User-Agent 生成器。
type UserAgentGenerator = fp.UserAgentGenerator

// OperatingSystem 操作系统类型。
type OperatingSystem = fp.OperatingSystem

// ClientProfile 客户端指纹配置。
type ClientProfile = fp.ClientProfile

// NewGenerator 创建 User-Agent 生成器。
func NewGenerator() *UserAgentGenerator {
	return fp.NewUserAgentGenerator()
}

// RandomOS 随机选择操作系统。
func RandomOS() OperatingSystem {
	return fp.RandomOS()
}

// GetByProfileName 根据 profile 名称获取 User-Agent。
func GetByProfileName(profileName string) (string, error) {
	return fp.GetUserAgentByProfileName(profileName)
}

// GetByProfileNameWithOS 根据 profile 名称和指定系统获取 User-Agent。
func GetByProfileNameWithOS(profileName string, os OperatingSystem) (string, error) {
	return fp.GetUserAgentByProfileNameWithOS(profileName, os)
}

// GetFromProfile 从 profile 获取 User-Agent。
func GetFromProfile(profile ClientProfile) (string, error) {
	return fp.GetUserAgentFromProfile(profile)
}

// GetFromProfileWithOS 从 profile 和指定系统获取 User-Agent。
func GetFromProfileWithOS(profile ClientProfile, os OperatingSystem) (string, error) {
	return fp.GetUserAgentFromProfileWithOS(profile, os)
}
