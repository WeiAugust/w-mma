package source

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type createRequest struct {
	Name            string `json:"name"`
	SourceType      string `json:"source_type"`
	Platform        string `json:"platform"`
	AccountID       string `json:"account_id"`
	SourceURL       string `json:"source_url"`
	ParserKind      string `json:"parser_kind"`
	Enabled         bool   `json:"enabled"`
	RightsDisplay   bool   `json:"rights_display"`
	RightsPlayback  bool   `json:"rights_playback"`
	RightsAISummary bool   `json:"rights_ai_summary"`
	RightsExpiresAt string `json:"rights_expires_at"`
	RightsProofURL  string `json:"rights_proof_url"`
}

type updateRequest struct {
	Name            string  `json:"name"`
	Platform        string  `json:"platform"`
	AccountID       *string `json:"account_id"`
	SourceURL       string  `json:"source_url"`
	ParserKind      string  `json:"parser_kind"`
	RightsDisplay   *bool   `json:"rights_display"`
	RightsPlayback  *bool   `json:"rights_playback"`
	RightsAISummary *bool   `json:"rights_ai_summary"`
	RightsExpiresAt *string `json:"rights_expires_at"`
	RightsProofURL  *string `json:"rights_proof_url"`
}

func RegisterAdminSourceRoutes(r *gin.Engine, svc *Service) {
	r.GET("/admin/sources", func(c *gin.Context) {
		items, err := svc.List(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	})

	r.POST("/admin/sources", func(c *gin.Context) {
		var req createRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		input := CreateInput{
			Name:            req.Name,
			SourceType:      req.SourceType,
			Platform:        req.Platform,
			AccountID:       req.AccountID,
			SourceURL:       req.SourceURL,
			ParserKind:      req.ParserKind,
			Enabled:         req.Enabled,
			RightsDisplay:   req.RightsDisplay,
			RightsPlayback:  req.RightsPlayback,
			RightsAISummary: req.RightsAISummary,
			RightsProofURL:  req.RightsProofURL,
		}
		if req.RightsExpiresAt != "" {
			expiresAt, err := time.Parse(time.RFC3339, req.RightsExpiresAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rights_expires_at"})
				return
			}
			input.RightsExpiresAt = expiresAt
		}

		item, err := svc.Create(c.Request.Context(), input)
		if err != nil {
			if errors.Is(err, ErrInvalidSourceType) {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, item)
	})

	r.PUT("/admin/sources/:id", func(c *gin.Context) {
		sourceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source id"})
			return
		}

		var req updateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		input := UpdateInput{
			Name:            req.Name,
			Platform:        req.Platform,
			AccountID:       req.AccountID,
			SourceURL:       req.SourceURL,
			ParserKind:      req.ParserKind,
			RightsDisplay:   req.RightsDisplay,
			RightsPlayback:  req.RightsPlayback,
			RightsAISummary: req.RightsAISummary,
			RightsProofURL:  req.RightsProofURL,
		}
		if req.RightsExpiresAt != nil {
			expiresAt, err := time.Parse(time.RFC3339, *req.RightsExpiresAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rights_expires_at"})
				return
			}
			input.RightsExpiresAt = &expiresAt
		}

		if err := svc.Update(c.Request.Context(), sourceID, input); err != nil {
			if errors.Is(err, ErrSourceNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	r.POST("/admin/sources/:id/toggle", func(c *gin.Context) {
		sourceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source id"})
			return
		}

		if err := svc.Toggle(c.Request.Context(), sourceID); err != nil {
			if errors.Is(err, ErrSourceNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}
