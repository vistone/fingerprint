package fingerprint

import "strings"

const clientHelloSpecNotImplementedMsg = "please implement this method"

// IsClientHelloSpecNotImplemented 判断错误是否表示 profile 未实现 ClientHelloSpec。
func IsClientHelloSpecNotImplemented(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), clientHelloSpecNotImplementedMsg)
}
