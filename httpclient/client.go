package httpclient

import (
	"context"

	"github.com/go-resty/resty/v2"
)

var client = NewClient()

func Client() *resty.Client {
	return client
}

func NewRequest(ctx context.Context) *resty.Request {
	return client.NewRequest().SetContext(ctx)
}

func ResetClientWithOptions(opts ...Option) {
	client = NewClient(opts...)
}

// NewClient 多客户端时使用
func NewClient(opts ...Option) *resty.Client {
	options := mergeOptions(opts...)
	c := resty.New()
	if options.Host != "" {
		c.SetBaseURL(options.Host)
	}
	if options.Timeout != 0 {
		c.SetTimeout(options.Timeout)
	}
	if options.Otel {
		c.OnBeforeRequest(OtelMiddleware())
	}
	return c
}
