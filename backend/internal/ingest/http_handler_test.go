package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/bajiaozhi/w-mma/backend/internal/source"
)

type capturePublisher struct {
	job FetchJob
	err error
}

func (p *capturePublisher) Enqueue(_ context.Context, job FetchJob) error {
	p.job = job
	return p.err
}

type fakeSourceReader struct {
	item source.DataSource
	err  error
}

func (f fakeSourceReader) Get(_ context.Context, _ int64) (source.DataSource, error) {
	return f.item, f.err
}

func TestAdminIngestFetch_UsesSourceURLWhenURLMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pub := &capturePublisher{}
	sourceReader := fakeSourceReader{
		item: source.DataSource{
			ID:         7,
			Name:       "UFC Events",
			SourceURL:  "https://www.ufc.com/events",
			ParserKind: "ufc_schedule",
		},
	}

	r := gin.New()
	RegisterAdminIngestRoutes(r, pub, sourceReader)

	body, _ := json.Marshal(map[string]any{
		"source_id": 7,
	})
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/ingest/fetch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
	if pub.job.URL != "https://www.ufc.com/events" {
		t.Fatalf("expected fallback source url, got %q", pub.job.URL)
	}
	if pub.job.ParserKind != "ufc_schedule" {
		t.Fatalf("expected parser kind from source, got %q", pub.job.ParserKind)
	}
}
