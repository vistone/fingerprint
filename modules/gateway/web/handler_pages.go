package web

import (
	"encoding/json"
	"net/http"
	"strings"

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

	// Extract profile ID from URL path: /api/admin/profiles/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/profiles/")
	profileID := strings.TrimSpace(path)

	if profileID == "" {
		http.Error(w, "Profile ID required", http.StatusBadRequest)
		return
	}

	// Find profile from loaded profiles
	var found profiles.ClientProfile
	foundOK := false
	h.mu.RLock()
	for i := range h.profiles {
		if h.profiles[i].ID == profileID {
			found = h.profiles[i]
			foundOK = true
			break
		}
	}
	h.mu.RUnlock()

	if !foundOK {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

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

	// 添加 TCP/IP 指纹信息
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

	// 添加 HTTP/3 (QUIC) 配置
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

	// 转换为前端期望的格式
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
		cfg := h.gateway.GetConfig()
		agentEnabled := cfg.AgentEnabled
		var agentCfg map[string]interface{}
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
			agentCfg = map[string]interface{}{
				"enabled":               agentEnabled,
				"sessionWindow":         sessionWindow,
				"maxObservations":       maxObs,
				"fpSwitchRateThreshold": fpThresh,
				"consistencyThreshold":  consThresh,
				"requestBurstThreshold": burstThresh,
				"activeSessions":        stats.ActiveSessions,
				"totalObservations":     stats.TotalObservations,
			}
		} else {
			agentCfg = map[string]interface{}{
				"enabled": false,
			}
		}

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
			"p3": map[string]interface{}{
				"enabled":       cfg.P3Enabled,
				"profileId":     cfg.P3ProfileID,
				"configDir":     cfg.P3ConfigDir,
				"proxyTarget":   cfg.P3ProxyTarget,
				"directProxy":   cfg.P3DirectProxy,
				"injectConsist": cfg.P3InjectConsist,
			},
			"scanner": map[string]interface{}{
				"useBrowser":     cfg.ScannerUseBrowser,
				"browserWS":      cfg.ScannerBrowserWS,
				"browserTimeout": int(cfg.ScannerBrowserTimeout.Seconds()),
			},
			"agent": agentCfg,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)

	case http.MethodPost:
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

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Helper functions
