package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin_ReturnsJWTAndMiddlewareProtectsAdminRoute(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate hash: %v", err)
	}

	repo := NewStaticUserRepository("admin", string(hash))
	svc := NewService(repo, "test-secret")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterAdminAuthRoutes(r, svc)
	r.GET("/admin/protected", RequireAdminAuth("test-secret"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	noTokenResp := httptest.NewRecorder()
	noTokenReq := httptest.NewRequest(http.MethodGet, "/admin/protected", nil)
	r.ServeHTTP(noTokenResp, noTokenReq)
	if noTokenResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", noTokenResp.Code)
	}

	loginBody, _ := json.Marshal(map[string]string{
		"username": "admin",
		"password": "secret",
	})
	loginResp := httptest.NewRecorder()
	loginReq := httptest.NewRequest(http.MethodPost, "/admin/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected 200 from login, got %d", loginResp.Code)
	}

	var payload struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if payload.Token == "" {
		t.Fatalf("expected non-empty token")
	}

	okResp := httptest.NewRecorder()
	okReq := httptest.NewRequest(http.MethodGet, "/admin/protected", nil)
	okReq.Header.Set("Authorization", "Bearer "+payload.Token)
	r.ServeHTTP(okResp, okReq)
	if okResp.Code != http.StatusOK {
		t.Fatalf("expected 200 with token, got %d", okResp.Code)
	}
}
