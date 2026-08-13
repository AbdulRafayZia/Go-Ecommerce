package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	httpTracerName = "gocommerce/http"
)

// HTTPTracing adds distributed tracing to HTTP requests
func HTTPTracing() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tracer := otel.Tracer(httpTracerName)

			// Extract trace context from headers
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Start a new span
			ctx, span := tracer.Start(ctx, r.Method+" "+r.URL.Path,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.scheme", r.URL.Scheme),
					attribute.String("http.host", r.Host),
					attribute.String("http.target", r.URL.Path),
					attribute.String("http.user_agent", r.UserAgent()),
					attribute.String("http.remote_addr", r.RemoteAddr),
				),
			)
			defer span.End()

			// Wrap response writer to capture status code
			wrapped := newResponseWriter(w)

			// Inject trace context into response headers
			otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(w.Header()))

			// Process request
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Add status code to span
			span.SetAttributes(attribute.Int("http.status_code", wrapped.statusCode))

			// Set span status based on HTTP status code
			if wrapped.statusCode >= 500 {
				span.SetStatus(codes.Error, http.StatusText(wrapped.statusCode))
			} else if wrapped.statusCode >= 400 {
				span.SetStatus(codes.Error, http.StatusText(wrapped.statusCode))
			} else {
				span.SetStatus(codes.Ok, "")
			}
		})
	}
}
