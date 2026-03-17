// Package gateway provides gRPC API support.
package gateway

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/vistone/fingerprint/modules/core"
	gatewayv1 "github.com/vistone/fingerprint/modules/gateway/proto/fingerprint/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// GRPCServer is a gRPC server wrapper.
type GRPCServer struct {
	server  *grpc.Server
	gateway *Gateway
}

type gatewayGRPCService struct {
	gatewayv1.UnimplementedFingerprintGatewayServer
	gateway *Gateway
}

func (s *gatewayGRPCService) Analyze(ctx context.Context, req *gatewayv1.AnalyzeRequest) (*gatewayv1.AnalyzeResponse, error) {
	if s.gateway == nil {
		return nil, status.Error(codes.Unavailable, "gateway unavailable")
	}

	analysisReq := ConvertToGatewayRequest(&GRPCAnalyzeRequest{
		ClientID:        req.GetClientId(),
		TLSVersion:      req.GetTlsVersion(),
		CipherSuites:    req.GetCipherSuites(),
		Extensions:      req.GetExtensions(),
		SupportedCurves: req.GetSupportedCurves(),
		Headers:         req.GetHeaders(),
		ClientIP:        req.GetClientIp(),
	})

	if strings.TrimSpace(analysisReq.ClientIP) == "" {
		analysisReq.ClientIP = extractClientIP(ctx)
	}

	analysisCtx, cancel := context.WithTimeout(ctx, AnalyzeTimeout)
	defer cancel()

	resp, err := s.gateway.Analyze(analysisCtx, analysisReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "analyze failed: %v", err)
	}

	legacyResp := ConvertFromGatewayResponse(resp)
	return &gatewayv1.AnalyzeResponse{
		FingerprintHash:  legacyResp.FingerprintHash,
		Protocol:         legacyResp.Protocol,
		BrowserFamily:    legacyResp.BrowserFamily,
		BrowserVersion:   legacyResp.BrowserVersion,
		Confidence:       legacyResp.Confidence,
		RiskLevel:        legacyResp.RiskLevel,
		RiskScore:        legacyResp.RiskScore,
		Ja3Hash:          legacyResp.JA3Hash,
		Ja4Fingerprint:   legacyResp.JA4Fingerprint,
		DefenseHints:     legacyResp.DefenseHints,
		Cached:           legacyResp.Cached,
		ProcessingTimeMs: legacyResp.ProcessingTimeMs,
	}, nil
}

func (s *gatewayGRPCService) Health(context.Context, *gatewayv1.HealthRequest) (*gatewayv1.HealthResponse, error) {
	return &gatewayv1.HealthResponse{
		Status:  "SERVING",
		Service: "fingerprint.gateway.v1.FingerprintGateway",
	}, nil
}

// NewGRPCServer creates a new gRPC server with plaintext transport.
func NewGRPCServer(gateway *Gateway) *GRPCServer {
	return NewGRPCServerWithTLS(gateway, nil)
}

// NewGRPCServerWithTLS creates a gRPC server and enables TLS when tlsConfig is not nil.
func NewGRPCServerWithTLS(gateway *Gateway, tlsConfig *tls.Config) *GRPCServer {
	// Create gRPC server with interceptors.
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(rateLimitInterceptor(gateway)),
	}
	if tlsConfig != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	s := grpc.NewServer(opts...)
	gatewayv1.RegisterFingerprintGatewayServer(s, &gatewayGRPCService{gateway: gateway})

	return &GRPCServer{
		server:  s,
		gateway: gateway,
	}
}

// rateLimitInterceptor is the rate limiting interceptor.
func rateLimitInterceptor(gateway *Gateway) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract client IP from context.
		clientIP := extractClientIP(ctx)

		// Check rate limit.
		if gateway != nil && gateway.limiter != nil && !gateway.limiter.Allow(clientIP) {
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

// extractClientIP extracts client IP from gRPC context.
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
	// Compatible with addresses without port.
	return strings.TrimSpace(addr)
}

// Start starts the gRPC server.
func (s *GRPCServer) Start(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	slog.Info("gRPC server starting", "port", port)
	return s.server.Serve(listener)
}

// Stop stops the gRPC server.
func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}

// ==================== Compatibility conversion types ====================

// GRPCAnalyzeRequest is the compatibility gRPC-format analysis request.
type GRPCAnalyzeRequest struct {
	ClientID        string
	TLSVersion      uint32
	CipherSuites    []uint32
	Extensions      []uint32
	SupportedCurves []uint32
	Headers         map[string]string
	ClientIP        string
}

// GRPCAnalyzeResponse is the compatibility gRPC-format analysis response.
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

