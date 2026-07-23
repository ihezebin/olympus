package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ihezebin/olympus/httpserver"
)

func TestReuseBodyAllowsMultipleReads(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ReuseBody())
	r.POST("/echo", func(c *gin.Context) {
		b1, _ := io.ReadAll(c.Request.Body)
		b2, _ := io.ReadAll(c.Request.Body)
		c.JSON(http.StatusOK, gin.H{"first": string(b1), "second": string(b2)})
	})

	req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewBufferString(`{"a":1}`))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["first"] != `{"a":1}` || got["second"] != `{"a":1}` {
		t.Fatalf("got=%v, want both reads equal body", got)
	}
}

func TestRecoveryUsesFrameworkCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Recovery())
	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Code != codeInternalServerError {
		t.Fatalf("code = %v, want %v", body.Code, codeInternalServerError)
	}
	if body.Code != int(httpserver.CodeInternalServerError) {
		t.Fatalf("code out of sync with httpserver.CodeInternalServerError: %v vs %v", body.Code, httpserver.CodeInternalServerError)
	}
}

func TestResponseBodyTruncationSuffix(t *testing.T) {
	rw := &responseWriter{Body: bytes.NewBufferString("hello")}
	got := responseBody(rw)
	if got != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
}
