// handler_advanced.go — 高级功能 API 端点
// 分析引擎 / ML引擎 / 防御系统 / 反检测引擎 / 插件系统 / 指纹工具
package web

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/vistone/fingerprint/modules/gateway"
	"github.com/vistone/fingerprint/modules/profiles"
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
