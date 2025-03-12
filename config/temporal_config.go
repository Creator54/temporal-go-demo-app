package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"go.temporal.io/sdk/client"
)

const (
	defaultNamespace = "default"
	defaultTaskQueue = "hello-world-task-queue"
)

// GetTemporalClient returns a configured Temporal client
func GetTemporalClient() (client.Client, error) {
	// Get configuration from environment variables or use defaults
	namespace := getEnvOrDefault("TEMPORAL_NAMESPACE", defaultNamespace)
	
	clientOptions := client.Options{
		Namespace: namespace,
	}

	// Check both HOST_URL and HOST_ADDRESS for backward compatibility
	hostAddr := os.Getenv("TEMPORAL_HOST_URL")
	if hostAddr == "" {
		hostAddr = os.Getenv("TEMPORAL_HOST_ADDRESS")
	}
	if hostAddr != "" {
		clientOptions.HostPort = hostAddr
	}

	// Configure mTLS if certificates are provided
	certPath := os.Getenv("TEMPORAL_TLS_CERT")
	keyPath := os.Getenv("TEMPORAL_TLS_KEY")
	if certPath != "" && keyPath != "" {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificates: %w", err)
		}

		// Create a certificate pool from the system CA certs
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("failed to load system cert pool: %w", err)
		}

		clientOptions.ConnectionOptions = client.ConnectionOptions{
			TLS: &tls.Config{
				Certificates: []tls.Certificate{cert},
				RootCAs:     certPool,
				MinVersion:  tls.VersionTLS12,
			},
		}
	}

	return client.NewClient(clientOptions)
}

// GetTaskQueue returns the task queue name from environment or default
func GetTaskQueue() string {
	return getEnvOrDefault("TEMPORAL_TASK_QUEUE", defaultTaskQueue)
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 