package ingest

import (
	"context"
	"testing"
)

type stubURLParser struct {
	summary string
}

func (s stubURLParser) ParseURL(_ context.Context, url string) (PendingRecord, error) {
	return PendingRecord{
		Title:     "title",
		Summary:   s.summary,
		SourceURL: url,
	}, nil
}

func TestParserRegistry_UsesParserByKind(t *testing.T) {
	registry := NewParserRegistry(
		stubURLParser{summary: "fallback"},
		map[string]URLParser{
			"ufc_schedule": stubURLParser{summary: "ufc"},
		},
	)

	rec, err := registry.Parse(context.Background(), FetchJob{
		URL:        "https://example.com",
		ParserKind: "ufc_schedule",
	})
	if err != nil {
		t.Fatalf("parse with parser kind: %v", err)
	}
	if rec.Summary != "ufc" {
		t.Fatalf("expected ufc parser selected, got %q", rec.Summary)
	}
}

func TestParserRegistry_FallsBackWhenKindMissing(t *testing.T) {
	registry := NewParserRegistry(
		stubURLParser{summary: "fallback"},
		map[string]URLParser{
			"ufc_schedule": stubURLParser{summary: "ufc"},
		},
	)

	rec, err := registry.Parse(context.Background(), FetchJob{
		URL:        "https://example.com",
		ParserKind: "unknown",
	})
	if err != nil {
		t.Fatalf("parse with unknown kind: %v", err)
	}
	if rec.Summary != "fallback" {
		t.Fatalf("expected fallback parser selected, got %q", rec.Summary)
	}
}
