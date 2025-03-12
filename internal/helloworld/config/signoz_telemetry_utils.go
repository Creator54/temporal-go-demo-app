package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
)

// SignozTelemetryUtils coordinates all telemetry components
type SignozTelemetryUtils struct {
	otelConfig *OpenTelemetryConfig
}

// NewSignozTelemetryUtils creates a new telemetry configuration
func NewSignozTelemetryUtils() *SignozTelemetryUtils {
	return &SignozTelemetryUtils{
		otelConfig: NewOpenTelemetryConfig(),
	}
}

// InitProvider initializes all telemetry providers
func (c *SignozTelemetryUtils) InitProvider(ctx context.Context) (func(), error) {
	// Create resource
	res, err := c.otelConfig.CreateResource(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Get TLS credentials
	creds := c.otelConfig.GetTLSCredentials()

	// Initialize metrics
	metricsConfig := &MetricsExporterConfig{
		ServiceName: c.otelConfig.ServiceName,
		Environment: c.otelConfig.Environment,
		Endpoint:    c.otelConfig.Endpoint,
		Headers:     c.otelConfig.Headers,
		TLSCreds:    creds,
	}
	meterProvider, err := NewMetricsExporter(ctx, metricsConfig, res)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}

	// Set global meter provider
	otel.SetMeterProvider(meterProvider)

	// Initialize tracing
	tracingConfig := &TracingExporterConfig{
		ServiceName: c.otelConfig.ServiceName,
		Environment: c.otelConfig.Environment,
		Endpoint:    c.otelConfig.Endpoint,
		Headers:     c.otelConfig.Headers,
		TLSCreds:    creds,
	}
	tracerProvider, err := NewTracingExporter(ctx, tracingConfig, res)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer provider: %w", err)
	}

	// Set global tracer provider
	otel.SetTracerProvider(tracerProvider)

	// Setup propagator
	SetupPropagator()

	// Return cleanup function
	cleanup := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
			log.Printf("[ERROR] Error shutting down tracer provider: %v\n", err)
		}

		if err := meterProvider.Shutdown(shutdownCtx); err != nil {
			log.Printf("[ERROR] Error shutting down meter provider: %v\n", err)
		}
	}

	return cleanup, nil
}
