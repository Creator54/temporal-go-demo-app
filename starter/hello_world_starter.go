package starter

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/creator54/temporal-go-demo-app/config"
	"go.temporal.io/sdk/client"
)

// StartWorkflow initiates the HelloWorld workflow
func StartWorkflow(name string) error {
	// Create the client
	c, err := config.GetTemporalClient()
	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}
	defer c.Close()

	// Create workflow options
	workflowID := fmt.Sprintf("hello-world-%v", time.Now().Unix())
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: config.GetTaskQueue(),
	}

	// Start the workflow using the registered name "HelloWorldWorkflow"
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, "HelloWorldWorkflow", name)
	if err != nil {
		return fmt.Errorf("unable to execute workflow: %w", err)
	}

	log.Printf("Started workflow with ID: %s\n", workflowID)

	// Wait for workflow completion
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		return fmt.Errorf("unable to get workflow result: %w", err)
	}

	log.Printf("Workflow result: %s\n", result)
	return nil
} 