package server

type GrpcServerConfig struct {
	Port         int
	EnableTLS    bool
	TLSCACert    string
	TLSServerKey string
}
