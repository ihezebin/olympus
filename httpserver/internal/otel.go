package internal

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func OtelExtractTrace(service string) gin.HandlerFunc {
	return otelgin.Middleware(service)
}

func OtelInjectTrace() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		propagator := otel.GetTextMapPropagator()
		propagator.Inject(ctx, propagation.HeaderCarrier(c.Writer.Header()))
		c.Next()
	}
}
