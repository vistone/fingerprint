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
// Anti-detection engine API
// =====================================================================

// handleAntiDetectStatus returns anti-detection engine status.
func (h *Handler) handleAntiDetectStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := h.gateway.GetConfig()
	injector := h.gateway.GetInjector()
	pm := h.gateway.GetProfileManager()

	status := map[string]interface{}{
		"enabled":       cfg.AntiDetectEnabled,
		"profileId":     cfg.AntiDetectProfileID,
		"configDir":     cfg.AntiDetectConfigDir,
		"proxyTarget":   cfg.AntiDetectProxyTarget,
		"directProxy":   cfg.AntiDetectDirectProxy,
		"injectConsist": cfg.AntiDetectInjectConsist,
		"injectorReady": injector != nil,
	}

	if pm != nil {
		profileList := pm.ListProfiles()
		status["availableProfiles"] = profileList
		status["profileCount"] = len(profileList)
	}

	// List available code generators.
	status["generators"] = []map[string]interface{}{
		{"id": "webgpu", "name": "WebGPU Override", "description": "Overrides navigator.gpu APIs to match target browser capabilities"},
		{"id": "media_devices", "name": "MediaDevices Override", "description": "Mocks enumerateDevices() with stable virtual device lists"},
		{"id": "permissions", "name": "Permissions API Override", "description": "Intercepts navigator.permissions.query() with consistent permission states"},
		{"id": "automation", "name": "Automation Hiding", "description": "Hides webdriver / __selenium / callPhantom style automation markers"},
		{"id": "cross_layer", "name": "Cross-Layer Consistency", "description": "Injects JS consistency checks across navigator / screen / canvas and TLS/HTTP layers"},
		{"id": "full", "name": "Full Anti-Detection", "description": "Combines all generators into a full anti-detection script"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleAntiDetectPreview previews anti-detection JavaScript.
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

	// Resolve profile.
	var profile *profiles.ClientProfile
	if req.ProfileID != "" {
		if p, ok := h.findProfile(req.ProfileID); ok {
			profile = &p
		}
	}
	if profile == nil {
		// Fall back to default profile.
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

// handleAntiDetectInjectTest tests HTML injection behavior.
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

// handleAntiDetectSDKPreview previews SDK JavaScript output.
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
// Plugin system API
// =====================================================================

// handlePluginsInfo returns plugin registry information.
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
				"description": "Analyzer plugin: runs custom analysis logic on fingerprint data",
				"icon":        "🔬",
			},
			{
				"type":        "transformer",
				"name":        "Transformer",
				"description": "Transformer plugin: converts, normalizes, or enriches fingerprint data",
				"icon":        "🔄",
			},
			{
				"type":        "exporter",
				"name":        "Exporter",
				"description": "Exporter plugin: ships results to external systems (Elasticsearch, Kafka, etc.)",
				"icon":        "📤",
			},
			{
				"type":        "validator",
				"name":        "Validator",
				"description": "Validator plugin: validates fingerprint integrity and correctness",
				"icon":        "✅",
			},
		},
		"extensionArchitecture": map[string]interface{}{
			"pipeline":    "Parser → Analyzer → Handler",
			"description": "Extension system uses a 3-stage pipeline: parse raw data -> analyze features -> handle outputs",
			"interfaces": []map[string]interface{}{
				{"name": "Parser", "method": "Parse(data []byte) (ExtensionData, error)", "description": "Parses raw byte data"},
				{"name": "Analyzer", "method": "Analyze(data ExtensionData) (AnalysisResult, error)", "description": "Analyzes structured extension data"},
				{"name": "Handler", "method": "Handle(event ExtensionEvent) (EventResult, error)", "description": "Processes events and produces outputs"},
			},
		},
		"registrationAPI": []map[string]interface{}{
			{"function": "RegisterExtension(metadata)", "description": "Registers extension metadata"},
			{"function": "RegisterParser(type, parser)", "description": "Registers a parser for an extension type"},
			{"function": "RegisterAnalyzer(type, analyzer)", "description": "Registers an analyzer for an extension type"},
			{"function": "RegisterHandler(type, handler)", "description": "Registers a handler for an extension type"},
			{"function": "RegisterPlugin(name, plugin)", "description": "Registers a full plugin"},
			{"function": "GetPlugin(name)", "description": "Retrieves a registered plugin"},
			{"function": "LoadPlugins(configPath)", "description": "Loads plugins from config file"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
