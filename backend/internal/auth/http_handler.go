package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func RegisterAdminAuthRoutes(r *gin.Engine, svc *Service) {
	r.POST("/admin/auth/login", func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, err := svc.Login(c.Request.Context(), req.Username, req.Password)
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidCredentials):
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			case errors.Is(err, ErrUserDisabled):
				c.JSON(http.StatusForbidden, gin.H{"error": "user disabled"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	r.POST("/admin/auth/logout", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}
