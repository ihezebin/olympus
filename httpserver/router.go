package httpserver

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ihezebin/openapi"
)

type Router interface {
	Group(...string) Router
	Use(...gin.HandlerFunc) Router
	Any(string, handler, ...OpenAPIOption)
	GET(string, handler, ...OpenAPIOption)
	POST(string, handler, ...OpenAPIOption)
	DELETE(string, handler, ...OpenAPIOption)
	PATCH(string, handler, ...OpenAPIOption)
	PUT(string, handler, ...OpenAPIOption)
	OPTIONS(string, handler, ...OpenAPIOption)
	HEAD(string, handler, ...OpenAPIOption)
}

type OpenAPIOption func(*openapi.Route)

func WithDescription(description string) OpenAPIOption {
	return func(route *openapi.Route) {
		route.HasDescription(description)
	}
}

func WithSummary(summary string) OpenAPIOption {
	return func(route *openapi.Route) {
		route.HasSummary(summary)
	}
}

func WithDeprecated() OpenAPIOption {
	return func(route *openapi.Route) {
		route.HasDeprecated(true)
	}
}

func WithResponseHeader(name string, param openapi.HeaderParam) OpenAPIOption {
	return func(route *openapi.Route) {
		route.HasResponseHeader(http.StatusOK, name, param)
	}
}

func mergeOpenAPIOptions(route *openapi.Route, options ...OpenAPIOption) *openapi.Route {
	for _, option := range options {
		option(route)
	}
	return route
}

type openapiRouter struct {
	prefix  string
	router  gin.IRouter
	openapi *openapi.API
}

func (r *openapiRouter) Group(prefixes ...string) Router {
	prefix := strings.Join(prefixes, "/")

	return &openapiRouter{
		prefix:  strings.ReplaceAll(strings.Join([]string{r.prefix, prefix}, "/"), "//", "/"),
		router:  r.router.Group(prefix),
		openapi: r.openapi,
	}
}

func (r *openapiRouter) mergePath(paths ...string) string {
	path := strings.ReplaceAll(strings.Join(paths, "/"), "//", "/")
	if strings.Contains(path, "//") {
		return r.mergePath(paths...)
	}

	return path
}

func (r *openapiRouter) Use(middleware ...gin.HandlerFunc) Router {
	r.router.Use(middleware...)
	return r
}

func (r *openapiRouter) handle(method string, path string, h handler, options ...OpenAPIOption) {
	requestBody, responseBody, query, params, requestHeader, responseHeader, handlerFunc := h()

	// register gin route
	r.router.Handle(method, path, handlerFunc)

	// handle path
	path = r.mergePath(r.prefix, path)

	// handle route
	route := r.openapi.Route(method, path)
	route = mergeOpenAPIOptions(route, options...)
	operationID := strings.ReplaceAll(path, "/", "_")
	operationID = strings.TrimLeft(operationID, "_")
	operationID = strings.TrimRight(operationID, "_")

	route.HasOperationID(operationID)

	if requestBody != nil {
		route.HasRequestModel(*requestBody)
	}

	route.HasResponseModel(http.StatusInternalServerError, openapi.ModelOf[Body[EmptyType]]())
	if responseBody != nil {
		route.HasResponseModel(http.StatusOK, *responseBody)
	}

	if len(query) > 0 {
		for k, v := range query {
			route.HasQueryParameter(k, v)
		}
	}

	// path 里面有，但是由于 uri tag 添加的 param 要删除
	realExistParam := make(map[string]bool)
	// path 里面包含 :id 格式的，添加 param
	if strings.Contains(path, ":") {
		for _, param := range strings.Split(path, "/") {
			if strings.Contains(param, ":") {
				realExistParam[param] = true
				param = strings.TrimLeft(param, ":")
				pathParam, ok := params[param]
				if !ok {
					pathParam = openapi.PathParam{
						Description: "",
						Type:        openapi.PrimitiveTypeString,
					}
				}
				params[param] = pathParam
			}
		}
	}

	if len(params) > 0 {
		for k, v := range params {
			if realExistParam[k] {
				route.HasPathParameter(k, v)
			}
		}
	}

	if len(requestHeader) > 0 {
		for k, v := range requestHeader {
			route.HasHeaderParameter(k, v)
		}
	}

	if len(responseHeader) > 0 {
		for k, v := range responseHeader {
			route.HasResponseHeader(http.StatusOK, k, v)
		}
	}
}

func (r *openapiRouter) Any(path string, h handler, options ...OpenAPIOption) {
	anyMethods := []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
		http.MethodTrace,
	}

	for _, method := range anyMethods {
		r.handle(method, path, h, options...)
	}
}

func (r *openapiRouter) GET(path string, h handler, options ...OpenAPIOption) {
	r.handle(http.MethodGet, path, h, options...)
}

func (r *openapiRouter) POST(path string, h handler, options ...OpenAPIOption) {
	r.handle(http.MethodPost, path, h, options...)
}

func (r *openapiRouter) DELETE(path string, h handler, options ...OpenAPIOption) {
	r.handle(http.MethodDelete, path, h, options...)
}

func (r *openapiRouter) PATCH(path string, h handler, options ...OpenAPIOption) {
	r.handle(http.MethodPatch, path, h, options...)
}

func (r *openapiRouter) PUT(path string, h handler, options ...OpenAPIOption) {
	r.handle(http.MethodPut, path, h, options...)
}

func (r *openapiRouter) OPTIONS(path string, h handler, options ...OpenAPIOption) {
	r.handle(http.MethodOptions, path, h, options...)
}

func (r *openapiRouter) HEAD(path string, h handler, options ...OpenAPIOption) {
	r.handle(http.MethodHead, path, h, options...)
}
