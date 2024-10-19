package configuration

type Configuration struct {
	KeystoreDir string

	GrpcPort    string
	MetricsPort string

	TLSCACert    string
	TLSServerKey string
}
