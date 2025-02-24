package httpserver

import (
	"context"
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ihezebin/openapi"

	"github.com/ihezebin/soup/httpserver/middleware"
)

var ctx = context.Background()

func TestServer(t *testing.T) {
	server := NewServer(
		WithPort(8000),
		WithServiceName("test_server"),
		WithMiddlewares(middleware.LoggingRequestWithoutHeader(), middleware.LoggingResponse()),
	)

	server.RegisterRoutes(&HelloRouter{})

	err := server.RegisterOpenAPIUI("/stoplight", StoplightUI)
	if err != nil {
		t.Fatal(err)
	}
	err = server.RegisterOpenAPIUI("/swagger", SwaggerUI)
	if err != nil {
		t.Fatal(err)
	}
	err = server.RegisterOpenAPIUI("/redoc", RedocUI)
	if err != nil {
		t.Fatal(err)
	}
	err = server.RegisterOpenAPIUI("/rapidoc", RapidocUI)
	if err != nil {
		t.Fatal(err)
	}

	if err := server.RunWithNotifySignal(ctx); err != nil {
		t.Fatal(err)
	}
}

type HelloRouter struct {
}

func (h *HelloRouter) RegisterRoutes(router Router) {
	group := router.Group("/hello")
	group.POST("/world", NewHandler(h.Hello))
	group.GET("/ping", NewHandler(h.Ping), WithOpenAPIDeprecated(), WithOpenAPIResponseHeader("Token", openapi.HeaderParam{
		Description: "认证 JWT",
	}))
}

type HelloReq struct {
	Content string `json:"content" form:"content"`
	Id      string `json:"id" form:"id"`
}

// HelloResp 测试的 hello 响应数据
type HelloResp struct {
	Message string `json:"message"`
}

func (h *HelloRouter) Hello(c *gin.Context, req *HelloReq) (resp *HelloResp, err error) {
	return &HelloResp{Message: req.Content}, nil
}

func (h *HelloRouter) Ping(c *gin.Context, req map[string]interface{}) (resp string, err error) {
	fmt.Println(req)
	return "pong", nil
}
