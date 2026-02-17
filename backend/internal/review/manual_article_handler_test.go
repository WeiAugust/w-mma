package review

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestManualArticleCreate_RequiresSourceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := NewMemoryRepository()
	r := gin.New()
	RegisterAdminManualArticleRoutes(r, repo)

	missingSource := []byte(`{"title":"manual","summary":"s","source_url":"https://example.com"}`)
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/admin/articles/manual", bytes.NewReader(missingSource))
	req1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when source_id missing, got %d", w1.Code)
	}

	okBody := []byte(`{"source_id":1,"title":"manual","summary":"s","source_url":"https://example.com"}`)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/admin/articles/manual", bytes.NewReader(okBody))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 with source_id, got %d", w2.Code)
	}
}
