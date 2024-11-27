package main

import (
	"fmt"
	"log/slog"
	"os"
	"sort"

	"github.com/Layr-Labs/cerberus/internal/configuration"
	"github.com/Layr-Labs/cerberus/internal/server"

	"github.com/urfave/cli/v2"
)

var (
	version = "development"

	keystoreDirFlag = &cli.StringFlag{
		Name:    "keystore-dir",
		Usage:   "Directory where the keystore files are stored",
		Value:   "./data/keystore",
		EnvVars: []string{"KEYSTORE_DIR"},
	}

	grpcPortFlag = &cli.StringFlag{
		Name:    "grpc-port",
		Usage:   "Port for the gRPC server",
		Value:   "50051",
		EnvVars: []string{"GRPC_PORT"},
	}

	metricsPortFlag = &cli.StringFlag{
		Name:    "metrics-port",
		Usage:   "Port for the metrics server",
		Value:   "9091",
		EnvVars: []string{"METRICS_PORT"},
	}

	logLevelFlag = &cli.StringFlag{
		Name:    "log-level",
		Usage:   "Log level - supported levels: debug, info, warn, error",
		Value:   "info",
		EnvVars: []string{"LOG_LEVEL"},
	}

	logFormatFlag = &cli.StringFlag{
		Name:    "log-format",
		Usage:   "Log format - supported formats: text, json",
		Value:   "text",
		EnvVars: []string{"LOG_FORMAT"},
	}

	// TLS flags to set up secure gRPC server, optional
	tlsCaCertFlag = &cli.StringFlag{
		Name:    "tls-ca-cert",
		Usage:   "TLS CA certificate",
		EnvVars: []string{"TLS_CA_CERT"},
	}

	tlsServerKeyFlag = &cli.StringFlag{
		Name:    "tls-server-key",
		Usage:   "TLS server key",
		EnvVars: []string{"TLS_SERVER_KEY"},
	}

	storageTypeFlag = &cli.StringFlag{
		Name:    "storage-type",
		Usage:   "Storage type - supported types: filesystem, aws-secret-manager",
		Value:   "filesystem",
		EnvVars: []string{"STORAGE_TYPE"},
	}

	awsRegionFlag = &cli.StringFlag{
		Name:    "aws-region",
		Usage:   "AWS region",
		Value:   "us-east-2",
		EnvVars: []string{"AWS_REGION"},
	}

	awsProfileFlag = &cli.StringFlag{
		Name:    "aws-profile",
		Usage:   "AWS profile",
		Value:   "default",
		EnvVars: []string{"AWS_PROFILE"},
	}

	awsAuthenticationModeFlag = &cli.StringFlag{
		Name:    "aws-authentication-mode",
		Usage:   "AWS authentication mode - supported modes: environment, specified",
		Value:   "environment",
		EnvVars: []string{"AWS_AUTHENTICATION_MODE"},
	}

	awsAccessKeyIDFlag = &cli.StringFlag{
		Name:    "aws-access-key-id",
		Usage:   "AWS access key ID",
		EnvVars: []string{"AWS_ACCESS_KEY_ID"},
	}

	awsSecretAccessKeyFlag = &cli.StringFlag{
		Name:    "aws-secret-access-key",
		Usage:   "AWS secret access key",
		EnvVars: []string{"AWS_SECRET_ACCESS_KEY"},
	}

	gcpProjectIDFlag = &cli.StringFlag{
		Name:    "gcp-project-id",
		Usage:   "Project ID for Google Cloud Platform",
		EnvVars: []string{"GCP_PROJECT_ID"},
	}
)

func main() {
	cli.AppHelpTemplate = fmt.Sprintf(`        
                   _                             
                  | |                            
  ___   ___  _ __ | |__    ___  _ __  _   _  ___ 
 / __| / _ \| '__|| '_ \  / _ \| '__|| | | |/ __|
| (__ |  __/| |   | |_) ||  __/| |   | |_| |\__ \
 \___| \___||_|   |_.__/  \___||_|    \__,_||___/

	
%s`, cli.AppHelpTemplate)
	app := cli.NewApp()

	app.Name = "cerberus"
	app.Usage = "Remote BLS Signer"
	app.Version = version
	app.Copyright = "(c) 2024 EigenLabs"

	app.Flags = []cli.Flag{
		keystoreDirFlag,
		grpcPortFlag,
		logFormatFlag,
		logLevelFlag,
		metricsPortFlag,
		tlsCaCertFlag,
		tlsServerKeyFlag,
		storageTypeFlag,
		awsRegionFlag,
		awsProfileFlag,
		awsAuthenticationModeFlag,
		awsAccessKeyIDFlag,
		awsSecretAccessKeyFlag,
		gcpProjectIDFlag,
	}
	sort.Sort(cli.FlagsByName(app.Flags))

	app.Action = start

	if err := app.Run(os.Args); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}

func start(c *cli.Context) error {
	keystoreDir := c.String(keystoreDirFlag.Name)
	grpcPort := c.String(grpcPortFlag.Name)
	metricsPort := c.String(metricsPortFlag.Name)
	logLevel := c.String(logLevelFlag.Name)
	logFormat := c.String(logFormatFlag.Name)
	tlsCaCert := c.String(tlsCaCertFlag.Name)
	tlsServerKey := c.String(tlsServerKeyFlag.Name)
	storageType := c.String(storageTypeFlag.Name)
	awsRegion := c.String(awsRegionFlag.Name)
	awsProfile := c.String(awsProfileFlag.Name)
	awsAuthenticationMode := c.String(awsAuthenticationModeFlag.Name)
	awsAccessKeyID := c.String(awsAccessKeyIDFlag.Name)
	awsSecretAccessKey := c.String(awsSecretAccessKeyFlag.Name)
	gcpProjectID := c.String(gcpProjectIDFlag.Name)

	cfg := &configuration.Configuration{
		KeystoreDir:           keystoreDir,
		GrpcPort:              grpcPort,
		MetricsPort:           metricsPort,
		TLSCACert:             tlsCaCert,
		TLSServerKey:          tlsServerKey,
		StorageType:           configuration.StorageType(storageType),
		AWSRegion:             awsRegion,
		AWSProfile:            awsProfile,
		AWSAuthenticationMode: configuration.AWSAuthenticationMode(awsAuthenticationMode),
		AWSAccessKeyID:        awsAccessKeyID,
		AWSSecretAccessKey:    awsSecretAccessKey,
		GCPProjectID:          gcpProjectID,
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %v", err)
	}

	sLogLevel := levelToLogLevel(logLevel)
	slogOptions := slog.HandlerOptions{AddSource: true, Level: sLogLevel}
	var logger *slog.Logger
	if logFormat == "json" {
		handler := slog.NewJSONHandler(os.Stdout, &slogOptions)
		logger = slog.New(handler)
	} else {
		handler := slog.NewTextHandler(os.Stdout, &slogOptions)
		logger = slog.New(handler)
	}
	logger.Info(fmt.Sprintf("Starting cerberus server version: %s", version))
	server.Start(cfg, logger)
	return nil
}

func levelToLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
