package httpclient

import "time"

type Options struct {
	Otel    bool
	Host    string
	Timeout time.Duration
	Debug   bool
}

type Option func(*Options)

func mergeOptions(opts ...Option) *Options {
	options := &Options{
		Otel:    true,
		Host:    "",
		Timeout: 10 * time.Second,
		Debug:   false,
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

func WithOtel(enabled bool) Option {
	return func(o *Options) {
		o.Otel = enabled
	}
}

func WithHost(host string) Option {
	return func(o *Options) {
		o.Host = host
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

func WithDebug(debug bool) Option {
	return func(o *Options) {
		o.Debug = debug
	}
}
