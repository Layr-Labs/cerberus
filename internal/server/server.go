package server

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	v1 "github.com/Layr-Labs/cerberus-api/pkg/api/v1"

	"github.com/Layr-Labs/cerberus/internal/configuration"
	"github.com/Layr-Labs/cerberus/internal/database"
	"github.com/Layr-Labs/cerberus/internal/database/repository/postgres"
	"github.com/Layr-Labs/cerberus/internal/metrics"
	"github.com/Layr-Labs/cerberus/internal/middleware"
	"github.com/Layr-Labs/cerberus/internal/services/kms"
	"github.com/Layr-Labs/cerberus/internal/services/signing"
	"github.com/Layr-Labs/cerberus/internal/store"
	"github.com/Layr-Labs/cerberus/internal/store/awssecretmanager"
	"github.com/Layr-Labs/cerberus/internal/store/filesystem"
	"github.com/Layr-Labs/cerberus/internal/store/googlesm"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	_ "github.com/lib/pq"
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

	var keystore store.Store
	switch config.StorageType {
	case configuration.FileSystemStorageType:
		keystore = filesystem.NewStore(config.KeystoreDir, logger)
	case configuration.AWSSecretManagerStorageType:
		switch config.AWSAuthenticationMode {
		case configuration.EnvironmentAWSAuthenticationMode:
			keystore, err = awssecretmanager.NewStoreWithEnv(
				config.AWSRegion,
				config.AWSProfile,
				logger,
			)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to create AWS Secret Manager store: %v", err))
				os.Exit(1)
			}
			logger.Info("Using environment credentials for AWS Secret Manager")
		case configuration.SpecifiedAWSAuthenticationMode:
			keystore, err = awssecretmanager.NewStoreWithSpecifiedCredentials(
				config.AWSRegion,
				config.AWSAccessKeyID,
				config.AWSSecretAccessKey,
				logger,
			)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to create AWS Secret Manager store: %v", err))
				os.Exit(1)
			}
			logger.Info("Using specified credentials for AWS Secret Manager")
		}
	case configuration.GoogleSecretManagerStorageType:
		keystore, err = googlesm.NewKeystore(config.GCPProjectID, logger)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to create Google Secret Manager store: %v", err))
			os.Exit(1)
		}
	default:
		logger.Error(fmt.Sprintf("Unsupported storage type: %s", config.StorageType))
		os.Exit(1)
	}

	var opts []grpc.ServerOption
	if config.TLSCACert != "" && config.TLSServerKey != "" {
		creds, err := credentials.NewServerTLSFromFile(config.TLSCACert, config.TLSServerKey)
		if err != nil {
			log.Fatalf("Failed to load TLS certificates: %v", err)
		}
		logger.Info("Server-side TLS support enabled")

		opts = append(opts, grpc.Creds(creds))
	}

	// Initialize database
	db, err := sql.Open("postgres", config.PostgresDatabaseURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to database: %v", err))
		os.Exit(1)
	}
	defer db.Close()

	if err := database.MigrateDB(config.PostgresDatabaseURL, logger); err != nil {
		logger.Error(fmt.Sprintf("Failed to migrate database: %v", err))
		os.Exit(1)
	}

	keyMetadataRepo := postgres.NewKeyMetadataRepository(db)

	// Register metrics middleware
	metricsMiddleware := middleware.NewMetricsMiddleware(registry, rpcMetrics)
	authInterceptor := middleware.AuthInterceptor("signer.v1.Signer", keyMetadataRepo)
	opts = append(
		opts,
		grpc.ChainUnaryInterceptor(metricsMiddleware.UnaryServerInterceptor(), authInterceptor),
	)

	s := grpc.NewServer(opts...)
	kmsService := kms.NewService(config, keystore, keyMetadataRepo, logger, rpcMetrics)
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
