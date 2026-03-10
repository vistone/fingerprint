// Fingerprint Gateway Server
// Main entry point for the fingerprint analysis gateway service
package main

import (
	"context"
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

var (
	httpPort = flag.Int("http-port", 8080, "HTTP server port")
	grpcPort = flag.Int("grpc-port", 9090, "gRPC server port")
	version  = flag.String("version", "v3.0.0", "Service version")
	health   = flag.Bool("health", false, "Run health check and exit")
)

// metricsMiddleware 记录请求指标的中间件
func metricsMiddleware(next http.HandlerFunc, endpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 创建一个响应记录器来捕获状态码
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next(rec, r)

		// 计算延迟
		latency := time.Since(start).Milliseconds()
		success := rec.statusCode < 400

		// 获取客户端 IP
		clientIP := getClientIP(r)

		// 构建请求记录
		req := web.RequestRecord{
			Timestamp: time.Now(),
			IP:        clientIP,
			Method:    r.Method,
			Path:      r.URL.Path,
			JA3:       r.Header.Get("X-JA3-Fingerprint"),
			Latency:   latency,
			Status:    rec.statusCode,
		}

		// 记录指标
		web.RecordAPIMetrics(req, success, "", "")
	}
}

// getClientIP 获取客户端 IP
func getClientIP(r *http.Request) string {
	// 检查 X-Forwarded-For 头
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// 取第一个 IP
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return xff
	}

	// 检查 X-Real-IP 头
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用 RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// responseRecorder 包装 http.ResponseWriter 以捕获状态码
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *responseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

func main() {
	flag.Parse()

	// Health check mode
	if *health {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", *httpPort))
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		os.Exit(0)
	}

	log.Printf("Starting Fingerprint Gateway %s", *version)

	// 初始化日志捕获系统（将 log 输出存入缓冲区供前端实时查看）
	web.InitLogCapture()

	// Create gateway with custom config
	config := *gateway.DefaultGatewayConfig
	config.Port = *httpPort
	if v := strings.TrimSpace(os.Getenv("FP_P3_ENABLED")); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			config.P3Enabled = parsed
		}
	}
	if v := strings.TrimSpace(os.Getenv("FP_P3_PROFILE")); v != "" {
		config.P3ProfileID = v
	}
	if v := strings.TrimSpace(os.Getenv("FP_P3_CONFIG_DIR")); v != "" {
		config.P3ConfigDir = v
	}
	if v := strings.TrimSpace(os.Getenv("FP_P3_PROXY_TARGET")); v != "" {
		config.P3ProxyTarget = v
	}
	if v := strings.TrimSpace(os.Getenv("FP_P3_INJECT_CONSISTENCY")); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			config.P3InjectConsist = parsed
		}
	}
	if v := strings.TrimSpace(os.Getenv("FP_P3_DIRECT_PROXY")); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			config.P3DirectProxy = parsed
		}
	}
	if v := strings.TrimSpace(os.Getenv("FP_SCANNER_USE_BROWSER")); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			config.ScannerUseBrowser = parsed
		}
	}
	if v := strings.TrimSpace(os.Getenv("FP_SCANNER_BROWSER_WS")); v != "" {
		config.ScannerBrowserWS = v
	}
	if v := strings.TrimSpace(os.Getenv("FP_SCANNER_BROWSER_TIMEOUT_MS")); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms > 0 {
			config.ScannerBrowserTimeout = time.Duration(ms) * time.Millisecond
		}
	}
	gw := gateway.NewGateway(&config)

	// 写入启动日志
	web.WriteLog("INFO", "gateway", "Fingerprint Gateway %s starting", *version)
	web.WriteLog("INFO", "gateway", "HTTP server will listen on port %d", *httpPort)
	web.WriteLog("INFO", "gateway", "gRPC server will listen on port %d", *grpcPort)
	if config.P3Enabled {
		web.WriteLog("INFO", "p3", "P3 anti-detection enabled, profile: %s", config.P3ProfileID)
	}
	if config.AgentEnabled {
		web.WriteLog("INFO", "agent", "Autonomous Security Agent enabled")
	}
	if config.ScannerUseBrowser {
		web.WriteLog("INFO", "scanner", "Browser-based scanner enabled, WS: %s", config.ScannerBrowserWS)
	}

	// Setup HTTP routes
	mux := http.NewServeMux()

	// API endpoints with metrics recording
	mux.HandleFunc("/api/v1/analyze", metricsMiddleware(gw.HTTPHandler, "analyze"))
	mux.HandleFunc("/api/v1/classify", metricsMiddleware(gw.HTTPHandler, "classify")) // Alias
	mux.HandleFunc("/api/v1/sdk.js", gw.SDKHandler)
	mux.HandleFunc("/api/v1/collect", gw.CollectHandler)

	// P3 Anti-Detection API endpoints
	mux.HandleFunc("/api/v1/p3/antidetect.js", gw.AntiDetectCodeHandler)
	mux.HandleFunc("/api/v1/p3/profiles", gw.ProfileListHandler)
	mux.HandleFunc("/api/v1/p3/profile", gw.ProfileDetailHandler)

	// JavaScript Fingerprint Detection Scanner
	mux.HandleFunc("/api/v1/scan", gw.V8ScannerHandler)
	// P3 Proxy mode (if proxy target configured)
	// 访问 /proxy/* 会代理到目标服务并自动注入 P3 代码
	mux.Handle("/proxy/", http.StripPrefix("/proxy", gw.InjectProxyHandler()))

	// P3 Direct Proxy mode
	// 启用后，客户端直接访问网关根路径即可被透明代理并自动注入。
	// API 路由优先级高于 "/"，因此 /api/* 不受影响。
	if config.P3DirectProxy && config.P3ProxyTarget != "" {
		mux.Handle("/", gw.InjectProxyHandler())
		log.Printf("P3 direct proxy mode enabled, target: %s", config.P3ProxyTarget)
	}

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"status":"ok","version":"%s"}`, *version)))
	})

	// Metrics endpoint (placeholder)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("# Fingerprint Gateway Metrics\n"))
	})

	// Admin web console
	webHandler := web.NewHandler(gw)
	webHandler.RegisterRoutes(mux)

	grpcServer := gateway.NewGRPCServer(gw)
	grpcErrCh := make(chan error, 2)

	// Start gRPC server in background
	go func() {
		log.Printf("gRPC server starting on port %d", *grpcPort)
		if err := grpcServer.Start(*grpcPort); err != nil {
			grpcErrCh <- err
			return
		}
		grpcErrCh <- nil
	}()

	// Start HTTP server
	addr := fmt.Sprintf(":%d", *httpPort)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	log.Printf("HTTP server starting on %s", addr)

	httpErrCh := make(chan error, 2)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			httpErrCh <- err
			return
		}
		httpErrCh <- nil
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("Received signal %s, shutting down gracefully...", sig.String())
	case err := <-httpErrCh:
		if err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	case err := <-grpcErrCh:
		if err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}
	grpcServer.Stop()
}
