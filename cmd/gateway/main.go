// Fingerprint Gateway Server.
// Main entry point for the fingerprint analysis gateway service.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	adminMux := http.NewServeMux()
	webHandler.RegisterRoutes(adminMux)
	registerProtectedAdminRoutes(mux, adminMux, config)
}

func registerProtectedAdminRoutes(mux *http.ServeMux, adminMux *http.ServeMux, config *gateway.GatewayConfig) {
	protectedHandler := protectAdminRoutes(adminMux, config)
	mux.Handle("/admin", protectedHandler)
	mux.Handle("/admin/", protectedHandler)
	mux.Handle("/api/admin/", protectedHandler)
}

func protectAdminRoutes(adminMux *http.ServeMux, config *gateway.GatewayConfig) http.Handler {
	if len(config.APIKeys) == 0 {
		return http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusUnauthorized)
			writeBytes(writer, []byte(`{"error":"missing API key"}`))
		})
	}

	return gateway.NewAPIKeyAuth(config.APIKeys, nil).Middleware(adminMux)
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
	serverRuntime := startServerRuntime(mux, gatewayService)
	waitForShutdownSignal(serverRuntime)
	shutdownServers(serverRuntime)
}

type runtimeServers struct {
	httpServer    *http.Server
	grpcServer    *gateway.GRPCServer
	httpErrorChan chan error
	grpcErrorChan chan error
}

func startServerRuntime(mux *http.ServeMux, gatewayService *gateway.Gateway) *runtimeServers {
	grpcServer := gateway.NewGRPCServer(gatewayService)

	tlsConfig, err := buildGRPCTLSConfigFromEnv()
	if err != nil {
		if errors.Is(err, errGRPCTLSNotConfigured) {
			log.Printf("gRPC TLS disabled: not configured")
		} else {
			log.Printf("gRPC TLS disabled: %v", err)
		}
	} else if tlsConfig != nil {
		grpcServer = gateway.NewGRPCServerWithTLS(gatewayService, tlsConfig)

		log.Printf("gRPC TLS enabled")
	}

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

	return &runtimeServers{
		httpServer:    httpServer,
		grpcServer:    grpcServer,
		httpErrorChan: httpErrorChannel,
		grpcErrorChan: grpcErrorChannel,
	}
}

func waitForShutdownSignal(runtime *runtimeServers) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	select {
	case signalValue := <-signalChannel:
		log.Printf("Received signal %s, shutting down gracefully...", signalValue.String())
	case err := <-runtime.httpErrorChan:
		if err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	case err := <-runtime.grpcErrorChan:
		if err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}
}

func shutdownServers(runtime *runtimeServers) {
	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err := runtime.httpServer.Shutdown(shutdownContext)
	if err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	runtime.grpcServer.Stop()
}
