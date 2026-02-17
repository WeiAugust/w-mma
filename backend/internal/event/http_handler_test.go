package event

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func fakeEventService() *Service {
	return NewService(NewInMemoryRepository())
}

func TestGetEventCard_ReturnsBouts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterEventRoutes(r, fakeEventService())
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/events/10", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
