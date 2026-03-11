// Package gateway provides gRPC API support.
// Note: this file uses a simplified gRPC implementation; full implementation requires protobuf code generation.
package gateway

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/vistone/fingerprint/modules/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// GRPCServer is a gRPC server wrapper
type GRPCServer struct {
	server  *grpc.Server
	gateway *Gateway
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(gateway *Gateway) *GRPCServer {
	// Create gRPC server with interceptors
	s := grpc.NewServer(
		grpc.UnaryInterceptor(rateLimitInterceptor(gateway)),
	)

	return &GRPCServer{
		server:  s,
		gateway: gateway,
	}
}

// rateLimitInterceptor is the rate limiting interceptor
func rateLimitInterceptor(gateway *Gateway) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract client IP from context
		clientIP := extractClientIP(ctx)

		// Check rate limit
		if !gateway.limiter.Allow(clientIP) {
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

// extractClientIP extracts client IP from gRPC context
func extractClientIP(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok || p == nil || p.Addr == nil {
		return "unknown"
	}
	addr := p.Addr.String()
	host, _, err := net.SplitHostPort(addr)
	if err == nil && host != "" {
		return host
	}
	// Compatible with addresses without port
	return strings.TrimSpace(addr)
}

// Start starts the gRPC server
func (s *GRPCServer) Start(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	fmt.Printf("gRPC server starting on port %d\n", port)
	return s.server.Serve(listener)
}

// Stop stops the gRPC server
func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}

// ==================== gRPC message conversion ====================

// GRPCAnalyzeRequest is the gRPC format analysis request
type GRPCAnalyzeRequest struct {
	ClientID        string
	TLSVersion      uint32
	CipherSuites    []uint32
	Extensions      []uint32
	SupportedCurves []uint32
	Headers         map[string]string
	ClientIP        string
}

// GRPCAnalyzeResponse is the gRPC format analysis response
type GRPCAnalyzeResponse struct {
	FingerprintHash  string
	Protocol         string
	BrowserFamily    string
	BrowserVersion   string
	Confidence       float64
	RiskLevel        string
	RiskScore        float64
	JA3Hash          string
	JA4Fingerprint   string
	DefenseHints     []string
	Cached           bool
	ProcessingTimeMs int64
}

// ConvertToGatewayRequest converts a gRPC request to a gateway request
func ConvertToGatewayRequest(req *GRPCAnalyzeRequest) *AnalyzeRequest {
	// Convert uint32 to uint16
	cipherSuites := make([]uint16, len(req.CipherSuites))
	for i, v := range req.CipherSuites {
		cipherSuites[i] = uint16(v)
	}

	extensions := make([]core.TLSExtension, len(req.Extensions))
	for i, v := range req.Extensions {
		extensions[i] = core.TLSExtension{Type: uint16(v)}
	}

	curves := make([]core.CurveID, len(req.SupportedCurves))
	for i, v := range req.SupportedCurves {
		curves[i] = core.CurveID(v)
	}

	// Build HTTP headers
	headers := &core.HTTPHeaders{
		UserAgent:      req.Headers["user-agent"],
		Accept:         req.Headers["accept"],
		AcceptLanguage: req.Headers["accept-language"],
		AcceptEncoding: req.Headers["accept-encoding"],
	}

	return &AnalyzeRequest{
		TLSVersion:      uint16(req.TLSVersion),
		CipherSuites:    cipherSuites,
		Extensions:      extensions,
		SupportedCurves: curves,
		Headers:         headers,
		ClientIP:        req.ClientIP,
	}
}

// ConvertFromGatewayResponse converts a gateway response to a gRPC response
func ConvertFromGatewayResponse(resp *AnalyzeResponse) *GRPCAnalyzeResponse {
	if resp == nil {
		return nil
	}

	result := &GRPCAnalyzeResponse{
		FingerprintHash:  resp.FingerprintHash,
		DefenseHints:     resp.DefenseHints,
		Cached:           resp.Cached,
		ProcessingTimeMs: resp.ProcessingTimeMs,
	}

	if resp.Classification != nil {
		result.Protocol = string(resp.Classification.Protocol)
		result.BrowserFamily = string(resp.Classification.Family)
		result.BrowserVersion = resp.Classification.Version
		result.Confidence = resp.Classification.Confidence
	}

	if resp.RiskAssessment != nil {
		result.RiskLevel = resp.RiskAssessment.Level.String()
		result.RiskScore = resp.RiskAssessment.Score
	}

	if resp.JA3 != nil {
		result.JA3Hash = resp.JA3.Hash
	}

	if resp.JA4 != nil {
		result.JA4Fingerprint = resp.JA4.Fingerprint
	}

	return result
}

// ==================== gRPC Client ====================

// GRPCClient is a gRPC client
type GRPCClient struct {
	conn *grpc.ClientConn
}

// NewGRPCClient creates a gRPC client
func NewGRPCClient(address string) (*GRPCClient, error) {
	// Use insecure connection (production should use TLS)
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &GRPCClient{conn: conn}, nil
}

// Close closes the connection
func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

// Connection returns the underlying connection (for advanced usage)
func (c *GRPCClient) Connection() *grpc.ClientConn {
	return c.conn
}

// ==================== Hybrid Server ====================

// HybridServer is an HTTP + gRPC hybrid server
type HybridServer struct {
	gateway  *Gateway
	httpPort int
	grpcPort int
}

// NewHybridServer creates a hybrid server
func NewHybridServer(gateway *Gateway, httpPort, grpcPort int) *HybridServer {
	return &HybridServer{
		gateway:  gateway,
		httpPort: httpPort,
		grpcPort: grpcPort,
	}
}

// Start starts the hybrid server
func (s *HybridServer) Start() error {
	// Start gRPC server (in background)
	grpcServer := NewGRPCServer(s.gateway)
	go func() {
		if err := grpcServer.Start(s.grpcPort); err != nil {
			fmt.Printf("gRPC server error: %v\n", err)
		}
	}()

	// Start HTTP server (blocking)
	return s.gateway.Start()
}

// Protocol represents the server protocol type
type Protocol string

const (
	// ProtocolHTTP is the HTTP protocol
	ProtocolHTTP Protocol = "http"
	// ProtocolGRPC is the gRPC protocol
	ProtocolGRPC Protocol = "grpc"
)

// ServerInfo contains server information
type ServerInfo struct {
	Protocol    Protocol
	Port        int
	Status      string
	Connections int
}

// GetServerInfo returns server information
func (s *HybridServer) GetServerInfo() []ServerInfo {
	return []ServerInfo{
		{ProtocolHTTP, s.httpPort, "running", 0},
		{ProtocolGRPC, s.grpcPort, "running", 0},
	}
}
