package summary

import (
	"context"
	"testing"
)

func TestCreateSummaryJob_NoAPIKey_DowngradesToManualRequired(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo, Config{Provider: "openai", APIKey: ""})

	job, err := svc.CreateArticleJob(context.Background(), 1, 101)
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if job.Status != StatusManualRequired {
		t.Fatalf("expected %s, got %s", StatusManualRequired, job.Status)
	}
}

func TestCreateSummaryJob_WithAPIKey_EnqueuesPending(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo, Config{Provider: "openai", APIKey: "key-123"})

	job, err := svc.CreateArticleJob(context.Background(), 1, 101)
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if job.Status != StatusPending {
		t.Fatalf("expected %s, got %s", StatusPending, job.Status)
	}
}
