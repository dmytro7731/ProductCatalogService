package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/product-catalog-service/internal/services"
	pb "github.com/product-catalog-service/proto/product/v1"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}

func run() error {
	// Load configuration from environment
	config := loadConfig()

	// Initialize context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Spanner client
	spannerClient, err := createSpannerClient(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create spanner client: %w", err)
	}
	defer spannerClient.Close()

	// Initialize dependency injection container
	container := services.NewContainer(spannerClient)

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	pb.RegisterProductServiceServer(grpcServer, container.ProductHandler)

	// Enable reflection for development
	reflection.Register(grpcServer)

	// Start gRPC server
	listener, err := net.Listen("tcp", config.GRPCAddress)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop()
		cancel()
	}()

	log.Printf("Starting gRPC server on %s", config.GRPCAddress)
	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Config holds the application configuration.
type Config struct {
	GRPCAddress     string
	SpannerProject  string
	SpannerInstance string
	SpannerDatabase string
	UseEmulator     bool
}

func loadConfig() Config {
	config := Config{
		GRPCAddress:     getEnv("GRPC_ADDRESS", ":50051"),
		SpannerProject:  getEnv("SPANNER_PROJECT", "test-project"),
		SpannerInstance: getEnv("SPANNER_INSTANCE", "test-instance"),
		SpannerDatabase: getEnv("SPANNER_DATABASE", "product-catalog"),
		UseEmulator:     getEnv("SPANNER_EMULATOR_HOST", "") != "",
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func createSpannerClient(ctx context.Context, config Config) (*spanner.Client, error) {
	database := fmt.Sprintf(
		"projects/%s/instances/%s/databases/%s",
		config.SpannerProject,
		config.SpannerInstance,
		config.SpannerDatabase,
	)

	client, err := spanner.NewClient(ctx, database)
	if err != nil {
		return nil, err
	}

	return client, nil
}
