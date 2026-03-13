// handler_advanced.go — 高级功能 API 端点
// 分析引擎 / ML引擎 / 防御系统 / 反检测引擎 / 插件系统 / 指纹工具
package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/defense"
	"github.com/vistone/fingerprint/modules/frontend"
	"github.com/vistone/fingerprint/modules/gateway"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/plugin"
	"github.com/vistone/fingerprint/modules/profiles"
	tlsmod "github.com/vistone/fingerprint/modules/tls"
)

// =====================================================================
// 分析引擎 API — 通过 Profile 运行完整分析管线
// =====================================================================

// handleAnalyzeProfile 基于 Profile 执行完整指纹分析
func (h *Handler) handleAnalyzeProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProfileID string `json:"profileId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if req.ProfileID == "" {
		http.Error(w, "profileId required", http.StatusBadRequest)
		return
	}

	// 查找 profile
	var profile profiles.ClientProfile
	found := false
	h.mu.RLock()
	for _, p := range h.profiles {
		if p.ID == req.ProfileID {
			profile = p
			found = true
			break
		}
	}
	h.mu.RUnlock()
	if !found {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	// 构建 AnalyzeRequest
	analyzeReq := &gateway.AnalyzeRequest{
		TLSVersion:      profile.TLSVersion,
		CipherSuites:    profile.CipherSuites,
		Extensions:      profile.Extensions,
		SupportedCurves: profile.SupportedCurves,
		Headers:         profile.Headers,
		Method:          "GET",
		Path:            "/",
		ClientIP:        "admin-console",
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.gateway.Analyze(ctx, analyzeReq)
	if err != nil {
		http.Error(w, "Analysis failed", http.StatusInternalServerError)
		return
	}

	// 构建丰富的结果
	result := map[string]interface{}{
		"profile": map[string]interface{}{
			"id":      profile.ID,
			"name":    profile.Name,
			"browser": profile.BrowserType,
			"version": profile.BrowserVersion,
			"os":      profile.OS,
		},
		"fingerprintHash":  resp.FingerprintHash,
		"cached":           resp.Cached,
		"processingTimeMs": resp.ProcessingTimeMs,
	}

	// 分类结果
	if resp.Classification != nil {
		result["classification"] = map[string]interface{}{
			"protocol":           resp.Classification.Protocol,
			"family":             resp.Classification.Family,
			"version":            resp.Classification.Version,
			"confidence":         resp.Classification.Confidence,
			"protocolConfidence": resp.Classification.ProtocolConfidence,
			"familyConfidence":   resp.Classification.FamilyConfidence,
			"versionConfidence":  resp.Classification.VersionConfidence,
			"labels":             resp.Classification.Labels,
		}
	}

	// 风险评估
	if resp.RiskAssessment != nil {
		factors := make([]map[string]interface{}, 0, len(resp.RiskAssessment.Factors))
		for _, f := range resp.RiskAssessment.Factors {
			factors = append(factors, map[string]interface{}{
				"name":        f.Name,
				"weight":      f.Weight,
				"description": f.Description,
			})
		}
		result["riskAssessment"] = map[string]interface{}{
			"level":       resp.RiskAssessment.Level,
			"score":       resp.RiskAssessment.Score,
			"factors":     factors,
			"suggestions": resp.RiskAssessment.Suggestions,
		}
	}

	// JA3/JA4/JA4H
	if resp.JA3 != nil {
		result["ja3"] = resp.JA3
	}
	if resp.JA4 != nil {
		result["ja4"] = resp.JA4
	}
	if resp.JA4H != nil {
		result["ja4h"] = resp.JA4H
	}

	// 检测发现
	result["findings"] = resp.Findings
	result["defenseHints"] = resp.DefenseHints

	// Agent 决策
	if resp.AgentDecision != nil {
		result["agentDecision"] = resp.AgentDecision
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// =====================================================================
// ML 引擎 API
// =====================================================================

// handleMLInfo 返回 ML 分类器模型信息
func (h *Handler) handleMLInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	classifier := h.gateway.GetClassifier()
	p, f, v := classifier.GetConfidenceThresholds()

	info := map[string]interface{}{
		"architecture": "3-Layer Hierarchical Classifier",
		"layers": []map[string]interface{}{
			{
				"name":        "Protocol Classifier",
				"description": "协议类型分类 (TLS / HTTP / HTTP2 / QUIC / HTTP3)",
				"level":       1,
				"threshold":   p,
				"weight":      0.3,
			},
			{
				"name":        "Family Classifier",
				"description": "浏览器家族识别 (Chrome / Firefox / Safari / Edge / Opera)",
				"level":       2,
				"threshold":   f,
				"weight":      0.3,
			},
			{
				"name":        "Version Classifier",
				"description": "具体版本识别 (Chrome 134 / Firefox 135 / Safari 18 ...)",
				"level":       3,
				"threshold":   v,
				"weight":      0.4,
			},
		},
		"featureTypes": []map[string]interface{}{
			{"name": "tls_version", "category": "TLS", "description": "TLS 协议版本号"},
			{"name": "cipher_suites", "category": "TLS", "description": "密码套件数量"},
			{"name": "extensions", "category": "TLS", "description": "TLS 扩展数量"},
			{"name": "http2_settings", "category": "HTTP", "description": "HTTP/2 设置帧哈希"},
			{"name": "http_headers", "category": "HTTP", "description": "HTTP 头部特征哈希"},
			{"name": "user_agent", "category": "HTTP", "description": "User-Agent 哈希"},
			{"name": "canvas", "category": "Frontend", "description": "Canvas 指纹哈希"},
			{"name": "webgl", "category": "Frontend", "description": "WebGL 渲染器指纹"},
			{"name": "audio", "category": "Frontend", "description": "AudioContext 指纹"},
			{"name": "fonts", "category": "Frontend", "description": "字体列表指纹"},
			{"name": "storage", "category": "Frontend", "description": "存储 API 指纹"},
			{"name": "webrtc", "category": "Frontend", "description": "WebRTC 配置指纹"},
			{"name": "hardware", "category": "Frontend", "description": "硬件信息指纹"},
			{"name": "timing", "category": "Frontend", "description": "时间精度指纹"},
			{"name": "headless_browser", "category": "Detection", "description": "无头浏览器检测"},
			{"name": "entropy", "category": "Detection", "description": "信息熵"},
			{"name": "tool_marker", "category": "Detection", "description": "自动化工具标记"},
			{"name": "behavior_pattern", "category": "Detection", "description": "行为一致性模式"},
		},
		"status": "trained",
	}

	// 添加 MLService 信息
	if svc := h.gateway.GetMLService(); svc != nil {
		st := svc.Stats()
		info["mlService"] = map[string]interface{}{
			"enabled":       true,
			"ready":         svc.IsReady(),
			"inferCount":    st.InferCount,
			"feedbackCount": st.FeedbackCount,
			"evolveCount":   st.EvolveCount,
			"modelReady":    st.ModelReady,
			"modelVersions": st.ModelVersions,
		}
		if st.LearnerStats != nil {
			info["learner"] = map[string]interface{}{
				"totalSamples":    st.LearnerStats.TotalSamples,
				"bufferFilled":    st.LearnerStats.BufferFilled,
				"peakAccuracy":    st.LearnerStats.PeakAccuracy,
				"recentAccuracy":  st.LearnerStats.RecentAccuracy,
				"driftDetected":   st.LearnerStats.DriftDetected,
				"driftEventCount": st.LearnerStats.DriftEventCount,
			}
		}
	} else {
		info["mlService"] = map[string]interface{}{"enabled": false}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// handleMLExtract 从 Profile 提取特征向量
func (h *Handler) handleMLExtract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProfileID string `json:"profileId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	profile, ok := h.findProfile(req.ProfileID)
	if !ok {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	extractor := h.gateway.GetExtractor()
	fv := extractor.ExtractFromProfile(&profile)

	// 转换特征列表
	features := make([]map[string]interface{}, 0, len(fv.Features))
	for ft, val := range fv.Features {
		features = append(features, map[string]interface{}{
			"name":  string(ft),
			"value": val,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"profileId":   req.ProfileID,
		"profileName": profile.Name,
		"features":    features,
		"totalCount":  len(features),
	})
}

// handleMLClassify 对 Profile 执行 ML 分类
func (h *Handler) handleMLClassify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProfileID string `json:"profileId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	profile, ok := h.findProfile(req.ProfileID)
	if !ok {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	extractor := h.gateway.GetExtractor()
	classifier := h.gateway.GetClassifier()

	fv := extractor.ExtractFromProfile(&profile)
	result := classifier.Classify(fv)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"profileId":   req.ProfileID,
		"profileName": profile.Name,
		"browser":     profile.BrowserType,
		"version":     profile.BrowserVersion,
		"classification": map[string]interface{}{
			"protocol":           result.Protocol,
			"family":             result.Family,
			"version":            result.Version,
			"confidence":         result.Confidence,
			"protocolConfidence": result.ProtocolConfidence,
			"familyConfidence":   result.FamilyConfidence,
			"versionConfidence":  result.VersionConfidence,
			"labels":             result.Labels,
		},
		"isHighConfidence": result.IsHighConfidence(),
	})
}

// handleMLBatch 批量分类所有 Profile
func (h *Handler) handleMLBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.mu.RLock()
	snapshot := append([]profiles.ClientProfile(nil), h.profiles...)
	h.mu.RUnlock()

	extractor := h.gateway.GetExtractor()
	classifier := h.gateway.GetClassifier()

	results := make([]map[string]interface{}, 0, len(snapshot))
	protocolDist := make(map[string]int)
	familyDist := make(map[string]int)
	highConfCount := 0

	for _, p := range snapshot {
		fv := extractor.ExtractFromProfile(&p)
		cr := classifier.Classify(fv)

		protocolDist[string(cr.Protocol)]++
		familyDist[string(cr.Family)]++
		if cr.IsHighConfidence() {
			highConfCount++
		}

		results = append(results, map[string]interface{}{
			"id":         p.ID,
			"name":       p.Name,
			"browser":    p.BrowserType,
			"version":    p.BrowserVersion,
			"protocol":   cr.Protocol,
			"family":     cr.Family,
			"mlVersion":  cr.Version,
			"confidence": cr.Confidence,
			"highConf":   cr.IsHighConfidence(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total":                len(results),
		"highConfidenceRate":   float64(highConfCount) / float64(max(len(results), 1)) * 100,
		"protocolDistribution": protocolDist,
		"familyDistribution":   familyDist,
		"results":              results,
	})
}

// =====================================================================
// 防御系统 API
// =====================================================================

// handleDefenseRules 返回所有检测规则
func (h *Handler) handleDefenseRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从 Detector 获取规则描述
	rules := []map[string]interface{}{
		{
			"name":        "headless_browser",
			"description": "检测无头浏览器 (Puppeteer, Playwright, Selenium WebDriver)",
			"feature":     "headless_browser",
			"threshold":   0.5,
			"riskScore":   0.7,
			"severity":    "high",
			"category":    "automation",
		},
		{
			"name":        "high_entropy",
			"description": "检测异常高信息熵值 — 指纹随机化或伪造工具",
			"feature":     "entropy",
			"threshold":   10.0,
			"riskScore":   0.5,
			"severity":    "medium",
			"category":    "anomaly",
		},
		{
			"name":        "automation_tool",
			"description": "检测自动化工具标记 (webdriver, __selenium, callPhantom)",
			"feature":     "tool_marker",
			"threshold":   0.3,
			"riskScore":   0.8,
			"severity":    "critical",
			"category":    "automation",
		},
		{
			"name":        "inconsistent_behavior",
			"description": "指纹行为模式不一致 — TLS/HTTP/JS 层信号矛盾",
			"feature":     "behavior_pattern",
			"threshold":   0.3,
			"riskScore":   0.6,
			"severity":    "medium",
			"category":    "consistency",
		},
	}

	// 从 Agent 取策略补充
	strategies := []map[string]interface{}{}
	if a := h.gateway.GetAgent(); a != nil {
		for _, s := range a.GetActiveStrategies() {
			strategies = append(strategies, map[string]interface{}{
				"id":       s.ID,
				"name":     s.Name,
				"action":   s.Action,
				"threat":   s.ThreatClass,
				"priority": s.Priority,
				"learned":  s.Learned,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"detectionRules":  rules,
		"agentStrategies": strategies,
		"riskLevels": []map[string]interface{}{
			{"level": "none", "threshold": 0, "color": "#4caf50"},
			{"level": "low", "threshold": 0.2, "color": "#8bc34a"},
			{"level": "medium", "threshold": 0.4, "color": "#ff9800"},
			{"level": "high", "threshold": 0.7, "color": "#f44336"},
			{"level": "critical", "threshold": 0.9, "color": "#9c27b0"},
		},
	})
}

// handleDefenseDetect 对 Profile 执行威胁检测
func (h *Handler) handleDefenseDetect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProfileID string `json:"profileId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	profile, ok := h.findProfile(req.ProfileID)
	if !ok {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	extractor := h.gateway.GetExtractor()
	fv := extractor.ExtractFromProfile(&profile)

	// 运行检测
	detector := defense.NewDetector()
	detection := detector.Detect(fv)

	// 运行风险评估
	classifier := h.gateway.GetClassifier()
	classification := classifier.Classify(fv)
	riskEngine := h.gateway.GetRiskEngine()
	risk := riskEngine.Evaluate(fv, classification)

	// 获取防御建议
	defenseSystem := defense.NewDefenseSystem()
	advice := defenseSystem.Analyze(fv, classification)

	findings := make([]map[string]interface{}, 0, len(detection.Findings))
	for _, f := range detection.Findings {
		findings = append(findings, map[string]interface{}{
			"rule":        f.Rule,
			"description": f.Description,
			"riskScore":   f.RiskScore,
		})
	}

	factors := make([]map[string]interface{}, 0)
	if risk != nil {
		for _, f := range risk.Factors {
			factors = append(factors, map[string]interface{}{
				"name":        f.Name,
				"weight":      f.Weight,
				"description": f.Description,
			})
		}
	}

	result := map[string]interface{}{
		"profileId":   req.ProfileID,
		"profileName": profile.Name,
		"detection": map[string]interface{}{
			"isThreat":  detection.IsThreat(),
			"riskScore": detection.RiskScore,
			"riskLevel": detection.RiskLevel,
			"findings":  findings,
			"labels":    detection.Labels,
		},
	}

	if risk != nil {
		result["riskAssessment"] = map[string]interface{}{
			"level":       risk.Level,
			"score":       risk.Score,
			"factors":     factors,
			"suggestions": risk.Suggestions,
		}
	}

	if advice != nil {
		result["defenseAdvice"] = map[string]interface{}{
			"detection":  advice.Detection,
			"risk":       advice.Risk,
			"protection": advice.Protection,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

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

// =====================================================================
// 指纹工具 API
// =====================================================================

// handleToolsJA3 计算 JA3 指纹
func (h *Handler) handleToolsJA3(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProfileID    string   `json:"profileId"`
		TLSVersion   uint16   `json:"tlsVersion"`
		CipherSuites []uint16 `json:"cipherSuites"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var spec core.ClientHelloSpec
	if req.ProfileID != "" {
		if p, ok := h.findProfile(req.ProfileID); ok {
			spec = core.ClientHelloSpec{
				TLSVersion:      p.TLSVersion,
				CipherSuites:    p.CipherSuites,
				Extensions:      p.Extensions,
				SupportedCurves: p.SupportedCurves,
			}
		} else {
			http.Error(w, "Profile not found", http.StatusNotFound)
			return
		}
	} else {
		spec = core.ClientHelloSpec{
			TLSVersion:   req.TLSVersion,
			CipherSuites: req.CipherSuites,
		}
	}

	ja3 := tlsmod.CalculateJA3(spec)
	ja4 := tlsmod.CalculateJA4(spec)

	result := map[string]interface{}{
		"ja3": map[string]interface{}{
			"hash": ja3.Hash,
			"raw":  ja3.RawString,
		},
		"ja4": map[string]interface{}{
			"fingerprint": ja4.Fingerprint,
		},
		"input": map[string]interface{}{
			"tlsVersion":   spec.TLSVersion,
			"cipherSuites": len(spec.CipherSuites),
			"extensions":   len(spec.Extensions),
			"curves":       len(spec.SupportedCurves),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleToolsValidate 验证 Profile 完整性
func (h *Handler) handleToolsValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProfileID string `json:"profileId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	profile, ok := h.findProfile(req.ProfileID)
	if !ok {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	validator := profiles.NewProfileValidator()
	result := validator.Validate(profile)

	// TCP/IP 验证
	tcpipResult := ""
	if profile.TCPIP != nil {
		tcpipResult = profiles.ValidateTCPIP(profile.TCPIP)
	}

	// Header 验证
	headerResult := map[string]interface{}{}
	if profile.Headers != nil {
		hvr := profiles.ValidateHeaders(profile.Headers)
		headerResult["missing"] = hvr.Missing
		headerResult["empty"] = hvr.Empty
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"profileId":   req.ProfileID,
		"profileName": profile.Name,
		"validation": map[string]interface{}{
			"valid":         result.Valid,
			"errors":        result.Errors,
			"warnings":      result.Warnings,
			"missingFields": result.MissingFields,
		},
		"tcpipValidation":  tcpipResult,
		"headerValidation": headerResult,
	})
}

// handleToolsCompare 比较两个 Profile
func (h *Handler) handleToolsCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProfileA string `json:"profileA"`
		ProfileB string `json:"profileB"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	profileA, okA := h.findProfile(req.ProfileA)
	profileB, okB := h.findProfile(req.ProfileB)
	if !okA || !okB {
		http.Error(w, "One or both profiles not found", http.StatusNotFound)
		return
	}

	// 提取特征并计算相似度
	extractor := h.gateway.GetExtractor()
	fvA := extractor.ExtractFromProfile(&profileA)
	fvB := extractor.ExtractFromProfile(&profileB)

	similarity := calculateSimilarity(fvA, fvB)

	// 详细比较
	comparison := map[string]interface{}{
		"a": map[string]interface{}{
			"id": profileA.ID, "name": profileA.Name,
			"browser": profileA.BrowserType, "version": profileA.BrowserVersion,
			"os":         profileA.OS,
			"tlsVersion": profileA.TLSVersion,
			"ciphers":    len(profileA.CipherSuites),
			"extensions": len(profileA.Extensions),
		},
		"b": map[string]interface{}{
			"id": profileB.ID, "name": profileB.Name,
			"browser": profileB.BrowserType, "version": profileB.BrowserVersion,
			"os":         profileB.OS,
			"tlsVersion": profileB.TLSVersion,
			"ciphers":    len(profileB.CipherSuites),
			"extensions": len(profileB.Extensions),
		},
		"similarity": similarity,
		"diffs":      buildProfileDiffs(profileA, profileB),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comparison)
}

// =====================================================================
// Helper functions
// =====================================================================

func (h *Handler) findProfile(id string) (profiles.ClientProfile, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, p := range h.profiles {
		if p.ID == id {
			return p, true
		}
	}
	return profiles.ClientProfile{}, false
}

func calculateSimilarity(a, b *core.FeatureVector) float64 {
	if a == nil || b == nil {
		return 0
	}
	// 收集所有 feature keys
	keys := make(map[core.FeatureType]bool)
	for k := range a.Features {
		keys[k] = true
	}
	for k := range b.Features {
		keys[k] = true
	}
	if len(keys) == 0 {
		return 1.0
	}

	matches := 0
	for k := range keys {
		va := a.Get(k)
		vb := b.Get(k)
		if va == vb {
			matches++
		} else if va != 0 && vb != 0 {
			// 相对误差 < 10% 算匹配
			ratio := va / vb
			if ratio < 0 {
				ratio = -ratio
			}
			if ratio > 0.9 && ratio < 1.1 {
				matches++
			}
		}
	}
	return float64(matches) / float64(len(keys))
}

func buildProfileDiffs(a, b profiles.ClientProfile) []map[string]interface{} {
	diffs := []map[string]interface{}{}

	if a.TLSVersion != b.TLSVersion {
		diffs = append(diffs, map[string]interface{}{
			"field": "TLS Version", "a": a.TLSVersion, "b": b.TLSVersion,
		})
	}
	if len(a.CipherSuites) != len(b.CipherSuites) {
		diffs = append(diffs, map[string]interface{}{
			"field": "Cipher Suites Count", "a": len(a.CipherSuites), "b": len(b.CipherSuites),
		})
	}
	if len(a.Extensions) != len(b.Extensions) {
		diffs = append(diffs, map[string]interface{}{
			"field": "Extensions Count", "a": len(a.Extensions), "b": len(b.Extensions),
		})
	}
	if string(a.BrowserType) != string(b.BrowserType) {
		diffs = append(diffs, map[string]interface{}{
			"field": "Browser", "a": a.BrowserType, "b": b.BrowserType,
		})
	}
	if string(a.OS) != string(b.OS) {
		diffs = append(diffs, map[string]interface{}{
			"field": "OS", "a": a.OS, "b": b.OS,
		})
	}
	if a.HTTP2Settings.InitialWindowSize != b.HTTP2Settings.InitialWindowSize {
		diffs = append(diffs, map[string]interface{}{
			"field": "H2 InitialWindowSize", "a": a.HTTP2Settings.InitialWindowSize, "b": b.HTTP2Settings.InitialWindowSize,
		})
	}
	if a.HTTP2Settings.MaxConcurrentStreams != b.HTTP2Settings.MaxConcurrentStreams {
		diffs = append(diffs, map[string]interface{}{
			"field": "H2 MaxConcurrentStreams", "a": a.HTTP2Settings.MaxConcurrentStreams, "b": b.HTTP2Settings.MaxConcurrentStreams,
		})
	}
	if len(a.PseudoHeaderOrder) > 0 && len(b.PseudoHeaderOrder) > 0 {
		aOrder := strings.Join(a.PseudoHeaderOrder, ",")
		bOrder := strings.Join(b.PseudoHeaderOrder, ",")
		if aOrder != bOrder {
			diffs = append(diffs, map[string]interface{}{
				"field": "Pseudo Header Order", "a": aOrder, "b": bOrder,
			})
		}
	}

	// TCP/IP 比较
	if a.TCPIP != nil && b.TCPIP != nil {
		if a.TCPIP.TTL != b.TCPIP.TTL {
			diffs = append(diffs, map[string]interface{}{
				"field": "TCP TTL", "a": a.TCPIP.TTL, "b": b.TCPIP.TTL,
			})
		}
		if a.TCPIP.WindowSize != b.TCPIP.WindowSize {
			diffs = append(diffs, map[string]interface{}{
				"field": "TCP Window Size", "a": a.TCPIP.WindowSize, "b": b.TCPIP.WindowSize,
			})
		}
	}

	return diffs
}

// =====================================================================
// MLService — 中央 AI 服务 API
// =====================================================================

// handleMLServiceStats 返回 MLService 统计信息
func (h *Handler) handleMLServiceStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	svc := h.gateway.GetMLService()
	if svc == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"enabled": false,
			"message": "MLService is not enabled. Set MLServiceEnabled=true in gateway config.",
		})
		return
	}

	st := svc.Stats()
	result := map[string]interface{}{
		"enabled":       true,
		"ready":         svc.IsReady(),
		"inferCount":    st.InferCount,
		"feedbackCount": st.FeedbackCount,
		"evolveCount":   st.EvolveCount,
		"modelReady":    st.ModelReady,
		"modelVersions": st.ModelVersions,
	}
	if st.LearnerStats != nil {
		result["learner"] = map[string]interface{}{
			"totalSamples":    st.LearnerStats.TotalSamples,
			"bufferFilled":    st.LearnerStats.BufferFilled,
			"peakAccuracy":    st.LearnerStats.PeakAccuracy,
			"recentAccuracy":  st.LearnerStats.RecentAccuracy,
			"driftDetected":   st.LearnerStats.DriftDetected,
			"driftEventCount": st.LearnerStats.DriftEventCount,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleMLServiceHealth 返回 MLService 健康状态
func (h *Handler) handleMLServiceHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	svc := h.gateway.GetMLService()
	if svc == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "disabled",
			"healthy": false,
		})
		return
	}

	st := svc.Stats()
	healthy := svc.IsReady()
	status := "ready"
	if !healthy {
		status = "not_trained"
	}

	components := []map[string]interface{}{
		{"name": "Pipeline", "ready": svc.IsReady()},
		{"name": "Store", "ready": svc.Store() != nil},
		{"name": "Learner", "ready": st.LearnerStats != nil},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     status,
		"healthy":    healthy,
		"components": components,
	})
}

// handleMLServiceInfer 对指定 Profile 执行 MLService 推理
func (h *Handler) handleMLServiceInfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	svc := h.gateway.GetMLService()
	if svc == nil {
		http.Error(w, "MLService not enabled", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		ProfileID string `json:"profileId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	profile, ok := h.findProfile(req.ProfileID)
	if !ok {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	result := svc.Infer(&profile, nil)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"profileId":   req.ProfileID,
		"profileName": profile.Name,
		"browser": map[string]interface{}{
			"family":     result.Browser.Family,
			"confidence": result.Browser.Confidence,
		},
		"forgery": map[string]interface{}{
			"forgeryProb": result.Forgery.ForgeryProb,
			"forgeryType": result.Forgery.ForgeryType,
		},
		"crossLayerDim": len(result.CrossFeatures),
		"embeddingDim":  len(result.Embedding),
	})
}

// handleMLServiceValidate 使用 MLService 验证指纹是否真实
func (h *Handler) handleMLServiceValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	svc := h.gateway.GetMLService()
	if svc == nil {
		http.Error(w, "MLService not enabled", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		ProfileID string `json:"profileId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	profile, ok := h.findProfile(req.ProfileID)
	if !ok {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	vr := svc.Validate(&profile)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"profileId":        req.ProfileID,
		"profileName":      profile.Name,
		"valid":            vr.Valid,
		"forgeryProb":      vr.ForgeryProb,
		"forgeryType":      vr.ForgeryType.String(),
		"consistencyScore": vr.ConsistencyScore,
		"browserFamily":    vr.BrowserFamily,
		"confidence":       vr.Confidence,
		"suggestions":      vr.Suggestions,
	})
}

// handleMLServiceGenerate 使用 MLService 生成 ML 引导的指纹
func (h *Handler) handleMLServiceGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	svc := h.gateway.GetMLService()
	if svc == nil {
		http.Error(w, "MLService not enabled", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		TargetBrowser  string  `json:"targetBrowser"`
		TargetOS       string  `json:"targetOS"`
		MaxAttempts    int     `json:"maxAttempts"`
		NoiseIntensity float64 `json:"noiseIntensity"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	genCfg := &ml.GenerateConfig{
		TargetBrowser:  req.TargetBrowser,
		TargetOS:       req.TargetOS,
		MaxAttempts:    req.MaxAttempts,
		NoiseIntensity: req.NoiseIntensity,
	}
	if genCfg.MaxAttempts <= 0 {
		genCfg.MaxAttempts = 10
	}
	if genCfg.NoiseIntensity <= 0 {
		genCfg.NoiseIntensity = 0.05
	}

	result, err := svc.Generate(genCfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"attempts":        result.Attempts,
		"sourceProfileID": result.SourceProfileID,
	}
	if result.Profile != nil {
		resp["profile"] = map[string]interface{}{
			"id":      result.Profile.ID,
			"name":    result.Profile.Name,
			"browser": result.Profile.BrowserType,
			"version": result.Profile.BrowserVersion,
			"os":      result.Profile.OS,
		}
	}
	if result.Validation != nil {
		resp["validation"] = map[string]interface{}{
			"valid":            result.Validation.Valid,
			"forgeryProb":      result.Validation.ForgeryProb,
			"consistencyScore": result.Validation.ConsistencyScore,
			"browserFamily":    result.Validation.BrowserFamily,
			"confidence":       result.Validation.Confidence,
			"suggestions":      result.Validation.Suggestions,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleMLServiceEvolve 触发 MLService 增量进化
func (h *Handler) handleMLServiceEvolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	svc := h.gateway.GetMLService()
	if svc == nil {
		http.Error(w, "MLService not enabled", http.StatusServiceUnavailable)
		return
	}

	h.trainingMu.Lock()
	if h.trainingActive {
		h.trainingMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Training already in progress (" + h.trainingPhase + ")",
		})
		return
	}
	h.trainingActive = true
	h.trainingPhase = "evolve"
	h.trainingStart = time.Now()
	h.trainingErr = ""
	h.trainingDone = false
	h.trainingResult = nil
	h.trainingMu.Unlock()

	registry := profiles.NewProfileRegistry()
	for _, p := range profiles.GetAll() {
		registry.Register(p)
	}

	go func() {
		metrics, err := svc.Evolve(registry)
		h.trainingMu.Lock()
		defer h.trainingMu.Unlock()
		h.trainingActive = false
		h.trainingDone = true
		if err != nil {
			h.trainingErr = err.Error()
		} else {
			h.trainingResult = map[string]interface{}{
				"phase":   "evolve",
				"success": true,
				"metrics": map[string]interface{}{
					"epoch":       metrics.Epoch,
					"valAccuracy": metrics.ValAccuracy,
					"encoderLoss": metrics.EncoderLoss,
					"classLoss":   metrics.ClassLoss,
					"forgeryLoss": metrics.ForgeryLoss,
					"threatLoss":  metrics.ThreatLoss,
					"forgeryAUC":  metrics.ForgeryAUC,
				},
			}
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Evolution started asynchronously",
	})
}

// handleMLServiceTrain 触发 MLService 全量训练（异步，通过 Python GPU 脚本）
func (h *Handler) handleMLServiceTrain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	svc := h.gateway.GetMLService()
	if svc == nil {
		http.Error(w, "MLService not enabled", http.StatusServiceUnavailable)
		return
	}

	h.trainingMu.Lock()
	if h.trainingActive {
		h.trainingMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Training already in progress (" + h.trainingPhase + ")",
		})
		return
	}
	h.trainingActive = true
	h.trainingPhase = "gpu-train"
	h.trainingStart = time.Now()
	h.trainingErr = ""
	h.trainingDone = false
	h.trainingResult = nil
	h.trainingMu.Unlock()

	go func() {
		err := h.runGPUTraining(svc)
		h.trainingMu.Lock()
		defer h.trainingMu.Unlock()
		h.trainingActive = false
		h.trainingDone = true
		if err != nil {
			h.trainingErr = err.Error()
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "GPU training started asynchronously (240 profiles)",
	})
}

