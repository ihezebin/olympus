package httpserver

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ihezebin/openapi"
	"github.com/pkg/errors"

	"github.com/ihezebin/olympus/logger"
)

type Handler[RequestT any, ResponseT any] func(c *gin.Context, req RequestT) (resp ResponseT, err error)

func newErrHandlerFunc[RequestT any, ResponseT any](err error) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data ResponseT
		body := &Body[ResponseT]{
			Data: data,
		}
		errx := ErrorWithInternalServer()
		errx.Err = err

		body.WithErr(errx)
		c.AbortWithStatusJSON(body.status, body)
	}
}

// ensureRequestAllocated 保证 RequestT 为指针类型时内层已分配，避免空 body 跳过 JSON 绑定后
// ShouldBindUri / ShouldBindHeader 对 nil 解引用 panic。
func ensureRequestAllocated(requestPtr any) {
	v := reflect.ValueOf(requestPtr)
	if v.Kind() != reflect.Ptr {
		return
	}
	v = v.Elem()
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
}

// allowEmptyJSONBody 这些方法常无 body；Content-Type 为 JSON 时空 body 的 EOF 可忽略。
func allowEmptyJSONBody(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodDelete:
		return true
	default:
		return false
	}
}

func newGinHandlerFunc[RequestT any, ResponseT any](handler Handler[RequestT, ResponseT], isRequestStruct bool) gin.HandlerFunc {

	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var err error

		body := &Body[ResponseT]{
			status: http.StatusOK,
			Code:   CodeOK,
		}

		requestPtr := new(RequestT)
		if isRequestStruct {
			ensureRequestAllocated(requestPtr)
			if err = c.ShouldBind(requestPtr); err != nil {
				// DELETE 等带 Content-Type: application/json 但无 body 时，Gin 选 JSON 绑定器，
				// 空 body 解码会返回 EOF；对通常无 body 的方法忽略并继续 URI/Header 绑定。
				if !(errors.Is(err, io.EOF) && allowEmptyJSONBody(c.Request.Method)) {
					logger.WithError(err).Errorf(ctx, "failed to bind, uri: %s", c.Request.RequestURI)
					body = body.WithErr(ErrorWithBadRequest())
					c.PureJSON(body.status, body)
					return
				}
			}
		} else { // map：优先 body，空则从 query 读
			if err = c.ShouldBindWith(requestPtr, mapBinding{}); err != nil {
				logger.WithError(err).Errorf(ctx, "failed to bind, uri: %s", c.Request.RequestURI)
				body = body.WithErr(ErrorWithBadRequest())
				c.PureJSON(body.status, body)
				return
			}
		}

		if isRequestStruct && len(c.Params) > 0 {
			if err = c.ShouldBindUri(requestPtr); err != nil {
				logger.WithError(err).Errorf(ctx, "failed to bind uri, uri: %s, param: %+v", c.Request.RequestURI, c.Params)
				body = body.WithErr(ErrorWithBadRequest())
				c.PureJSON(body.status, body)
				return
			}
		}

		if isRequestStruct && len(c.Request.Header) > 0 {
			if err = c.ShouldBindHeader(requestPtr); err != nil {
				logger.WithError(err).Errorf(ctx, "failed to bind header, uri: %s, header: %+v", c.Request.RequestURI, c.Request.Header)
				body = body.WithErr(ErrorWithBadRequest())
				c.PureJSON(body.status, body)
				return
			}
		}

		var response ResponseT
		response, err = handler(c, *requestPtr)
		if c.Writer.Written() {
			return
		}

		// handle error
		if err != nil {
			var errx *Err
			if errors.As(err, &errx) {
				body = body.WithErr(errx)
			} else {
				body = body.WithErr(ErrorWithInternalServer())
			}
		} else {
			body.Data = response
		}

		// handle success
		c.PureJSON(body.status, body)
	}
}

// handlerGenerator return [requestBody, responseBody, query, params, requestHeader, responseHeader, gin.HandlerFunc]
type handlerGenerator func() (*openapi.Model, *openapi.Model, map[string]openapi.QueryParam, map[string]openapi.PathParam, map[string]openapi.HeaderParam, map[string]openapi.HeaderParam, gin.HandlerFunc)

