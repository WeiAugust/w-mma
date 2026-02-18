package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAdminRoutes_RequireJWT_ExceptLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r)

	noTokenResp := httptest.NewRecorder()
	noTokenReq := httptest.NewRequest(http.MethodGet, "/admin/sources", nil)
	r.ServeHTTP(noTokenResp, noTokenReq)
	if noTokenResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for protected route without token, got %d", noTokenResp.Code)
	}

	loginBody, _ := json.Marshal(map[string]string{
		"username": "admin",
		"password": "admin123456",
	})
	loginResp := httptest.NewRecorder()
	loginReq := httptest.NewRequest(http.MethodPost, "/admin/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected login route open and return 200, got %d", loginResp.Code)
	}

	var payload struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if payload.Token == "" {
		t.Fatalf("expected login token")
	}

	authedResp := httptest.NewRecorder()
	authedReq := httptest.NewRequest(http.MethodGet, "/admin/sources", nil)
	authedReq.Header.Set("Authorization", "Bearer "+payload.Token)
	r.ServeHTTP(authedResp, authedReq)
	if authedResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for protected route with token, got %d", authedResp.Code)
	}
}

func TestLoginRoute_CORSPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r)

	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/admin/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for CORS preflight, got %d", resp.Code)
	}
	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected Access-Control-Allow-Origin header, got %q", got)
	}
	if got := resp.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatalf("expected Access-Control-Allow-Headers header")
	}
	if got := resp.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatalf("expected Access-Control-Allow-Methods header")
	}
}
