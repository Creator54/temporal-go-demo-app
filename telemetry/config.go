package telemetry

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultServiceName = "temporal-hello-world"
	defaultEnvironment = "development"
)

// Config holds the OpenTelemetry configuration
type Config struct {
	ServiceName string
	Environment string
}

// NewConfig creates a new OpenTelemetry configuration
func NewConfig() *Config {
	// Parse service name from resource attributes or use default
	serviceName := defaultServiceName
	resourceAttrs := os.Getenv("OTEL_RESOURCE_ATTRIBUTES")
	if resourceAttrs != "" {
		attrs := strings.Split(resourceAttrs, ",")
		for _, attr := range attrs {
			kv := strings.Split(attr, "=")
			if len(kv) == 2 && kv[0] == "service.name" {
				serviceName = kv[1]
				break
			}
		}
	}

	return &Config{
		ServiceName: serviceName,
		Environment: defaultEnvironment,
	}
}

// InitProvider initializes the OpenTelemetry provider with tracing and metrics
func (c *Config) InitProvider(ctx context.Context) (func(), error) {
	// Get endpoint configuration
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4317"
	}

	// Determine if using TLS based on endpoint scheme
	useTLS := strings.HasPrefix(endpoint, "https://")

	// Strip http:// or https:// prefix
	endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "http://"), "https://")
	// log.Printf("[DEBUG] Using OTLP endpoint: %s (TLS: %v)", endpoint, useTLS)

	// Create resource with service information
	hostname, _ := os.Hostname()
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(c.ServiceName),
			semconv.ServiceVersion("1.0.0"),
			semconv.DeploymentEnvironment(c.Environment),
			attribute.String("library.language", "go"),
			attribute.String("host.name", hostname),
			attribute.String("service.instance.id", fmt.Sprintf("%s-%d", hostname, os.Getpid())),
		),
		resource.WithSchemaURL(semconv.SchemaURL),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}
	// log.Printf("[DEBUG] Created OpenTelemetry resource with attributes: %+v", res.Attributes())

	// Set up gRPC connection options
	var creds credentials.TransportCredentials
	if useTLS {
		creds = credentials.NewClientTLSFromCert(nil, "")
	} else {
		creds = insecure.NewCredentials()
	}

	// Get headers for authentication
	headers := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")
	var metadata map[string]string
	if headers != "" {
		metadata = make(map[string]string)
		for _, header := range strings.Split(headers, ",") {
			parts := strings.SplitN(header, "=", 2)
			if len(parts) == 2 {
				metadata[parts[0]] = parts[1]
			}
		}
	}

	// Initialize trace exporter
	// log.Printf("[DEBUG] Creating trace exporter...")
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithHeaders(metadata),
		otlptracegrpc.WithTLSCredentials(creds),
		otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
			Enabled:         true,
			InitialInterval: 1 * time.Second,
			MaxInterval:     5 * time.Second,
			MaxElapsedTime:  30 * time.Second,
		}),
	)
	if err != nil {
		log.Printf("[ERROR] Failed to create trace exporter: %v", err)
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}
	// log.Println("[DEBUG] Trace exporter created successfully")

	// Configure trace provider with more frequent batching
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(traceExporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
			sdktrace.WithMaxQueueSize(2048),
		),
	)
	otel.SetTracerProvider(tracerProvider)
	// log.Println("[DEBUG] Tracer provider configured and set globally")

	// Initialize metrics exporter
	// log.Printf("[DEBUG] Creating metrics exporter...")
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithHeaders(metadata),
		otlpmetricgrpc.WithTLSCredentials(creds),
		otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetryConfig{
			Enabled:         true,
			InitialInterval: 1 * time.Second,
			MaxInterval:     5 * time.Second,
			MaxElapsedTime:  30 * time.Second,
		}),
	)
	if err != nil {
		log.Printf("[ERROR] Failed to create metric exporter: %v", err)
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}
	// log.Println("[DEBUG] Metrics exporter created successfully")

	// Configure metrics provider with more frequent reporting
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(10*time.Second),
		)),
	)
	otel.SetMeterProvider(meterProvider)
	// log.Println("[DEBUG] Meter provider configured and set globally")

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	// log.Println("[DEBUG] Text map propagator configured and set globally")

	// Return cleanup function
	cleanup := func() {
		// log.Println("[DEBUG] Starting OpenTelemetry shutdown...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
