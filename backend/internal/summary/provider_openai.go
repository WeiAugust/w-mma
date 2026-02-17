package summary

import (
	"context"
	"errors"
	"strings"
)

var ErrAPIKeyRequired = errors.New("summary api key is required")

type Summarizer interface {
	Summarize(ctx context.Context, content string) (string, error)
}

type OpenAIProvider struct {
	apiKey string
}

func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{apiKey: apiKey}
}

func (p *OpenAIProvider) Summarize(_ context.Context, content string) (string, error) {
	if strings.TrimSpace(p.apiKey) == "" {
		return "", ErrAPIKeyRequired
	}
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return "", nil
	}
	if len(trimmed) <= 120 {
		return trimmed, nil
	}
	return trimmed[:120], nil
}
