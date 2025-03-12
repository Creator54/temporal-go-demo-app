package config

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

// TracingExporterConfig holds configuration for tracing export
type TracingExporterConfig struct {
	ServiceName string
	Environment string
	Endpoint    string
	Headers     map[string]string
	TLSCreds    credentials.TransportCredentials
}

// NewTracingExporter creates and configures the tracing exporter
func NewTracingExporter(ctx context.Context, cfg *TracingExporterConfig, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	// Configure trace exporter
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithHeaders(cfg.Headers),
		otlptracegrpc.WithTLSCredentials(cfg.TLSCreds),
		otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
			Enabled:         true,
			InitialInterval: 1 * time.Second,
			MaxInterval:     5 * time.Second,
			MaxElapsedTime:  30 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Configure tracer provider with more frequent batching
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
			sdktrace.WithMaxQueueSize(2048),
		),
	)

	return tracerProvider, nil
}

// GetSpanAttributes returns common span attributes for the application
func GetSpanAttributes(workflowType, workflowID, taskQueue string) map[string]string {
	return map[string]string{
		"workflow.type":       workflowType,
		"workflow.id":         workflowID,
		"workflow.task_queue": taskQueue,
		"service.name":        "temporal-hello-world",
	}
}
