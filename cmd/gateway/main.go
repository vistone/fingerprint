// Fingerprint Gateway Server
// Main entry point for the fingerprint analysis gateway service
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/vistone/fingerprint/modules/gateway"
	"github.com/vistone/fingerprint/modules/gateway/web"
)

var (
	httpPort = flag.Int("http-port", 8080, "HTTP server port")
	grpcPort = flag.Int("grpc-port", 9090, "gRPC server port")
	version  = flag.String("version", "v3.0.0", "Service version")
	health   = flag.Bool("health", false, "Run health check and exit")
)

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

	// Create gateway with custom config
	config := *gateway.DefaultGatewayConfig
	config.Port = *httpPort
	gw := gateway.NewGateway(&config)

	// Setup HTTP routes
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/v1/analyze", gw.HTTPHandler)
	mux.HandleFunc("/api/v1/classify", gw.HTTPHandler) // Alias
	mux.HandleFunc("/api/v1/sdk.js", gw.SDKHandler)
	mux.HandleFunc("/api/v1/collect", gw.CollectHandler)

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

	// Start gRPC server in background
	go func() {
		grpcServer := gateway.NewGRPCServer(gw)
		log.Printf("gRPC server starting on port %d", *grpcPort)
		if err := grpcServer.Start(*grpcPort); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Start HTTP server
	addr := fmt.Sprintf(":%d", *httpPort)
	log.Printf("HTTP server starting on %s", addr)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down gracefully...")
		os.Exit(0)
	}()

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
