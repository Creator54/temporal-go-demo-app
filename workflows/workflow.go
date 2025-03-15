package workflows

import "go.temporal.io/sdk/workflow"

// HelloWorldWorkflow defines the interface for our workflow
type HelloWorldWorkflow interface {
	SayHello(ctx workflow.Context, name string) (string, error)
} 
