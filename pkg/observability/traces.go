package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

type TracesConfig struct {
	Endpoint       string `yaml:"endpoint" env-required:"true" env:"TRACES_ENDPOINT"`
	ServiceName    string `yaml:"service_name" env-required:"true" env:"TRACES_SERVICE_NAME"`
	ServiceVersion string `yaml:"service_version" env-required:"true" env:"TRACES_SERVICE_VERSION"`
}

func NewTracerProvider(ctx context.Context, cfg TracesConfig) (*trace.TracerProvider, error) {
	const op = "observability.NewTracerProvider"

	res, err := resource.New(ctx, resource.WithAttributes(
		attribute.String("service.name", cfg.ServiceName),
		attribute.String("service.version", cfg.ServiceVersion),
	))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	exporter, err := otlptracegrpc.New(
		ctx,
		// otlptracegrpc.WithEndpointURL(cfg.Endpoint),
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
	)

	return tp, nil
}
