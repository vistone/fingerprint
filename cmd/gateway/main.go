// Fingerprint Gateway Server.
// Main entry point for the fingerprint analysis gateway service.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/vistone/fingerprint/modules/gateway"
	"github.com/vistone/fingerprint/modules/gateway/web"
)

const (
	defaultHTTPPort             = 8080
	defaultGRPCPort             = 9090
	httpSuccessUpperBound       = 400
	closedLoopDefaultCandidates = 5
	closedLoopDefaultNoise      = 0.1
	serverErrorChannelSize      = 2
	healthCheckTimeout          = 3 * time.Second
	shutdownTimeout             = 10 * time.Second
	httpReadHeaderTimeout       = 5 * time.Second
)

var (
	httpPort = flag.Int("http-port", defaultHTTPPort, "HTTP server port")
	grpcPort = flag.Int("grpc-port", defaultGRPCPort, "gRPC server port")
	version  = flag.String("version", "v3.0.0", "Service version")
	health   = flag.Bool("health", false, "Run health check and exit")
)

// metricsMiddleware records request metrics for API handlers.
func metricsMiddleware(nextHandler http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		startTime := time.Now()

		recorder := &responseRecorder{
			ResponseWriter: writer,
			statusCode:     http.StatusOK,
		}

		nextHandler(recorder, request)

		latencyMilliseconds := time.Since(startTime).Milliseconds()
		success := recorder.statusCode < httpSuccessUpperBound
		clientIP := getClientIP(request)

		record := web.RequestRecord{
			Timestamp:      time.Now(),
			IP:             clientIP,
			Method:         request.Method,
			Path:           request.URL.Path,
			Classification: "",
			JA3:            request.Header.Get("X-Ja3-Fingerprint"),
			Latency:        latencyMilliseconds,
			Status:         recorder.statusCode,
		}

		web.RecordAPIMetrics(record, success, "", "")
	}
}

// getClientIP resolves client IP from forwarded headers or remote address.
func getClientIP(request *http.Request) string {
	if xForwardedFor := request.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		firstIP, _, hasComma := strings.Cut(xForwardedFor, ",")
		if hasComma {
			return strings.TrimSpace(firstIP)
		}

		return xForwardedFor
	}

	if xRealIP := request.Header.Get("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}

	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		return request.RemoteAddr
	}

	return host
}

// responseRecorder wraps http.ResponseWriter to capture status code.
type responseRecorder struct {
	http.ResponseWriter

	statusCode int
}

func (recorder *responseRecorder) WriteHeader(code int) {
	recorder.statusCode = code
	recorder.ResponseWriter.WriteHeader(code)
}

func buildGatewayConfig(port int) gateway.GatewayConfig {
	config := *gateway.DefaultGatewayConfig
	config.Port = port
	applyGatewayEnv(&config)
	applyClosedLoopEnv(&config)

	return config
}

func applyGatewayEnv(config *gateway.GatewayConfig) {
	applyAntiDetectEnv(config)
	applyScannerEnv(config)
	applyMLEnv(config)
	applyRiskAndProxyEnv(config)
}

func applyAntiDetectEnv(config *gateway.GatewayConfig) {
	if parsed, ok := readEnvBool("FP_ANTIDETECT_ENABLED"); ok {
		config.AntiDetectEnabled = parsed
	}

	if profileID, ok := readEnvString("FP_ANTIDETECT_PROFILE"); ok {
		config.AntiDetectProfileID = profileID
	}

	if configDir, ok := readEnvString("FP_ANTIDETECT_CONFIG_DIR"); ok {
		config.AntiDetectConfigDir = configDir
	}

	if proxyTarget, ok := readEnvString("FP_ANTIDETECT_PROXY_TARGET"); ok {
		config.AntiDetectProxyTarget = proxyTarget
	}

	if parsed, ok := readEnvBool("FP_ANTIDETECT_INJECT_CONSISTENCY"); ok {
		config.AntiDetectInjectConsist = parsed
	}

	if parsed, ok := readEnvBool("FP_ANTIDETECT_DIRECT_PROXY"); ok {
		config.AntiDetectDirectProxy = parsed
	}
}

func applyScannerEnv(config *gateway.GatewayConfig) {
	if parsed, ok := readEnvBool("FP_SCANNER_USE_BROWSER"); ok {
		config.ScannerUseBrowser = parsed
	}

	if browserWS, ok := readEnvString("FP_SCANNER_BROWSER_WS"); ok {
		config.ScannerBrowserWS = browserWS
	}

	if timeoutMilliseconds, ok := readEnvInt("FP_SCANNER_BROWSER_TIMEOUT_MS"); ok && timeoutMilliseconds > 0 {
		config.ScannerBrowserTimeout = time.Duration(timeoutMilliseconds) * time.Millisecond
	}
}

