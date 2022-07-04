package server

import (
	"context"
	"testing"

	"github.com/kroksys/user-service-example/pkg/pb/v1"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	grpcAddr = "localhost:1988"
	apiAddr  = "localhost:1989"
)

func TestServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	grpcServer, err := StartGrpcServer(ctx, grpcAddr)
	if err != nil {
		t.Fatalf("Error starting GRPC server: %v\n", err)
	}
	defer grpcServer.GracefulStop()

	httpServer, err := StartHTTPServer(ctx, apiAddr, grpcAddr)
	if err != nil {
		t.Fatalf("Error starting HTTP server: %v\n", err)
	}
	defer httpServer.Shutdown(context.Background())

	// Test healthcheck
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("TestServer could not create grpc client: %v", err)
	}
	defer conn.Close()
	c := healthpb.NewHealthClient(conn)

	resp, err := c.Check(context.Background(), &healthpb.HealthCheckRequest{
		Service: pb.UserService_ServiceDesc.ServiceName,
	})
	if err != nil {
		t.Fatalf("TestServer: healthchek service responded with error: %v", err)
	}
	if resp.Status != healthpb.HealthCheckResponse_SERVING {
		t.Errorf("TestServer: healthchek service responded with a status other than SERVING. Got status: %s", resp.Status.String())
	}
}
