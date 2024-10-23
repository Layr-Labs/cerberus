package server

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/Layr-Labs/cerberus/internal/middleware"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	v1 "github.com/Layr-Labs/cerberus-api/pkg/api/v1"

	"github.com/Layr-Labs/cerberus/internal/configuration"
	"github.com/Layr-Labs/cerberus/internal/metrics"
	"github.com/Layr-Labs/cerberus/internal/services/kms"
	"github.com/Layr-Labs/cerberus/internal/services/signing"
	"github.com/Layr-Labs/cerberus/internal/store/filesystem"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

func Start(config *configuration.Configuration, logger *slog.Logger) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.GrpcPort))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to listen: %v", err))
		os.Exit(1)
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())
	rpcMetrics := metrics.NewRPCServerMetrics("cerberus", registry)

	go startMetricsServer(registry, config.MetricsPort, logger)

	keystore := filesystem.NewStore(config.KeystoreDir, logger)

	var opts []grpc.ServerOption
	if config.TLSCACert != "" && config.TLSServerKey != "" {
		creds, err := credentials.NewServerTLSFromFile(config.TLSCACert, config.TLSServerKey)
		if err != nil {
			log.Fatalf("Failed to load TLS certificates: %v", err)
		}
		logger.Info("Server-side TLS support enabled")

		opts = append(opts, grpc.Creds(creds))
	}

	// Register metrics middleware
	metricsMiddleware := middleware.NewMetricsMiddleware(registry, rpcMetrics)
	opts = append(opts, grpc.UnaryInterceptor(metricsMiddleware.UnaryServerInterceptor()))

	s := grpc.NewServer(opts...)
	kmsService := kms.NewService(config, keystore, logger, rpcMetrics)
	signingService := signing.NewService(config, keystore, logger, rpcMetrics)

	v1.RegisterKeyManagerServer(s, kmsService)
	v1.RegisterSignerServer(s, signingService)

	// Register the reflection service
	reflection.Register(s)

	logger.Info(fmt.Sprintf("Starting gRPC server on port %s...", config.GrpcPort))
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func startMetricsServer(r *prometheus.Registry, port string, logger *slog.Logger) {
	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	logger.Info(fmt.Sprintf("Starting metrics server on port %s...", port))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		logger.Error(fmt.Sprintf("Failed to start metrics server: %v", err))
	}
}
