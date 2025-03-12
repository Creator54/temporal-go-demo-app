package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/creator54/temporal-go-demo-app/internal/helloworld/config"
	"github.com/creator54/temporal-go-demo-app/internal/helloworld/starter"
	"github.com/creator54/temporal-go-demo-app/internal/helloworld/worker"
)

func main() {
	// Parse command line flags
	workerMode := flag.Bool("worker", false, "Run in worker mode")
	flag.Parse()

	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	if *workerMode {
		log.SetPrefix("[WORKER] ")
	} else {
		log.SetPrefix("[WORKFLOW] ")
	}

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("[INFO] Received signal %v, initiating shutdown...", sig)
		cancel()
	}()

	// Initialize OpenTelemetry
	log.Println("[INFO] Initializing OpenTelemetry...")
	telemetry := config.NewSignozTelemetryUtils()
	cleanup, err := telemetry.InitProvider(ctx)
	if err != nil {
		log.Fatalf("[ERROR] Failed to initialize OpenTelemetry: %v", err)
	}
	defer func() {
		log.Println("[INFO] Cleaning up OpenTelemetry resources...")
		cleanup()
	}()
	log.Println("[INFO] OpenTelemetry initialized successfully")

	if *workerMode {
		// Start the worker
		log.Println("[INFO] Starting worker process...")
		worker.StartWorker()
	} else {
		// Get workflow input name from environment
		name := os.Getenv("WORKFLOW_NAME")
		if name == "" {
			name = "Temporal"
		}

		log.Printf("[INFO] Starting workflow with input: %s", name)

		// Create a context with timeout for workflow execution
		execCtx, execCancel := context.WithTimeout(ctx, 30*time.Second)
		defer execCancel()

		// Start the workflow using the context
		if err := starter.StartWorkflow(execCtx, name); err != nil {
			if execCtx.Err() == context.DeadlineExceeded {
				log.Fatalf("[ERROR] Workflow execution timed out after 30 seconds")
			}
			log.Fatalf("[ERROR] Failed to start workflow: %v", err)
		}

		log.Println("[INFO] Workflow completed successfully")
	}
}
