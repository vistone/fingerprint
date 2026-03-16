package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server hosts the metrics HTTP endpoint
type Server struct {
	server *http.Server
	addr   string
}

// NewServer creates a metrics server
// addr: listen address, e.g. ":8080" or "0.0.0.0:9090"
func NewServer(addr string) *Server {
	if addr == "" {
		addr = ":8080"
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/", indexHandler)

	return &Server{
		addr: addr,
		server: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

// Start runs the metrics server (non-blocking)
func (s *Server) Start() error {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("metrics server error: %v\n", err)
		}
	}()
	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Addr returns server address
func (s *Server) Addr() string {
	return s.addr
}

// healthHandler serves health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// indexHandler redirects root path to /metrics
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/metrics", http.StatusMovedPermanently)
		return
	}
	http.NotFound(w, r)
}

// DefaultServer is the default metrics server instance
var DefaultServer *Server

// InitDefaultServer initializes default metrics server
func InitDefaultServer(addr string) {
	DefaultServer = NewServer(addr)
	DefaultServer.Start()
}

// StopDefaultServer stops default metrics server
func StopDefaultServer() error {
	if DefaultServer == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return DefaultServer.Stop(ctx)
}
