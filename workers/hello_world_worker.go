package workers

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// HelloWorldWorkflow is a basic workflow that calls an activity
func HelloWorldWorkflow(ctx workflow.Context, name string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("[DEBUG] Starting HelloWorldWorkflow", "name", name)

	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var result string
	logger.Info("[DEBUG] Executing HelloWorldActivity")
	err := workflow.ExecuteActivity(ctx, HelloWorldActivity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("[ERROR] Activity execution failed", "error", err)
		return "", err
	}
	logger.Info("[DEBUG] Activity execution completed successfully", "result", result)
	return result, nil
}

// HelloWorldActivity is a basic activity that returns a greeting
func HelloWorldActivity(ctx context.Context, name string) (string, error) {
	// Get the current span from context
	span := trace.SpanFromContext(ctx)
	log.Printf("[DEBUG] Activity span: TraceID=%s, SpanID=%s", span.SpanContext().TraceID(), span.SpanContext().SpanID())

	result := fmt.Sprintf("Hello %s!", name)
	log.Printf("[DEBUG] Activity executed with result: %s", result)
	return result, nil
}

func StartWorker() {
	log.Println("[DEBUG] Starting worker initialization...")

	// Create the client options with tracing interceptor
	log.Println("[DEBUG] Creating tracing interceptor...")
	tracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
	if err != nil {
		log.Fatalln("[ERROR] Unable to create tracing interceptor:", err)
	}
	log.Println("[DEBUG] Tracing interceptor created successfully")

	// Get the current tracer
	tracer := otel.GetTracerProvider().Tracer("temporal-worker")
	log.Printf("[DEBUG] Using tracer: %v", tracer)

	clientOptions := client.Options{
		HostPort:     client.DefaultHostPort,
		Interceptors: []interceptor.ClientInterceptor{tracingInterceptor},
	}

	// Initialize the Temporal client
	log.Println("[DEBUG] Creating Temporal client...")
	temporalClient, err := client.NewClient(clientOptions)
	if err != nil {
		log.Fatalln("[ERROR] Unable to create Temporal client:", err)
	}
	defer temporalClient.Close()
	log.Println("[DEBUG] Temporal client created successfully")

	// Create the worker options with tracing interceptor
	workerOptions := worker.Options{
		EnableLoggingInReplay: true,
		Interceptors:          []interceptor.WorkerInterceptor{tracingInterceptor},
	}

	// Create a new worker
	log.Println("[DEBUG] Creating worker...")
	w := worker.New(temporalClient, "hello-world-task-queue", workerOptions)
	log.Println("[DEBUG] Worker created successfully")

	// Register workflow and activity
	log.Println("[DEBUG] Registering workflow and activity...")
	w.RegisterWorkflow(HelloWorldWorkflow)
	w.RegisterActivity(HelloWorldActivity)
	log.Println("[DEBUG] Workflow and activity registered successfully")

	// Start the worker
	log.Println("[DEBUG] Starting worker...")
	err = w.Start()
	if err != nil {
		log.Fatalln("[ERROR] Unable to start worker:", err)
	}
	log.Println("[DEBUG] Worker started successfully")

	// Handle graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan
	log.Println("[DEBUG] Received shutdown signal")
	w.Stop()
	log.Println("[DEBUG] Worker stopped successfully")
}
