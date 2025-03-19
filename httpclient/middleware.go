package httpclient

import (
	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func OtelMiddleware() resty.RequestMiddleware {
	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTextMapPropagator(propagator)

	return func(client *resty.Client, request *resty.Request) error {
		ctx := request.Context()

		propagator.Inject(ctx, propagation.HeaderCarrier(request.Header))
		return nil
	}
}
