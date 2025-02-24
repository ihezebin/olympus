package httpserver

import "github.com/gin-gonic/gin"

type ServerOptions struct {
	Port            uint              `json:"port" yaml:"port" toml:"port"`
	Daemon          bool              `json:"daemon" yaml:"daemon" toml:"daemon"`
	Middlewares     []gin.HandlerFunc `json:"middlewares" yaml:"middlewares" toml:"middlewares"`
	ServiceName     string            `json:"service_name" yaml:"service_name" toml:"service_name"`
	HiddenRoutesLog bool              `json:"hidden_routes_log" yaml:"hidden_routes_log" toml:"hidden_routes_log"`
	Metrics         bool              `json:"metrics" yaml:"metrics" toml:"metrics"`
	Pprof           bool              `json:"pprof" yaml:"pprof" toml:"pprof"`
}

type ServerOption func(*ServerOptions)

func mergeServerOptions(opts ...ServerOption) *ServerOptions {
	opt := &ServerOptions{
		Port: 8080,
	}
	for _, o := range opts {
		o(opt)
	}
	return opt
}

func WithPort(port uint) ServerOption {
	return func(o *ServerOptions) {
		o.Port = port
	}
}

func WithDaemon(daemon bool) ServerOption {
	return func(o *ServerOptions) {
		o.Daemon = daemon
	}
}

func WithMiddlewares(middlewares ...gin.HandlerFunc) ServerOption {
	return func(o *ServerOptions) {
		o.Middlewares = middlewares
	}
}

func WithMetrics() ServerOption {
	return func(o *ServerOptions) {
		o.Metrics = true
	}
}

func WithPprof() ServerOption {
	return func(o *ServerOptions) {
		o.Pprof = true
	}
}

const (
	ServiceNameKey = "service"
)

func WithServiceName(name string) ServerOption {
	return func(o *ServerOptions) {
		o.ServiceName = name
	}
}

func WithHiddenRoutesLog() ServerOption {
	return func(o *ServerOptions) {
		o.HiddenRoutesLog = true
	}
}
