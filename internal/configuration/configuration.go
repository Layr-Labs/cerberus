package configuration

import "fmt"

type StorageType string

type AWSAuthenticationMode string

const (
	FileSystemStorageType          StorageType = "filesystem"
	AWSSecretManagerStorageType    StorageType = "aws-secrets-manager"
	GoogleSecretManagerStorageType StorageType = "google-secrets-manager"

	EnvironmentAWSAuthenticationMode AWSAuthenticationMode = "environment"
	SpecifiedAWSAuthenticationMode   AWSAuthenticationMode = "specified"
)

type Configuration struct {
	StorageType StorageType

	// FileSystem storage parameters
	KeystoreDir string

	// AWS Secrets Manager storage parameters
	AWSRegion             string
	AWSProfile            string
	AWSAuthenticationMode AWSAuthenticationMode
	AWSAccessKeyID        string
	AWSSecretAccessKey    string

	// Google Secrets Manager storage parameters
	GCPProjectID string

	GrpcPort    int
	MetricsPort int
	AdminPort   int

	EnableAdmin bool

	TLSCACert    string
	TLSServerKey string

	// Postgres database parameters
	PostgresDatabaseURL string
}

func (s *Configuration) Validate() error {
	if s.StorageType == "" {
		return fmt.Errorf("storage type is required")
	}

	switch s.StorageType {
	case FileSystemStorageType:
		if s.KeystoreDir == "" {
			return fmt.Errorf("keystore directory is required")
		}
	case AWSSecretManagerStorageType:
		if s.AWSRegion == "" {
			return fmt.Errorf("AWS region is required")
		}

		if s.AWSAuthenticationMode == SpecifiedAWSAuthenticationMode {
			if s.AWSAccessKeyID == "" {
				return fmt.Errorf("AWS access key ID is required")
			}
			if s.AWSSecretAccessKey == "" {
				return fmt.Errorf("AWS secret access key is required")
			}
		}
	case GoogleSecretManagerStorageType:
		if s.GCPProjectID == "" {
			return fmt.Errorf("GCP project ID is required")
		}
	default:
		return fmt.Errorf("unsupported storage type: %s", s.StorageType)
	}

	if s.GrpcPort == 0 {
		return fmt.Errorf("gRPC port is required")
	}

	if s.MetricsPort == 0 {
		return fmt.Errorf("metrics port is required")
	}

	if s.TLSCACert != "" && s.TLSServerKey == "" {
		return fmt.Errorf("TLS server key is required when TLS CA certificate is provided")
	}

	if s.TLSServerKey != "" && s.TLSCACert == "" {
		return fmt.Errorf("TLS CA certificate is required when TLS server key is provided")
	}

	if s.PostgresDatabaseURL == "" {
		return fmt.Errorf("postgres database URL is required")
	}

	return nil
}
