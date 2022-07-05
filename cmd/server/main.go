package main

import (
	"context"
	"log"
	"os"

	"github.com/kroksys/user-service-example/pkg/db"
	"github.com/kroksys/user-service-example/pkg/service"
)

var (
	// Database connection string used to connect to database
	connectionString = "user:userpw@tcp(localhost:3306)/users?parseTime=true"

	// gRPC server address
	grpcAddr = "localhost:9000"

	// HTTP server address
	apiAddr = "localhost:9001"
)

func main() {

	// Read environment variables
	if os.Getenv("USERSERVICE_CONNECTION_STRING") != "" {
		connectionString = os.Getenv("USERSERVICE_CONNECTION_STRING")
	}
	if os.Getenv("USERSERVICE_GRPC_ADDR") != "" {
		grpcAddr = os.Getenv("USERSERVICE_GRPC_ADDR")
	}
	if os.Getenv("USERSERVICE_HTTP_ADDR") != "" {
		apiAddr = os.Getenv("USERSERVICE_HTTP_ADDR")
	}

	// Connect to database
	err := db.Connect(connectionString)
	if err != nil {
		log.Printf("connectionString: %s\n", connectionString)
		log.Fatalf("Error connecting to database: %s\n", err.Error())
	}

	// Migrate user model to database
	err = db.Migrate()
	if err != nil {
		log.Fatalf("Error migrating user to database: %s\n", err.Error())
	}

	// Context for both servers to be stopped
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start grpc server
	log.Printf("Starting gRPC server with addr: %s\n", grpcAddr)
	grpcServer, err := service.StartGrpcServer(ctx, grpcAddr)
	if err != nil {
		log.Fatalf("Error starting GRPC server: %v\n", err)
	}
	defer grpcServer.GracefulStop()

	// Start HTTP server
	log.Printf("Starting HTTP server with addr: %s\n", apiAddr)
	httpServer, err := service.StartHTTPServer(ctx, apiAddr, grpcAddr)
	if err != nil {
		log.Fatalf("Error starting HTTP server: %v\n", err)
	}
	defer httpServer.Shutdown(context.Background())

	// Wait for conext to be canceled
	<-ctx.Done()

	log.Println("Server stopped")
}
