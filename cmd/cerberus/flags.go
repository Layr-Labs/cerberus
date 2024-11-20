package main

import "github.com/urfave/cli/v2"

var (
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
)
