package workers

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/creator54/temporal-go-demo-app/config"
	"github.com/creator54/temporal-go-demo-app/workflows/impl"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// StartWorker initializes and starts a Temporal worker
func StartWorker() error {
	// Create the client
	c, err := config.GetTemporalClient()
	if err != nil {
		return err
	}
	defer c.Close()

	// Create a worker
	w := worker.New(c, config.GetTaskQueue(), worker.Options{})

	// Register workflow with the name "HelloWorldWorkflow"
	w.RegisterWorkflowWithOptions(
		new(impl.HelloWorldWorkflowImpl).SayHello,
		workflow.RegisterOptions{Name: "HelloWorldWorkflow"},
	)

	// Start the worker
	err = w.Start()
	if err != nil {
		return err
	}

	log.Printf("Worker started for task queue: %s\n", config.GetTaskQueue())

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down worker...")
	w.Stop()
	return nil
} 