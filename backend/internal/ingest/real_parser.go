package ingest

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var titlePattern = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
var spacePattern = regexp.MustCompile(`\s+`)

// HTTPParser fetches and parses real web pages into pending records.
type HTTPParser struct {
	client *http.Client
}

func NewHTTPParser(client *http.Client) *HTTPParser {
	if client == nil {
		client = &http.Client{}
	}
	return &HTTPParser{client: client}
}

func (p *HTTPParser) Parse(ctx context.Context, url string) (PendingRecord, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return PendingRecord{}, err
	}

	res, err := p.client.Do(req)
	if err != nil {
		return PendingRecord{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return PendingRecord{}, err
	}

	title := extractTitle(string(body))
	if title == "" {
		title = url
	}

	return PendingRecord{
		Title:     title,
		Summary:   "抓取自真实来源页面",
		SourceURL: url,
	}, nil
}

func extractTitle(html string) string {
	m := titlePattern.FindStringSubmatch(html)
	if len(m) < 2 {
		return ""
	}
	title := strings.TrimSpace(m[1])
	title = spacePattern.ReplaceAllString(title, " ")
	return title
}
