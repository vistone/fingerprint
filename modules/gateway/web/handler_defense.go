package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/vistone/fingerprint/modules/frontend"
	"github.com/vistone/fingerprint/modules/plugin"
	"github.com/vistone/fingerprint/modules/profiles"
)

// =====================================================================
// 反检测引擎 API
// =====================================================================

// handleAntiDetectStatus 返回反检测引擎状态
func (h *Handler) handleAntiDetectStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := h.gateway.GetConfig()
	injector := h.gateway.GetInjector()
	pm := h.gateway.GetProfileManager()

	status := map[string]interface{}{
		"enabled":       cfg.P3Enabled,
		"profileId":     cfg.P3ProfileID,
		"configDir":     cfg.P3ConfigDir,
		"proxyTarget":   cfg.P3ProxyTarget,
		"directProxy":   cfg.P3DirectProxy,
		"injectConsist": cfg.P3InjectConsist,
		"injectorReady": injector != nil,
	}

	if pm != nil {
		profileList := pm.ListProfiles()
		status["availableProfiles"] = profileList
		status["profileCount"] = len(profileList)
	}

	// 列出可用的代码生成器
	status["generators"] = []map[string]interface{}{
		{"id": "webgpu", "name": "WebGPU Override", "description": "重写 navigator.gpu API 以匹配目标浏览器的 WebGPU 能力"},
		{"id": "media_devices", "name": "MediaDevices Override", "description": "伪造 enumerateDevices() 返回一致的虚拟设备列表"},
		{"id": "permissions", "name": "Permissions API Override", "description": "拦截 navigator.permissions.query() 返回一致的权限状态"},
		{"id": "automation", "name": "Automation Hiding", "description": "隐藏 webdriver / __selenium / callPhantom 等自动化标记"},
		{"id": "cross_layer", "name": "Cross-Layer Consistency", "description": "注入 JS 层一致性校验 — 确保 navigator / screen / canvas 与 TLS/HTTP 层一致"},
		{"id": "full", "name": "Full Anti-Detection", "description": "组合以上全部生成器的完整反检测代码"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleAntiDetectPreview 预览反检测代码
func (h *Handler) handleAntiDetectPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProfileID string `json:"profileId"`
		Generator string `json:"generator"` // webgpu, media_devices, permissions, automation, cross_layer, full
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 获取 profile
	var profile *profiles.ClientProfile
	if req.ProfileID != "" {
		if p, ok := h.findProfile(req.ProfileID); ok {
			profile = &p
		}
	}
	if profile == nil {
		// 使用默认 profile
		pm := h.gateway.GetProfileManager()
		if pm != nil {
			var err error
			profile, err = pm.GetDefaultProfile()
			if err != nil {
				http.Error(w, "No default profile available", http.StatusServiceUnavailable)
				return
			}
		}
	}
	if profile == nil {
		http.Error(w, "No profile available", http.StatusServiceUnavailable)
		return
	}

	gen := frontend.NewJSAntiDetectCodeGenerator(profile)

	var code string
	var generatorName string
	switch strings.ToLower(req.Generator) {
	case "webgpu":
		code = gen.GenerateWebGPUCode()
		generatorName = "WebGPU Override"
	case "media_devices":
		code = gen.GenerateMediaDevicesCode()
		generatorName = "MediaDevices Override"
	case "permissions":
		code = gen.GeneratePermissionsCode()
		generatorName = "Permissions API Override"
	case "automation":
		code = gen.GenerateAutomationCode()
		generatorName = "Automation Hiding"
	case "cross_layer":
		code = gen.GenerateCrossLayerConsistencyCode()
		generatorName = "Cross-Layer Consistency"
	case "full":
		code = gen.GenerateFullAntiDetectionCode()
		generatorName = "Full Anti-Detection"
	default:
		code = gen.GenerateFullAntiDetectionCode()
		generatorName = "Full Anti-Detection"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"generator":   generatorName,
		"generatorId": req.Generator,
		"profileId":   profile.ID,
		"profileName": profile.Name,
		"code":        code,
		"codeLength":  len(code),
	})
}

