package queue

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type StreamMessage struct {
	ID     string
	Values map[string]any
}

type StreamQueue struct {
	client   redis.Cmdable
	stream   string
	group    string
	consumer string
}

func NewStreamQueue(client redis.Cmdable, stream string, group string, consumer string) *StreamQueue {
	return &StreamQueue{client: client, stream: stream, group: group, consumer: consumer}
}

func (q *StreamQueue) EnsureGroup(ctx context.Context) error {
	err := q.client.XGroupCreateMkStream(ctx, q.stream, q.group, "$").Err()
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "BUSYGROUP") {
		return nil
	}
	return err
}

func (q *StreamQueue) Publish(ctx context.Context, values map[string]any) error {
	if len(values) == 0 {
		return errors.New("values cannot be empty")
	}
	return q.client.XAdd(ctx, &redis.XAddArgs{Stream: q.stream, Values: values}).Err()
}

func (q *StreamQueue) Consume(ctx context.Context, handler func(StreamMessage) error) error {
	streams, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    q.group,
		Consumer: q.consumer,
		Streams:  []string{q.stream, ">"},
		Count:    1,
		Block:    1 * time.Second,
	}).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}
	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return nil
	}

	msg := streams[0].Messages[0]
	wrapped := StreamMessage{ID: msg.ID, Values: msg.Values}
	if err := handler(wrapped); err != nil {
		return err
	}
	return q.client.XAck(ctx, q.stream, q.group, msg.ID).Err()
}

func (q *StreamQueue) Stream() string {
	return q.stream
}

func (q *StreamQueue) Group() string {
	return q.group
}

func (q *StreamQueue) String() string {
	return fmt.Sprintf("stream=%s group=%s consumer=%s", q.stream, q.group, q.consumer)
}