func applyMLEnv(config *gateway.GatewayConfig) {
	if parsed, ok := readEnvBool("FP_ML_SERVICE_ENABLED"); ok {
		config.MLServiceEnabled = parsed
	}

	if classifierPath, ok := readEnvString("FP_ML_CLASSIFIER_PATH"); ok {
		config.MLClassifierPath = classifierPath
	}

	if trainingDataPath, ok := readEnvString("FP_ML_TRAINING_DATA"); ok {
		config.MLTrainingData = trainingDataPath
	}
}

func applyRiskAndProxyEnv(config *gateway.GatewayConfig) {
	if threshold, ok := readEnvFloat("FP_RISK_THRESHOLD"); ok && threshold > 0 {
		config.RiskThreshold = threshold
	}

	if trustedProxies, ok := readEnvString("FP_TRUSTED_PROXIES"); ok {
		config.TrustedProxies = splitCSV(trustedProxies)
	}

	if pluginConfigPath, ok := readEnvString("FP_PLUGIN_CONFIG_PATH"); ok {
		config.PluginConfigPath = pluginConfigPath
	}
}

func applyClosedLoopEnv(config *gateway.GatewayConfig) {
	if parsed, ok := readEnvBool("FP_CLOSED_LOOP_ENABLED"); ok {
		config.ClosedLoopEnabled = parsed
	}

	if !config.ClosedLoopEnabled {
		return
	}

	closedLoopConfig := &gateway.ClosedLoopConfig{
		Enabled:            true,
		TrainingInterval:   time.Hour,
		CandidatesPerCycle: closedLoopDefaultCandidates,
		NoiseIntensity:     closedLoopDefaultNoise,
	}

	if duration, ok := readEnvDuration("FP_CLOSED_LOOP_TRAINING_INTERVAL"); ok && duration > 0 {
		closedLoopConfig.TrainingInterval = duration
	}

	if candidates, ok := readEnvInt("FP_CLOSED_LOOP_CANDIDATES"); ok && candidates > 0 {
		closedLoopConfig.CandidatesPerCycle = candidates
	}

	if noise, ok := readEnvFloat("FP_CLOSED_LOOP_NOISE"); ok && noise >= 0 && noise <= 1 {
		closedLoopConfig.NoiseIntensity = noise
	}

	config.ClosedLoopConfig = closedLoopConfig
}

func readEnvString(name string) (string, bool) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", false
	}

	return value, true
}

func readEnvBool(name string) (bool, bool) {
	value, ok := readEnvString(name)
	if !ok {
		return false, false
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, false
	}

	return parsed, true
}

func readEnvInt(name string) (int, bool) {
	value, ok := readEnvString(name)
	if !ok {
		return 0, false
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}

	return parsed, true
}

func readEnvFloat(name string) (float64, bool) {
	value, ok := readEnvString(name)
	if !ok {
		return 0, false
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}

	return parsed, true
}

func readEnvDuration(name string) (time.Duration, bool) {
	value, ok := readEnvString(name)
	if !ok {
		return 0, false
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, false
	}

	return parsed, true
}

func writeStartupLogs(config *gateway.GatewayConfig, modules *runtimeModules) {
	web.WriteLog("INFO", "gateway", "Fingerprint Gateway %s starting", *version)
	web.WriteLog("INFO", "gateway", "HTTP server will listen on port %d", *httpPort)
	web.WriteLog("INFO", "gateway", "gRPC server will listen on port %d", *grpcPort)

	if config.AntiDetectEnabled {
		web.WriteLog("INFO", "antidetect", "Anti-detection enabled, profile: %s", config.AntiDetectProfileID)
	}

	if config.AgentEnabled {
		web.WriteLog("INFO", "agent", "Autonomous Security Agent enabled")
	}

	if config.ScannerUseBrowser {
		web.WriteLog("INFO", "scanner", "Browser-based scanner enabled, WS: %s", config.ScannerBrowserWS)
	}

	if config.MLServiceEnabled {
		web.WriteLog("INFO", "ml", "MLService enabled")
	}

	if config.PluginConfigPath != "" {
		web.WriteLog("INFO", "plugin", "Plugin system enabled, config: %s", config.PluginConfigPath)
	}

	if config.ClosedLoopEnabled {
		web.WriteLog("INFO", "closedloop", "Adversarial closed-loop training enabled")
	}

	if modules.waf != nil {
		web.WriteLog("INFO", "waf", "WAF integrated into gateway request pipeline")
	}

	if modules.crawler != nil {
		crawlerConfig := modules.crawler.ConfigSnapshot()
		web.WriteLog("INFO", "crawler", "Crawler integrated into gateway runtime, targets: %d", len(crawlerConfig.TargetURLs))
	}

	if config.RiskThreshold > 0 {
		web.WriteLog("INFO", "risk", "Risk threshold set to %.2f", config.RiskThreshold)
	}
}

