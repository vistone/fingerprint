package policy

import fp "github.com/vistone/fingerprint"

// PermissionDirective 权限指令。
type PermissionDirective = fp.PermissionDirective

// PermissionsPolicy Permissions-Policy 解析结果。
type PermissionsPolicy = fp.PermissionsPolicy

// PermissionsPolicyAnalyzer 权限策略分析器。
type PermissionsPolicyAnalyzer = fp.PermissionsPolicyAnalyzer

// NewPermissionsPolicyAnalyzer 创建权限策略分析器。
func NewPermissionsPolicyAnalyzer() *PermissionsPolicyAnalyzer {
	return fp.NewPermissionsPolicyAnalyzer()
}
