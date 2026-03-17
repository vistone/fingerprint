package waf

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/vistone/fingerprint/modules/agent"
	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/ml"
)

type riskAggregation struct {
	detectionLayers []string
	riskFactors     []core.RiskFactor
	totalRisk       float64
	riskLevel       core.RiskLevel
}

// Analyze analyzes the request
func (w *WAF) Analyze(ctx context.Context, req *http.Request) *WAFResult {
	w.mu.RLock()
	defer w.mu.RUnlock()

	clientIP := w.getClientIP(req)
	recordAndReturn := func(result *WAFResult) *WAFResult {
		if result == nil {
			return nil
		}
		w.recordDecision(WAFDecision{
			Timestamp:       time.Now(),
			Action:          result.Action,
			Reason:          result.Reason,
			RiskScore:       result.RiskScore,
			ClientIP:        clientIP,
			Method:          req.Method,
			Path:            req.URL.Path,
			DetectionLayers: result.DetectionLayers,
		})
		return result
	}

	if earlyResult := w.runEarlyChecks(req, clientIP); earlyResult != nil {
		return recordAndReturn(earlyResult)
	}

	aggregation := w.aggregateRisk(req, clientIP)
	w.applyMLAdjustment(req, aggregation)
	agentDecision := w.runAgentDecision(ctx, clientIP, aggregation)

	result := &WAFResult{
		RiskScore:       aggregation.totalRisk,
		RiskLevel:       aggregation.riskLevel,
		DetectionLayers: aggregation.detectionLayers,
		RiskFactors:     aggregation.riskFactors,
		FingerprintInfo: w.extractFingerprintInfo(req),
	}
	w.applyDecisionPolicy(result, clientIP, aggregation, agentDecision)
	w.updateStatsAndLearning(clientIP, result)

	return recordAndReturn(result)
}

func (w *WAF) runEarlyChecks(req *http.Request, clientIP string) *WAFResult {
	if !w.config.Enabled {
		return &WAFResult{Action: ActionAllow, Reason: "waf_disabled"}
	}
	if w.isWhitelisted(req) {
		return &WAFResult{Action: ActionAllow, Reason: "whitelisted"}
	}
	if action, reason := w.isBlacklisted(req); action != ActionAllow {
		return &WAFResult{Action: action, Reason: reason, BlockDuration: w.config.BlockDuration}
	}
	if w.blockList.IsBlocked(clientIP) {
		return &WAFResult{
			Action:        ActionBlock,
			Reason:        "in_blocklist",
			BlockDuration: w.blockList.RemainingTime(clientIP),
		}
	}
	if !w.rateLimiter.Allow(clientIP) {
		w.stats.ThrottledRequests++
		return &WAFResult{Action: ActionThrottle, Reason: "rate_limit_exceeded"}
	}

	return nil
}

func (w *WAF) aggregateRisk(req *http.Request, clientIP string) *riskAggregation {
	result := &riskAggregation{
		detectionLayers: make([]string, 0),
		riskFactors:     make([]core.RiskFactor, 0),
	}

	if w.config.NetworkLayerEnabled && w.networkEngine != nil {
		if layerResult := w.networkEngine.Analyze(req); layerResult.Score > 0 {
			result.detectionLayers = append(result.detectionLayers, "network")
			result.riskFactors = append(result.riskFactors, layerResult.Factors...)
			result.totalRisk += layerResult.Score
		}
	}
	if w.config.TLSLayerEnabled && w.tlsEngine != nil {
		if layerResult := w.tlsEngine.Analyze(req); layerResult.Score > 0 {
			result.detectionLayers = append(result.detectionLayers, "tls")
			result.riskFactors = append(result.riskFactors, layerResult.Factors...)
			result.totalRisk += layerResult.Score
		}
	}
	if w.config.HTTPLayerEnabled && w.httpEngine != nil {
		if layerResult := w.httpEngine.Analyze(req); layerResult.Score > 0 {
			result.detectionLayers = append(result.detectionLayers, "http")
			result.riskFactors = append(result.riskFactors, layerResult.Factors...)
			result.totalRisk += layerResult.Score
		}
	}
	if w.config.BehaviorLayerEnabled && w.behaviorEngine != nil {
		if layerResult := w.behaviorEngine.Analyze(clientIP, req); layerResult.Score > 0 {
			result.detectionLayers = append(result.detectionLayers, "behavior")
			result.riskFactors = append(result.riskFactors, layerResult.Factors...)
			result.totalRisk += layerResult.Score
		}
	}
	if w.config.DeviceLayerEnabled && w.deviceEngine != nil {
		if layerResult := w.deviceEngine.Analyze(req); layerResult.Score > 0 {
			result.detectionLayers = append(result.detectionLayers, "device")
			result.riskFactors = append(result.riskFactors, layerResult.Factors...)
			result.totalRisk += layerResult.Score
		}
	}

	result.riskLevel = core.RiskLevelFromScore(result.totalRisk)
	return result
}

