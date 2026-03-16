package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vistone/fingerprint/modules/agent"
	"github.com/vistone/fingerprint/modules/client"
	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/gateway"
	"github.com/vistone/fingerprint/modules/profiles"
)

func (h *Handler) handleLogStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := globalLogBuffer.Subscribe()
	defer globalLogBuffer.Unsubscribe(ch)

	// Send an initial connection acknowledgement event.
	fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(entry)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// handleAgentStatus returns Agent stats and status
func (h *Handler) handleAgentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a := h.gateway.GetAgent()
	if a == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"enabled": false,
			"status":  "disabled",
		})
		return
	}

	stats := a.Stats()
	strategies := a.GetActiveStrategies()
	kb := a.Knowledge()
	kbStats := kb.Stats()

	result := map[string]interface{}{
		"enabled": true,
		"status":  "running",
		"stats":   stats,
		"strategySummary": map[string]interface{}{
			"total":   len(strategies),
			"learned": countLearnedStrategies(strategies),
		},
		"knowledge": map[string]interface{}{
			"totalBrowsers": kbStats.TotalKnownBrowsers,
			"totalVersions": kbStats.TotalKnownVersions,
			"totalProfiles": kbStats.TotalKnownProfiles,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleAgentKnowledge returns the knowledge base data
func (h *Handler) handleAgentKnowledge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a := h.gateway.GetAgent()
	if a == nil {
		http.Error(w, "Agent not enabled", http.StatusServiceUnavailable)
		return
	}

	kb := a.Knowledge()
	kbStats := kb.Stats()

	// Build browser family details for the response.
	families := []map[string]interface{}{}
	browserTypes := []core.BrowserType{
		core.BrowserChrome, core.BrowserFirefox, core.BrowserSafari,
		core.BrowserEdge, core.BrowserOpera, core.BrowserBrave,
	}
	for _, bt := range browserTypes {
		bk := kb.GetBrowserKnowledge(bt)
		if bk == nil {
			continue
		}
		versions := []map[string]interface{}{}
		for _, v := range bk.Versions {
			versions = append(versions, map[string]interface{}{
				"version":      v.Version,
				"versionMajor": v.VersionMajor,
				"supportedOS":  v.SupportedOS,
				"tlsVersion":   v.TLSVersion,
				"cipherSuites": len(v.CipherSuites),
				"extensions":   len(v.Extensions),
				"h2WindowSize": v.H2InitialWindowSize,
				"releasedYear": v.ReleasedYear,
				"deprecated":   v.Deprecated,
			})
		}
		families = append(families, map[string]interface{}{
			"family":       string(bt),
			"marketShare":  bk.MarketShare,
			"cipherSuites": len(bk.CommonCipherSuites),
			"extensions":   len(bk.CommonExtensions),
			"versions":     versions,
		})
	}

	result := map[string]interface{}{
		"stats":    kbStats,
		"families": families,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleAgentStrategies returns active strategy list
func (h *Handler) handleAgentStrategies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a := h.gateway.GetAgent()
	if a == nil {
		http.Error(w, "Agent not enabled", http.StatusServiceUnavailable)
		return
	}

	strategies := a.GetActiveStrategies()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"strategies": strategies,
		"total":      len(strategies),
	})
}

func countLearnedStrategies(strategies []agent.StrategyInfo) int {
	count := 0
	for _, s := range strategies {
		if s.Learned {
			count++
		}
	}
	return count
}

// handleClientTest runs a request using a selected fingerprint profile.
func (h *Handler) handleClientTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request payload.
	var req struct {
		ProfileID string `json:"profileId"`
		URL       string `json:"url"`
		Method    string `json:"method"`
		Body      string `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields.
	if req.ProfileID == "" {
		http.Error(w, "profileId is required", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	// Resolve requested profile.
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
		http.Error(w, fmt.Sprintf("Profile not found: %s", req.ProfileID), http.StatusNotFound)
		return
	}

	// Validate profile completeness before execution.
	validator := profiles.NewProfileValidator()
	validationResult := validator.Validate(profile)

	// Execute request and return trace details.
	result := testWithProfile(profile, req.URL, req.Method, req.Body, validationResult)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// testWithProfile executes a request with the selected profile and returns full traces.
func testWithProfile(profile profiles.ClientProfile, url, method, body string, validationResult profiles.ProfileValidationResult) map[string]interface{} {
	// Use ExecuteProxyRequest to collect request/response traces.
	result := client.ExecuteProxyRequest(profile, url, method, body, nil)

	// Build response payload.
	response := map[string]interface{}{
		"success":       result.Success,
		"error":         result.Error,
		"errorType":     result.ErrorType,
		"errorCode":     result.ErrorCode,
		"errorDetails":  result.ErrorDetails,
		"profileUsed":   result.ProfileUsed,
		"requestTrace":  result.RequestTrace,
		"responseTrace": result.ResponseTrace,
	}

	// Attach validation details when warnings or errors exist.
	validation := map[string]interface{}{
		"valid": validationResult.Valid,
	}

	if len(validationResult.Warnings) > 0 {
		validation["warnings"] = validationResult.Warnings
	}

	if len(validationResult.Errors) > 0 {
		validation["errors"] = validationResult.Errors
	}

	if len(validationResult.MissingFields) > 0 {
		validation["missing_fields"] = validationResult.MissingFields
	}

	response["validation"] = validation

	return response
}

// getMLServiceConfig builds the MLService section for admin status.
func (h *Handler) getMLServiceConfig(cfg *gateway.GatewayConfig) map[string]interface{} {
	result := map[string]interface{}{
		"enabled": cfg.MLServiceEnabled,
	}
	if svc := h.gateway.GetMLService(); svc != nil {
		result["ready"] = svc.IsReady()
		st := svc.Stats()
		result["inferCount"] = st.InferCount
		result["feedbackCount"] = st.FeedbackCount
		result["evolveCount"] = st.EvolveCount
		result["modelVersions"] = st.ModelVersions
	}
	if cfg.MLServiceConfig != nil {
		result["modelStorePath"] = cfg.MLServiceConfig.ModelStorePath
		result["maxStoreVersions"] = cfg.MLServiceConfig.MaxStoreVersions
		result["driftThreshold"] = cfg.MLServiceConfig.DriftThreshold
		result["forgeryThreshold"] = cfg.MLServiceConfig.ValidationForgeryThreshold
		result["consistencyMin"] = cfg.MLServiceConfig.ValidationConsistencyMin
	}
	return result
}
