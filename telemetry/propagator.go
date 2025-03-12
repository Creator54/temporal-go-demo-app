package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

const (
	tracingHeaderKey = "traceparent"
)

type TracingContextPropagator struct {
	propagator propagation.TextMapPropagator
	dc         converter.DataConverter
}

func NewTracingContextPropagator() *TracingContextPropagator {
	return &TracingContextPropagator{
		propagator: otel.GetTextMapPropagator(),
		dc:         converter.GetDefaultDataConverter(),
	}
}

func (p *TracingContextPropagator) Inject(ctx context.Context, carrier workflow.HeaderWriter) error {
	textCarrier := propagation.MapCarrier{}
	p.propagator.Inject(ctx, textCarrier)

	if traceparent, ok := textCarrier[tracingHeaderKey]; ok {
		payload, err := p.dc.ToPayload(traceparent)
		if err != nil {
			return err
		}
		carrier.Set(tracingHeaderKey, payload)
	}
	return nil
}

func (p *TracingContextPropagator) Extract(ctx context.Context, carrier workflow.HeaderReader) (context.Context, error) {
	textCarrier := propagation.MapCarrier{}

	if err := carrier.ForEachKey(func(key string, value *commonpb.Payload) error {
		if key == tracingHeaderKey && value != nil {
			var v string
			if err := p.dc.FromPayload(value, &v); err != nil {
				return err
			}
			textCarrier[key] = v
		}
		return nil
	}); err != nil {
		return ctx, err
	}

	return p.propagator.Extract(ctx, textCarrier), nil
}

func (p *TracingContextPropagator) InjectFromWorkflow(ctx workflow.Context, carrier workflow.HeaderWriter) error {
	// No-op for now as we can't propagate trace context from workflow
	return nil
}

func (p *TracingContextPropagator) ExtractToWorkflow(ctx workflow.Context, carrier workflow.HeaderReader) (workflow.Context, error) {
	// No-op for now as we can't propagate trace context to workflow
	return ctx, nil
}
