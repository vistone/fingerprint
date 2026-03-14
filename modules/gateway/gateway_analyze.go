package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/vistone/fingerprint/modules/agent"
	"github.com/vistone/fingerprint/modules/core"
	"github.com/vistone/fingerprint/modules/defense"
	"github.com/vistone/fingerprint/modules/ml"
	"github.com/vistone/fingerprint/modules/network/ja4t"
	"github.com/vistone/fingerprint/modules/network/tcp"
)

func (g *Gateway) Analyze(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error) {
	start := time.Now()

	fingerprintHash := g.generateFingerprintHash(req)

	// Check cache
	if g.config.CacheEnabled {
		if cached, ok := g.cache.Get(fingerprintHash); ok {
			resp := cloneAnalyzeResponse(cached)
			resp.Cached = true
			resp.CacheTime = time.Now()
			resp.ProcessingTimeMs = time.Since(start).Milliseconds()
			return resp, nil
		}
	}

	// Core analysis
	features := g.extractFeatures(req)
	classification := g.classifier.Classify(features)
	risk := g.riskEngine.Evaluate(features, classification)
	detector := defense.NewDetector()
	detection := detector.Detect(features)

	// Fingerprints
	ja3, ja4 := g.calculateTLSFingerprints(req)
	ja4h := g.calculateHTTPFingerprint(req)
	ja4tInfo, netAnalysis := g.analyzeNetworkLayer(req, features)

	// Plugin pipeline
	pluginFindings := g.runPluginPipeline(ctx, req, fingerprintHash, classification, risk)

	riskBlocked := g.config.RiskThreshold > 0 && risk != nil && risk.Score >= g.config.RiskThreshold

	response := &AnalyzeResponse{
		FingerprintHash:  fingerprintHash,
		Classification:   classification,
		RiskAssessment:   risk,
		RiskBlocked:      riskBlocked,
		Findings:         detection.Findings,
		JA3:              ja3,
		JA4:              ja4,
		JA4H:             ja4h,
		JA4T:             ja4tInfo,
		NetworkAnalysis:  netAnalysis,
		PluginResults:    pluginFindings,
		DefenseHints:     risk.Suggestions,
		Cached:           false,
		ProcessingTimeMs: time.Since(start).Milliseconds(),
	}

	// Post-processing enrichment
	g.enrichWithMLValidation(response, features)

	if g.agent != nil {
		obs := &agent.Observation{
			ID:              fingerprintHash,
			ClientID:        req.ClientIP,
			Timestamp:       time.Now(),
			Features:        features,
			Classification:  classification,
			Detection:       detection,
			RiskAssessment:  risk,
			FingerprintHash: fingerprintHash,
		}
		response.AgentDecision = g.agent.Process(ctx, obs)
	}

	// Store in cache
	if g.config.CacheEnabled {
		g.cache.Set(fingerprintHash, response)
	}

	return response, nil
}

// analyzeNetworkLayer computes JA4T transport fingerprint and TCP/IP analysis.
func (g *Gateway) analyzeNetworkLayer(req *AnalyzeRequest, features *core.FeatureVector) (*JA4TInfo, *NetworkAnalysisResult) {
	var ja4tInfo *JA4TInfo
	var netAnalysis *NetworkAnalysisResult

	if req.TCPData != nil {
		synData := ja4t.TCPSYNData{
			WindowSize:  req.TCPData.WindowSize,
			Options:     req.TCPData.Options,
			MSS:         req.TCPData.MSS,
			WindowScale: req.TCPData.WindowScale,
			TTL:         req.TCPData.TTL,
			DF:          req.TCPData.DF,
		}
		ja4tResult := ja4t.ComputeJA4T(synData)
		if ja4tResult != nil {
			ja4tInfo = &JA4TInfo{
				Hash:      ja4tResult.Hash,
				Raw:       ja4tResult.RawFingerprint,
				Anomalies: ja4tResult.AnomalyFlags,
				GuessedOS: ja4tResult.ProbableOS,
			}
			features.Set(core.FeatureType("tcp_window_size"), float64(synData.WindowSize))
			features.Set(core.FeatureType("tcp_mss"), float64(synData.MSS))
			features.Set(core.FeatureType("tcp_ttl"), float64(synData.TTL))
			if synData.DF {
				features.Set(core.FeatureType("tcp_df"), 1.0)
			}
		}
	}

	if len(req.TCPPackets) > 0 {
		analyzer := tcp.NewTCPIPAnalyzer()
		for _, pkt := range req.TCPPackets {
			analyzer.AddPacket(pkt)
		}
		tcpResult, err := analyzer.AnalyzeStream()
		if err == nil {
			netAnalysis = &NetworkAnalysisResult{
				OS:             tcpResult.OS,
				OSFamily:       tcpResult.OSFamily,
				OSConfidence:   tcpResult.Confidence,
				IsVPN:          tcpResult.IsVPN,
				IsProxy:        tcpResult.IsProxy,
				IsNAT:          tcpResult.IsNAT,
				NetworkRisk:    tcpResult.RiskScore,
				InitialTTL:     tcpResult.InitialTTL,
				MSS:            tcpResult.MSS,
				AnomaliesFound: tcpResult.AnomaliesFound,
			}
			features.Set(core.FeatureType("network_risk"), tcpResult.RiskScore)
			if tcpResult.IsVPN {
				features.Set(core.FeatureType("is_vpn"), 1.0)
			}
			if tcpResult.IsProxy {
				features.Set(core.FeatureType("is_proxy"), 1.0)
			}
		}
	}

	return ja4tInfo, netAnalysis
}

