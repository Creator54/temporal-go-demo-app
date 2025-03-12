package workflows

import "go.temporal.io/sdk/workflow"

// HelloWorldWorkflow defines the interface for our workflow
type HelloWorldWorkflow interface {
	// SayHello matches the Java example's method name
	SayHello(ctx workflow.Context, name string) (string, error)
} 