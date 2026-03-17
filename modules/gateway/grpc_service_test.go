package gateway

import (
	"context"
	"net"
	"testing"
	"time"

	gatewayv1 "github.com/vistone/fingerprint/modules/gateway/proto/fingerprint/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestGRPCService_AnalyzeAndHealth(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	gw := NewGateway(nil)
	defer gw.Close()

	srv := NewGRPCServer(gw)
	go func() {
		_ = srv.server.Serve(lis)
	}()
	defer srv.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	client := gatewayv1.NewFingerprintGatewayClient(conn)

	health, err := client.Health(ctx, &gatewayv1.HealthRequest{})
	if err != nil {
		t.Fatalf("health failed: %v", err)
	}
	if health.GetStatus() != "SERVING" {
		t.Fatalf("unexpected health status: %q", health.GetStatus())
	}

	resp, err := client.Analyze(ctx, &gatewayv1.AnalyzeRequest{
		TlsVersion:   0x0303,
		CipherSuites: []uint32{0x1301, 0x1302},
		Extensions:   []uint32{0, 23, 65281},
		Headers: map[string]string{
			"user-agent": "Mozilla/5.0",
			"accept":     "*/*",
		},
	})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	if resp.GetFingerprintHash() == "" {
		t.Fatal("empty fingerprint hash")
	}
}
