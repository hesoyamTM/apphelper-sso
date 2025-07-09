package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsConfig struct {
	Port int `yaml:"port" env-default:"6004" env:"METRICS_PORT"`
}

type Metrics struct {
	Config MetricsConfig

	RequestCounter  *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	ErrorCounter    *prometheus.CounterVec
}

func NewMetrics(ctx context.Context, config MetricsConfig) (*Metrics, error) {
	const op = "metrics.NewMetrics"

	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "The total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)
	if err := prometheus.Register(requestCounter); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "The HTTP request latencies in seconds",
			Buckets: prometheus.ExponentialBuckets(0.0001, 2, 20),
		},
		[]string{"method", "path", "status_code"},
	)
	if err := prometheus.Register(requestDuration); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	errorCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "The total number of HTTP errors",
		},
		[]string{"method", "path", "status_code"},
	)
	if err := prometheus.Register(errorCounter); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := prometheus.Register(collectors.NewBuildInfoCollector()); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Metrics{
		Config: config,

		RequestCounter:  requestCounter,
		RequestDuration: requestDuration,
		ErrorCounter:    errorCounter,
	}, nil
}

func (m *Metrics) Start(ctx context.Context) error {
	http.Handle("/metrics", promhttp.Handler())
	logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("metrics server is running on port %d", m.Config.Port))
	return http.ListenAndServe(fmt.Sprintf(":%d", m.Config.Port), nil)
}