func NewHandler[RequestT any, ResponseT any](handler Handler[RequestT, ResponseT]) handlerGenerator {
	return func() (*openapi.Model, *openapi.Model, map[string]openapi.QueryParam, map[string]openapi.PathParam, map[string]openapi.HeaderParam, map[string]openapi.HeaderParam, gin.HandlerFunc) {
		responseBodyModel := openapi.ModelOf[Body[ResponseT]]()

		request := new(RequestT)
		requestType := reflect.TypeOf(request).Elem()

		for requestType.Kind() == reflect.Ptr {
			requestType = requestType.Elem()
		}

		//  RequestT 必须是结构体或者 map
		if requestType.Kind() != reflect.Struct && requestType.Kind() != reflect.Map {
			err := fmt.Errorf("request type must be struct or map, but got %T", request)
			return nil, &responseBodyModel, nil, nil, nil, nil, newErrHandlerFunc[RequestT, ResponseT](err)
		}

		isRequestStruct := requestType.Kind() == reflect.Struct
		ginHandleFunc := newGinHandlerFunc(handler, isRequestStruct)
		if requestType.Kind() == reflect.Map {
			requestBodyModel := openapi.ModelFromType(requestType)
			return &requestBodyModel, &responseBodyModel, nil, nil, nil, nil, ginHandleFunc
		}

		// 通过反射获取 request 的字段和 tag
		var requestBodyStructFields []reflect.StructField
		query := map[string]openapi.QueryParam{}
		params := map[string]openapi.PathParam{}
		requestHeader := map[string]openapi.HeaderParam{}
		responseHeader := map[string]openapi.HeaderParam{}

		for i := 0; i < requestType.NumField(); i++ {
			field := requestType.Field(i)
			fieldType := field.Type

			tagDescription := field.Tag.Get("description")
			if tagDescription == "" {
				tagDescription = field.Tag.Get("desc")
			}

			isRequired := false
			isAllowEmpty := false
			tagOpenApi := field.Tag.Get("openapi")
			if tagOpenApi != "" { // required,empty
				parts := strings.Split(tagOpenApi, ",")
				for _, part := range parts {
					if part == "required" {
						isRequired = true
					}
					if part == "empty" {
						isAllowEmpty = true
					}
				}
			}

			tagJson := field.Tag.Get("json")
			if tagJson != "" {
				requestBodyStructFields = append(requestBodyStructFields, field)
			}

			primitiveType := openapi.PrimitiveTypeString
			switch fieldType.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				primitiveType = openapi.PrimitiveTypeInteger
			case reflect.Bool:
				primitiveType = openapi.PrimitiveTypeBool
			case reflect.Float64, reflect.Float32:
				primitiveType = openapi.PrimitiveTypeFloat64
			}

			tagQuery := field.Tag.Get("query")
			if tagQuery == "" {
				tagQuery = field.Tag.Get("form")
			}
			if tagQuery != "" {
				queryParam := openapi.QueryParam{
					Description: tagDescription,
					Required:    isRequired,
					AllowEmpty:  isAllowEmpty,
					Type:        primitiveType,
				}

				query[tagQuery] = queryParam
			}

			tagUri := field.Tag.Get("uri")
			if tagUri != "" {
				param := openapi.PathParam{
					Description: tagDescription,
					Type:        primitiveType,
				}

				params[tagUri] = param
			}

			tagHeader := field.Tag.Get("header")
			if tagHeader != "" {
				headerParam := openapi.HeaderParam{
					Description: tagDescription,
					Type:        primitiveType,
					Required:    isRequired,
				}

				requestHeader[tagHeader] = headerParam
			}
		}
		var requestBodyModel *openapi.Model = nil

		if len(requestBodyStructFields) > 0 {
			requestBodyType := reflect.StructOf(requestBodyStructFields)
			requestBodyModel = &openapi.Model{
				Type: requestBodyType,
			}
		}

		return requestBodyModel, &responseBodyModel, query, params, requestHeader, responseHeader, ginHandleFunc
	}
}
