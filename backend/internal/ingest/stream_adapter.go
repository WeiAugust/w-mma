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
	values := map[string]any{
		"source_id": strconv.FormatInt(job.SourceID, 10),
		"url":       job.URL,
	}
	if job.ParserKind != "" {
		values["parser_kind"] = job.ParserKind
	}
	return p.queue.Publish(ctx, values)
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

		parserKind := "generic"
		if rawParserKind, ok := msg.Values["parser_kind"]; ok {
			if parsed := fmt.Sprint(rawParserKind); parsed != "" {
				parserKind = parsed
			}
		}

		return handler(FetchJob{SourceID: sourceID, URL: url, ParserKind: parserKind})
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
