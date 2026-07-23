package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type deleteByIDReq struct {
	ID string `uri:"id"`
}

type deleteByIDResp struct {
	ID string `json:"id"`
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
