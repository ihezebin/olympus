package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/ihezebin/openapi"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ihezebin/olympus/httpserver/internal"
	"github.com/ihezebin/olympus/logger"
)

type server struct {
	*http.Server
	options   *ServerOptions
	engine    *gin.Engine
	openapi   *openapi.API
	shutdowns []ShutdownFunc
}

type ShutdownFunc func(context.Context) error

func NewServer(opts ...ServerOption) *server {
	serverOptions := mergeServerOptions(opts...)

	// 隐藏路由日志
	if serverOptions.HiddenRoutesLog {
		gin.DefaultWriter = io.Discard
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	// 中间件
	engine.Use(serverOptions.Middlewares...)

	shutdowns := make([]ShutdownFunc, 0)
	if serverOptions.Otel {
		shutdowns = append(shutdowns, serverOptions.OtelInit()...)
		engine.Use(internal.OtelExtractTrace(serverOptions.ServiceName))
		engine.Use(internal.OtelInjectTrace())
	}

	// 设置服务名称
	serviceName := "olympus httpserver"
	if serverOptions.ServiceName != "" {
		serviceName = serverOptions.ServiceName
	}

	engine.Use(func(c *gin.Context) {
		c.Set(ServiceNameKey, serviceName)
		c.Next()
	})
	//默认的健康检查接口
	engine.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	if serverOptions.Metrics {
		engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	if serverOptions.Pprof {
		pprof.Register(engine)
	}

	openapiOpts := make([]openapi.APIOpts, 0)
	if serverOptions.OpenAPInfo != nil {
		openapiOpts = append(openapiOpts, openapi.WithInfo(*serverOptions.OpenAPInfo))
	}
	if len(serverOptions.OpenAPIServers) > 0 {
		openapiOpts = append(openapiOpts, openapi.WithServer(serverOptions.OpenAPIServers...))
	}
	openApi := openapi.NewAPI(serviceName, openapiOpts...)
	openApi.RegisterModel(openapi.ModelOf[Body[any]]())

	kernel := &http.Server{
		Handler: engine,
		Addr:    fmt.Sprintf(":%d", serverOptions.Port),
	}

	shutdowns = append(shutdowns, kernel.Shutdown)

	server := &server{
		Server:    kernel,
		options:   serverOptions,
		engine:    engine,
		openapi:   openApi,
		shutdowns: shutdowns,
	}

	return server
}

func (s *server) Name() string {
	return fmt.Sprintf("httpserver[%s]", s.options.ServiceName)
}

func (s *server) Engine() *gin.Engine {
	return s.engine
}

func (s *server) OpenAPI() *openapi.API {
	return s.openapi
}

type RegisterRoutes interface {
	RegisterRoutes(router Router)
}

func (s *server) RegisterRoutes(routers ...RegisterRoutes) {
	for _, router := range routers {
		router.RegisterRoutes(&openapiRouter{
			router:  s.engine,
			openapi: s.openapi,
			prefix:  "",
		})
	}
}

func (s *server) RegisterOpenAPIUI(path string, ui OpenAPIUIBuilder) error {
	if path == "" {
		path = "/openapi"
	}
	spec, err := s.OpenAPI().Spec()
	if err != nil {
		return errors.Wrap(err, "get openapi spec err")
	}

	specStr, err := json.Marshal(spec)
	if err != nil {
		return errors.Wrap(err, "marshal openapi spec err")
	}
	s.engine.GET(path, func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(ui.HTML(string(specStr), s.options.ServiceName)))
	})

	return nil
}

func (s *server) Run(ctx context.Context) error {
	run := func(options *ServerOptions) error {
		logger.Infof(ctx, "http server is starting in port: %d", options.Port)
		if err := s.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.WithError(err).Error(ctx, "http server ListenAndServe err")
				return err
			}
			logger.Info(ctx, "http server closed")

			return nil
		}
		return nil
	}

	if s.options.Daemon {
		go run(s.options)
	} else {
		return run(s.options)
	}

	return nil
}

func (s *server) RunWithNotifySignal(ctx context.Context) error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	go func() {
		<-signalChan
		s.Close(ctx)
	}()

	return s.Run(ctx)
}

func (s *server) Close(ctx context.Context) error {
	for _, shutdown := range s.shutdowns {
		if err := shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}
