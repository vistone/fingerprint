package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/vistone/fingerprint/modules/gateway"
	"github.com/vistone/fingerprint/modules/profiles"
)

func (h *Handler) handleProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := strings.ToLower(r.URL.Query().Get("q"))
	browser := strings.ToLower(r.URL.Query().Get("browser"))
	os := strings.ToLower(r.URL.Query().Get("os"))

	h.mu.RLock()
	profileSnapshot := append([]profiles.ClientProfile(nil), h.profiles...)
	h.mu.RUnlock()
	filtered := filterProfiles(profileSnapshot, query, browser, os)

	response := map[string]interface{}{
		"profiles": filtered,
		"total":    len(profileSnapshot),
		"filtered": len(filtered),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleProfileDetail returns a single profile detail
func (h *Handler) handleProfileDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	profileID := extractProfileIDFromPath(r.URL.Path)
	if profileID == "" {
		http.Error(w, "Profile ID required", http.StatusBadRequest)
		return
	}

	found, foundOK := h.findProfile(profileID)
	if !foundOK {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	response := buildProfileDetailResponse(found)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func extractProfileIDFromPath(path string) string {
	return strings.TrimSpace(strings.TrimPrefix(path, "/api/admin/profiles/"))
}

func buildProfileDetailResponse(found profiles.ClientProfile) map[string]interface{} {
	response := map[string]interface{}{
		"id":              found.ID,
		"name":            found.Name,
		"description":     found.Description,
		"browserType":     found.BrowserType,
		"browserVersion":  found.BrowserVersion,
		"os":              found.OS,
		"osVersion":       found.OSVersion,
		"osArch":          found.OSArch,
		"osBitness":       found.OSBitness,
		"tlsVersion":      found.TLSVersion,
		"cipherSuites":    found.CipherSuites,
		"extensions":      found.Extensions,
		"supportedCurves": found.SupportedCurves,
		"supportedPoints": found.SupportedPoints,
		"http2Settings": map[string]interface{}{
			"headerTableSize":      found.HTTP2Settings.HeaderTableSize,
			"enablePush":           found.HTTP2Settings.EnablePush,
			"maxConcurrentStreams": found.HTTP2Settings.MaxConcurrentStreams,
			"initialWindowSize":    found.HTTP2Settings.InitialWindowSize,
			"maxFrameSize":         found.HTTP2Settings.MaxFrameSize,
			"maxHeaderListSize":    found.HTTP2Settings.MaxHeaderListSize,
		},
		"http2Priorities":   found.HTTP2Priorities,
		"pseudoHeaderOrder": found.PseudoHeaderOrder,
		"connectionFlow":    found.ConnectionFlow,
		"headers":           found.Headers,
		"metadata":          found.Metadata,
	}

	if found.TCPIP != nil {
		response["tcpip"] = map[string]interface{}{
			"ipVersion":        found.TCPIP.IPVersion,
			"ttl":              found.TCPIP.TTL,
			"df":               found.TCPIP.DF,
			"flags":            found.TCPIP.Flags,
			"windowSize":       found.TCPIP.WindowSize,
			"mss":              found.TCPIP.MSS,
			"windowScale":      found.TCPIP.WindowScale,
			"sackPermitted":    found.TCPIP.SAckPermitted,
			"timestamps":       found.TCPIP.Timestamps,
			"noOperation":      found.TCPIP.NoOperation,
			"endOfOptions":     found.TCPIP.EndOfOptions,
			"optionsSignature": found.TCPIP.OptionsSignature,
			"ja4t":             found.TCPIP.JA4T,
		}
	}

	if found.HTTP3Settings != nil {
		response["http3Settings"] = map[string]interface{}{
			"quicVersion":            found.HTTP3Settings.QUICVersion,
			"initialMaxData":         found.HTTP3Settings.InitialMaxData,
			"initialMaxStreamData":   found.HTTP3Settings.InitialMaxStreamData,
			"initialMaxStreamsBidi":  found.HTTP3Settings.InitialMaxStreamsBidi,
			"initialMaxStreamsUni":   found.HTTP3Settings.InitialMaxStreamsUni,
			"maxUDPPayloadSize":      found.HTTP3Settings.MaxUDPPayloadSize,
			"ackDelayExponent":       found.HTTP3Settings.AckDelayExponent,
			"maxAckDelay":            found.HTTP3Settings.MaxAckDelay,
			"disableActiveMigration": found.HTTP3Settings.DisableActiveMigration,
		}
		response["quicVersions"] = found.QUICVersions
		response["http3Supported"] = true
	} else {
		response["http3Supported"] = false
	}

	return response
}

// handleAnalytics returns analytics data
func (h *Handler) handleAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.mu.RLock()
	profileSnapshot := append([]profiles.ClientProfile(nil), h.profiles...)
	h.mu.RUnlock()

	analytics := map[string]interface{}{
		"browserDistribution": getBrowserDistribution(profileSnapshot),
		"osDistribution":      getOSDistribution(profileSnapshot),
		"tcpipDistribution":   getTCPIPDistribution(profileSnapshot),
		"topFingerprints":     getTopFingerprints(),
		"trafficData":         getTrafficData(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// handleRequests returns request logs
func (h *Handler) handleRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	records := GetRecentRequests()

	// Transform to frontend-expected shape.
	requests := make([]map[string]interface{}, len(records))
	for i, req := range records {
		requests[i] = map[string]interface{}{
			"timestamp":      req.Timestamp.Unix(),
			"ip":             req.IP,
			"method":         req.Method,
			"path":           req.Path,
			"ja3":            req.JA3,
			"classification": req.Classification,
			"latency":        req.Latency,
			"status":         req.Status,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": requests,
	})
}

// handleLogs returns system logs from the real log buffer
func (h *Handler) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	level := r.URL.Query().Get("level")
	entries := globalLogBuffer.GetFiltered(level)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  entries,
		"total": len(entries),
	})
}

