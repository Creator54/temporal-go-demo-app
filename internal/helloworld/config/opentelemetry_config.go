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
	useTLS := false

	if endpoint == "" {
		endpoint = "localhost:4317"
	} else {
		// Check if this is a cloud endpoint that needs TLS
		if strings.Contains(endpoint, "https://") || strings.Contains(endpoint, "signoz.cloud") {
			useTLS = true
			// For gRPC, we don't need the https:// prefix
			endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")
		} else {
			// For non-TLS endpoint
			endpoint = strings.TrimPrefix(strings.TrimPrefix(endpoint, "http://"), "https://")
		}
	}

	// Debug output
	headers := parseHeaders()
	fmt.Printf("[DEBUG] OTEL endpoint: %s\n", endpoint)
	fmt.Printf("[DEBUG] OTEL headers: %v\n", headers)
	fmt.Printf("[DEBUG] OTEL using TLS: %v\n", useTLS)

	return &OpenTelemetryConfig{
		ServiceName: "temporal-hello-world",
		Environment: "development",
		Endpoint:    endpoint,
		Headers:     headers,
		UseTLS:      useTLS,
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
	fmt.Printf("[DEBUG] Getting TLS credentials, UseTLS: %v\n", c.UseTLS)
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

// GetSpanAttributes returns a map of common attributes for spans
func GetSpanAttributes(workflowType, workflowID, taskQueue string) map[string]string {
	attrs := make(map[string]string)
	attrs["workflow_type"] = workflowType
	attrs["workflow_id"] = workflowID
	attrs["task_queue"] = taskQueue
	attrs["namespace"] = "default" // Default namespace, can be overridden if needed
	return attrs
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
