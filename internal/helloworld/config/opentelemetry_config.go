package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// OpenTelemetryConfig holds OpenTelemetry configuration
type OpenTelemetryConfig struct {
	ServiceName string
	Environment string
	Endpoint    string
	Headers     map[string]string
	UseTLS      bool
}

// NewOpenTelemetryConfig creates a new OpenTelemetry configuration
func NewOpenTelemetryConfig() *OpenTelemetryConfig {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4317"
	} else {
		// Strip http:// or https:// prefix
		endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "http://"), "https://")
	}

	return &OpenTelemetryConfig{
		ServiceName: "temporal-hello-world",
		Environment: "development",
		Endpoint:    endpoint,
		Headers:     parseHeaders(),
		UseTLS:      false,
	}
}

// CreateResource creates a new resource with common attributes
func (c *OpenTelemetryConfig) CreateResource(ctx context.Context) (*resource.Resource, error) {
	hostname, _ := os.Hostname()
	return resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(c.ServiceName),
			semconv.ServiceVersion("1.0.0"),
			semconv.DeploymentEnvironment(c.Environment),
			semconv.HostName(hostname),
			semconv.ServiceInstanceID(fmt.Sprintf("%s-%d", hostname, os.Getpid())),
		),
		resource.WithSchemaURL(semconv.SchemaURL),
	)
}

// GetTLSCredentials returns the appropriate TLS credentials
func (c *OpenTelemetryConfig) GetTLSCredentials() credentials.TransportCredentials {
	if c.UseTLS {
		return credentials.NewClientTLSFromCert(nil, "")
	}
	return insecure.NewCredentials()
}

// SetupPropagator configures the global propagator
func SetupPropagator() {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
}

// parseHeaders parses OTEL_EXPORTER_OTLP_HEADERS environment variable
func parseHeaders() map[string]string {
	headers := make(map[string]string)
	headerStr := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")
	if headerStr != "" {
		// Parse headers in format "key1=value1,key2=value2"
		pairs := strings.Split(headerStr, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}
	return headers
}
