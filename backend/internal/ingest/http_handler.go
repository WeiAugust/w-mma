package ingest

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bajiaozhi/w-mma/backend/internal/source"
)

type fetchRequest struct {
	SourceID   int64  `json:"source_id"`
	URL        string `json:"url"`
	ParserKind string `json:"parser_kind"`
}

type SourceReader interface {
	Get(ctx context.Context, sourceID int64) (source.DataSource, error)
}

func RegisterAdminIngestRoutes(r *gin.Engine, publisher FetchPublisher, sourceReader ...SourceReader) {
	var reader SourceReader
	if len(sourceReader) > 0 {
		reader = sourceReader[0]
	}

	r.POST("/admin/ingest/fetch", func(c *gin.Context) {
		var req fetchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.SourceID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "source_id is required"})
			return
		}

		url := req.URL
		parserKind := req.ParserKind
		if (url == "" || parserKind == "") && reader != nil {
			item, err := reader.Get(c.Request.Context(), req.SourceID)
			if err != nil && !errors.Is(err, source.ErrSourceNotFound) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if err == nil {
				if url == "" {
					url = item.SourceURL
				}
				if parserKind == "" {
					parserKind = item.ParserKind
				}
			}
		}

		if url == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
			return
		}
		if parserKind == "" {
			parserKind = "generic"
		}

		if err := publisher.Enqueue(c.Request.Context(), FetchJob{SourceID: req.SourceID, URL: url, ParserKind: parserKind}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}
