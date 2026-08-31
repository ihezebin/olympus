package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	rw := &responseWriter{Body: bytes.NewBufferString("hello"), ResponseWriter: c.Writer}
	got := responseBody(rw)
	if got != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
}

func TestResponseBodySkipsMedia(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	rw := &responseWriter{Body: new(bytes.Buffer), ResponseWriter: c.Writer}
	c.Writer = rw
	c.Header("Content-Type", "image/png")

	payload := []byte("fake-png-binary-payload")
	if _, err := rw.Write(payload); err != nil {
		t.Fatalf("write: %v", err)
	}

	got := responseBody(rw)
	want := fmt.Sprintf("[skipped body, content-length=%d]", len(payload))
	if got != want {
		t.Fatalf("body=%q, want %q", got, want)
	}
	if rw.Body.Len() != 0 {
		t.Fatalf("media body should not be buffered, got len=%d", rw.Body.Len())
	}
	if !bytes.Equal(rec.Body.Bytes(), payload) {
		t.Fatalf("response payload mismatch")
	}
}

func TestSkipRequestBodyRead(t *testing.T) {
	cases := []struct {
		ct   string
		skip bool
	}{
		{"application/json", false},
		{"text/plain", false},
		{"multipart/form-data; boundary=----x", true},
		{"multipart/mixed; boundary=----x", true},
		{"application/octet-stream", true},
		{"video/mp4", true},
		{"audio/mpeg", true},
		{"image/png", true},
		{"VIDEO/MP4", true}, // ParseMediaType lowercases type
	}
	for _, tc := range cases {
		if got := skipRequestBodyRead(tc.ct); got != tc.skip {
			t.Fatalf("skipRequestBodyRead(%q)=%v, want %v", tc.ct, got, tc.skip)
		}
	}
}

func TestRequestBodySkipsMultipartWithoutConsuming(t *testing.T) {
	gin.SetMode(gin.TestMode)
	payload := []byte("fake-binary-payload")
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(payload))
	c.Request.Header.Set("Content-Type", "multipart/form-data; boundary=----x")
	c.Request.ContentLength = int64(len(payload))

	got := requestBody(c)
	want := "[skipped body, content-length=19]"
	if got != want {
		t.Fatalf("body=%q, want %q", got, want)
	}

	// body 未被消费，后续仍可读
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		t.Fatalf("read after skip: %v", err)
	}
	if !bytes.Equal(data, payload) {
		t.Fatalf("body consumed unexpectedly: got %q", data)
	}
}
