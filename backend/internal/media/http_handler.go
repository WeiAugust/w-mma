package media

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type attachRequest struct {
	MediaType string `json:"media_type"`
	URL       string `json:"url"`
	CoverURL  string `json:"cover_url"`
	Title     string `json:"title"`
	SortNo    int    `json:"sort_no"`
}

func RegisterAdminMediaRoutes(r *gin.Engine, svc *Service) {
	registerAttachRoute(r, svc, "article", "/admin/articles/:id/media")
	registerAttachRoute(r, svc, "event", "/admin/events/:id/media")
	registerAttachRoute(r, svc, "fighter", "/admin/fighters/:id/media")
}

func registerAttachRoute(r *gin.Engine, svc *Service, ownerType string, path string) {
	r.POST(path, func(c *gin.Context) {
		ownerID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner id"})
			return
		}

		var req attachRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		item, err := svc.Attach(c.Request.Context(), AttachInput{
			OwnerType: ownerType,
			OwnerID:   ownerID,
			MediaType: req.MediaType,
			URL:       req.URL,
			CoverURL:  req.CoverURL,
			Title:     req.Title,
			SortNo:    req.SortNo,
		})
		if err != nil {
			if errors.Is(err, ErrInvalidOwnerType) || errors.Is(err, ErrInvalidMediaType) {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, item)
	})
}
