package review

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAdminReviewHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := newFakeReviewRepo()
	svc := NewService(repo)

	r := gin.New()
	RegisterAdminReviewRoutes(r, svc)

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/admin/review/pending", nil)
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/admin/review/101/approve?reviewer_id=9001", nil)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
}
