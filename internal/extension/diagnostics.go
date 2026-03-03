package extension

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// RegistryStatus 注册表状态快照
type RegistryStatus struct {
	Timestamp           time.Time
	TotalExtensions     int
	RegisteredParsers   int
	RegisteredAnalyzers int
	RegisteredHandlers  int
	CustomPlugins       int
	ExtensionDetails    []ExtensionStatusDetail
	LoadedExtensions    []string
	MissingParsers      []string
	MissingAnalyzers    []string
}

// ExtensionStatusDetail 扩展状态详情
type ExtensionStatusDetail struct {
	Type           ExtensionType
	Name           string
	Category       string
	HasParser      bool
	HasAnalyzer    bool
	HandlersCount  int
	IsExperimental bool
	RFC            string
}

// RegistryDiagnostics 扩展注册表诊断工具
type RegistryDiagnostics struct {
	registry *ExtensionRegistry
	mu       sync.RWMutex
}

// NewRegistryDiagnostics 创建诊断工具
func NewRegistryDiagnostics(registry *ExtensionRegistry) *RegistryDiagnostics {
	if registry == nil {
		registry = GetRegistry()
	}
	return &RegistryDiagnostics{
		registry: registry,
	}
}

// GetStatus 获取当前注册表的完整状态快照
func (rd *RegistryDiagnostics) GetStatus() RegistryStatus {
	rd.registry.mu.RLock()
	defer rd.registry.mu.RUnlock()

	status := RegistryStatus{
		Timestamp:           time.Now(),
		TotalExtensions:     len(rd.registry.metadata),
		RegisteredParsers:   len(rd.registry.parsers),
		RegisteredAnalyzers: len(rd.registry.analyzers),
		CustomPlugins:       len(rd.registry.customPlugins),
		ExtensionDetails:    make([]ExtensionStatusDetail, 0),
		LoadedExtensions:    make([]string, 0),
		MissingParsers:      make([]string, 0),
		MissingAnalyzers:    make([]string, 0),
	}

	// 统计处理器
	handlersCount := 0
	for _, handlers := range rd.registry.handlers {
		handlersCount += len(handlers)
	}
	status.RegisteredHandlers = handlersCount

	// 按 type 排序扩展
	types := make([]ExtensionType, 0, len(rd.registry.metadata))
	for extType := range rd.registry.metadata {
		types = append(types, extType)
	}
	sort.Slice(types, func(i, j int) bool {
		return types[i] < types[j]
	})

	// 收集扩展详情
	for _, extType := range types {
		meta := rd.registry.metadata[extType]
		detail := ExtensionStatusDetail{
			Type:           extType,
			Name:           meta.Name,
			Category:       meta.Category,
			HasParser:      rd.registry.parsers[extType] != nil,
			HasAnalyzer:    rd.registry.analyzers[extType] != nil,
			HandlersCount:  len(rd.registry.handlers[extType]),
			IsExperimental: meta.IsExperimental,
			RFC:            meta.RFC,
		}

		status.ExtensionDetails = append(status.ExtensionDetails, detail)

		if detail.HasParser && detail.HasAnalyzer {
			status.LoadedExtensions = append(status.LoadedExtensions, meta.Name)
		}

		// 记录缺失的parser/analyzer
		if !detail.HasParser {
			status.MissingParsers = append(status.MissingParsers, fmt.Sprintf("%s (type:%d)", meta.Name, extType))
		}
		if !detail.HasAnalyzer {
			status.MissingAnalyzers = append(status.MissingAnalyzers, fmt.Sprintf("%s (type:%d)", meta.Name, extType))
		}
	}

	return status
}

// ValidateRequiredExtensions 验证必需的扩展是否已加载
func (rd *RegistryDiagnostics) ValidateRequiredExtensions(required []ExtensionType) (valid bool, missing []ExtensionType) {
	rd.registry.mu.RLock()
	defer rd.registry.mu.RUnlock()

	missing = make([]ExtensionType, 0)
	for _, extType := range required {
		if rd.registry.parsers[extType] == nil || rd.registry.analyzers[extType] == nil {
			missing = append(missing, extType)
		}
	}

	return len(missing) == 0, missing
}

