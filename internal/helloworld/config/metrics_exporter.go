package config

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc/credentials"
)

// MetricsExporterConfig holds configuration for metrics export
type MetricsExporterConfig struct {
	ServiceName string
	Environment string
	Endpoint    string
	Headers     map[string]string
	TLSCreds    credentials.TransportCredentials
}

// NewMetricsExporter creates and configures the metrics exporter
func NewMetricsExporter(ctx context.Context, cfg *MetricsExporterConfig, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	// Configure metrics exporter
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.Endpoint),
		otlpmetricgrpc.WithHeaders(cfg.Headers),
		otlpmetricgrpc.WithTLSCredentials(cfg.TLSCreds),
		otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetryConfig{
			Enabled:         true,
			InitialInterval: 1 * time.Second,
			MaxInterval:     5 * time.Second,
			MaxElapsedTime:  30 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}

	// Configure periodic reader with more frequent reporting
	reader := sdkmetric.NewPeriodicReader(
		exporter,
		sdkmetric.WithInterval(10*time.Second),
	)

	// Create meter provider
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)

	// Register common metrics
	meter := mp.Meter("temporal-hello-world")
	if err := RegisterCommonMetrics(meter); err != nil {
		return nil, fmt.Errorf("failed to register common metrics: %w", err)
	}

	return mp, nil
}

// RegisterCommonMetrics registers metrics that are common across the application
func RegisterCommonMetrics(meter metric.Meter, attrs ...attribute.KeyValue) error {
	// Example metrics (add more as needed)
	_, err := meter.Int64Counter(
		"temporal_workflow_completed",
		metric.WithDescription("Number of completed workflows"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create workflow completed counter: %w", err)
	}

	_, err = meter.Int64Counter(
		"temporal_workflow_started",
		metric.WithDescription("Number of started workflows"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create workflow started counter: %w", err)
	}

	_, err = meter.Int64UpDownCounter(
		"temporal_worker_task_slots_available",
		metric.WithDescription("Number of available task slots"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return fmt.Errorf("failed to create task slots counter: %w", err)
	}

	return nil
}