// handleAntiDetectInjectTest 测试 HTML 注入
func (h *Handler) handleAntiDetectInjectTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		HTML      string `json:"html"`
		ProfileID string `json:"profileId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.HTML == "" {
		req.HTML = `<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body><h1>Hello World</h1></body>
</html>`
	}

	injector := h.gateway.GetInjector()
	if injector == nil {
		http.Error(w, "Anti-detection injector not enabled", http.StatusServiceUnavailable)
		return
	}

	injected := injector.InjectIntoHTML([]byte(req.HTML))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"originalLength": len(req.HTML),
		"injectedLength": len(injected),
		"injected":       string(injected),
		"deltaBytes":     len(injected) - len(req.HTML),
	})
}

// handleAntiDetectSDKPreview 预览 SDK JavaScript
func (h *Handler) handleAntiDetectSDKPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sdk := h.gateway.GetSDK()
	if sdk == nil {
		http.Error(w, "SDK not available", http.StatusServiceUnavailable)
		return
	}

	coreJS := sdk.GenerateJSCore()
	injectorJS := sdk.GenerateJSInjector("/api/v1/collect")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"coreJS": map[string]interface{}{
			"code":   coreJS,
			"length": len(coreJS),
		},
		"injectorJS": map[string]interface{}{
			"code":   injectorJS,
			"length": len(injectorJS),
		},
	})
}

// =====================================================================
// 插件系统 API
// =====================================================================

// handlePluginsInfo 返回插件注册表信息
func (h *Handler) handlePluginsInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	registryStats := plugin.GetRegistryStats()

	info := map[string]interface{}{
		"registry": registryStats,
		"pluginTypes": []map[string]interface{}{
			{
				"type":        "analyzer",
				"name":        "Analyzer",
				"description": "分析插件 — 对指纹数据执行自定义分析逻辑",
				"icon":        "🔬",
			},
			{
				"type":        "transformer",
				"name":        "Transformer",
				"description": "转换插件 — 转换、标准化或增强指纹数据格式",
				"icon":        "🔄",
			},
			{
				"type":        "exporter",
				"name":        "Exporter",
				"description": "导出插件 — 将结果导出到外部系统 (Elasticsearch, Kafka, etc.)",
				"icon":        "📤",
			},
			{
				"type":        "validator",
				"name":        "Validator",
				"description": "验证插件 — 检验指纹数据完整性和有效性",
				"icon":        "✅",
			},
		},
		"extensionArchitecture": map[string]interface{}{
			"pipeline":    "Parser → Analyzer → Handler",
			"description": "扩展系统采用三阶段管道: 解析原始数据 → 分析提取特征 → 处理生成结果",
			"interfaces": []map[string]interface{}{
				{"name": "Parser", "method": "Parse(data []byte) (ExtensionData, error)", "description": "解析原始字节数据"},
				{"name": "Analyzer", "method": "Analyze(data ExtensionData) (AnalysisResult, error)", "description": "分析结构化数据"},
				{"name": "Handler", "method": "Handle(event ExtensionEvent) (EventResult, error)", "description": "处理事件和生成输出"},
			},
		},
		"registrationAPI": []map[string]interface{}{
			{"function": "RegisterExtension(metadata)", "description": "注册新扩展元数据"},
			{"function": "RegisterParser(type, parser)", "description": "注册解析器到指定扩展类型"},
			{"function": "RegisterAnalyzer(type, analyzer)", "description": "注册分析器到指定扩展类型"},
			{"function": "RegisterHandler(type, handler)", "description": "注册处理器到指定扩展类型"},
			{"function": "RegisterPlugin(name, plugin)", "description": "注册完整插件"},
			{"function": "GetPlugin(name)", "description": "获取已注册插件"},
			{"function": "LoadPlugins(configPath)", "description": "从配置文件加载插件"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
