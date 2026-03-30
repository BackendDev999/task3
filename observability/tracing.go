package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Interactor[Req any, Res any] interface {
	Execute(ctx context.Context, req Req) (Res, error)
}

type TracingDecorator[Req any, Res any] struct {
	inner    Interactor[Req, Res]
	tracer   trace.Tracer
	name     string
	attrFn   func(Req) []attribute.KeyValue
	resultFn func(Res) []attribute.KeyValue
}

func NewTracingDecorator[Req any, Res any](
	inner Interactor[Req, Res],
	tracer trace.Tracer,
	name string,
	attrFn func(Req) []attribute.KeyValue,
	resultFn func(Res) []attribute.KeyValue,
) *TracingDecorator[Req, Res] {
	return &TracingDecorator[Req, Res]{
		inner:    inner,
		tracer:   tracer,
		name:     name,
		attrFn:   attrFn,
		resultFn: resultFn,
	}
}

func (d *TracingDecorator[Req, Res]) Execute(ctx context.Context, req Req) (Res, error) {
	ctx, span := d.tracer.Start(ctx, d.name)
	defer span.End()

	if d.attrFn != nil {
		span.SetAttributes(d.attrFn(req)...)
	}

	res, err := d.inner.Execute(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		var zero Res
		return zero, err
	}

	if d.resultFn != nil {
		span.SetAttributes(d.resultFn(res)...)
	}

	return res, nil
}

type DomainTracer interface {
	OnOrderCompleting(orderID string)
	OnOrderCompleted(orderID string)
	OnOrderFailed(orderID string, err error)
}

type SpanDomainTracer struct {
	span trace.Span
}

func NewSpanDomainTracer(ctx context.Context) DomainTracer {
	return SpanDomainTracer{span: trace.SpanFromContext(ctx)}
}

func (t SpanDomainTracer) OnOrderCompleting(orderID string) {
	t.span.AddEvent("domain.order.completing", trace.WithAttributes(
		attribute.String("order.id", orderID),
	))
}

func (t SpanDomainTracer) OnOrderCompleted(orderID string) {
	t.span.AddEvent("domain.order.completed", trace.WithAttributes(
		attribute.String("order.id", orderID),
	))
}

func (t SpanDomainTracer) OnOrderFailed(orderID string, err error) {
	t.span.RecordError(fmt.Errorf("order %s: %w", orderID, err))
}
