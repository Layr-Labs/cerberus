package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	SubsystemRPCServer = "rpc_server"

	MetricRequestTotal           = "request_total"
	MetricRequestDurationSeconds = "request_duration_seconds"

	MethodLabelName = "method"
	CodeLabelName   = "code"
)

type Recorder interface {
	RecordRPCServerRequest(method string) func(code string)
}

type RPCServerMetrics struct {
	RPCServerRequestTotal           *prometheus.CounterVec
	RPCServerRequestDurationSeconds *prometheus.SummaryVec
}

func NewRPCServerMetrics(ns string, registry *prometheus.Registry) *RPCServerMetrics {
	m := &RPCServerMetrics{
		RPCServerRequestTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: SubsystemRPCServer,
			Name:      MetricRequestTotal,
			Help:      "Total number of RPC server requests with status codes",
		}, []string{MethodLabelName, CodeLabelName}),
		RPCServerRequestDurationSeconds: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace:  ns,
			Subsystem:  SubsystemRPCServer,
			Name:       MetricRequestDurationSeconds,
			Help:       "Duration of RPC server requests in seconds.",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.01, 0.99: 0.001},
		}, []string{MethodLabelName}),
	}
	registry.MustRegister(m.RPCServerRequestTotal)
	registry.MustRegister(m.RPCServerRequestDurationSeconds)
	return m
}

func (m *RPCServerMetrics) RecordRPCServerRequest(method string) func(code string) {
	timer := prometheus.NewTimer(m.RPCServerRequestDurationSeconds.WithLabelValues(method))
	return func(code string) {
		m.RPCServerRequestTotal.WithLabelValues(method, code).Inc()
		timer.ObserveDuration()
	}
}

type NoopRPCMetrics struct{}

func NewNoopRPCMetrics() *NoopRPCMetrics {
	return &NoopRPCMetrics{}
}

func (NoopRPCMetrics) RecordRPCServerRequest(method string) func(code string) {
	return func(code string) {}
}

var _ Recorder = (*NoopRPCMetrics)(nil)