// profileSampleJSON represents one profile sample for the Python training script.
type profileSampleJSON struct {
	Features    []float64 `json:"features"`
	FamilyLabel int       `json:"family_label"`
	BrowserType string    `json:"browser_type"`
}

// exportProfileFeatures encodes all profiles to a JSON file for the Python GPU trainer.
func exportProfileFeatures(allProfiles []profiles.ClientProfile, outputPath string) (int, error) {
	labelMap := map[core.BrowserType]int{
		core.BrowserChrome:  0,
		core.BrowserFirefox: 1,
		core.BrowserSafari:  2,
		core.BrowserEdge:    3,
		core.BrowserOpera:   4,
		core.BrowserBrave:   5,
		core.BrowserSamsung: 6,
	}

	var samples []profileSampleJSON
	for i := range allProfiles {
		p := &allProfiles[i]
		features := ml.EncodeFingerprint(p)
		label, ok := labelMap[p.BrowserType]
		if !ok {
			label = 0
		}
		samples = append(samples, profileSampleJSON{
			Features:    features,
			FamilyLabel: label,
			BrowserType: string(p.BrowserType),
		})
	}

	data, err := json.Marshal(map[string]interface{}{"samples": samples})
	if err != nil {
		return 0, fmt.Errorf("marshal profile features: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0750); err != nil {
		return 0, fmt.Errorf("create dir: %w", err)
	}
	if err := os.WriteFile(outputPath, data, 0600); err != nil {
		return 0, fmt.Errorf("write profile features: %w", err)
	}
	return len(samples), nil
}

// runGPUTraining executes the Python GPU training script as a subprocess.
func (h *Handler) runGPUTraining(svc *ml.MLService) error {
	allProfiles := profiles.GetAll()

	// Paths
	inputPath := "/tmp/ml_train_input.json"
	outputPath := "/models/weights.json"
	progressPath := "/tmp/ml_training_progress.json"

	// Phase 1: Export profile features
	h.trainingMu.Lock()
	h.trainingPhase = "exporting profiles"
	h.trainingMu.Unlock()

	nProfiles, err := exportProfileFeatures(allProfiles, inputPath)
	if err != nil {
		return fmt.Errorf("export profiles: %w", err)
	}

	// Clean up input file when done
	defer os.Remove(inputPath)
	defer os.Remove(progressPath)

	// Phase 2: Run Python GPU training
	h.trainingMu.Lock()
	h.trainingPhase = fmt.Sprintf("gpu-training (%d profiles)", nProfiles)
	h.trainingMu.Unlock()

	// Determine training script path
	scriptPath := "/app/gpu_train.py"
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		// Fallback for local development
		scriptPath = "training/gpu_train.py"
	}

	cmd := exec.Command("python3", scriptPath,
		"--input", inputPath,
		"--output", outputPath,
		"--progress", progressPath,
		"--epochs", "200",
	)
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

	// Capture stdout/stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gpu training failed: %w\nOutput: %s", err, string(output))
	}

	// Phase 3: Load weights into pipeline
	h.trainingMu.Lock()
	h.trainingPhase = "loading weights"
	h.trainingMu.Unlock()

	if err := svc.Pipeline().LoadWeights(outputPath); err != nil {
		return fmt.Errorf("load weights: %w", err)
	}

	// Read final progress for result
	var finalProgress map[string]interface{}
	if progressData, err := os.ReadFile(progressPath); err == nil {
		json.Unmarshal(progressData, &finalProgress)
	}

	st := svc.Stats()
	result := map[string]interface{}{
		"phase":         "gpu-train",
		"success":       true,
		"modelReady":    st.ModelReady,
		"modelVersions": st.ModelVersions,
		"profiles":      nProfiles,
		"output":        string(output),
	}
	if finalProgress != nil {
		result["gpuProgress"] = finalProgress
	}

	h.trainingMu.Lock()
	h.trainingResult = result
	h.trainingMu.Unlock()

	return nil
}

