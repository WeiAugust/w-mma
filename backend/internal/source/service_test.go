package source

import (
	"context"
	"testing"
	"time"
)

func TestSourceService_CreateListUpdateToggle(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo)

	created, err := svc.Create(context.Background(), CreateInput{
		Name:            "UFC 官网赛程",
		SourceType:      "schedule",
		Platform:        "ufc",
		SourceURL:       "https://www.ufc.com/events",
		ParserKind:      "rss",
		Enabled:         true,
		RightsDisplay:   true,
		RightsPlayback:  false,
		RightsAISummary: false,
		RightsExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected source id")
	}

	items, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("list source: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 source, got %d", len(items))
	}

	err = svc.Update(context.Background(), created.ID, UpdateInput{
		Name:            "UFC Official",
		RightsDisplay:   boolPtr(true),
		RightsPlayback:  boolPtr(true),
		RightsAISummary: boolPtr(true),
	})
	if err != nil {
		t.Fatalf("update source: %v", err)
	}

	err = svc.Toggle(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("toggle source: %v", err)
	}

	updated, err := svc.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get source: %v", err)
	}
	if updated.Enabled {
		t.Fatalf("expected source disabled after toggle")
	}
	if !updated.RightsPlayback || !updated.RightsAISummary {
		t.Fatalf("expected rights update persisted")
	}
}

func boolPtr(v bool) *bool {
	return &v
}
