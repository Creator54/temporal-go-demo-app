package starter

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/creator54/temporal-go-demo-app/internal/helloworld/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.temporal.io/sdk/client"
)

// StartWorkflow initiates the HelloWorld workflow
func StartWorkflow(ctx context.Context, name string) error {
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
	attrs := config.GetSpanAttributes("temporal", workflowID, workflowOptions.TaskQueue)
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
	workflowAttrs := config.GetSpanAttributes("HelloWorldWorkflow", workflowID, workflowOptions.TaskQueue)
	workflowAttrs["workflow.input"] = name
	for k, v := range workflowAttrs {
		workflowSpan.SetAttributes(attribute.String(k, v))
	}
	defer workflowSpan.End()

	fmt.Printf("Starting workflow with input: %s\n", name)

	// Start the workflow
	we, err := c.ExecuteWorkflow(ctx, workflowOptions, "SayHello", name)
	if err != nil {
		return fmt.Errorf("unable to execute workflow: %w", err)
	}

	log.Printf("Started workflow with ID: %s\n", workflowID)

	// Wait for workflow completion
	var result string
	err = we.Get(ctx, &result)
	if err != nil {
		return fmt.Errorf("unable to get workflow result: %w", err)
	}

	log.Printf("Workflow result: %s\n", result)

	// Sleep briefly to ensure spans are exported
	time.Sleep(2 * time.Second)
	return nil
}
