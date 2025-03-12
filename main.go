package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/creator54/temporal-go-demo-app/telemetry"
	"github.com/creator54/temporal-go-demo-app/workers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/client"
)

func main() {
	log.Println("[DEBUG] Starting application...")

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("[DEBUG] Received shutdown signal")
		cancel()
	}()

	// Initialize OpenTelemetry
	log.Println("[DEBUG] Initializing OpenTelemetry...")
	config := telemetry.NewConfig()
	cleanup, err := config.InitProvider(ctx)
	if err != nil {
		log.Fatalf("[ERROR] Failed to initialize OpenTelemetry: %v", err)
	}
	defer cleanup()

	// Create a test span
	tr := otel.GetTracerProvider().Tracer("test-tracer")
	ctx, span := tr.Start(ctx, "TestConnection")
	span.SetAttributes(
		attribute.String("test.attribute", "test-value"),
		attribute.String("service.name", "temporal-hello-world"),
	)
	log.Printf("[DEBUG] Created test span: TraceID=%s, SpanID=%s", span.SpanContext().TraceID(), span.SpanContext().SpanID())

	// Add an event to the span
	span.AddEvent("test.event", trace.WithAttributes(
		attribute.String("event.type", "test"),
		attribute.String("event.message", "Testing OpenTelemetry connection"),
	))

	// Sleep for a moment to allow the span to be exported
	time.Sleep(2 * time.Second)
	span.End()
	log.Println("[DEBUG] Test span completed")

	// Start the worker in a goroutine
	log.Println("[DEBUG] Starting worker...")
	go workers.StartWorker()

	// Create client options
	log.Println("[DEBUG] Creating Temporal client...")
	clientOptions := client.Options{
		HostPort: client.DefaultHostPort,
	}

	// Create Temporal client
	c, err := client.NewClient(clientOptions)
	if err != nil {
		log.Fatalln("[ERROR] Unable to create client:", err)
	}
	defer c.Close()
	log.Println("[DEBUG] Temporal client created successfully")

	workflowOptions := client.StartWorkflowOptions{
		ID:        "hello-world-" + fmt.Sprint(time.Now().Unix()),
		TaskQueue: "hello-world-task-queue",
	}

	name := os.Getenv("WORKFLOW_NAME")
	if name == "" {
		name = "Temporal"
	}

	// Create a new span for workflow execution
	tr = otel.GetTracerProvider().Tracer("temporal-workflow")
	ctx, span = tr.Start(ctx, "ExecuteWorkflow")
	defer span.End()

	log.Printf("[DEBUG] Starting workflow with name '%s'...\n", name)
	log.Printf("[DEBUG] Workflow span: TraceID=%s, SpanID=%s", span.SpanContext().TraceID(), span.SpanContext().SpanID())

	we, err := c.ExecuteWorkflow(ctx, workflowOptions, workers.HelloWorldWorkflow, name)
	if err != nil {
		log.Fatalln("[ERROR] Unable to execute workflow:", err)
	}
	log.Printf("[DEBUG] Workflow started with ID: %s\n", we.GetID())

	// Wait for workflow completion
	var result string
	err = we.Get(ctx, &result)
	if err != nil {
		log.Fatalln("[ERROR] Unable to get workflow result:", err)
	}
	log.Printf("[DEBUG] Workflow completed successfully with result: %s\n", result)

	// Sleep for a moment to allow final spans to be exported
	time.Sleep(2 * time.Second)
}
