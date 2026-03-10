// translated comment
// translated comment
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

// translated comment
type GRPCServer struct {
	server  *grpc.Server
	gateway *Gateway
}

// translated comment
func NewGRPCServer(gateway *Gateway) *GRPCServer {
	// translated comment
	s := grpc.NewServer(
		grpc.UnaryInterceptor(rateLimitInterceptor(gateway)),
	)

	return &GRPCServer{
		server:  s,
		gateway: gateway,
	}
}

// translated comment
func rateLimitInterceptor(gateway *Gateway) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// translated comment
		clientIP := extractClientIP(ctx)

		// translated comment
		if !gateway.limiter.Allow(clientIP) {
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

// translated comment
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
	// translated comment
	return strings.TrimSpace(addr)
}

// translated comment
func (s *GRPCServer) Start(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	fmt.Printf("gRPC server starting on port %d\n", port)
	return s.server.Serve(listener)
}

// translated comment
func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}

// ==================== gRPC messageconvert ====================

// GRPCAnalyzeRequest gRPC formatanalyzerequest
type GRPCAnalyzeRequest struct {
	ClientID        string
	TLSVersion      uint32
	CipherSuites    []uint32
	Extensions      []uint32
	SupportedCurves []uint32
	Headers         map[string]string
	ClientIP        string
}

// GRPCAnalyzeResponse gRPC formatanalyzeresponse
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

// translated comment
func ConvertToGatewayRequest(req *GRPCAnalyzeRequest) *AnalyzeRequest {
	// translated comment
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

	// translated comment
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

// translated comment
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

// translated comment

// translated comment
type GRPCClient struct {
	conn *grpc.ClientConn
}

// translated comment
func NewGRPCClient(address string) (*GRPCClient, error) {
	// translated comment
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &GRPCClient{conn: conn}, nil
}

// Close closeconnect
func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

// translated comment
func (c *GRPCClient) Connection() *grpc.ClientConn {
	return c.conn
}

// translated comment

// translated comment
type HybridServer struct {
	gateway  *Gateway
	httpPort int
	grpcPort int
}

// translated comment
func NewHybridServer(gateway *Gateway, httpPort, grpcPort int) *HybridServer {
	return &HybridServer{
		gateway:  gateway,
		httpPort: httpPort,
		grpcPort: grpcPort,
	}
}

// translated comment
func (s *HybridServer) Start() error {
	// translated comment
	grpcServer := NewGRPCServer(s.gateway)
	go func() {
		if err := grpcServer.Start(s.grpcPort); err != nil {
			fmt.Printf("gRPC server error: %v\n", err)
		}
	}()

	// translated comment
	return s.gateway.Start()
}

// translated comment
type Protocol string

const (
	// ProtocolHTTP HTTP protocol
	ProtocolHTTP Protocol = "http"
	// ProtocolGRPC gRPC protocol
	ProtocolGRPC Protocol = "grpc"
)

// translated comment
type ServerInfo struct {
	Protocol    Protocol
	Port        int
	Status      string
	Connections int
}

// translated comment
func (s *HybridServer) GetServerInfo() []ServerInfo {
	return []ServerInfo{
		{ProtocolHTTP, s.httpPort, "running", 0},
		{ProtocolGRPC, s.grpcPort, "running", 0},
	}
}