// handleConfig handles configuration requests - reads/writes real GatewayConfig
func (h *Handler) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleConfigGet(w)

	case http.MethodPost:
		h.handleConfigPost(w, r)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleConfigGet(w http.ResponseWriter) {
	cfg := h.gateway.GetConfig()
	config := map[string]interface{}{
		"server": map[string]interface{}{
			"endpoint": cfg.Endpoint,
			"port":     cfg.Port,
		},
		"rateLimit": map[string]interface{}{
			"enabled":   true,
			"rps":       cfg.RateLimitRequests,
			"burstSize": cfg.RateLimitBurst,
			"window":    int(cfg.RateLimitWindow.Seconds()),
		},
		"cache": map[string]interface{}{
			"enabled": cfg.CacheEnabled,
			"size":    cfg.CacheSize,
			"ttl":     int(cfg.CacheTTL.Minutes()),
		},
		"ml": map[string]interface{}{
			"enabled":       true,
			"riskThreshold": cfg.RiskThreshold,
		},
		"mlService": h.getMLServiceConfig(cfg),
		"antiDetect": map[string]interface{}{
			"enabled":       cfg.AntiDetectEnabled,
			"profileId":     cfg.AntiDetectProfileID,
			"configDir":     cfg.AntiDetectConfigDir,
			"proxyTarget":   cfg.AntiDetectProxyTarget,
			"directProxy":   cfg.AntiDetectDirectProxy,
			"injectConsist": cfg.AntiDetectInjectConsist,
		},
		"scanner": map[string]interface{}{
			"useBrowser":     cfg.ScannerUseBrowser,
			"browserWS":      cfg.ScannerBrowserWS,
			"browserTimeout": int(cfg.ScannerBrowserTimeout.Seconds()),
		},
		"agent": h.buildAgentConfigState(cfg),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func (h *Handler) buildAgentConfigState(cfg *gateway.GatewayConfig) map[string]interface{} {
	agentEnabled := cfg.AgentEnabled
	if a := h.gateway.GetAgent(); a != nil {
		stats := a.Stats()
		sessionWindow := 30
		maxObs := 500
		fpThresh := 3.0
		consThresh := 0.4
		burstThresh := 10.0
		if cfg.AgentConfig != nil {
			sessionWindow = int(cfg.AgentConfig.SessionWindow.Minutes())
			maxObs = cfg.AgentConfig.MaxObservations
			fpThresh = cfg.AgentConfig.FPSwitchRateThreshold
			consThresh = cfg.AgentConfig.ConsistencyThreshold
			burstThresh = cfg.AgentConfig.RequestBurstThreshold
		}
		return map[string]interface{}{
			"enabled":               agentEnabled,
			"sessionWindow":         sessionWindow,
			"maxObservations":       maxObs,
			"fpSwitchRateThreshold": fpThresh,
			"consistencyThreshold":  consThresh,
			"requestBurstThreshold": burstThresh,
			"activeSessions":        stats.ActiveSessions,
			"totalObservations":     stats.TotalObservations,
		}
	}

	return map[string]interface{}{"enabled": false}
}

func (h *Handler) handleConfigPost(w http.ResponseWriter, r *http.Request) {
	var newConfig map[string]interface{}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&newConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.applyConfigUpdate(newConfig)
	WriteLog("INFO", "config", "Configuration updated via admin console")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}

// Helper functions
