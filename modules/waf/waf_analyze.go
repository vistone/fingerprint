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

	if !w.config.Enabled {
		return recordAndReturn(&WAFResult{Action: ActionAllow, Reason: "waf_disabled"})
	}

	// 1. Whitelist check
	if w.isWhitelisted(req) {
		return recordAndReturn(&WAFResult{Action: ActionAllow, Reason: "whitelisted"})
	}

	// 2. Blacklist check
	if action, reason := w.isBlacklisted(req); action != ActionAllow {
		return recordAndReturn(&WAFResult{
			Action:        action,
			Reason:        reason,
			BlockDuration: w.config.BlockDuration,
		})
	}

	// 3. Block list check
	if w.blockList.IsBlocked(clientIP) {
		return recordAndReturn(&WAFResult{
			Action:        ActionBlock,
			Reason:        "in_blocklist",
			BlockDuration: w.blockList.RemainingTime(clientIP),
		})
	}

	// 4. Rate limit check
	if !w.rateLimiter.Allow(clientIP) {
		w.stats.ThrottledRequests++
		return recordAndReturn(&WAFResult{
			Action: ActionThrottle,
			Reason: "rate_limit_exceeded",
		})
	}

	// 5. Multi-layer detection
	detectionLayers := make([]string, 0)
	riskFactors := make([]core.RiskFactor, 0)
	var totalRisk float64

	// L1: Network layer detection
	if w.config.NetworkLayerEnabled && w.networkEngine != nil {
		if result := w.networkEngine.Analyze(req); result.Score > 0 {
			detectionLayers = append(detectionLayers, "network")
			riskFactors = append(riskFactors, result.Factors...)
			totalRisk += result.Score
		}
	}

	// L2: TLS layer detection
	if w.config.TLSLayerEnabled && w.tlsEngine != nil {
		if result := w.tlsEngine.Analyze(req); result.Score > 0 {
			detectionLayers = append(detectionLayers, "tls")
			riskFactors = append(riskFactors, result.Factors...)
			totalRisk += result.Score
		}
	}

	// L3: HTTP layer detection
	if w.config.HTTPLayerEnabled && w.httpEngine != nil {
		if result := w.httpEngine.Analyze(req); result.Score > 0 {
			detectionLayers = append(detectionLayers, "http")
			riskFactors = append(riskFactors, result.Factors...)
			totalRisk += result.Score
		}
	}

	// L4: Behavior layer detection
	if w.config.BehaviorLayerEnabled && w.behaviorEngine != nil {
		if result := w.behaviorEngine.Analyze(clientIP, req); result.Score > 0 {
			detectionLayers = append(detectionLayers, "behavior")
			riskFactors = append(riskFactors, result.Factors...)
			totalRisk += result.Score
		}
	}

	// L5: Device layer detection
	if w.config.DeviceLayerEnabled && w.deviceEngine != nil {
		if result := w.deviceEngine.Analyze(req); result.Score > 0 {
			detectionLayers = append(detectionLayers, "device")
			riskFactors = append(riskFactors, result.Factors...)
			totalRisk += result.Score
		}
	}

	// 6. Comprehensive risk scoring
	riskLevel := core.RiskLevelFromScore(totalRisk)

	// 7. ML verification
	if w.mlService != nil && w.mlService.IsReady() && w.learningPipeline != nil {
		// Build a lightweight profile from request fingerprint info for ML inference.
		fpInfo := w.extractFingerprintInfo(req)
		if fpInfo != nil && fpInfo.JA3 != "" {
			riskAdjustment := w.learningPipeline.RunInference(nil, riskFactors)
			totalRisk += riskAdjustment
			riskLevel = core.RiskLevelFromScore(totalRisk)
		}
	}

	// 8. Autonomous agent decision
	var agentDecision *agent.Decision
	if w.agent != nil {
		obs := &agent.Observation{
			ClientID:  clientIP,
			Timestamp: time.Now(),
			RiskAssessment: &core.RiskAssessment{
				Score:   totalRisk,
				Level:   riskLevel,
				Factors: riskFactors,
			},
		}
		agentDecision = w.agent.Process(ctx, obs)
	}

	// 9. Decision logic
	result := &WAFResult{
		RiskScore:       totalRisk,
		RiskLevel:       riskLevel,
		DetectionLayers: detectionLayers,
		RiskFactors:     riskFactors,
		FingerprintInfo: w.extractFingerprintInfo(req),
	}

	// Determine response based on operating mode
	switch w.config.Mode {
	case WAFModeLearning:
		result.Action = ActionMonitor
		result.Reason = "learning_mode"

	case WAFModeDetection:
		if totalRisk >= w.config.RiskThreshold {
			result.Action = ActionMonitor
			result.Reason = fmt.Sprintf("suspicious_activity_detected: %v", detectionLayers)
		} else {
			result.Action = ActionAllow
		}

	case WAFModeAggressive:
		if totalRisk >= w.config.RiskThreshold*0.5 {
			result.Action = ActionBlock
			result.Reason = "aggressive_mode"
			w.blockList.Block(clientIP, "aggressive_mode")
			w.stats.BlockedRequests++
		}

	default: // WAFModeProtection
		if agentDecision != nil {
			// Prioritize agent decision
			result.Action = WAFAction(agentDecision.Action)
			result.Reason = "agent_decision"
		} else if totalRisk >= w.config.RiskThreshold {
			if totalRisk >= 0.9 {
				result.Action = ActionBlock
				result.Reason = fmt.Sprintf("high_risk: %v", detectionLayers)
				w.blockList.Block(clientIP, fmt.Sprintf("high_risk: %v", detectionLayers))
				w.stats.BlockedRequests++
			} else if totalRisk >= 0.8 {
				result.Action = ActionChallenge
				result.Reason = fmt.Sprintf("medium_risk: %v", detectionLayers)
				result.ChallengeToken = w.generateChallengeToken(clientIP)
				w.stats.ChallengedRequests++
			} else {
				result.Action = ActionMonitor
				result.Reason = fmt.Sprintf("low_risk: %v", detectionLayers)
				w.stats.MonitoredRequests++
			}
		} else {
			result.Action = ActionAllow
			result.Reason = "clean"
		}
	}

	w.stats.TotalRequests++
	if result.Action == ActionAllow {
		w.stats.AllowedRequests++
	}

	// Feed detection result to ML learning pipeline
	if w.learningPipeline != nil && totalRisk > 0 {
		w.learningPipeline.FeedDetection(&ml.WAFDetectionFeedback{
			ClientIP:        clientIP,
			RiskScore:       totalRisk,
			DetectionLayers: detectionLayers,
			Blocked:         result.Action == ActionBlock,
			FingerprintID:   result.FingerprintInfo.JA3,
			Timestamp:       time.Now(),
		})
	}

	return recordAndReturn(result)
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
