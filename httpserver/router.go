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
	Any(string, handlerGenerator, ...OpenAPIOption)
	AnyWithOptions(string, handlerGenerator, ...RouterOption)
	GET(string, handlerGenerator, ...OpenAPIOption)
	GetWithOptions(string, handlerGenerator, ...RouterOption)
	POST(string, handlerGenerator, ...OpenAPIOption)
	PostWithOptions(string, handlerGenerator, ...RouterOption)
	DELETE(string, handlerGenerator, ...OpenAPIOption)
	DeleteWithOptions(string, handlerGenerator, ...RouterOption)
	PATCH(string, handlerGenerator, ...OpenAPIOption)
	PatchWithOptions(string, handlerGenerator, ...RouterOption)
	PUT(string, handlerGenerator, ...OpenAPIOption)
	PutWithOptions(string, handlerGenerator, ...RouterOption)
	OPTIONS(string, handlerGenerator, ...OpenAPIOption)
	OptionsWithOptions(string, handlerGenerator, ...RouterOption)
	HEAD(string, handlerGenerator, ...OpenAPIOption)
	HeadWithOptions(string, handlerGenerator, ...RouterOption)
	Kernel() gin.IRouter
}

type openapiRouter struct {
	prefix    string
	ginRouter gin.IRouter
	openapi   *openapi.API
}

func (r *openapiRouter) Kernel() gin.IRouter {
	return r.ginRouter
}

func (r *openapiRouter) Group(prefixes ...string) Router {
	ginPrefix := joinPathSegments(prefixes...)

	return &openapiRouter{
		prefix:    joinHTTPPaths(r.prefix, ginPrefix),
		ginRouter: r.ginRouter.Group(ginPrefix),
		openapi:   r.openapi,
	}
}

func (r *openapiRouter) Use(middleware ...gin.HandlerFunc) Router {
	r.ginRouter.Use(middleware...)
	return r
}

func (r *openapiRouter) handle(method string, path string, h handlerGenerator, routerOptions *RouterOptions) {
	requestBody, responseBody, query, params, requestHeader, responseHeader, handlerFunc := h()

	ginFuncs := make([]gin.HandlerFunc, 0, 1)
	ginFuncs = append(ginFuncs, routerOptions.PreMiddlewares...)
	ginFuncs = append(ginFuncs, handlerFunc)
	ginFuncs = append(ginFuncs, routerOptions.PostMiddlewares...)

	// gin 相对路径：保留空字符串表示 group 根；非空则规范化（去多余斜杠、去尾斜杠）
	ginPath := path
	if path != "" && path != "/" {
		ginPath = joinPathSegments(path)
	} else if path == "/" {
		ginPath = ""
	}

	// register gin route
	r.ginRouter.Handle(method, ginPath, ginFuncs...)

	// OpenAPI / PathRegister 使用规范化后的完整绝对路径（无尾斜杠）
	fullPath := joinHTTPPaths(r.prefix, path)

	if routerOptions.PathRegister != nil {
		routerOptions.PathRegister(method, fullPath)
	}

	// OpenAPI 使用 {id} 风格 path
	openapiPath := ginPathToOpenAPIPath(fullPath)
	route := r.openapi.Route(method, openapiPath)
	route = mergeOpenAPIOptions(route, routerOptions.OpenAPIOptions...)
	operationID := strings.ReplaceAll(strings.Trim(openapiPath, "/"), "/", "_")
	operationID = strings.ReplaceAll(operationID, "{", "")
	operationID = strings.ReplaceAll(operationID, "}", "")
	operationID = method + "_" + operationID

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

	// path 中真实存在的 :param / {param}，才注册到 OpenAPI
	realExistParam := make(map[string]bool)
	for _, segment := range strings.Split(fullPath, "/") {
		var name string
		switch {
		case strings.HasPrefix(segment, ":"):
			name = strings.TrimPrefix(segment, ":")
		case strings.HasPrefix(segment, "*"):
			name = strings.TrimPrefix(segment, "*")
		case strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}"):
			name = strings.TrimSuffix(strings.TrimPrefix(segment, "{"), "}")
			if i := strings.IndexByte(name, ':'); i >= 0 { // {id:\d+}
				name = name[:i]
			}
		default:
			continue
		}
		if name == "" {
			continue
		}
		realExistParam[name] = true
		pathParam, ok := params[name]
		if !ok {
			pathParam = openapi.PathParam{
				Description: "",
				Type:        openapi.PrimitiveTypeString,
			}
		}
		params[name] = pathParam
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

func (r *openapiRouter) Any(path string, h handlerGenerator, options ...OpenAPIOption) {
	r.AnyWithOptions(path, h, WithOpenAPIOptions(options...))
}

func (r *openapiRouter) AnyWithOptions(path string, h handlerGenerator, options ...RouterOption) {
	anyMethods := []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
		http.MethodTrace,
	}

	routerOptions := mergeRouterOptions(options...)

	for _, method := range anyMethods {
		r.handle(method, path, h, routerOptions)
	}
}

