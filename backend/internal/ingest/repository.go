package ingest

import "context"

// PendingRecord is an article candidate waiting for review.
type PendingRecord struct {
	SourceID  int64
	Title     string
	Summary   string
	SourceURL string
}

// Repository persists pending ingest records.
type Repository interface {
	SavePending(ctx context.Context, rec PendingRecord) error
}
