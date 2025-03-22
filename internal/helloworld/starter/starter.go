package starter

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/creator54/temporal-go-demo-app/internal/helloworld/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.temporal.io/sdk/client"
)

// Metrics for workflow execution
var (
	workflowStartCounter     metric.Int64Counter
	workflowCompletionCounter metric.Int64Counter
	metricsInitialized       bool
)

// Initialize workflow counters
func initializeStarterMetrics() {
	if metricsInitialized {
		return
	}

	// Get meter
	meter := otel.GetMeterProvider().Meter("temporal-starter")
	log.Println("Initializing starter metrics...")

	// Create workflow start counter
	var err error
	workflowStartCounter, err = meter.Int64Counter(
		"workflow_started_count_total",
		metric.WithDescription("Total workflow executions started"),
		metric.WithUnit("{execution}"),
	)
	if err != nil {
		log.Printf("Failed to create workflow start counter: %v", err)
		return
	}

	// Create workflow completion counter
	workflowCompletionCounter, err = meter.Int64Counter(
		"workflow_completed_count_total",
		metric.WithDescription("Total workflow executions completed"),
		metric.WithUnit("{execution}"),
	)
	if err != nil {
		log.Printf("Failed to create workflow completion counter: %v", err)
		return
	}

	metricsInitialized = true
	log.Println("Starter metrics initialized successfully")
}

// StartWorkflow initiates the HelloWorld workflow
func StartWorkflow(ctx context.Context, name string) error {
	// Initialize metrics if not already initialized
	initializeStarterMetrics()

	// Create the client options
	clientOptions := client.Options{
		HostPort: client.DefaultHostPort,
	}

	// Initialize the Temporal client
	c, err := client.NewClient(clientOptions)
	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}
	defer c.Close()

	// Create workflow options
	workflowID := fmt.Sprintf("hello-world-%v", time.Now().Unix())
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "hello-world-task-queue",
	}

	// Create a tracer
	tr := otel.GetTracerProvider().Tracer("temporal-workflow")

	// Create the parent StartWorkflow span
	ctx, startSpan := tr.Start(ctx, "StartWorkflow")
	attrs := config.GetTracingAttributes("temporal", workflowID, workflowOptions.TaskQueue)
	for k, v := range attrs {
		startSpan.SetAttributes(attribute.String(k, v))
	}
	defer startSpan.End()

	// Create ExecuteWorkflow span as child of StartWorkflow
	ctx, executeSpan := tr.Start(ctx, "ExecuteWorkflow")
	for k, v := range attrs {
		executeSpan.SetAttributes(attribute.String(k, v))
	}
	defer executeSpan.End()

	// Create StartWorkflow:HelloWorldWorkflow span as child of ExecuteWorkflow
	ctx, workflowSpan := tr.Start(ctx, "StartWorkflow:HelloWorldWorkflow")
	workflowAttrs := config.GetTracingAttributes("HelloWorldWorkflow", workflowID, workflowOptions.TaskQueue)
	workflowAttrs["workflow.input"] = name
	for k, v := range workflowAttrs {
		workflowSpan.SetAttributes(attribute.String(k, v))
	}
	defer workflowSpan.End()

	fmt.Printf("Starting workflow with input: %s\n", name)

	// Record workflow start metric
	if metricsInitialized {
		workflowStartCounter.Add(ctx, 1, metric.WithAttributes())
	}

	// Start the workflow
	we, err := c.ExecuteWorkflow(ctx, workflowOptions, "SayHello", name)
	if err != nil {
		// Record failure metric
		config.RecordFailure(ctx, "HelloWorldWorkflow", workflowID, we.GetRunID(), "default")
		return fmt.Errorf("unable to execute workflow: %w", err)
	}

	log.Printf("Started workflow with ID: %s\n", workflowID)

	// Wait for workflow completion
	var result string
	err = we.Get(ctx, &result)

	// Record appropriate metrics based on workflow outcome
	if metricsInitialized {
		workflowCompletionCounter.Add(ctx, 1, metric.WithAttributes())
	}

	if err != nil {
		// Record workflow failure
		config.RecordFailure(ctx, "HelloWorldWorkflow", workflowID, we.GetRunID(), "default")
		return fmt.Errorf("unable to get workflow result: %w", err)
	}

	// Record workflow success
	config.RecordSuccess(ctx, "HelloWorldWorkflow", workflowID, we.GetRunID(), "default")

	log.Printf("Workflow result: %s\n", result)

	// Sleep briefly to ensure spans are exported
	time.Sleep(2 * time.Second)
	return nil
}
