package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// translated comment
type Server struct {
	server *http.Server
	addr   string
}

// translated comment
// translated comment
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

// translated comment
func (s *Server) Start() error {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("metrics server error: %v\n", err)
		}
	}()
	return nil
}

// translated comment
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// translated comment
func (s *Server) Addr() string {
	return s.addr
}

// translated comment
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// translated comment
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/metrics", http.StatusMovedPermanently)
		return
	}
	http.NotFound(w, r)
}

// translated comment
var DefaultServer *Server

// translated comment
func InitDefaultServer(addr string) {
	DefaultServer = NewServer(addr)
	DefaultServer.Start()
}

// translated comment
func StopDefaultServer() error {
	if DefaultServer == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return DefaultServer.Stop(ctx)
}
