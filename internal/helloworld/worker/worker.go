package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/creator54/temporal-go-demo-app/internal/helloworld/workflow/impl"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// HelloWorldActivity is a basic activity that returns a greeting
func HelloWorldActivity(ctx context.Context, name string) (string, error) {
	result := fmt.Sprintf("Hello %s!", name)
	return result, nil
}

// StartWorker initializes and starts a Temporal worker
func StartWorker() {
	log.Println("[INFO] Initializing Temporal worker...")

	// Create the client options
	clientOptions := client.Options{
		HostPort: client.DefaultHostPort,
	}

	// Initialize the Temporal client
	temporalClient, err := client.NewClient(clientOptions)
	if err != nil {
		log.Fatalf("[ERROR] Unable to create Temporal client: %v", err)
	}
	defer func() {
		log.Println("[INFO] Closing Temporal client...")
		temporalClient.Close()
	}()

	// Create the worker options with default settings
	workerOptions := worker.Options{
		EnableLoggingInReplay: true,
		// Use reasonable defaults for concurrency
		MaxConcurrentActivityExecutionSize: 20,
	}

	// Create a new worker
	w := worker.New(temporalClient, "hello-world-task-queue", workerOptions)

	// Register workflow and activity
	w.RegisterWorkflow((&impl.HelloWorldImpl{}).SayHello)
	w.RegisterActivity(HelloWorldActivity)

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the worker
	log.Println("[INFO] Starting worker...")
	if err := w.Start(); err != nil {
		log.Fatalf("[ERROR] Unable to start worker: %v", err)
	}
	log.Println("[INFO] Worker started successfully")

	// Handle graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	log.Println("[INFO] Shutdown signal received, stopping worker...")

	// Create a context with timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	// Stop accepting new tasks
	w.Stop()

	// Wait for in-flight workflows to complete or timeout
	select {
	case <-shutdownCtx.Done():
		if shutdownCtx.Err() == context.DeadlineExceeded {
			log.Println("[WARN] Graceful shutdown timed out, forcing exit...")
		}
	case <-time.After(100 * time.Millisecond): // Small buffer to ensure cleanup messages are printed
		log.Println("[INFO] Worker stopped gracefully")
	}
}
