package impl

import (
	"fmt"

	"github.com/creator54/temporal-go-demo-app/internal/helloworld/workflow"
	sdkworkflow "go.temporal.io/sdk/workflow"
)

// HelloWorldImpl implements the HelloWorld interface
type HelloWorldImpl struct{}

// Verify HelloWorldImpl implements HelloWorld
var _ workflow.HelloWorld = (*HelloWorldImpl)(nil)

// SayHello is a simple workflow that returns a greeting
func (w *HelloWorldImpl) SayHello(ctx sdkworkflow.Context, name string) (string, error) {
	logger := sdkworkflow.GetLogger(ctx)
	logger.Info("Executing HelloWorldWorkflow", "name", name)

	greeting := fmt.Sprintf("Hello %s!", name)
	return greeting, nil
}
