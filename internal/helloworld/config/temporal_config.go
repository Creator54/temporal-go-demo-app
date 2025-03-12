package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"go.temporal.io/sdk/client"
)

// GetTaskQueue returns the task queue name
func GetTaskQueue() string {
	return "hello-world-task-queue"
}

// GetTemporalClient creates a new Temporal client with default options
func GetTemporalClient() (client.Client, error) {
	return GetTemporalClientWithOptions(client.Options{})
}

// GetTemporalClientWithOptions creates a new Temporal client with custom options
func GetTemporalClientWithOptions(options client.Options) (client.Client, error) {
	// Get Temporal address from environment variable or use default
	address := os.Getenv("TEMPORAL_ADDRESS")
	if address == "" {
		address = "localhost:7233"
	}

	// Check if we need to use TLS
	if os.Getenv("TEMPORAL_MTLS_TLS_CERT") != "" && os.Getenv("TEMPORAL_MTLS_TLS_KEY") != "" {
		cert, err := tls.LoadX509KeyPair(
			os.Getenv("TEMPORAL_MTLS_TLS_CERT"),
			os.Getenv("TEMPORAL_MTLS_TLS_KEY"),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert and key: %w", err)
		}

		// Load CA cert if provided
		var certPool *x509.CertPool
		if caCertPath := os.Getenv("TEMPORAL_MTLS_TLS_CA"); caCertPath != "" {
			certPool = x509.NewCertPool()
			caCert, err := os.ReadFile(caCertPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load CA cert: %w", err)
			}
			certPool.AppendCertsFromPEM(caCert)
		}

		options.ConnectionOptions = client.ConnectionOptions{
			TLS: &tls.Config{
				Certificates: []tls.Certificate{cert},
				RootCAs:      certPool,
			},
		}
	}

	// Set the address in the options
	options.HostPort = address

	// Create the client
	return client.Dial(options)
}