// runPluginPipeline executes plugin analyzers and validators.
func (g *Gateway) runPluginPipeline(ctx context.Context, req *AnalyzeRequest, fingerprintHash string, classification *ml.ClassificationResult, risk *core.RiskAssessment) []PluginFinding {
	if g.pluginManager == nil {
		return nil
	}

	pluginData := map[string]interface{}{
		"fingerprint_hash": fingerprintHash,
		"client_ip":        req.ClientIP,
		"tls_version":      req.TLSVersion,
		"cipher_suites":    req.CipherSuites,
		"classification":   classification,
		"risk":             risk,
	}

	var findings []PluginFinding
	if results, err := g.pluginManager.ExecuteAnalyzers(ctx, pluginData); err == nil {
		for _, r := range results {
			msg := fmt.Sprintf("score=%.2f confidence=%.2f", r.Score, r.Confidence)
			category := ""
			if cat, ok := r.Labels["category"]; ok {
				category = cat
			}
			pluginName := ""
			if name, ok := r.Labels["plugin"]; ok {
				pluginName = name
			}
			findings = append(findings, PluginFinding{
				PluginName: pluginName,
				Category:   category,
				Message:    msg,
				RiskScore:  r.Score,
			})
		}
	}

	if vResult, err := g.pluginManager.ExecuteValidators(ctx, pluginData); err == nil && vResult != nil && !vResult.Valid {
		for _, issue := range vResult.Errors {
			findings = append(findings, PluginFinding{
				PluginName: "validator",
				Category:   "validation",
				Message:    issue,
			})
		}
		for _, warn := range vResult.Warnings {
			findings = append(findings, PluginFinding{
				PluginName: "validator",
				Category:   "warning",
				Message:    warn,
			})
		}
	}

	return findings
}

// enrichWithMLValidation adds ML service validation to the response.
func (g *Gateway) enrichWithMLValidation(response *AnalyzeResponse, features *core.FeatureVector) {
	if g.mlService == nil || !g.mlService.IsReady() {
		return
	}

	result := g.mlService.InferFromFeatures(features, nil)
	if result == nil {
		return
	}

	vr := &ml.ValidationResult{
		Valid:            !result.Forgery.IsForgery,
		ForgeryProb:      result.Forgery.ForgeryProb,
		ForgeryType:      result.Forgery.ForgeryType,
		ConsistencyScore: 1.0 - result.Forgery.ForgeryProb,
		BrowserFamily:    string(result.Browser.Family),
		Confidence:       result.Browser.Confidence,
	}
	response.MLValidation = vr
	if !vr.Valid {
		response.DefenseHints = append(response.DefenseHints,
			fmt.Sprintf("ML: forgery detected (type=%s, prob=%.2f)", vr.ForgeryType, vr.ForgeryProb))
	}
}

// extractFeatures extracts features
func (g *Gateway) extractFeatures(req *AnalyzeRequest) *core.FeatureVector {
	fv := core.NewFeatureVector()

	// TLS features
	fv.Set(core.FeatureTLSVersion, float64(req.TLSVersion))
	fv.Set(core.FeatureCipherSuites, float64(len(req.CipherSuites)))
	fv.Set(core.FeatureExtensions, float64(len(req.Extensions)))

	// HTTP features
	if req.Headers != nil {
		httpFV := g.extractor.ExtractFromHTTPHeaders(req.Headers)
		for ft, v := range httpFV.Features {
			fv.Set(ft, v)
		}
	}

	// Frontend features
	if req.Frontend != nil {
		frontendFV := g.extractor.ExtractFromFrontend(*req.Frontend)
		for ft, v := range frontendFV.Features {
			fv.Set(ft, v)
		}
	}

	return fv
}