func registerRoutes(mux *http.ServeMux, gatewayService *gateway.Gateway, modules *runtimeModules, config *gateway.GatewayConfig) {
	mux.HandleFunc("/api/v1/analyze", protectWithWAF(modules.waf, gatewayService.HTTPHandler))
	mux.HandleFunc("/api/v1/classify", protectWithWAF(modules.waf, gatewayService.HTTPHandler))
	mux.HandleFunc("/api/v1/sdk.js", gatewayService.SDKHandler)
	mux.HandleFunc("/api/v1/collect", protectWithWAF(modules.waf, gatewayService.CollectHandler))

	mux.HandleFunc("/api/v1/antidetect/antidetect.js", gatewayService.AntiDetectCodeHandler)
	mux.HandleFunc("/api/v1/antidetect/profiles", gatewayService.ProfileListHandler)
	mux.HandleFunc("/api/v1/antidetect/profile", gatewayService.ProfileDetailHandler)
	mux.HandleFunc("/api/v1/scan", protectWithWAF(modules.waf, gatewayService.V8ScannerHandler))
	mux.Handle("/proxy/", http.StripPrefix("/proxy", gatewayService.InjectProxyHandler()))

	if config.AntiDetectDirectProxy && config.AntiDetectProxyTarget != "" {
		mux.Handle("/", gatewayService.InjectProxyHandler())
		log.Printf("Direct proxy mode enabled, target: %s", config.AntiDetectProxyTarget)
	}

	mux.HandleFunc("/api/v1/closedloop/status", closedLoopStatusHandler(gatewayService))
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/metrics", metricsHandler)

	webHandler := web.NewHandler(gatewayService)
	webHandler.RegisterRoutes(mux)
}

func closedLoopStatusHandler(gatewayService *gateway.Gateway) http.HandlerFunc {
	return func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		stats := gatewayService.ClosedLoopStats()
		if stats == nil {
			writeBytes(writer, []byte(`{"enabled":false}`))

			return
		}

		data := fmt.Sprintf(
			`{"enabled":true,"cyclesCompleted":%d,"profilesGenerated":%d,"detectionsProcessed":%d,"modelsEvolved":%d,"lastCycleTime":"%s"}`,
			stats.CyclesCompleted,
			stats.ProfilesGenerated,
			stats.DetectionsProcessed,
			stats.ModelsEvolved,
			stats.LastCycleTime.Format(time.RFC3339),
		)
		writeBytes(writer, []byte(data))
	}
}

func healthHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	payload := fmt.Sprintf(`{"status":"ok","version":"%s"}`, *version)
	writeBytes(writer, []byte(payload))
}

func metricsHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	writeBytes(writer, []byte("# Fingerprint Gateway Metrics\n"))
}

func writeBytes(writer http.ResponseWriter, payload []byte) {
	_, err := writer.Write(payload)
	if err != nil {
		log.Printf("response write error: %v", err)
	}
}

func runHealthCheck() int {
	healthURL := fmt.Sprintf("http://localhost:%d/health", *httpPort)
	client := new(http.Client)
	client.Timeout = healthCheckTimeout

	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, healthURL, nil)
	if err != nil {
		return 1
	}

	response, err := client.Do(request)
	if err != nil {
		return 1
	}

	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		return 1
	}

	return 0
}

func main() {
	flag.Parse()

	if *health {
		os.Exit(runHealthCheck())
	}

	log.Printf("Starting Fingerprint Gateway %s", *version)
	web.InitLogCapture()

	config := buildGatewayConfig(*httpPort)
	gatewayService := gateway.NewGateway(&config)
	modules := configureRuntimeModules(&config)
	attachRuntimeModules(gatewayService, modules)
	writeStartupLogs(&config, modules)

	mux := http.NewServeMux()
	registerRoutes(mux, gatewayService, modules, &config)

	grpcServer := gateway.NewGRPCServer(gatewayService)
	grpcErrorChannel := make(chan error, serverErrorChannelSize)

	go func() {
		log.Printf("gRPC server starting on port %d", *grpcPort)

		err := grpcServer.Start(*grpcPort)
		if err != nil {
			grpcErrorChannel <- err

			return
		}

		grpcErrorChannel <- nil
	}()

	address := fmt.Sprintf(":%d", *httpPort)
	httpServer := new(http.Server)
	httpServer.Addr = address
	httpServer.Handler = mux
	httpServer.ReadHeaderTimeout = httpReadHeaderTimeout

	log.Printf("HTTP server starting on %s", address)

	httpErrorChannel := make(chan error, serverErrorChannelSize)

	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			httpErrorChannel <- err

			return
		}

		httpErrorChannel <- nil
	}()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	select {
	case signalValue := <-signalChannel:
		log.Printf("Received signal %s, shutting down gracefully...", signalValue.String())
	case err := <-httpErrorChannel:
		if err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	case err := <-grpcErrorChannel:
		if err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err := httpServer.Shutdown(shutdownContext)
	if err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	grpcServer.Stop()
}
