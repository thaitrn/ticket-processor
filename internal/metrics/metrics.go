package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	MessagesProduced   prometheus.Counter
	MessagesConsumed   prometheus.Counter
	ProcessingDuration prometheus.Histogram
	ErrorsCount        prometheus.Counter
	ConsumerLag        prometheus.Gauge
}

func NewMetrics(namespace string) *Metrics {
	return &Metrics{
		MessagesProduced: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "messages_produced_total",
			Help:      "The total number of produced messages",
		}),
		MessagesConsumed: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "messages_consumed_total",
			Help:      "The total number of consumed messages",
		}),
		ProcessingDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "message_processing_duration_seconds",
			Help:      "Message processing duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		}),
		ErrorsCount: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "errors_total",
			Help:      "The total number of errors",
		}),
		ConsumerLag: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "consumer_lag",
			Help:      "The current consumer lag",
		}),
	}
}
