package web

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/profiles"
)

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
