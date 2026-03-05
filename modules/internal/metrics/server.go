package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server 指标 HTTP 服务器
type Server struct {
	server *http.Server
	addr   string
}

// NewServer 创建指标服务器
// addr: 监听地址，如 ":8080" 或 "0.0.0.0:9090"
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

// Start 启动指标服务器（非阻塞）
func (s *Server) Start() error {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("metrics server error: %v\n", err)
		}
	}()
	return nil
}

// Stop 优雅关闭服务器
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Addr 返回服务器地址
func (s *Server) Addr() string {
	return s.addr
}

// healthHandler 健康检查端点
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// indexHandler 根路径重定向到 /metrics
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/metrics", http.StatusMovedPermanently)
		return
	}
	http.NotFound(w, r)
}

// DefaultServer 默认指标服务器实例
var DefaultServer *Server

// InitDefaultServer 初始化默认指标服务器
func InitDefaultServer(addr string) {
	DefaultServer = NewServer(addr)
	DefaultServer.Start()
}

// StopDefaultServer 关闭默认指标服务器
func StopDefaultServer() error {
	if DefaultServer == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return DefaultServer.Stop(ctx)
}
