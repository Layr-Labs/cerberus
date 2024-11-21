package configuration

import "fmt"

type StorageType string

type AWSAuthenticationMode string

const (
	FileSystemStorageType       StorageType = "filesystem"
	AWSSecretManagerStorageType StorageType = "aws-secrets-manager"

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

	GrpcPort    string
	MetricsPort string

	TLSCACert    string
	TLSServerKey string
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
	default:
		return fmt.Errorf("unsupported storage type: %s", s.StorageType)
	}

	if s.GrpcPort == "" {
		return fmt.Errorf("gRPC port is required")
	}

	if s.MetricsPort == "" {
		return fmt.Errorf("metrics port is required")
	}

	if s.TLSCACert != "" && s.TLSServerKey == "" {
		return fmt.Errorf("TLS server key is required when TLS CA certificate is provided")
	}

	if s.TLSServerKey != "" && s.TLSCACert == "" {
		return fmt.Errorf("TLS CA certificate is required when TLS server key is provided")
	}

	return nil
}
