package config

import (
	"context"
	"log"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	// Metric counters
	workflowSuccess       metric.Int64Counter
	workflowFailed        metric.Int64Counter
	workflowTimeout       metric.Int64Counter
	workflowTerminate     metric.Int64Counter
	workflowCancel        metric.Int64Counter
	serviceRequests       metric.Int64Counter
	serviceErrors         metric.Int64Counter
	serviceErrorWithType  metric.Int64Counter
	restarts              metric.Int64Counter
	scheduleToStartTimeout metric.Int64Counter
	startToCloseTimeout   metric.Int64Counter

	// Attributes constants
	WorkflowTypeKey     = attribute.Key("workflow_type")
	WorkflowIDKey       = attribute.Key("workflow_id")
	RunIDKey            = attribute.Key("run_id")
	NamespaceKey        = attribute.Key("namespace")
	OperationKey        = attribute.Key("operation")
	ServiceTypeKey      = attribute.Key("temporal_service_type")
	ErrorTypeKey        = attribute.Key("error_type")

	once sync.Once
)

// InitializeMetrics initializes all workflow metrics
func InitializeMetrics(meter metric.Meter) error {
	var err error

	// Initialize workflow state metrics
	workflowSuccess, err = meter.Int64Counter(
		"workflow_success",
		metric.WithDescription("Count of successfully completed workflows"),
		metric.WithUnit("{execution}"),
	)
	if err != nil {
		return err
	}

	workflowFailed, err = meter.Int64Counter(
		"workflow_failed",
		metric.WithDescription("Count of failed workflows"),
		metric.WithUnit("{execution}"),
	)
	if err != nil {
		return err
	}

	workflowTimeout, err = meter.Int64Counter(
		"workflow_timeout",
		metric.WithDescription("Count of timed out workflows"),
		metric.WithUnit("{execution}"),
	)
	if err != nil {
		return err
	}

	workflowTerminate, err = meter.Int64Counter(
		"workflow_terminate",
		metric.WithDescription("Count of terminated workflows"),
		metric.WithUnit("{execution}"),
	)
	if err != nil {
		return err
	}

	workflowCancel, err = meter.Int64Counter(
		"workflow_cancel",
		metric.WithDescription("Count of canceled workflows"),
		metric.WithUnit("{execution}"),
	)
	if err != nil {
		return err
	}

	// Initialize service request metrics
	serviceRequests, err = meter.Int64Counter(
		"service_requests",
		metric.WithDescription("Count of service requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return err
	}

	serviceErrors, err = meter.Int64Counter(
		"service_errors",
		metric.WithDescription("Count of service errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	serviceErrorWithType, err = meter.Int64Counter(
		"service_error_with_type",
		metric.WithDescription("Count of service errors by type"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return err
	}

	restarts, err = meter.Int64Counter(
		"restarts",
		metric.WithDescription("Count of service restarts"),
		metric.WithUnit("{restart}"),
	)
	if err != nil {
		return err
	}

	// Initialize timeout metrics
	scheduleToStartTimeout, err = meter.Int64Counter(
		"schedule_to_start_timeout",
		metric.WithDescription("Schedule to start timeout metrics"),
		metric.WithUnit("{timeout}"),
	)
	if err != nil {
		return err
	}

	startToCloseTimeout, err = meter.Int64Counter(
		"start_to_close_timeout",
		metric.WithDescription("Start to close timeout metrics"),
		metric.WithUnit("{timeout}"),
	)
	if err != nil {
		return err
	}

	return nil
}

// Ensure metrics are initialized
func ensureMetricsInitialized() {
	once.Do(func() {
		log.Println("Initializing metrics counters...")
		meter := otel.GetMeterProvider().Meter("temporal-metrics")
		if err := InitializeMetrics(meter); err != nil {
			log.Printf("Failed to initialize metrics: %v", err)
			// If there's an error, just log it and continue
			// This way the application doesn't crash if metrics fail
			// In a production environment, you might want to handle this differently
		} else {
			log.Println("Metrics counters initialized successfully")
		}
	})
}

// Cleanup resets all metric counters
func Cleanup() {
	// In Go, metric providers are managed by the SDK, so we don't need to explicitly
	// null out counters like in Java. We can just reset if needed.
}

// --- Workflow Completion Metrics ---

// RecordSuccess records a successful workflow completion
func RecordSuccess(ctx context.Context, workflowType, workflowID, runID, namespace string) {
	ensureMetricsInitialized()
	workflowSuccess.Add(ctx, 1, 
		metric.WithAttributes(
			WorkflowTypeKey.String(workflowType),
			WorkflowIDKey.String(workflowID),
			RunIDKey.String(runID),
			NamespaceKey.String(namespace),
			OperationKey.String("CompletionStats"),
		),
	)
}

// RecordFailure records a workflow failure
func RecordFailure(ctx context.Context, workflowType, workflowID, runID, namespace string) {
	ensureMetricsInitialized()
	workflowFailed.Add(ctx, 1, 
		metric.WithAttributes(
			WorkflowTypeKey.String(workflowType),
			WorkflowIDKey.String(workflowID),
			RunIDKey.String(runID),
			NamespaceKey.String(namespace),
			OperationKey.String("CompletionStats"),
		),
	)
}

// RecordTimeout records a workflow timeout
func RecordTimeout(ctx context.Context, workflowType, workflowID, runID, namespace string) {
	ensureMetricsInitialized()
	workflowTimeout.Add(ctx, 1, 
		metric.WithAttributes(
			WorkflowTypeKey.String(workflowType),
			WorkflowIDKey.String(workflowID),
			RunIDKey.String(runID),
			NamespaceKey.String(namespace),
			OperationKey.String("CompletionStats"),
		),
	)
}

// RecordTermination records a workflow termination
func RecordTermination(ctx context.Context, workflowType, workflowID, runID, namespace string) {
	ensureMetricsInitialized()
	workflowTerminate.Add(ctx, 1, 
		metric.WithAttributes(
			WorkflowTypeKey.String(workflowType),
			WorkflowIDKey.String(workflowID),
			RunIDKey.String(runID),
			NamespaceKey.String(namespace),
			OperationKey.String("CompletionStats"),
		),
	)
}

// RecordCancellation records a workflow cancellation
func RecordCancellation(ctx context.Context, workflowType, workflowID, runID, namespace string) {
	ensureMetricsInitialized()
	workflowCancel.Add(ctx, 1, 
		metric.WithAttributes(
			WorkflowTypeKey.String(workflowType),
			WorkflowIDKey.String(workflowID),
			RunIDKey.String(runID),
			NamespaceKey.String(namespace),
			OperationKey.String("CompletionStats"),
		),
	)
}

// --- Activity Task Metrics ---

// RecordAddActivityTask records adding an activity task
func RecordAddActivityTask(ctx context.Context) {
	ensureMetricsInitialized()
	serviceRequests.Add(ctx, 1, metric.WithAttributes(OperationKey.String("AddActivityTask")))
}

// RecordRecordActivityTaskStarted records starting an activity task
func RecordRecordActivityTaskStarted(ctx context.Context) {
	ensureMetricsInitialized()
	serviceRequests.Add(ctx, 1, metric.WithAttributes(OperationKey.String("RecordActivityTaskStarted")))
}

// RecordResponseActivityCompleted records completing an activity task
func RecordResponseActivityCompleted(ctx context.Context) {
	ensureMetricsInitialized()
	serviceRequests.Add(ctx, 1, metric.WithAttributes(OperationKey.String("RespondActivityTaskCompleted")))
}

// RecordRespondActivityTaskFailed records failing an activity task
func RecordRespondActivityTaskFailed(ctx context.Context) {
	ensureMetricsInitialized()
	serviceRequests.Add(ctx, 1, metric.WithAttributes(OperationKey.String("RespondActivityTaskFailed")))
}

// RecordRespondActivityTaskCanceled records canceling an activity task
func RecordRespondActivityTaskCanceled(ctx context.Context) {
	ensureMetricsInitialized()
	serviceRequests.Add(ctx, 1, metric.WithAttributes(OperationKey.String("RespondActivityTaskCanceled")))
}

// --- Workflow Task Metrics ---

// RecordAddWorkflowTask records adding a workflow task
func RecordAddWorkflowTask(ctx context.Context) {
	ensureMetricsInitialized()
	serviceRequests.Add(ctx, 1, metric.WithAttributes(OperationKey.String("AddWorkflowTask")))
}

// RecordRecordWorkflowTaskStarted records starting a workflow task
func RecordRecordWorkflowTaskStarted(ctx context.Context) {
	ensureMetricsInitialized()
	serviceRequests.Add(ctx, 1, metric.WithAttributes(OperationKey.String("RecordWorkflowTaskStarted")))
}

// RecordRespondWorkflowTaskCompleted records completing a workflow task
func RecordRespondWorkflowTaskCompleted(ctx context.Context) {
	ensureMetricsInitialized()
	serviceRequests.Add(ctx, 1, metric.WithAttributes(OperationKey.String("RespondWorkflowTaskCompleted")))
}

// RecordRespondWorkflowTaskFailed records failing a workflow task
func RecordRespondWorkflowTaskFailed(ctx context.Context) {
	ensureMetricsInitialized()
	serviceRequests.Add(ctx, 1, metric.WithAttributes(OperationKey.String("RespondWorkflowTaskFailed")))
}

// RecordTimerActiveTaskWorkflowTimeout records a workflow timeout
func RecordTimerActiveTaskWorkflowTimeout(ctx context.Context) {
	ensureMetricsInitialized()
	serviceRequests.Add(ctx, 1, metric.WithAttributes(OperationKey.String("TimerActiveTaskWorkflowTimeout")))
}

// --- Error Metrics ---

// RecordAddActivityTaskError records an error adding an activity task
func RecordAddActivityTaskError(ctx context.Context) {
	ensureMetricsInitialized()
	serviceErrors.Add(ctx, 1, metric.WithAttributes(OperationKey.String("AddActivityTask")))
}

// RecordRecordActivityTaskStartedError records an error starting an activity task
func RecordRecordActivityTaskStartedError(ctx context.Context) {
	ensureMetricsInitialized()
	serviceErrors.Add(ctx, 1, metric.WithAttributes(OperationKey.String("RecordActivityTaskStarted")))
}

// RecordRespondActivityTaskCompletedError records an error completing an activity task
func RecordRespondActivityTaskCompletedError(ctx context.Context) {
	ensureMetricsInitialized()
	serviceErrors.Add(ctx, 1, metric.WithAttributes(OperationKey.String("RespondActivityTaskCompleted")))
}

// RecordRespondActivityTaskFailedError records an error failing an activity task
func RecordRespondActivityTaskFailedError(ctx context.Context) {
	ensureMetricsInitialized()
	serviceErrors.Add(ctx, 1, metric.WithAttributes(OperationKey.String("RespondActivityTaskFailed")))
}

// RecordRespondActivityTaskCanceledError records an error canceling an activity task
func RecordRespondActivityTaskCanceledError(ctx context.Context) {
	ensureMetricsInitialized()
	serviceErrors.Add(ctx, 1, metric.WithAttributes(OperationKey.String("RespondActivityTaskCanceled")))
}

// RecordAddWorkflowTaskError records an error adding a workflow task
func RecordAddWorkflowTaskError(ctx context.Context) {
	ensureMetricsInitialized()
	serviceErrors.Add(ctx, 1, metric.WithAttributes(OperationKey.String("AddWorkflowTask")))
}

// RecordRecordWorkflowTaskStartedError records an error starting a workflow task
func RecordRecordWorkflowTaskStartedError(ctx context.Context) {
	ensureMetricsInitialized()
	serviceErrors.Add(ctx, 1, metric.WithAttributes(OperationKey.String("RecordWorkflowTaskStarted")))
}

// RecordRespondWorkflowTaskCompletedError records an error completing a workflow task
func RecordRespondWorkflowTaskCompletedError(ctx context.Context) {
	ensureMetricsInitialized()
	serviceErrors.Add(ctx, 1, metric.WithAttributes(OperationKey.String("RespondWorkflowTaskCompleted")))
}

// RecordRespondWorkflowTaskFailedError records an error failing a workflow task
func RecordRespondWorkflowTaskFailedError(ctx context.Context) {
	ensureMetricsInitialized()
	serviceErrors.Add(ctx, 1, metric.WithAttributes(OperationKey.String("RespondWorkflowTaskFailed")))
}

// --- Error Type Metrics ---

// RecordValidationError records a validation error
func RecordValidationError(ctx context.Context) {
	ensureMetricsInitialized()
	serviceErrorWithType.Add(ctx, 1, metric.WithAttributes(ErrorTypeKey.String("validation")))
}

// RecordTimeoutError records a timeout error
func RecordTimeoutError(ctx context.Context) {
	ensureMetricsInitialized()
	serviceErrorWithType.Add(ctx, 1, metric.WithAttributes(ErrorTypeKey.String("timeout")))
}

// RecordBusinessRuleError records a business rule error
func RecordBusinessRuleError(ctx context.Context) {
	ensureMetricsInitialized()
	serviceErrorWithType.Add(ctx, 1, metric.WithAttributes(ErrorTypeKey.String("business_rule")))
}

// RecordSystemError records a system error
func RecordSystemError(ctx context.Context) {
	ensureMetricsInitialized()
	serviceErrorWithType.Add(ctx, 1, metric.WithAttributes(ErrorTypeKey.String("system")))
}

// --- Timeout Metrics ---

// RecordScheduleToStartWorkflowTimeout records a schedule to start workflow timeout
func RecordScheduleToStartWorkflowTimeout(ctx context.Context) {
	ensureMetricsInitialized()
	scheduleToStartTimeout.Add(ctx, 1, metric.WithAttributes(OperationKey.String("TimerActiveTaskWorkflowTimeout")))
}

// RecordStartToCloseWorkflowTimeout records a start to close workflow timeout
func RecordStartToCloseWorkflowTimeout(ctx context.Context) {
	ensureMetricsInitialized()
	startToCloseTimeout.Add(ctx, 1, metric.WithAttributes(OperationKey.String("TimerActiveTaskWorkflowTimeout")))
}

// --- Service Metrics ---

// RecordServiceRestart records a service restart
func RecordServiceRestart(ctx context.Context, serviceType string) {
	ensureMetricsInitialized()
	restarts.Add(ctx, 1, metric.WithAttributes(ServiceTypeKey.String(serviceType)))
} 