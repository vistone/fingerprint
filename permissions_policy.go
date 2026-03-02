package fingerprint

import (
	pp "github.com/vistone/fingerprint/http/policy"
)

// PermissionDirective 权限指令（兼容别名）。
type PermissionDirective = pp.PermissionDirective

// PermissionsPolicy 权限策略（兼容别名）。
type PermissionsPolicy = pp.PermissionsPolicy

// PermissionsPolicyAnalyzer 权限策略分析器（兼容别名）。
type PermissionsPolicyAnalyzer = pp.PermissionsPolicyAnalyzer

// NewPermissionsPolicyAnalyzer 创建新的权限策略分析器（兼容入口）。
func NewPermissionsPolicyAnalyzer() *PermissionsPolicyAnalyzer {
	return pp.NewPermissionsPolicyAnalyzer()
}
