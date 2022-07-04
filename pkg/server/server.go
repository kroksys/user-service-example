package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/kroksys/user-service-example/pkg/pb/v1"
	"github.com/kroksys/user-service-example/pkg/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// Starts grpc server listening on provided addr.
// Registers healthcheck and user service.
// Returns grpc.Server that should be used to defer server.GracefulStop().
func StartGrpcServer(ctx context.Context, addr string) (*grpc.Server, error) {
	// Try to open TCP port for grpc server
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()

	// Register user service
	pb.RegisterUserServiceServer(server, service.UserService{})

	// Register healthckech service and setting status to serving
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(pb.UserService_ServiceDesc.ServiceName, healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(server, healthServer)

	go func() {
		if err := server.Serve(lis); err != nil {
			log.Printf("grpc server stopped with error: %v\n", err)
		}
	}()

	return server, nil
}

// Starts HTTP gin API server that proxies the requests to grpc server.
// returns http.Server and error. The server repose souhld be used to defer server.Shutdown().
func StartHTTPServer(ctx context.Context, addr, grpcAddr string) (*http.Server, error) {
	conn, err := grpc.DialContext(
		ctx,
		grpcAddr,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial a grpc server: %v", err)
	}

	// Grpc to Rest API
	mux := runtime.NewServeMux()
	err = pb.RegisterUserServiceHandler(ctx, mux, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %v", err)
	}

	// Gin HTTP server
	gin.SetMode(gin.ReleaseMode)
	server := gin.New()
	server.Use(gin.Logger())
	server.Group("v1/*{grpc_gateway}").Any("", gin.WrapH(mux))

	// Start the server
	srv := &http.Server{
		Addr:    addr,
		Handler: server,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("http server stopped with error: %v\n", err)
		}
	}()

	return srv, nil
}
