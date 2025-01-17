package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	v1 "github.com/Layr-Labs/cerberus-api/pkg/api/v1"

	"github.com/Layr-Labs/cerberus/internal/configuration"
	"github.com/Layr-Labs/cerberus/internal/services/admin"
	"github.com/Layr-Labs/cerberus/internal/services/kms"
	"github.com/Layr-Labs/cerberus/internal/services/signing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	_ "github.com/lib/pq"
)

type Server struct {
	resources *SharedResources
	servers   []*grpc.Server
	wg        sync.WaitGroup
}

// RegisterService represents a function type for service registration
type RegisterService func(*grpc.Server, *SharedResources)

// AddServiceOnPort adds a new gRPC service on the specified port
func (s *Server) AddServiceOnPort(
	serverCfg *GrpcServerConfig,
	registerService RegisterService,
) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", serverCfg.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %v", serverCfg.Port, err)
	}

	var opts []grpc.ServerOption
	if serverCfg.TLSCACert != "" && serverCfg.TLSServerKey != "" {
		creds, err := credentials.NewServerTLSFromFile(serverCfg.TLSCACert, serverCfg.TLSServerKey)
		if err != nil {
			log.Fatalf("Failed to load TLS certificates: %v", err)
		}
		s.resources.Logger.Info("Server-side TLS support enabled")

		opts = append(opts, grpc.Creds(creds))
	}

	opts = append(
		opts,
		grpc.ChainUnaryInterceptor(s.resources.GrpcMiddleware...),
	)

	grpcServer := grpc.NewServer(opts...)

	// Register the service with shared resources
	registerService(grpcServer, s.resources)

	// Add to servers list
	s.servers = append(s.servers, grpcServer)

	// Start the server in a goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Failed to serve on port %d: %v", serverCfg.Port, err)
		}
	}()

	return nil
}

// Start starts all registered services
func (s *Server) Start(ctx context.Context) error {
	// Create channel for shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either context cancellation or shutdown signal
	select {
	case <-ctx.Done():
		log.Println("Context cancelled, initiating shutdown...")
	case sig := <-sigChan:
		log.Printf("Received signal %v, initiating shutdown...", sig)
	}

	// Create context with timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for _, server := range s.servers {
		wg.Add(1)
		go func(srv *grpc.Server) {
			defer wg.Done()
			srv.GracefulStop()
		}(server)
	}

	// Wait for all servers to stop
	serverDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(serverDone)
	}()

	// Wait for server shutdown with timeout
	select {
	case <-serverDone:
		log.Println("All gRPC servers stopped successfully")
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout reached, forcing server stop")
		for _, server := range s.servers {
			server.Stop()
		}
	}

	return nil
}

func Start(config *configuration.Configuration, logger *slog.Logger) {
	server := NewServer(config, logger)

	kmsService := kms.NewService(
		config,
		server.resources.KeyStore,
		server.resources.KeyMetadataRepo,
		logger,
		server.resources.RpcMetrics,
	)
	signingService := signing.NewService(
		config,
		server.resources.KeyStore,
		logger,
		server.resources.RpcMetrics,
	)

	logger.Info(fmt.Sprintf("Starting gRPC server on port %d...", config.GrpcPort))
	err := server.AddServiceOnPort(&GrpcServerConfig{
		Port: config.GrpcPort,
	}, func(s *grpc.Server, resources *SharedResources) {
		v1.RegisterKeyManagerServer(s, kmsService)
		v1.RegisterSignerServer(s, signingService)

		// Register reflection service
		reflection.Register(s)
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to add service on port %d: %v", config.GrpcPort, err))
		os.Exit(1)
	}

	if config.EnableAdmin {
		adminService := admin.NewService(
			config,
			logger,
			server.resources.RpcMetrics,
			server.resources.KeyMetadataRepo,
		)

		logger.Info(fmt.Sprintf("Starting Admin server on port %d...", config.AdminPort))
		err = server.AddServiceOnPort(&GrpcServerConfig{
			Port: config.AdminPort,
		}, func(s *grpc.Server, resources *SharedResources) {
			v1.RegisterAdminServer(s, adminService)

			// Register reflection service
			reflection.Register(s)
		})

		if err != nil {
			logger.Error(
				fmt.Sprintf("Failed to admin service on port %d: %v", config.AdminPort, err),
			)
			os.Exit(1)
		}
	}

	// Start all services
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start all services
	if err := server.Start(ctx); err != nil {
		logger.Error(fmt.Sprintf("Failed to start servers: %v", err))
		os.Exit(1)
	}

}

// NewServer creates a new Server instance with shared resources
func NewServer(config *configuration.Configuration, logger *slog.Logger) *Server {
	return &Server{
		resources: NewSharedResources(config, logger),
		servers:   make([]*grpc.Server, 0),
	}
}
