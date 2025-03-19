package httpclient

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
)

func TestClient(t *testing.T) {
	ctx := context.Background()

	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(ctx)

	tracer := otel.Tracer("olympus/httpclient")
	ctx, span := tracer.Start(ctx, "test")
	defer span.End()

	t.Log("traceId", span.SpanContext().TraceID())

	response, err := NewRequest(ctx).Get("http://127.0.0.1:8000/hello/ping")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", response.Body())
}
