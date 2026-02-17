package ingest

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bajiaozhi/w-mma/backend/internal/queue"
)

const (
	FetchStreamName = "stream:ingest:fetch"
)

type FetchPublisher interface {
	Enqueue(ctx context.Context, job FetchJob) error
}

type StreamPublisher struct {
	queue *queue.StreamQueue
}

func NewStreamPublisher(queue *queue.StreamQueue) *StreamPublisher {
	return &StreamPublisher{queue: queue}
}

func (p *StreamPublisher) Enqueue(ctx context.Context, job FetchJob) error {
	return p.queue.Publish(ctx, map[string]any{
		"source_id": strconv.FormatInt(job.SourceID, 10),
		"url":       job.URL,
	})
}

type StreamConsumer struct {
	queue *queue.StreamQueue
}

func NewStreamConsumer(queue *queue.StreamQueue) *StreamConsumer {
	return &StreamConsumer{queue: queue}
}

func (c *StreamConsumer) ConsumeOnce(ctx context.Context, handler func(FetchJob) error) error {
	return c.queue.Consume(ctx, func(msg queue.StreamMessage) error {
		rawSourceID, ok := msg.Values["source_id"]
		if !ok {
			return fmt.Errorf("missing source_id")
		}
		rawURL, ok := msg.Values["url"]
		if !ok {
			return fmt.Errorf("missing url")
		}

		sourceID, err := toInt64(rawSourceID)
		if err != nil {
			return err
		}
		url := fmt.Sprint(rawURL)
		if url == "" {
			return fmt.Errorf("empty url")
		}

		return handler(FetchJob{SourceID: sourceID, URL: url})
	})
}

func toInt64(v any) (int64, error) {
	s := fmt.Sprint(v)
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid int64 %q: %w", s, err)
	}
	return n, nil
}
