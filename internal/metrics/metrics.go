package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	SuccessLabel = "success"
	FailureLabel = "failure"

	SubsystemRPCServer = "rpc_server"

	MetricRequestTotal           = "request_total"
	MetricRequestDurationSeconds = "request_duration_seconds"
	MetricResponseTotal          = "response_total"

	MethodLabelName = "method"
	StatusLabelName = "status"
)

type Recorder interface {
	RecordRPCServerRequest(method string) func()
	RecordRPCServerResponse(method string, status string)
}

type RPCServerMetrics struct {
	RPCServerRequestTotal           *prometheus.CounterVec
	RPCServerRequestDurationSeconds *prometheus.SummaryVec
	RPCServerResponseTotal          *prometheus.CounterVec
}

func NewRPCServerMetrics(ns string, registry *prometheus.Registry) *RPCServerMetrics {
	m := &RPCServerMetrics{
		RPCServerRequestTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: SubsystemRPCServer,
			Name:      MetricRequestTotal,
			Help:      "Total number of RPC server requests.",
		}, []string{MethodLabelName}),
		RPCServerRequestDurationSeconds: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace:  ns,
			Subsystem:  SubsystemRPCServer,
			Name:       MetricRequestDurationSeconds,
			Help:       "Duration of RPC server requests in seconds.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.01, 0.99: 0.001},
		}, []string{MethodLabelName}),
		RPCServerResponseTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: SubsystemRPCServer,
			Name:      MetricResponseTotal,
			Help:      "Total number of RPC server responses.",
		}, []string{MethodLabelName, StatusLabelName}),
	}
	registry.MustRegister(m.RPCServerRequestTotal)
	registry.MustRegister(m.RPCServerRequestDurationSeconds)
	registry.MustRegister(m.RPCServerResponseTotal)
	return m
}

func (m *RPCServerMetrics) RecordRPCServerRequest(method string) func() {
	m.RPCServerRequestTotal.WithLabelValues(method).Inc()
	timer := prometheus.NewTimer(m.RPCServerRequestDurationSeconds.WithLabelValues(method))
	return func() {
		timer.ObserveDuration()
	}
}

func (m *RPCServerMetrics) RecordRPCServerResponse(method string, status string) {
	m.RPCServerResponseTotal.WithLabelValues(method, status).Inc()
}

type NoopRPCMetrics struct{}

func NewNoopRPCMetrics() *NoopRPCMetrics {
	return &NoopRPCMetrics{}
}

func (NoopRPCMetrics) RecordRPCServerRequest(method string) func() {
	return func() {}
}

func (NoopRPCMetrics) RecordRPCServerResponse(method string, status string) {}

var _ Recorder = (*NoopRPCMetrics)(nil)
