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

type URLParser interface {
	ParseURL(ctx context.Context, url string) (PendingRecord, error)
}

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

func (p *HTTPParser) ParseURL(ctx context.Context, url string) (PendingRecord, error) {
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

type ParserRegistry struct {
	fallback URLParser
	parsers  map[string]URLParser
}

func NewParserRegistry(fallback URLParser, parsers map[string]URLParser) *ParserRegistry {
	if fallback == nil {
		fallback = NewHTTPParser(nil)
	}
	if parsers == nil {
		parsers = map[string]URLParser{}
	}
	return &ParserRegistry{fallback: fallback, parsers: parsers}
}

func NewDefaultParserRegistry(client *http.Client) *ParserRegistry {
	base := NewHTTPParser(client)
	return NewParserRegistry(base, map[string]URLParser{
		"generic":      NewLabeledParser("通用来源", base),
		"ufc_schedule": NewLabeledParser("UFC 官方来源", base),
		"one_schedule": NewLabeledParser("ONE 官方来源", base),
		"pfl_schedule": NewLabeledParser("PFL 官方来源", base),
		"jck_schedule": NewLabeledParser("JCK 官方来源", base),
		"wba_schedule": NewLabeledParser("WBA 官方来源", base),
		"wbc_schedule": NewLabeledParser("WBC 官方来源", base),
		"ibf_schedule": NewLabeledParser("IBF 官方来源", base),
		"wbo_schedule": NewLabeledParser("WBO 官方来源", base),
	})
}

func (r *ParserRegistry) Parse(ctx context.Context, job FetchJob) (PendingRecord, error) {
	selected := r.fallback
	if parser, ok := r.parsers[job.ParserKind]; ok {
		selected = parser
	}
	return selected.ParseURL(ctx, job.URL)
}

type labeledParser struct {
	label string
	base  URLParser
}

func NewLabeledParser(label string, base URLParser) URLParser {
	if base == nil {
		base = NewHTTPParser(nil)
	}
	return &labeledParser{label: label, base: base}
}

func (p *labeledParser) ParseURL(ctx context.Context, url string) (PendingRecord, error) {
	rec, err := p.base.ParseURL(ctx, url)
	if err != nil {
		return PendingRecord{}, err
	}
	rec.Summary = "抓取自" + p.label
	return rec, nil
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
