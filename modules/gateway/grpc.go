// Package gateway 提供 gRPC API 支持
// 注意：此文件使用简化的 gRPC 实现，完整实现需要 protobuf 代码生成
package gateway

import (
	"context"
	"fmt"
	"net"

	"github.com/vistone/fingerprint/modules/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer gRPC 服务器包装器
type GRPCServer struct {
	server  *grpc.Server
	gateway *Gateway
}

// NewGRPCServer 创建新的 gRPC 服务器
func NewGRPCServer(gateway *Gateway) *GRPCServer {
	// 创建带拦截器的 gRPC 服务器
	s := grpc.NewServer(
		grpc.UnaryInterceptor(rateLimitInterceptor(gateway)),
	)
	
	return &GRPCServer{
		server:  s,
		gateway: gateway,
	}
}

// rateLimitInterceptor 限流拦截器
func rateLimitInterceptor(gateway *Gateway) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 从上下文中获取客户端 IP
		clientIP := extractClientIP(ctx)
		
		// 检查限流
		if !gateway.limiter.Allow(clientIP) {
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}
		
		return handler(ctx, req)
	}
}

// extractClientIP 从 gRPC 上下文中提取客户端 IP
func extractClientIP(ctx context.Context) string {
	// 简化实现，实际应该解析 peer 信息
	return "unknown"
}

// Start 启动 gRPC 服务器
func (s *GRPCServer) Start(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	
	fmt.Printf("gRPC server starting on port %d\n", port)
	return s.server.Serve(listener)
}

// Stop 停止 gRPC 服务器
func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}

// ==================== gRPC 消息转换 ====================

// GRPCAnalyzeRequest gRPC 格式的分析请求
type GRPCAnalyzeRequest struct {
	ClientID        string
	TLSVersion      uint32
	CipherSuites    []uint32
	Extensions      []uint32
	SupportedCurves []uint32
	Headers         map[string]string
	ClientIP        string
}

// GRPCAnalyzeResponse gRPC 格式的分析响应
type GRPCAnalyzeResponse struct {
	FingerprintHash string
	Protocol        string
	BrowserFamily   string
	BrowserVersion  string
	Confidence      float64
	RiskLevel       string
	RiskScore       float64
	JA3Hash         string
	JA4Fingerprint  string
	DefenseHints    []string
	Cached          bool
	ProcessingTimeMs int64
}

// ConvertToGatewayRequest 转换 gRPC 请求为网关请求
func ConvertToGatewayRequest(req *GRPCAnalyzeRequest) *AnalyzeRequest {
	// 转换 uint32 为 uint16
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
	
	// 构建 HTTP 头
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

// ConvertFromGatewayResponse 转换网关响应为 gRPC 响应
func ConvertFromGatewayResponse(resp *AnalyzeResponse) *GRPCAnalyzeResponse {
	return &GRPCAnalyzeResponse{
		FingerprintHash:  resp.FingerprintHash,
		Protocol:         string(resp.Classification.Protocol),
		BrowserFamily:    string(resp.Classification.Family),
		BrowserVersion:   resp.Classification.Version,
		Confidence:       resp.Classification.Confidence,
		RiskLevel:        resp.RiskAssessment.Level.String(),
		RiskScore:        resp.RiskAssessment.Score,
		JA3Hash:          resp.JA3.Hash,
		JA4Fingerprint:   resp.JA4.Fingerprint,
		DefenseHints:     resp.DefenseHints,
		Cached:           resp.Cached,
		ProcessingTimeMs: resp.ProcessingTimeMs,
	}
}

// ==================== gRPC 客户端 ====================

// GRPCClient gRPC 客户端
type GRPCClient struct {
	conn   *grpc.ClientConn
}

// NewGRPCClient 创建 gRPC 客户端
func NewGRPCClient(address string) (*GRPCClient, error) {
	// 使用不安全的连接（生产环境应该使用 TLS）
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	
	return &GRPCClient{conn: conn}, nil
}

// Close 关闭连接
func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

// Connection 获取底层连接（用于高级用法）
func (c *GRPCClient) Connection() *grpc.ClientConn {
	return c.conn
}

// ==================== 混合服务器 ====================

// HybridServer HTTP + gRPC 混合服务器
type HybridServer struct {
	gateway *Gateway
	httpPort int
	grpcPort int
}

// NewHybridServer 创建混合服务器
func NewHybridServer(gateway *Gateway, httpPort, grpcPort int) *HybridServer {
	return &HybridServer{
		gateway:  gateway,
		httpPort: httpPort,
		grpcPort: grpcPort,
	}
}

// Start 启动混合服务器
func (s *HybridServer) Start() error {
	// 启动 gRPC 服务器（在后台）
	grpcServer := NewGRPCServer(s.gateway)
	go func() {
		if err := grpcServer.Start(s.grpcPort); err != nil {
			fmt.Printf("gRPC server error: %v\n", err)
		}
	}()
	
	// 启动 HTTP 服务器（阻塞）
	return s.gateway.Start()
}

// Protocol 服务器协议类型
type Protocol string

const (
	// ProtocolHTTP HTTP 协议
	ProtocolHTTP Protocol = "http"
	// ProtocolGRPC gRPC 协议
	ProtocolGRPC Protocol = "grpc"
)

// ServerInfo 服务器信息
type ServerInfo struct {
	Protocol    Protocol
	Port        int
	Status      string
	Connections int
}

// GetServerInfo 获取服务器信息
func (s *HybridServer) GetServerInfo() []ServerInfo {
	return []ServerInfo{
		{ProtocolHTTP, s.httpPort, "running", 0},
		{ProtocolGRPC, s.grpcPort, "running", 0},
	}
}