func (w *WAF) applyMLAdjustment(req *http.Request, agg *riskAggregation) {
	if w.mlService == nil || !w.mlService.IsReady() || w.learningPipeline == nil {
		return
	}
	fpInfo := w.extractFingerprintInfo(req)
	if fpInfo == nil || fpInfo.JA3 == "" {
		return
	}

	agg.totalRisk += w.learningPipeline.RunInference(nil, agg.riskFactors)
	agg.riskLevel = core.RiskLevelFromScore(agg.totalRisk)
}

func (w *WAF) runAgentDecision(ctx context.Context, clientIP string, agg *riskAggregation) *agent.Decision {
	if w.agent == nil {
		return nil
	}

	obs := &agent.Observation{
		ClientID:  clientIP,
		Timestamp: time.Now(),
		RiskAssessment: &core.RiskAssessment{
			Score:   agg.totalRisk,
			Level:   agg.riskLevel,
			Factors: agg.riskFactors,
		},
	}
	return w.agent.Process(ctx, obs)
}

func (w *WAF) applyDecisionPolicy(result *WAFResult, clientIP string, agg *riskAggregation, agentDecision *agent.Decision) {
	switch w.config.Mode {
	case WAFModeLearning:
		result.Action = ActionMonitor
		result.Reason = "learning_mode"

	case WAFModeDetection:
		if agg.totalRisk >= w.config.RiskThreshold {
			result.Action = ActionMonitor
			result.Reason = fmt.Sprintf("suspicious_activity_detected: %v", agg.detectionLayers)
		} else {
			result.Action = ActionAllow
		}

	case WAFModeAggressive:
		if agg.totalRisk >= w.config.RiskThreshold*0.5 {
			result.Action = ActionBlock
			result.Reason = "aggressive_mode"
			w.blockList.Block(clientIP, "aggressive_mode")
			w.stats.BlockedRequests++
		}

	default: // WAFModeProtection
		w.applyProtectionPolicy(result, clientIP, agg, agentDecision)
	}
}

func (w *WAF) applyProtectionPolicy(result *WAFResult, clientIP string, agg *riskAggregation, agentDecision *agent.Decision) {
	if agentDecision != nil {
		result.Action = WAFAction(agentDecision.Action)
		result.Reason = "agent_decision"
		return
	}
	if agg.totalRisk < w.config.RiskThreshold {
		result.Action = ActionAllow
		result.Reason = "clean"
		return
	}

	if agg.totalRisk >= 0.9 {
		result.Action = ActionBlock
		result.Reason = fmt.Sprintf("high_risk: %v", agg.detectionLayers)
		w.blockList.Block(clientIP, fmt.Sprintf("high_risk: %v", agg.detectionLayers))
		w.stats.BlockedRequests++
		return
	}
	if agg.totalRisk >= 0.8 {
		result.Action = ActionChallenge
		result.Reason = fmt.Sprintf("medium_risk: %v", agg.detectionLayers)
		result.ChallengeToken = w.generateChallengeToken(clientIP)
		w.stats.ChallengedRequests++
		return
	}

	result.Action = ActionMonitor
	result.Reason = fmt.Sprintf("low_risk: %v", agg.detectionLayers)
	w.stats.MonitoredRequests++
}

func (w *WAF) updateStatsAndLearning(clientIP string, result *WAFResult) {
	w.stats.TotalRequests++
	if result.Action == ActionAllow {
		w.stats.AllowedRequests++
	}
	if w.learningPipeline == nil || result.RiskScore <= 0 {
		return
	}

	w.learningPipeline.FeedDetection(&ml.WAFDetectionFeedback{
		ClientIP:        clientIP,
		RiskScore:       result.RiskScore,
		DetectionLayers: result.DetectionLayers,
		Blocked:         result.Action == ActionBlock,
		FingerprintID:   result.FingerprintInfo.JA3,
		Timestamp:       time.Now(),
	})
}

// Helper methods
func (w *WAF) isWhitelisted(req *http.Request) bool {
	clientIP := w.getClientIP(req)
	path := req.URL.Path

	for _, ip := range w.config.WhitelistIPs {
		if ip == clientIP {
			return true
		}
	}

	for _, p := range w.config.WhitelistPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	return false
}

func (w *WAF) isBlacklisted(req *http.Request) (WAFAction, string) {
	clientIP := w.getClientIP(req)
	path := req.URL.Path

	for _, ip := range w.config.BlacklistIPs {
		if ip == clientIP {
			return ActionBlock, "ip_blacklisted"
		}
	}

	for _, p := range w.config.BlacklistPaths {
		if strings.HasPrefix(path, p) {
			return ActionBlock, "path_blacklisted"
		}
	}

	return ActionAllow, ""
}

func (w *WAF) getClientIP(req *http.Request) string {
	// X-Forwarded-For
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return xff
	}

	// X-Real-IP
	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// RemoteAddr
	host, _, _ := net.SplitHostPort(req.RemoteAddr)
	if host != "" {
		return host
	}

	return req.RemoteAddr
}

func (w *WAF) extractFingerprintInfo(req *http.Request) *FingerprintInfo {
	return &FingerprintInfo{
		JA3:       req.Header.Get("X-JA3-Fingerprint"),
		JA4:       req.Header.Get("X-JA4-Fingerprint"),
		SessionID: req.Header.Get("X-Session-ID"),
	}
}
