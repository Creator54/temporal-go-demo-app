package workflow

import "go.temporal.io/sdk/workflow"

// HelloWorld defines the interface for our workflow
type HelloWorld interface {
	// SayHello matches the Java example's method name
	SayHello(ctx workflow.Context, name string) (string, error)
}
