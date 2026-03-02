// Package fingerprint 提供 JA3 指纹识别
package fingerprint

import (
	"fmt"

	tls "github.com/bogdanfinn/utls"
	tj "github.com/vistone/fingerprint/tls/ja3"
)

// JA3Result JA3 指纹结果（兼容别名）。
type JA3Result = tj.JA3Result

// 当第一次调用 ja3 相关函数时，注册 MappedTLSClients
var ja3Initialized = false

func ensureJA3Initialized() {
	if !ja3Initialized {
		// 使用反射将根包的 ClientProfile map 传递给子包
		// 这避免了类型不兼容问题
		tj.InitMappedTLSClientsRaw(MappedTLSClients)
		ja3Initialized = true
	}
}

// ComputeJA3FromSpec 从 TLS ClientHello 规范计算 JA3 指纹（兼容入口）。
func ComputeJA3FromSpec(spec tls.ClientHelloSpec) (*JA3Result, error) {
	ensureJA3Initialized()
	return tj.ComputeJA3FromSpec(spec)
}

// ComputeJA3FromProfile 从 ClientProfile 计算 JA3 指纹（兼容入口）。
func ComputeJA3FromProfile(profile ClientProfile) (*JA3Result, error) {
	ensureJA3Initialized()
	// 直接实现而不是转发，以避免类型问题
	spec, err := profile.GetClientHelloSpec()
	if err != nil {
		return nil, err
	}
	return ComputeJA3FromSpec(spec)
}

// ComputeJA3ByProfileName 根据指纹名称计算 JA3 指纹（兼容入口）。
func ComputeJA3ByProfileName(profileName string) (*JA3Result, error) {
	// 直接在根包中使用 MappedTLSClients，避免子包的访问问题
	profile, ok := MappedTLSClients[profileName]
	if !ok {
		return nil, fmt.Errorf("指纹 %s 不存在", profileName)
	}
	return ComputeJA3FromProfile(profile)
}

// MatchJA3 检查两个 JA3 哈希是否匹配（兼容入口）。
func MatchJA3(hash1, hash2 string) bool {
	return tj.MatchJA3(hash1, hash2)
}

// FindProfileByJA3 根据 JA3 哈希查找匹配的 ClientProfile 名称（兼容入口）。
func FindProfileByJA3(ja3Hash string) []string {
	ensureJA3Initialized()
	return tj.FindProfileByJA3(ja3Hash)
}