func (r *openapiRouter) GET(path string, h handlerGenerator, options ...OpenAPIOption) {
	r.GetWithOptions(path, h, WithOpenAPIOptions(options...))
}

func (r *openapiRouter) GetWithOptions(path string, h handlerGenerator, options ...RouterOption) {
	routerOptions := mergeRouterOptions(options...)

	r.handle(http.MethodGet, path, h, routerOptions)
}

func (r *openapiRouter) POST(path string, h handlerGenerator, options ...OpenAPIOption) {
	r.PostWithOptions(path, h, WithOpenAPIOptions(options...))
}

func (r *openapiRouter) PostWithOptions(path string, h handlerGenerator, options ...RouterOption) {
	routerOptions := mergeRouterOptions(options...)
	r.handle(http.MethodPost, path, h, routerOptions)
}

func (r *openapiRouter) DELETE(path string, h handlerGenerator, options ...OpenAPIOption) {
	r.DeleteWithOptions(path, h, WithOpenAPIOptions(options...))
}

func (r *openapiRouter) DeleteWithOptions(path string, h handlerGenerator, options ...RouterOption) {
	routerOptions := mergeRouterOptions(options...)
	r.handle(http.MethodDelete, path, h, routerOptions)
}

func (r *openapiRouter) PATCH(path string, h handlerGenerator, options ...OpenAPIOption) {
	r.PatchWithOptions(path, h, WithOpenAPIOptions(options...))
}

func (r *openapiRouter) PatchWithOptions(path string, h handlerGenerator, options ...RouterOption) {
	routerOptions := mergeRouterOptions(options...)
	r.handle(http.MethodPatch, path, h, routerOptions)
}

func (r *openapiRouter) PUT(path string, h handlerGenerator, options ...OpenAPIOption) {
	r.PutWithOptions(path, h, WithOpenAPIOptions(options...))
}

func (r *openapiRouter) PutWithOptions(path string, h handlerGenerator, options ...RouterOption) {
	routerOptions := mergeRouterOptions(options...)
	r.handle(http.MethodPut, path, h, routerOptions)
}

func (r *openapiRouter) OPTIONS(path string, h handlerGenerator, options ...OpenAPIOption) {
	r.OptionsWithOptions(path, h, WithOpenAPIOptions(options...))
}

func (r *openapiRouter) OptionsWithOptions(path string, h handlerGenerator, options ...RouterOption) {
	routerOptions := mergeRouterOptions(options...)
	r.handle(http.MethodOptions, path, h, routerOptions)
}

func (r *openapiRouter) HEAD(path string, h handlerGenerator, options ...OpenAPIOption) {
	r.HeadWithOptions(path, h, WithOpenAPIOptions(options...))
}

func (r *openapiRouter) HeadWithOptions(path string, h handlerGenerator, options ...RouterOption) {
	routerOptions := mergeRouterOptions(options...)
	r.handle(http.MethodHead, path, h, routerOptions)
}
