package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	BackgroundWorkerRunningMetricName  = "background_worker_running"
	BackgroundWorkerRestartsMetricName = "background_worker_restarts_total"

	BackgroundWorkerLabelName = "worker"
)

var (
	backgroundWorkerRunning = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Name:      BackgroundWorkerRunningMetricName,
		Help:      "1 while the background worker loop is executing, 0 otherwise",
	}, []string{BackgroundWorkerLabelName})

	backgroundWorkerRestarts = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Name:      BackgroundWorkerRestartsMetricName,
		Help:      "Number of times a background worker was restarted after an unexpected exit or panic",
	}, []string{BackgroundWorkerLabelName})
)

func SetBackgroundWorkerRunning(worker string, running bool) {
	value := 0.0
	if running {
		value = 1
	}
	backgroundWorkerRunning.WithLabelValues(worker).Set(value)
}

func IncBackgroundWorkerRestarts(worker string) {
	backgroundWorkerRestarts.WithLabelValues(worker).Inc()
}
