package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ihezebin/openapi"
)

type deleteByIDReq struct {
	ID string `uri:"id"`
}

type deleteByIDResp struct {
	ID string `json:"id"`
}

type jsonNameReq struct {
	Name string `json:"name" binding:"required"`
}

func TestDeleteWithEmptyJSONBodyBindsURI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	_, _, _, _, _, _, ginHandler := NewHandler(func(c *gin.Context, req deleteByIDReq) (deleteByIDResp, error) {
		return deleteByIDResp{ID: req.ID}, nil
	})()

	r.DELETE("/items/:id", ginHandler)

	req := httptest.NewRequest(http.MethodDelete, "/items/abc-123", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusOK, w.Body.String())
	}

	var body Body[deleteByIDResp]
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body.Code != CodeOK {
		t.Fatalf("code = %v, want %v, message=%s", body.Code, CodeOK, body.Message)
	}
	if body.Data.ID != "abc-123" {
		t.Fatalf("id = %q, want %q", body.Data.ID, "abc-123")
	}
}

func TestPointerRequestEmptyJSONBodyNoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	_, _, _, _, _, _, ginHandler := NewHandler(func(c *gin.Context, req *jsonNameReq) (string, error) {
		if req == nil {
			return "nil", nil
		}
		return req.Name, nil
	})()
	r.POST("/names", ginHandler)

	req := httptest.NewRequest(http.MethodPost, "/names", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestBadJSONReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	_, _, _, _, _, _, ginHandler := NewHandler(func(c *gin.Context, req deleteByIDReq) (deleteByIDResp, error) {
		return deleteByIDResp{ID: req.ID}, nil
	})()
	r.DELETE("/items/:id", ginHandler)

	req := httptest.NewRequest(http.MethodDelete, "/items/x", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusBadRequest, w.Body.String())
	}
	var body Body[deleteByIDResp]
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Code != CodeBadRequest {
		t.Fatalf("code = %v, want %v", body.Code, CodeBadRequest)
	}
}

func TestNewErrorSetsHTTPStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	_, _, _, _, _, _, ginHandler := NewHandler(func(c *gin.Context, req deleteByIDReq) (string, error) {
		return "", NewError(CodeForbidden, "nope")
	})()
	r.GET("/items/:id", ginHandler)

	req := httptest.NewRequest(http.MethodGet, "/items/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusForbidden, w.Body.String())
	}
}

func TestErrorWithCodeSetsHTTPStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	_, _, _, _, _, _, ginHandler := NewHandler(func(c *gin.Context, req deleteByIDReq) (string, error) {
		return "", ErrorWithCode(CodeNotFound)
	})()
	r.GET("/items/:id", ginHandler)

	req := httptest.NewRequest(http.MethodGet, "/items/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body=%s", w.Code, http.StatusNotFound, w.Body.String())
	}
}

func TestJoinHTTPPaths(t *testing.T) {
	cases := []struct {
		parts []string
		want  string
	}{
		{[]string{}, "/"},
		{[]string{""}, "/"},
		{[]string{"/"}, "/"},
		{[]string{"/api", ""}, "/api"},
		{[]string{"/api/", ""}, "/api"},
		{[]string{"/api", "/"}, "/api"},
		{[]string{"/api/", "/items/:id/"}, "/api/items/:id"},
		{[]string{"api", "items", ":id"}, "/api/items/:id"},
		{[]string{"/hello/", "/world/"}, "/hello/world"},
	}
	for _, tc := range cases {
		got := joinHTTPPaths(tc.parts...)
		if got != tc.want {
			t.Fatalf("joinHTTPPaths(%q) = %q, want %q", tc.parts, got, tc.want)
		}
	}

	if got := ginPathToOpenAPIPath("/api/items/:id"); got != "/api/items/{id}" {
		t.Fatalf("ginPathToOpenAPIPath = %q, want /api/items/{id}", got)
	}
	if got := ginPathToOpenAPIPath("/files/*filepath"); got != "/files/{filepath}" {
		t.Fatalf("ginPathToOpenAPIPath wildcard = %q", got)
	}
}

func TestPathRegisterNoTrailingSlashOnEmptyRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := openapi.NewAPI("test")
	engine := gin.New()
	router := &openapiRouter{ginRouter: engine, openapi: api, prefix: ""}

	var registered []string
	group := router.Group("/api/")
	group.GetWithOptions("", NewHandler(func(c *gin.Context, req map[string]any) (string, error) {
		return "ok", nil
	}), WithPathRegister(func(method, path string) {
		registered = append(registered, method+" "+path)
	}))

	if len(registered) != 1 {
		t.Fatalf("registered = %v, want 1 entry", registered)
	}
	if registered[0] != "GET /api" {
		t.Fatalf("registered path = %q, want %q", registered[0], "GET /api")
	}
}

type openAPIPathParamRouter struct{}

func (openAPIPathParamRouter) RegisterRoutes(router Router) {
	router.DELETE("/items/:id", NewHandler(func(c *gin.Context, req deleteByIDReq) (deleteByIDResp, error) {
		return deleteByIDResp{ID: req.ID}, nil
	}))
}

func TestOpenAPIRegistersPathParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	s, err := NewServer(nil, WithPort(0), WithServiceName("t"), WithHiddenRoutesLog())
	if err != nil {
		t.Fatal(err)
	}
	s.RegisterRoutes(openAPIPathParamRouter{})

	spec, err := s.OpenAPI().Spec()
	if err != nil {
		t.Fatalf("spec: %v", err)
	}
	pathItem := spec.Paths.Find("/items/{id}")
	if pathItem == nil {
		t.Fatalf("path /items/{id} not found")
	}
	op := pathItem.Delete
	if op == nil {
		t.Fatal("delete operation missing")
	}
	found := false
	for _, p := range op.Parameters {
		if p.Value != nil && p.Value.In == "path" && p.Value.Name == "id" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("path parameter id not registered, params=%+v", op.Parameters)
	}
}
