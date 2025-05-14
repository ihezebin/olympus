package internal

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func OtelExtractTrace(service string) gin.HandlerFunc {

	return func(c *gin.Context) {
		// 如果没有 traceId 则生成一个
		ctx := c.Request.Context()
		traceId := c.Request.Header.Get("traceparent")
		if traceId == "" {
			ctx, span := otel.GetTracerProvider().Tracer(service).Start(ctx, c.FullPath())
			defer span.End()
			c.Request = c.Request.WithContext(ctx)
		}
		otelgin.Middleware(service)(c)
	}
}

func OtelInjectTrace() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		propagator := otel.GetTextMapPropagator()
		propagator.Inject(ctx, propagation.HeaderCarrier(c.Writer.Header()))
		c.Next()
	}
}
