package fingerprint

import (
	"fmt"

	tls "github.com/bogdanfinn/utls"
	tj "github.com/vistone/fingerprint/tls/ja4"
)

// JA4Result JA4 指纹结果（兼容别名）。
type JA4Result = tj.JA4Result

// JA4Signature JA4 指纹签名输入（兼容别名）。
type JA4Signature = tj.JA4Signature

// 当第一次调用 ja4 相关函数时，注册 MappedTLSClients
var ja4Initialized = false

func ensureJA4Initialized() {
	if !ja4Initialized {
		tj.InitMappedTLSClients(MappedTLSClients)
		ja4Initialized = true
	}
}

// ComputeJA4FromSpec 从 TLS ClientHello 规范计算 JA4 指纹（兼容入口）。
func ComputeJA4FromSpec(spec tls.ClientHelloSpec) (*JA4Result, error) {
	ensureJA4Initialized()
	return tj.ComputeJA4FromSpec(spec)
}

// ComputeJA4FromProfile 从 ClientProfile 计算 JA4 指纹（兼容入口）。
func ComputeJA4FromProfile(profile ClientProfile) (*JA4Result, error) {
	ensureJA4Initialized()
	spec, err := profile.GetClientHelloSpec()
	if err != nil {
		return nil, err
	}
	return ComputeJA4FromSpec(spec)
}

// ComputeJA4ByProfileName 根据指纹名称计算 JA4 指纹（兼容入口）。
func ComputeJA4ByProfileName(profileName string) (*JA4Result, error) {
	ensureJA4Initialized()
	profile, ok := MappedTLSClients[profileName]
	if !ok {
		return nil, fmt.Errorf("指纹 %s 不存在", profileName)
	}
	return ComputeJA4FromProfile(profile)
}
