package goinsight

import (
	"context"
	"fmt"
)

func (c *Client) StartTrace(ctx context.Context, operation string) (context.Context, *TraceContext, error) {
	trace := Trace{
		ServiceName: c.serviceName,
	}

	resp, err := c.sendTrace(trace)
	if err != nil {
		return ctx, nil, err
	}

	traceCtx := &TraceContext{
		TraceID: resp["id"].(string),
	}

	// Start root span
	span := Span{
		TraceID:   traceCtx.TraceID,
		Service:   c.serviceName,
		Operation: operation,
	}

	spanResp, err := c.sendSpan(span)
	if err != nil {
		return ctx, traceCtx, err
	}

	traceCtx.SpanID = spanResp["id"].(string)

	newCtx := context.WithValue(ctx, "go-insight-trace", traceCtx)

	return newCtx, traceCtx, nil
}

func (c *Client) StartSpan(ctx context.Context, operation string) (context.Context, error) {
	traceCtx := GetTraceFromContext(ctx)
	if traceCtx == nil {
		return ctx, fmt.Errorf("no trace context found")
	}

	span := Span{
		TraceID:   traceCtx.TraceID,
		ParentID:  traceCtx.SpanID,
		Service:   c.serviceName,
		Operation: operation,
	}

	resp, err := c.sendSpan(span)
	if err != nil {
		return ctx, err
	}

	newTraceCtx := &TraceContext{
		TraceID: traceCtx.TraceID,
		SpanID:  resp["id"].(string),
	}

	newCtx := context.WithValue(ctx, "go-insight-trace", newTraceCtx)
	return newCtx, nil
}

func (c *Client) FinishSpan(ctx context.Context) error {
	traceCtx := GetTraceFromContext(ctx)
	if traceCtx == nil {
		return fmt.Errorf("no trace context found")
	}

	return c.endSpan(traceCtx.SpanID)
}

func (c *Client) FinishTrace(ctx context.Context) error {
	traceCtx := GetTraceFromContext(ctx)
	if traceCtx == nil {
		return fmt.Errorf("no trace context found")
	}

	return c.endTrace(traceCtx.TraceID)
}

func GetTraceFromContext(ctx context.Context) *TraceContext {
	if traceCtx, ok := ctx.Value("go-insight-trace").(*TraceContext); ok {
		return traceCtx
	}
	return nil
}
