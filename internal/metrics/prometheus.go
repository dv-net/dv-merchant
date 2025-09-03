package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricsNamespace = "backend"
)

const (
	ProcessingRequestDurationMetricName = "processing_request_duration"
	ProcessingHTTPStatusCodeMetricName  = "processing_response_code_http_total"
	ProcessingRPCStatusCodeMetricName   = "processing_response_code_rpc_total"
)

type ProcessingStatusCodeLabel string

const (
	ProcessingStatusCodeLabelName ProcessingStatusCodeLabel = "code"
)

const (
	ProcessingRequestDurationLabelService = "service"
	ProcessingRequestDurationLabelMethod  = "method"
)

type PrometheusMetrics struct {
	processingHTTPResponseStatusCode *prometheus.CounterVec
	processingRPCResponseStatusCode  *prometheus.CounterVec
	processingRequestDurationSeconds *prometheus.HistogramVec
}

func New() (*PrometheusMetrics, error) {
	processingHTTPResponseStatusCode := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metricsNamespace,
		Name:        ProcessingHTTPStatusCodeMetricName,
		Help:        "processing response status codes http",
		ConstLabels: nil,
	}, []string{string(ProcessingStatusCodeLabelName)})
	if err := prometheus.DefaultRegisterer.Register(processingHTTPResponseStatusCode); err != nil {
		return nil, err
	}

	processingRPCResponseStatusCode := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   metricsNamespace,
		Name:        ProcessingRPCStatusCodeMetricName,
		Help:        "processing response status codes rpc",
		ConstLabels: nil,
	}, []string{string(ProcessingStatusCodeLabelName)})

	if err := prometheus.DefaultRegisterer.Register(processingRPCResponseStatusCode); err != nil {
		return nil, err
	}

	processingRequestDurationSeconds := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   metricsNamespace,
		Name:        ProcessingRequestDurationMetricName,
		Help:        "request duration to processing in seconds",
		ConstLabels: nil,
		Buckets:     []float64{0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1.0, 2.5, 5.0, 7.5, 10.0, 20.0, 30.0, 50.0, 100.0},
	}, []string{ProcessingRequestDurationLabelService, ProcessingRequestDurationLabelMethod})
	if err := prometheus.DefaultRegisterer.Register(processingRequestDurationSeconds); err != nil {
		return nil, err
	}

	return &PrometheusMetrics{
		processingRPCResponseStatusCode:  processingRPCResponseStatusCode,
		processingHTTPResponseStatusCode: processingHTTPResponseStatusCode,
		processingRequestDurationSeconds: processingRequestDurationSeconds,
	}, nil
}

func (m *PrometheusMetrics) RegisterMetrics(collectors ...prometheus.Collector) {
	prometheus.DefaultRegisterer.MustRegister(collectors...)
}

func (m *PrometheusMetrics) ProcessingRPCResponseCode(status string) {
	m.processingRPCResponseStatusCode.WithLabelValues(status).Inc()
}

func (m *PrometheusMetrics) ProcessingHTTPResponseCode(status string) {
	m.processingHTTPResponseStatusCode.WithLabelValues(status).Inc()
}

func (m *PrometheusMetrics) ProcessingRequestDuration(service, methodName string, duration float64) {
	m.processingRequestDurationSeconds.WithLabelValues(service, methodName).Observe(duration)
}

func (m *PrometheusMetrics) Namespace() string {
	return metricsNamespace
}
