package impl

import (
	"fmt"

	"github.com/creator54/temporal-go-demo-app/workflows"
	"go.temporal.io/sdk/workflow"
)

// HelloWorldWorkflowImpl implements the HelloWorldWorkflow interface
type HelloWorldWorkflowImpl struct{}

// Verify HelloWorldWorkflowImpl implements HelloWorldWorkflow
var _ workflows.HelloWorldWorkflow = (*HelloWorldWorkflowImpl)(nil)

// SayHello is a simple workflow that returns a greeting
func (w *HelloWorldWorkflowImpl) SayHello(ctx workflow.Context, name string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Executing HelloWorldWorkflow", "name", name)

	greeting := fmt.Sprintf("Hello %s!", name)
	return greeting, nil
}
