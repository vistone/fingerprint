package web

import (
	"encoding/json"
	"net/http"

	"github.com/vistone/fingerprint/modules/defense"
)

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
