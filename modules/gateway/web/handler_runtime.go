package web

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) handleCrawlerStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cr := h.gateway.GetCrawler()
	if cr == nil {
		writeRuntimeJSON(w, map[string]interface{}{"enabled": false})
		return
	}

	cfg := cr.ConfigSnapshot()
	stats := cr.GetStats()
	writeRuntimeJSON(w, map[string]interface{}{
		"enabled":         true,
		"running":         cr.Running(),
		"targets":         cfg.TargetURLs,
		"workerCount":     cfg.Workers,
		"profileStrategy": cfg.ProfileStrategy,
		"stealthMode":     cfg.StealthMode,
		"stats": map[string]interface{}{
			"totalRequests":   stats.TotalRequests.Load(),
			"successRequests": stats.SuccessRequests.Load(),
			"failedRequests":  stats.FailedRequests.Load(),
			"blockedRequests": stats.BlockedRequests.Load(),
			"successRate":     stats.SuccessRate(),
			"blockRate":       stats.BlockRate(),
		},
	})
}

func (h *Handler) handleWAFStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	wafInst := h.gateway.GetWAF()
	if wafInst == nil {
		writeRuntimeJSON(w, map[string]interface{}{"enabled": false})
		return
	}

	cfg := wafInst.ConfigSnapshot()
	stats := wafInst.Stats()
	result := map[string]interface{}{
		"enabled":       true,
		"mode":          cfg.Mode,
		"riskThreshold": cfg.RiskThreshold,
		"stats": map[string]interface{}{
			"totalRequests":      stats.TotalRequests,
			"allowedRequests":    stats.AllowedRequests,
			"blockedRequests":    stats.BlockedRequests,
			"challengedRequests": stats.ChallengedRequests,
			"throttledRequests":  stats.ThrottledRequests,
			"monitoredRequests":  stats.MonitoredRequests,
		},
	}
	if learning := wafInst.LearningPipelineStats(); learning != nil {
		result["learning"] = map[string]interface{}{
			"inferRuns":        learning.InferencesRun,
			"detectionsFed":    learning.DetectionsFed,
			"samplesProcessed": learning.SamplesProcessed,
			"lastFeedbackTime": learning.LastFeedbackTime,
		}
	}
	writeRuntimeJSON(w, result)
}

func (h *Handler) appendRuntimeStats(stats map[string]interface{}) {
	if cr := h.gateway.GetCrawler(); cr != nil {
		cfg := cr.ConfigSnapshot()
		cstats := cr.GetStats()
		stats["crawler"] = map[string]interface{}{
			"enabled":         true,
			"running":         cr.Running(),
			"targets":         len(cfg.TargetURLs),
			"workers":         cfg.Workers,
			"blockedRequests": cstats.BlockedRequests.Load(),
			"successRate":     cstats.SuccessRate(),
		}
	} else {
		stats["crawler"] = map[string]interface{}{"enabled": false}
	}

	if wafInst := h.gateway.GetWAF(); wafInst != nil {
		wstats := wafInst.Stats()
		stats["waf"] = map[string]interface{}{
			"enabled":         true,
			"mode":            wafInst.ConfigSnapshot().Mode,
			"totalRequests":   wstats.TotalRequests,
			"blockedRequests": wstats.BlockedRequests,
			"allowedRequests": wstats.AllowedRequests,
		}
	} else {
		stats["waf"] = map[string]interface{}{"enabled": false}
	}

	if systemStatus, ok := stats["systemStatus"].(map[string]interface{}); ok {
		systemStatus["crawlerIntegrated"] = h.gateway.GetCrawler() != nil
		systemStatus["wafIntegrated"] = h.gateway.GetWAF() != nil
		systemStatus["crawlerRunning"] = h.gateway.GetCrawler() != nil && h.gateway.GetCrawler().Running()
		systemStatus["closedLoopEnabled"] = h.gateway.ClosedLoopStats() != nil
	}

	if cl := h.gateway.ClosedLoopStats(); cl != nil {
		stats["closedLoop"] = map[string]interface{}{
			"enabled":             true,
			"cyclesCompleted":     cl.CyclesCompleted,
			"profilesGenerated":   cl.ProfilesGenerated,
			"detectionsProcessed": cl.DetectionsProcessed,
			"modelsEvolved":       cl.ModelsEvolved,
			"lastCycleTime":       cl.LastCycleTime,
		}
	} else {
		stats["closedLoop"] = map[string]interface{}{"enabled": false}
	}
}

func writeRuntimeJSON(w http.ResponseWriter, payload map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}
