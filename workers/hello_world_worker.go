package workers

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// HelloWorldWorkflow is a basic workflow that calls an activity
func HelloWorldWorkflow(ctx workflow.Context, name string) (string, error) {
	logger := workflow.GetLogger(ctx)

	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var result string
	err := workflow.ExecuteActivity(ctx, HelloWorldActivity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("[ERROR] Activity execution failed", "error", err)
		return "", err
	}
	return result, nil
}

// HelloWorldActivity is a basic activity that returns a greeting
func HelloWorldActivity(ctx context.Context, name string) (string, error) {
	result := fmt.Sprintf("Hello %s!", name)
	return result, nil
}

func StartWorker() {
	// Create the client options with tracing interceptor
	tracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
	if err != nil {
		log.Fatalln("[ERROR] Unable to create tracing interceptor:", err)
	}

	clientOptions := client.Options{
		HostPort:     client.DefaultHostPort,
		Interceptors: []interceptor.ClientInterceptor{tracingInterceptor},
	}

	// Initialize the Temporal client
	temporalClient, err := client.NewClient(clientOptions)
	if err != nil {
		log.Fatalln("[ERROR] Unable to create Temporal client:", err)
	}
	defer temporalClient.Close()

	// Create the worker options with tracing interceptor
	workerOptions := worker.Options{
		EnableLoggingInReplay: true,
		Interceptors:          []interceptor.WorkerInterceptor{tracingInterceptor},
	}

	// Create a new worker
	w := worker.New(temporalClient, "hello-world-task-queue", workerOptions)

	// Register workflow and activity
	w.RegisterWorkflow(HelloWorldWorkflow)
	w.RegisterActivity(HelloWorldActivity)

	// Start the worker
	err = w.Start()
	if err != nil {
		log.Fatalln("[ERROR] Unable to start worker:", err)
	}

	// Handle graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan
	w.Stop()
}
