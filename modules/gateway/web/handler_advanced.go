// handler_advanced.go defines advanced admin API endpoints.
// Analysis engine / ML engine / defense / anti-detection / plugins / tools
package web

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/gateway"
	"github.com/vistone/fingerprint/modules/profiles"
)

// =====================================================================
// Analysis Engine API — runs full analysis pipeline by profile.
// =====================================================================

// handleAnalyzeProfile runs full fingerprint analysis for a selected profile.
func (h *Handler) handleAnalyzeProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req, ok := decodeAnalyzeProfileRequest(w, r)
	if !ok {
		return
	}

	profile, found := h.findProfile(req.ProfileID)
	if !found {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	resp, err := h.runProfileAnalysis(r.Context(), profile)
	if err != nil {
		http.Error(w, "Analysis failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(buildAnalyzeProfileResponse(req.ProfileID, profile, resp))
}

func decodeAnalyzeProfileRequest(w http.ResponseWriter, r *http.Request) (struct {
	ProfileID string `json:"profileId"`
}, bool) {
	var req struct {
		ProfileID string `json:"profileId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return req, false
	}
	if req.ProfileID == "" {
		http.Error(w, "profileId required", http.StatusBadRequest)
		return req, false
	}
	return req, true
}

func (h *Handler) runProfileAnalysis(parent context.Context, profile profiles.ClientProfile) (*gateway.AnalyzeResponse, error) {
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

	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()

	return h.gateway.Analyze(ctx, analyzeReq)
}

func buildAnalyzeProfileResponse(profileID string, profile profiles.ClientProfile, resp *gateway.AnalyzeResponse) map[string]interface{} {
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

	if resp.RiskAssessment != nil {
		result["riskAssessment"] = map[string]interface{}{
			"level":       resp.RiskAssessment.Level,
			"score":       resp.RiskAssessment.Score,
			"factors":     riskFactorsToMaps(resp.RiskAssessment.Factors),
			"suggestions": resp.RiskAssessment.Suggestions,
		}
	}

	if resp.JA3 != nil {
		result["ja3"] = resp.JA3
	}
	if resp.JA4 != nil {
		result["ja4"] = resp.JA4
	}
	if resp.JA4H != nil {
		result["ja4h"] = resp.JA4H
	}

	// Detection findings.
	result["findings"] = resp.Findings
	result["defenseHints"] = resp.DefenseHints

	// Agent decision.
	if resp.AgentDecision != nil {
		result["agentDecision"] = resp.AgentDecision
	}

	return result
}

func riskFactorsToMaps(factors []core.RiskFactor) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(factors))
	for _, factor := range factors {
		result = append(result, map[string]interface{}{
			"name":        factor.Name,
			"weight":      factor.Weight,
			"description": factor.Description,
		})
	}
	return result
}

// =====================================================================
// ML Engine API
// =====================================================================

// handleMLInfo returns ML classifier model metadata.
func (h *Handler) handleMLInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	classifier := h.gateway.GetClassifier()
	p, f, v := classifier.GetConfidenceThresholds()

	info := map[string]interface{}{
		"architecture": "3-Layer Hierarchical Classifier",
		"layers":       mlInfoLayers(p, f, v),
		"featureTypes": mlInfoFeatureTypes(),
		"status":       "trained",
	}

	h.attachMLInfoServiceState(info)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func mlInfoLayers(p, f, v float64) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "Protocol Classifier",
			"description": "Protocol classification (TLS / HTTP / HTTP2 / QUIC / HTTP3)",
			"level":       1,
			"threshold":   p,
			"weight":      0.3,
		},
		{
			"name":        "Family Classifier",
			"description": "Browser family recognition (Chrome / Firefox / Safari / Edge / Opera)",
			"level":       2,
			"threshold":   f,
			"weight":      0.3,
		},
		{
			"name":        "Version Classifier",
			"description": "Version recognition (Chrome 134 / Firefox 135 / Safari 18 ...)",
			"level":       3,
			"threshold":   v,
			"weight":      0.4,
		},
	}
}

func mlInfoFeatureTypes() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "tls_version", "category": "TLS", "description": "TLS protocol version"},
		{"name": "cipher_suites", "category": "TLS", "description": "Cipher suite count"},
		{"name": "extensions", "category": "TLS", "description": "TLS extension count"},
		{"name": "http2_settings", "category": "HTTP", "description": "HTTP/2 SETTINGS hash"},
		{"name": "http_headers", "category": "HTTP", "description": "HTTP header feature hash"},
		{"name": "user_agent", "category": "HTTP", "description": "User-Agent hash"},
		{"name": "canvas", "category": "Frontend", "description": "Canvas fingerprint hash"},
		{"name": "webgl", "category": "Frontend", "description": "WebGL renderer fingerprint"},
		{"name": "audio", "category": "Frontend", "description": "AudioContext fingerprint"},
		{"name": "fonts", "category": "Frontend", "description": "Font list fingerprint"},
		{"name": "storage", "category": "Frontend", "description": "Storage API fingerprint"},
		{"name": "webrtc", "category": "Frontend", "description": "WebRTC config fingerprint"},
		{"name": "hardware", "category": "Frontend", "description": "Hardware info fingerprint"},
		{"name": "timing", "category": "Frontend", "description": "Timing precision fingerprint"},
		{"name": "headless_browser", "category": "Detection", "description": "Headless browser detection"},
		{"name": "entropy", "category": "Detection", "description": "Information entropy"},
		{"name": "tool_marker", "category": "Detection", "description": "Automation tool marker"},
		{"name": "behavior_pattern", "category": "Detection", "description": "Behavior consistency pattern"},
	}
}

func (h *Handler) attachMLInfoServiceState(info map[string]interface{}) {
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
		return
	}

	info["mlService"] = map[string]interface{}{"enabled": false}
}

// handleMLExtract extracts feature vectors from a selected profile.
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

	// Convert feature map to list form.
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

// handleMLClassify runs ML classification for a selected profile.
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

// handleMLBatch classifies all profiles in batch.
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
