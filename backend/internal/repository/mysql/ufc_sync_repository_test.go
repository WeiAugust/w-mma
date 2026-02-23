package mysqlrepo

import (
	"testing"
	"time"

	"github.com/bajiaozhi/w-mma/backend/internal/model"
)

func TestMergeFighterUpdatesByDate_PreservesExistingWhenIncomingIsPartial(t *testing.T) {
	existing := []model.FighterUpdate{
		{
			FighterID:   669,
			Content:     "2025-10-04 · 胜 · KO/TKO终结 · 第1回合 1:20",
			PublishedAt: time.Date(2025, 10, 4, 0, 0, 0, 0, time.UTC),
		},
		{
			FighterID:   669,
			Content:     "2025-03-08 · 负 · 一致判定 · 第5回合 5:00",
			PublishedAt: time.Date(2025, 3, 8, 0, 0, 0, 0, time.UTC),
		},
		{
			FighterID:   669,
			Content:     "2024-10-05 · 胜 · KO/TKO终结 · 第4回合 4:32",
			PublishedAt: time.Date(2024, 10, 5, 0, 0, 0, 0, time.UTC),
		},
	}
	incoming := []model.FighterUpdate{
		{
			FighterID:   669,
			Content:     "2025-10-04 · 胜 · KO/TKO终结 · 第1回合 1:20",
			PublishedAt: time.Date(2025, 10, 4, 0, 0, 0, 0, time.UTC),
		},
	}

	merged := mergeFighterUpdatesByDate(existing, incoming)
	if len(merged) != 3 {
		t.Fatalf("expected existing history kept when incoming is partial, got %+v", merged)
	}
}

func TestMergeFighterUpdatesByDate_IncomingOverridesSameDate(t *testing.T) {
	existing := []model.FighterUpdate{
		{
			FighterID:   669,
			Content:     "2025-03-08 · 胜 · 一致判定 · 第5回合 5:00",
			PublishedAt: time.Date(2025, 3, 8, 0, 0, 0, 0, time.UTC),
		},
	}
	incoming := []model.FighterUpdate{
		{
			FighterID:   669,
			Content:     "2025-03-08 · 负 · 一致判定 · 第5回合 5:00",
			PublishedAt: time.Date(2025, 3, 8, 0, 0, 0, 0, time.UTC),
		},
	}

	merged := mergeFighterUpdatesByDate(existing, incoming)
	if len(merged) != 1 {
		t.Fatalf("expected one merged row, got %+v", merged)
	}
	if merged[0].Content != "2025-03-08 · 负 · 一致判定 · 第5回合 5:00" {
		t.Fatalf("expected incoming row to override same date, got %+v", merged)
	}
}