// handleMLServiceTrainingStatus 查询异步训练/进化状态
func (h *Handler) handleMLServiceTrainingStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.trainingMu.Lock()
	active := h.trainingActive
	phase := h.trainingPhase
	start := h.trainingStart
	done := h.trainingDone
	errMsg := h.trainingErr
	result := h.trainingResult
	h.trainingMu.Unlock()

	resp := map[string]interface{}{
		"active": active,
		"phase":  phase,
		"done":   done,
	}
	if !start.IsZero() {
		resp["elapsed"] = time.Since(start).Seconds()
		resp["startTime"] = start.Format(time.RFC3339)
	}
	if errMsg != "" {
		resp["error"] = errMsg
	}
	if result != nil {
		resp["result"] = result
	}

	// Read GPU training progress file if available
	if active {
		if progressData, err := os.ReadFile("/tmp/ml_training_progress.json"); err == nil {
			var gpuProgress map[string]interface{}
			if json.Unmarshal(progressData, &gpuProgress) == nil {
				resp["gpuProgress"] = gpuProgress
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleMLServiceFeedback 提交反馈样本给在线学习系统
func (h *Handler) handleMLServiceFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	svc := h.gateway.GetMLService()
	if svc == nil {
		http.Error(w, "MLService not enabled", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		ProfileID string  `json:"profileId"`
		Label     string  `json:"label"`
		Reward    float64 `json:"reward"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var prof *profiles.ClientProfile
	if req.ProfileID != "" {
		if p, ok := h.findProfile(req.ProfileID); ok {
			prof = &p
		}
	}

	sample := &ml.FeedbackSample{
		Profile: prof,
		Label:   req.Label,
		Reward:  req.Reward,
	}
	svc.Feedback(sample)

	st := svc.Stats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"feedbackCount": st.FeedbackCount,
	})
}

// Ensure unused imports are referenced at package level.
var (
	_ = (*defense.RiskEngine)(nil)
	_ = (*frontend.SDK)(nil)
	_ = plugin.GetRegistryStats
	_ = (*tlsmod.JA3Result)(nil)
	_ = (*ml.MLService)(nil)
	_ = fmt.Sprintf
	_ = os.Remove
	_ = (*exec.Cmd)(nil)
	_ = filepath.Dir
)