// calculateTLSFingerprints calculates TLS fingerprints
func (g *Gateway) calculateTLSFingerprints(req *AnalyzeRequest) (*JA3Info, *JA4Info) {
	spec := core.ClientHelloSpec{
		TLSVersion:      req.TLSVersion,
		CipherSuites:    req.CipherSuites,
		Extensions:      req.Extensions,
		SupportedCurves: req.SupportedCurves,
	}

	// JA3
	ja3Result := calculateJA3(spec)

	// JA4
	ja4Result := calculateJA4(spec)

	return &JA3Info{
			Hash: ja3Result.Hash,
			Raw:  ja3Result.RawString,
		}, &JA4Info{
			Fingerprint: ja4Result.Fingerprint,
		}
}

// calculateHTTPFingerprint calculates HTTP fingerprint
func (g *Gateway) calculateHTTPFingerprint(req *AnalyzeRequest) *JA4HInfo {
	if req.Headers == nil {
		return nil
	}

	// Simplified JA4H calculation
	headerMap := req.Headers.ToMap()
	headers := make([]string, 0, len(headerMap))
	for name := range headerMap {
		headers = append(headers, name)
	}

	return &JA4HInfo{
		Fingerprint: fmt.Sprintf("ja4h_%s_%d", req.Method, len(headers)),
		Headers:     headers,
	}
}

// generateFingerprintHash generates a fingerprint hash
func (g *Gateway) generateFingerprintHash(req *AnalyzeRequest) string {
	data := fmt.Sprintf("%d:%v:%v:%s:%s",
		req.TLSVersion,
		req.CipherSuites,
		req.Extensions,
		req.ClientIP,
		req.Method,
	)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

// HTTPHandler is the HTTP handler function
func (g *Gateway) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	// Rate limit check
	clientIP := g.getClientIP(r)
	if !g.limiter.Allow(clientIP) {
		writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Limit request body size to prevent DoS
	r.Body = http.MaxBytesReader(w, r.Body, core.MaxRequestBodySize)

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// Set client IP
	req.ClientIP = clientIP

	// Execute analysis
	ctx, cancel := context.WithTimeout(r.Context(), AnalyzeTimeout)
	defer cancel()

	response, err := g.Analyze(ctx, &req)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SDKHandler is the SDK script handler function
func (g *Gateway) SDKHandler(w http.ResponseWriter, r *http.Request) {
	// Rate limit check
	clientIP := g.getClientIP(r)
	if !g.limiter.Allow(clientIP) {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Generate SDK code
	js := g.sdk.GenerateJSInjector(g.config.Endpoint + "/collect")

	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte(js))
}

// CollectHandler is the frontend data collection handler function
func (g *Gateway) CollectHandler(w http.ResponseWriter, r *http.Request) {
	// Use SDK's handler function
	g.sdk.HandleCollect(w, r)
}

// getClientIP gets the client IP
// Only trust proxy headers when RemoteAddr is in the TrustedProxies list
func (g *Gateway) getClientIP(r *http.Request) string {
	remoteIP := r.RemoteAddr
	if host, _, err := net.SplitHostPort(remoteIP); err == nil && host != "" {
		remoteIP = host
	}

	// Only read proxy headers from trusted proxy requests
	if g.isTrustedProxy(remoteIP) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if idx := strings.Index(xff, ","); idx != -1 {
				return strings.TrimSpace(xff[:idx])
			}
			return strings.TrimSpace(xff)
		}
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return strings.TrimSpace(xri)
		}
	}

	return remoteIP
}

// isTrustedProxy checks if IP is in the trusted proxy list
func (g *Gateway) isTrustedProxy(ip string) bool {
	for _, trusted := range g.config.TrustedProxies {
		if trusted == ip {
			return true
		}
	}
	return false
}

// Start starts the gateway service
func (g *Gateway) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc(g.config.Endpoint+"/analyze", g.HTTPHandler)
	mux.HandleFunc(g.config.Endpoint+"/sdk.js", g.SDKHandler)
	mux.HandleFunc(g.config.Endpoint+"/collect", g.CollectHandler)
	mux.HandleFunc(g.config.Endpoint+"/scan", g.V8ScannerHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "ok"}`))
	})

	addr := fmt.Sprintf(":%d", g.config.Port)
	slog.Info("Fingerprint Gateway starting", "addr", addr)
	return http.ListenAndServe(addr, mux)
}

// RateLimiter is the rate limiter