// GetExtensionStatus 获取指定扩展的详细状态
func (rd *RegistryDiagnostics) GetExtensionStatus(extType ExtensionType) (*ExtensionStatusDetail, error) {
	rd.registry.mu.RLock()
	defer rd.registry.mu.RUnlock()

	meta, exists := rd.registry.metadata[extType]
	if !exists {
		return nil, fmt.Errorf("extension type %d not found in registry", extType)
	}

	return &ExtensionStatusDetail{
		Type:           extType,
		Name:           meta.Name,
		Category:       meta.Category,
		HasParser:      rd.registry.parsers[extType] != nil,
		HasAnalyzer:    rd.registry.analyzers[extType] != nil,
		HandlersCount:  len(rd.registry.handlers[extType]),
		IsExperimental: meta.IsExperimental,
		RFC:            meta.RFC,
	}, nil
}

// ListExtensions 获取所有已注册的扩展列表
func (rd *RegistryDiagnostics) ListExtensions() []ExtensionStatusDetail {
	status := rd.GetStatus()
	return status.ExtensionDetails
}

// CountByCategory 按分类统计扩展数量
func (rd *RegistryDiagnostics) CountByCategory() map[string]int {
	rd.registry.mu.RLock()
	defer rd.registry.mu.RUnlock()

	counts := make(map[string]int)
	for _, meta := range rd.registry.metadata {
		category := meta.Category
		if category == "" {
			category = "unknown"
		}
		counts[category]++
	}

	return counts
}

// GetDiagnosticReport 获取完整的诊断报告（文本格式）
func (rd *RegistryDiagnostics) GetDiagnosticReport() string {
	status := rd.GetStatus()

	report := fmt.Sprintf("=== Extension Registry Diagnostic Report ===\n")
	report += fmt.Sprintf("Timestamp: %s\n\n", status.Timestamp.Format(time.RFC3339))

	report += fmt.Sprintf("Summary:\n")
	report += fmt.Sprintf("  Total Extensions Registered: %d\n", status.TotalExtensions)
	report += fmt.Sprintf("  Parsers Ready: %d\n", status.RegisteredParsers)
	report += fmt.Sprintf("  Analyzers Ready: %d\n", status.RegisteredAnalyzers)
	report += fmt.Sprintf("  Handlers Registered: %d\n", status.RegisteredHandlers)
	report += fmt.Sprintf("  Custom Plugins: %d\n\n", status.CustomPlugins)

	if len(status.LoadedExtensions) > 0 {
		report += fmt.Sprintf("Fully Loaded Extensions (%d):\n", len(status.LoadedExtensions))
		for _, name := range status.LoadedExtensions {
			report += fmt.Sprintf("  ✓ %s\n", name)
		}
		report += "\n"
	}

	if len(status.MissingParsers) > 0 {
		report += fmt.Sprintf("Extensions Missing Parsers (%d):\n", len(status.MissingParsers))
		for _, name := range status.MissingParsers {
			report += fmt.Sprintf("  ✗ %s\n", name)
		}
		report += "\n"
	}

	if len(status.MissingAnalyzers) > 0 {
		report += fmt.Sprintf("Extensions Missing Analyzers (%d):\n", len(status.MissingAnalyzers))
		for _, name := range status.MissingAnalyzers {
			report += fmt.Sprintf("  ✗ %s\n", name)
		}
		report += "\n"
	}

	categories := rd.CountByCategory()
	if len(categories) > 0 {
		report += "Extensions by Category:\n"
		for category := range categories {
			report += fmt.Sprintf("  %s: %d\n", category, categories[category])
		}
	}

	return report
}

// HealthCheck 执行健康检查，返回任何发现的问题
func (rd *RegistryDiagnostics) HealthCheck() (healthy bool, issues []string) {
	status := rd.GetStatus()

	issues = make([]string, 0)

	if status.TotalExtensions == 0 {
		issues = append(issues, "no extensions registered")
	}

	if len(status.MissingParsers) > 0 {
		issues = append(issues, fmt.Sprintf("%d extension(s) missing parser", len(status.MissingParsers)))
	}

	if len(status.MissingAnalyzers) > 0 {
		issues = append(issues, fmt.Sprintf("%d extension(s) missing analyzer", len(status.MissingAnalyzers)))
	}

	return len(issues) == 0, issues
}
