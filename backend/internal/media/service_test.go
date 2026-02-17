package media

import (
	"context"
	"testing"
)

func TestAttachMedia_And_ListByOwner(t *testing.T) {
	repo := NewInMemoryRepository()
	svc := NewService(repo)

	_, err := svc.Attach(context.Background(), AttachInput{
		OwnerType: "article",
		OwnerID:   1001,
		MediaType: "image",
		URL:       "https://img.example.com/a.jpg",
		SortNo:    1,
	})
	if err != nil {
		t.Fatalf("attach image media: %v", err)
	}

	_, err = svc.Attach(context.Background(), AttachInput{
		OwnerType: "article",
		OwnerID:   1001,
		MediaType: "video",
		URL:       "https://video.example.com/a.mp4",
		CoverURL:  "https://img.example.com/a-cover.jpg",
		SortNo:    2,
	})
	if err != nil {
		t.Fatalf("attach video media: %v", err)
	}

	items, err := svc.ListByOwner(context.Background(), "article", 1001)
	if err != nil {
		t.Fatalf("list media by owner: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 media items, got %d", len(items))
	}
	if items[0].MediaType != "image" || items[1].MediaType != "video" {
		t.Fatalf("expected ordered media list by sort_no")
	}
}
