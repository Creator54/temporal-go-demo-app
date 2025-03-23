package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

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
	address := os.Getenv("TEMPORAL_HOST_URL")
	log.Printf("[DEBUG] TEMPORAL_HOST_URL value: %q", address)
	
	if address == "" {
		address = "localhost:7233"
		log.Printf("[DEBUG] Using default address: %q", address)
	}

	// Check if we need to use TLS
	certPath := os.Getenv("TEMPORAL_TLS_CERT")
	keyPath := os.Getenv("TEMPORAL_TLS_KEY")
	log.Printf("[DEBUG] TLS config - Cert: %q, Key: %q", certPath, keyPath)
	
	if certPath != "" && keyPath != "" {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert and key: %w", err)
		}

		options.ConnectionOptions = client.ConnectionOptions{
			TLS: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		}
		log.Println("[DEBUG] TLS configuration applied")
	}

	// Set the namespace if provided
	namespace := os.Getenv("TEMPORAL_NAMESPACE")
	log.Printf("[DEBUG] TEMPORAL_NAMESPACE value: %q", namespace)
	if namespace != "" {
		options.Namespace = namespace
		log.Printf("[DEBUG] Namespace set to: %q", namespace)
	}

	// Set the address in the options
	options.HostPort = address
	log.Printf("[DEBUG] Final connection address: %q", address)

	// Create the client
	return client.Dial(options)
}

// RegisterDashboardMetrics registers a timer to periodically emit metrics for dashboard visualization
func RegisterDashboardMetrics(ctx context.Context) {
	// Create a ticker that runs every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	// Add a delay for the first time
	initialDelay := time.NewTimer(1 * time.Second)

	go func() {
		// Wait for initial delay
		<-initialDelay.C
		
		for {
			select {
			case <-ticker.C:
				// Generate random workflow IDs and run IDs for demo purposes
				workflowID := fmt.Sprintf("workflow-%d", rand.Int63n(1000))
				runID := fmt.Sprintf("run-%d", rand.Int63n(1000))
				namespace := "default"
				
				// Record simulated metrics for all workflow states
				// Success metrics (most common)
				for i := 0; i < 5; i++ {
					RecordSuccess(ctx, "HelloWorldWorkflow", workflowID+fmt.Sprintf("-%d", i), runID+fmt.Sprintf("-%d", i), namespace)
				}
				
				// Failure metrics
				for i := 0; i < 2; i++ {
					RecordFailure(ctx, "HelloWorldWorkflow", workflowID+fmt.Sprintf("-failed-%d", i), runID+fmt.Sprintf("-failed-%d", i), namespace)
				}
				
				// Timeout metrics
				RecordTimeout(ctx, "HelloWorldWorkflow", workflowID+"-timeout", runID+"-timeout", namespace)
				
				// Termination metrics
				RecordTermination(ctx, "HelloWorldWorkflow", workflowID+"-terminate", runID+"-terminate", namespace)
				
				// Cancellation metrics
				RecordCancellation(ctx, "HelloWorldWorkflow", workflowID+"-cancel", runID+"-cancel", namespace)
				
				// Record various service requests
				RecordAddActivityTask(ctx)
				RecordAddWorkflowTask(ctx)
				RecordResponseActivityCompleted(ctx)
				RecordRespondWorkflowTaskCompleted(ctx)
				
				// Record some errors
				RecordSystemError(ctx)
				RecordTimeoutError(ctx)
				RecordBusinessRuleError(ctx)
				RecordValidationError(ctx)
				
				// Record timeout metrics
				RecordScheduleToStartWorkflowTimeout(ctx)
				RecordStartToCloseWorkflowTimeout(ctx)
				
				// Record service restarts
				RecordServiceRestart(ctx, "worker")
				
				log.Println("Emitted periodic metrics for dashboard visualization")
				
			case <-ctx.Done():
				ticker.Stop()
				log.Println("Stopping dashboard metrics emission")
				return
			}
		}
	}()
	
	log.Println("Registered dashboard metrics timer")
}
