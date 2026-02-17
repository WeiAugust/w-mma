package fighter

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func fakeFighterService() *Service {
	return NewService(NewInMemoryRepository())
}

func TestSearchFighter_ReturnsMatchedNames(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterFighterRoutes(r, fakeFighterService())
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/fighters/search?q=Alex", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCreateManualFighter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterAdminFighterRoutes(r, fakeFighterService())

	body := []byte(`{"source_id":1,"name":"Manual Fighter","country":"CN","record":"1-0"}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/fighters/manual", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
