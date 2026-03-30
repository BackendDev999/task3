package observability

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"

	"go.opentelemetry.io/otel/trace"
)

type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	With(args ...any) Logger
}

type SlogLogger struct {
	inner *slog.Logger
}

func NewSlogLogger(inner *slog.Logger) SlogLogger {
	return SlogLogger{inner: inner}
}

func (l SlogLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.withTrace(ctx).InfoContext(ctx, msg, args...)
}

func (l SlogLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.withTrace(ctx).ErrorContext(ctx, msg, args...)
}

func (l SlogLogger) With(args ...any) Logger {
	return SlogLogger{inner: l.inner.With(args...)}
}

func (l SlogLogger) withTrace(ctx context.Context) *slog.Logger {
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return l.inner
	}

	return l.inner.With(
		slog.String("trace_id", spanCtx.TraceID().String()),
		slog.String("span_id", spanCtx.SpanID().String()),
	)
}

type LoggingDecorator[Req any, Res any] struct {
	inner      Interactor[Req, Res]
	log        Logger
	name       string
	redactReq  func(Req) any
	redactResp func(Res) any
}

func NewLoggingDecorator[Req any, Res any](
	inner Interactor[Req, Res],
	log Logger,
	name string,
	redactReq func(Req) any,
	redactResp func(Res) any,
) *LoggingDecorator[Req, Res] {
	return &LoggingDecorator[Req, Res]{
		inner:      inner,
		log:        log,
		name:       name,
		redactReq:  redactReq,
		redactResp: redactResp,
	}
}

func (d *LoggingDecorator[Req, Res]) Execute(ctx context.Context, req Req) (Res, error) {
	reqPayload := any(req)
	if d.redactReq != nil {
		reqPayload = d.redactReq(req)
	}

	d.log.InfoContext(ctx, "usecase.started",
		"usecase", d.name,
		"request", reqPayload,
	)

	res, err := d.inner.Execute(ctx, req)
	if err != nil {
		d.log.ErrorContext(ctx, "usecase.failed",
			"usecase", d.name,
			"request", reqPayload,
			"error", err.Error(),
			"stacktrace", string(debug.Stack()),
		)
		var zero Res
		return zero, err
	}

	respPayload := any(res)
	if d.redactResp != nil {
		respPayload = d.redactResp(res)
	}

	d.log.InfoContext(ctx, "usecase.completed",
		"usecase", d.name,
		"response", respPayload,
	)

	return res, nil
}

type CreateOrderRequest struct {
	CustomerID    string
	CustomerEmail string
	CardToken     string
	AmountCents   int64
}

func RedactCreateOrderRequest(req CreateOrderRequest) map[string]any {
	return map[string]any{
		"customer_id":    req.CustomerID,
		"customer_email": redactEmail(req.CustomerEmail),
		"card_token":     "[REDACTED]",
		"amount_cents":   req.AmountCents,
	}
}

func redactEmail(email string) string {
	for i := range email {
		if email[i] == '@' {
			if i <= 1 {
				return "***" + email[i:]
			}
			return email[:1] + "***" + email[i:]
		}
	}
	return "[REDACTED]"
}

func WrapError(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}
