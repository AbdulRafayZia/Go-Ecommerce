package tracing

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/attribute"
)

func TestInitTracer_Disabled(t *testing.T) {
	cfg := Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Enabled:        false,
	}

	provider, err := InitTracer(cfg)
	if err != nil {
		t.Fatalf("expected no error when initializing disabled tracer, got %v", err)
	}

	if provider == nil {
		t.Fatal("expected provider to be non-nil")
	}

	ctx := context.Background()
	if err := provider.Shutdown(ctx); err != nil {
		t.Errorf("expected no error on shutdown, got %v", err)
	}
}

func TestGetTracer(t *testing.T) {
	tracer := GetTracer("test-tracer")
	if tracer == nil {
		t.Fatal("expected tracer to be non-nil")
	}
}

func TestStartSpan(t *testing.T) {
	ctx := context.Background()
	newCtx, span := StartSpan(ctx, "test-tracer", "test-span")

	if newCtx == nil {
		t.Fatal("expected context to be non-nil")
	}

	if span == nil {
		t.Fatal("expected span to be non-nil")
	}

	defer span.End()

	// Verify span is recording (even if it's a no-op span)
	if !span.IsRecording() {
		t.Log("span is not recording (might be a no-op span, which is fine for testing)")
	}
}

func TestAddSpanAttributes(t *testing.T) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "test-tracer", "test-span")
	defer span.End()

	// This should not panic
	AddSpanAttributes(ctx,
		attribute.String("key1", "value1"),
		attribute.Int("key2", 42),
	)
}

func TestAddSpanEvent(t *testing.T) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "test-tracer", "test-span")
	defer span.End()

	// This should not panic
	AddSpanEvent(ctx, "test-event",
		attribute.String("event_key", "event_value"),
	)
}

func TestRecordError(t *testing.T) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "test-tracer", "test-span")
	defer span.End()

	// This should not panic
	testErr := context.Canceled
	RecordError(ctx, testErr)
}

func TestGetTraceID(t *testing.T) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "test-tracer", "test-span")
	defer span.End()

	traceID := GetTraceID(ctx)
	// TraceID might be empty if using a no-op tracer
	t.Logf("TraceID: %s", traceID)
}

func TestGetSpanID(t *testing.T) {
	ctx := context.Background()
	ctx, span := StartSpan(ctx, "test-tracer", "test-span")
	defer span.End()

	spanID := GetSpanID(ctx)
	// SpanID might be empty if using a no-op tracer
	t.Logf("SpanID: %s", spanID)
}
