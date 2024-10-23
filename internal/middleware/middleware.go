package middleware

import (
	"context"

	"github.com/Layr-Labs/cerberus/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type MetricsMiddleware struct {
	registry *prometheus.Registry
	recorder metrics.Recorder
}

func NewMetricsMiddleware(
	registry *prometheus.Registry,
	recorder metrics.Recorder,
) *MetricsMiddleware {
	return &MetricsMiddleware{
		registry: registry,
		recorder: recorder,
	}
}

// UnaryServerInterceptor returns a new unary server interceptor for metrics
func (m *MetricsMiddleware) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		recordDuration := m.recorder.RecordRPCServerRequest(info.FullMethod)

		// Handle request
		resp, err := handler(ctx, req)

		// Get status code
		code := status.Code(err)

		// Record response
		recordDuration(code.String())

		return resp, err
	}
}