// ConvertToGatewayRequest converts a gRPC request to a gateway request.
func ConvertToGatewayRequest(req *GRPCAnalyzeRequest) *AnalyzeRequest {
	// Convert uint32 to uint16.
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

	// Build HTTP headers.
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

// ConvertFromGatewayResponse converts a gateway response to a gRPC response.
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

// GRPCClient is a gRPC client.
type GRPCClient struct {
	conn   *grpc.ClientConn
	client gatewayv1.FingerprintGatewayClient
}

// NewGRPCClient creates a plaintext gRPC client.
func NewGRPCClient(address string) (*GRPCClient, error) {
	return newGRPCClient(address, insecure.NewCredentials(), 5*time.Second)
}

// NewGRPCClientWithTLS creates a TLS-enabled gRPC client.
func NewGRPCClientWithTLS(address string, tlsConfig *tls.Config) (*GRPCClient, error) {
	if tlsConfig == nil {
		return nil, fmt.Errorf("tls config is required")
	}
	return newGRPCClient(address, credentials.NewTLS(tlsConfig), 5*time.Second)
}

func newGRPCClient(address string, creds credentials.TransportCredentials, timeout time.Duration) (*GRPCClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(creds), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &GRPCClient{
		conn:   conn,
		client: gatewayv1.NewFingerprintGatewayClient(conn),
	}, nil
}

// Analyze calls the gRPC Analyze endpoint.
func (c *GRPCClient) Analyze(ctx context.Context, req *GRPCAnalyzeRequest) (*GRPCAnalyzeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	resp, err := c.client.Analyze(ctx, &gatewayv1.AnalyzeRequest{
		ClientId:        req.ClientID,
		TlsVersion:      req.TLSVersion,
		CipherSuites:    req.CipherSuites,
		Extensions:      req.Extensions,
		SupportedCurves: req.SupportedCurves,
		Headers:         req.Headers,
		ClientIp:        req.ClientIP,
	})
	if err != nil {
		return nil, err
	}
	return &GRPCAnalyzeResponse{
		FingerprintHash:  resp.GetFingerprintHash(),
		Protocol:         resp.GetProtocol(),
		BrowserFamily:    resp.GetBrowserFamily(),
		BrowserVersion:   resp.GetBrowserVersion(),
		Confidence:       resp.GetConfidence(),
		RiskLevel:        resp.GetRiskLevel(),
		RiskScore:        resp.GetRiskScore(),
		JA3Hash:          resp.GetJa3Hash(),
		JA4Fingerprint:   resp.GetJa4Fingerprint(),
		DefenseHints:     resp.GetDefenseHints(),
		Cached:           resp.GetCached(),
		ProcessingTimeMs: resp.GetProcessingTimeMs(),
	}, nil
}

// Health calls the gRPC health endpoint.
func (c *GRPCClient) Health(ctx context.Context) (*gatewayv1.HealthResponse, error) {
	return c.client.Health(ctx, &gatewayv1.HealthRequest{})
}

// Close closes the connection.
func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

// Connection returns the underlying connection (for advanced usage).
func (c *GRPCClient) Connection() *grpc.ClientConn {
	return c.conn
}

// ==================== Hybrid Server ====================

// HybridServer is an HTTP + gRPC hybrid server.
type HybridServer struct {
	gateway  *Gateway
	httpPort int
	grpcPort int
}

// NewHybridServer creates a hybrid server.
func NewHybridServer(gateway *Gateway, httpPort, grpcPort int) *HybridServer {
	return &HybridServer{
		gateway:  gateway,
		httpPort: httpPort,
		grpcPort: grpcPort,
	}
}

// Start starts the hybrid server.
func (s *HybridServer) Start() error {
	// Start gRPC server (in background).
	grpcServer := NewGRPCServer(s.gateway)
	go func() {
		if err := grpcServer.Start(s.grpcPort); err != nil {
			slog.Error("gRPC server error", "error", err)
		}
	}()

	// Start HTTP server (blocking).
	return s.gateway.Start()
}

// Protocol represents the server protocol type.
type Protocol string

const (
	// ProtocolHTTP is the HTTP protocol.
	ProtocolHTTP Protocol = "http"
	// ProtocolGRPC is the gRPC protocol.
	ProtocolGRPC Protocol = "grpc"
)

// ServerInfo contains server information.
type ServerInfo struct {
	Protocol    Protocol
	Port        int
	Status      string
	Connections int
}

// GetServerInfo returns server information.
func (s *HybridServer) GetServerInfo() []ServerInfo {
	return []ServerInfo{
		{ProtocolHTTP, s.httpPort, "running", 0},
		{ProtocolGRPC, s.grpcPort, "running", 0},
	}
}
