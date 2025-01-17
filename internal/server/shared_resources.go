package server

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Layr-Labs/cerberus/internal/configuration"
	"github.com/Layr-Labs/cerberus/internal/database"
	"github.com/Layr-Labs/cerberus/internal/database/repository"
	"github.com/Layr-Labs/cerberus/internal/database/repository/postgres"
	"github.com/Layr-Labs/cerberus/internal/metrics"
	"github.com/Layr-Labs/cerberus/internal/middleware"
	"github.com/Layr-Labs/cerberus/internal/store"
	"github.com/Layr-Labs/cerberus/internal/store/awssecretmanager"
	"github.com/Layr-Labs/cerberus/internal/store/filesystem"
	"github.com/Layr-Labs/cerberus/internal/store/googlesm"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"google.golang.org/grpc"

	_ "github.com/lib/pq"
)

type SharedResources struct {
	KeyMetadataRepo repository.KeyMetadataRepository
	KeyStore        store.Store
	GrpcMiddleware  []grpc.UnaryServerInterceptor
	RpcMetrics      *metrics.RPCServerMetrics
	Logger          *slog.Logger

	// Private fields
	db *sql.DB
}

func NewSharedResources(
	config *configuration.Configuration,
	logger *slog.Logger,
) *SharedResources {

	// Initialize store
	keystore, err := initializeStore(config, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to initialize store: %v", err))
		os.Exit(1)
	}

	// Initialize database
	db, err := initializeDatabase(config, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to initialize database: %v", err))
		os.Exit(1)
	}

	// Initialize key metadata repository
	keyMetadataRepo := postgres.NewKeyMetadataRepository(db)

	// Initialize prometheus registry
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())
	rpcMetrics := metrics.NewRPCServerMetrics("cerberus", registry)

	// Start metrics server
	go startMetricsServer(registry, config.MetricsPort, logger)

	// Initialize grpc middleware
	grpcMiddleware := initializeGrpcMiddleware(registry, rpcMetrics, keyMetadataRepo)

	return &SharedResources{
		db:              db,
		KeyMetadataRepo: keyMetadataRepo,
		KeyStore:        keystore,
		GrpcMiddleware:  grpcMiddleware,
		RpcMetrics:      rpcMetrics,
		Logger:          logger,
	}
}

func initializeDatabase(
	config *configuration.Configuration,
	logger *slog.Logger,
) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.PostgresDatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := database.MigrateDB(config.PostgresDatabaseURL, logger); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

func initializeStore(
	config *configuration.Configuration,
	logger *slog.Logger,
) (store.Store, error) {
	var keystore store.Store
	var err error
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
				return nil, fmt.Errorf("failed to create AWS Secret Manager store: %w", err)
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
				return nil, fmt.Errorf("failed to create AWS Secret Manager store: %w", err)
			}
			logger.Info("Using specified credentials for AWS Secret Manager")
		}
	case configuration.GoogleSecretManagerStorageType:
		keystore, err = googlesm.NewKeystore(config.GCPProjectID, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create Google Secret Manager store: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.StorageType)
	}

	return keystore, nil
}

func initializeGrpcMiddleware(
	registry *prometheus.Registry,
	rpcMetrics *metrics.RPCServerMetrics,
	keyMetadataRepo repository.KeyMetadataRepository,
) []grpc.UnaryServerInterceptor {
	metricsMiddleware := middleware.NewMetricsMiddleware(registry, rpcMetrics)
	authInterceptor := middleware.AuthInterceptor("signer.v1.Signer", keyMetadataRepo)
	return []grpc.UnaryServerInterceptor{
		metricsMiddleware.UnaryServerInterceptor(),
		authInterceptor,
	}
}

func startMetricsServer(r *prometheus.Registry, port int, logger *slog.Logger) {
	http.Handle("/metrics", promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	logger.Info(fmt.Sprintf("Starting metrics server on port %d...", port))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		logger.Error(fmt.Sprintf("Failed to start metrics server: %v", err))
	}
}
