package observability

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
)

type MetricsConfig struct {
	Port int `yaml:"port" env-default:"6004" env:"METRICS_PORT"`
}

func NewMeterProvider(ctx context.Context) (*metric.MeterProvider, error) {
	const op = "observability.NewMeterProvider"

	exporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(exporter),
	)

	return meterProvider, nil
}

func serveMetrics(ctx context.Context, cfg MetricsConfig) {
	const op = "observability.serveMetrics"
	log := logger.GetLoggerFromCtx(ctx)

	http.Handle("/metrics", promhttp.Handler())
	log.Info(ctx, fmt.Sprintf("metrics server is running on port %d", cfg.Port))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil); err != nil {
		log.Error(ctx, fmt.Sprintf("%s: %w", op, err))
	}
}
