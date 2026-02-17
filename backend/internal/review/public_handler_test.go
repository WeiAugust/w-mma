package review

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakePublishedRepo struct {
	items []PendingArticle
}

func (f *fakePublishedRepo) ListPublished(context.Context) ([]PendingArticle, error) {
	return f.items, nil
}

type fakePlaybackPolicy struct {
	canPlay bool
}

func (f *fakePlaybackPolicy) CanPlay(context.Context, int64) bool {
	return f.canPlay
}

func TestPublicResponse_FiltersPlaybackWhenNoRights(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	repo := &fakePublishedRepo{
		items: []PendingArticle{
			{
				ID:       1,
				SourceID: 11,
				Title:    "t",
				VideoURL: "https://video.example.com/a.mp4",
			},
		},
	}
	RegisterPublicContentRoutes(r, repo, &fakePlaybackPolicy{canPlay: false})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/articles", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var payload struct {
		Items []PendingArticle `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(payload.Items))
	}
	if payload.Items[0].CanPlay {
		t.Fatalf("expected can_play=false")
	}
	if payload.Items[0].VideoURL != "" {
		t.Fatalf("expected video url removed when no playback rights")
	}
}
