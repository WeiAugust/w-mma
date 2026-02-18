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

	items, err := svc.List(context.Background(), ListFilter{})
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

func TestSourceService_SoftDeleteRestoreAndFilter(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo)

	created, err := svc.Create(context.Background(), CreateInput{
		Name:            "UFC Official",
		SourceType:      "schedule",
		Platform:        "ufc",
		SourceURL:       "https://www.ufc.com/events",
		ParserKind:      "ufc_schedule",
		Enabled:         true,
		RightsDisplay:   true,
		RightsPlayback:  false,
		RightsAISummary: true,
		IsBuiltin:       true,
	})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}

	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("delete source: %v", err)
	}

	activeItems, err := svc.List(context.Background(), ListFilter{})
	if err != nil {
		t.Fatalf("list active items: %v", err)
	}
	if len(activeItems) != 0 {
		t.Fatalf("expected 0 active items after delete, got %d", len(activeItems))
	}

	deletedItems, err := svc.List(context.Background(), ListFilter{IncludeDeleted: true, IsBuiltin: boolPtr(true)})
	if err != nil {
		t.Fatalf("list include deleted: %v", err)
	}
	if len(deletedItems) != 1 {
		t.Fatalf("expected 1 item include deleted, got %d", len(deletedItems))
	}
	if deletedItems[0].DeletedAt == nil {
		t.Fatalf("expected deleted_at to be set")
	}

	if err := svc.Restore(context.Background(), created.ID); err != nil {
		t.Fatalf("restore source: %v", err)
	}

	restored, err := svc.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get restored source: %v", err)
	}
	if restored.DeletedAt != nil {
		t.Fatalf("expected restored source to clear deleted_at")
	}
}

func boolPtr(v bool) *bool {
	return &v
}
