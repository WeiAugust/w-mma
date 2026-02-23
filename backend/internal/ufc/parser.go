package ufc

import (
	"context"
	"errors"
	stdhtml "html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	eventHrefPattern            = regexp.MustCompile(`href=["']((?:https?://[^"'#?]+)?/event/[^"'#?]+)["']`)
	athleteHrefPattern          = regexp.MustCompile(`href=["']((?:https?://[^"'#?]+)?/athlete/[^"'#?]+)["']`)
	h1Pattern                   = regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`)
	titlePattern                = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	tagPattern                  = regexp.MustCompile(`(?is)<[^>]+>`)
	spacePattern                = regexp.MustCompile(`\s+`)
	metaOGImagePattern          = regexp.MustCompile(`(?is)<meta[^>]+property=["']og:image["'][^>]+content=["']([^"']+)["']`)
	eventHeroImagePattern       = regexp.MustCompile(`(?is)<img[^>]+src=["']([^"']*/images/styles/background_image[^"']+)["']`)
	recordPattern               = regexp.MustCompile(`\b\d{1,2}-\d{1,2}(?:-\d{1,2})?\b`)
	mainCardPattern             = regexp.MustCompile(`(?i)\bmain\s+card\b`)
	prelimsPattern              = regexp.MustCompile(`(?i)\bprelims\b|\bearly\s+prelims\b`)
	weightClassPattern          = regexp.MustCompile(`(?i)(women['’]s\s+strawweight|women['’]s\s+flyweight|women['’]s\s+bantamweight|women['’]s\s+featherweight|strawweight|flyweight|bantamweight|featherweight|lightweight|welterweight|middleweight|light\s+heavyweight|heavyweight|catchweight)\s+bout`)
	rankingPattern              = regexp.MustCompile(`#\s*\d+`)
	fightCardTimeMainPattern    = regexp.MustCompile(`(?is)field--name-fight-card-time-main[^>]*>.*?<time[^>]+datetime=["']([^"']+)["']`)
	fightCardTimePrelimsPattern = regexp.MustCompile(`(?is)field--name-fight-card-time-prelims[^>]*>.*?<time[^>]+datetime=["']([^"']+)["']`)
	fightCardTimeEarlyPattern   = regexp.MustCompile(`(?is)field--name-fight-card-time-early-prelims[^>]*>.*?<time[^>]+datetime=["']([^"']+)["']`)
	fightCardTimePattern        = regexp.MustCompile(`(?is)field--name-fight-card-time[^>]*>.*?<time[^>]+datetime=["']([^"']+)["']`)
	eventLiveFinalPattern       = regexp.MustCompile(`(?is)"eventLiveStats"\s*:\s*\{[^{}]*"final"\s*:\s*(true|false)`)
	fightBlockPattern           = regexp.MustCompile(`(?is)<div[^>]+class=["'][^"']*\bc-listing-fight\b[^"']*["'][^>]*>`)
	roundTextPattern            = regexp.MustCompile(`(?is)c-listing-fight__result-text\s+round[^>]*>([^<]*)<`)
	timeTextPattern             = regexp.MustCompile(`(?is)c-listing-fight__result-text\s+time[^>]*>([^<]*)<`)
	methodTextPattern           = regexp.MustCompile(`(?is)c-listing-fight__result-text\s+method[^>]*>([^<]*)<`)
	drawOutcomePattern          = regexp.MustCompile(`(?is)c-listing-fight__outcome--(?:draw|no-contest|nc)`)
	eventCardBlockPattern       = regexp.MustCompile(`(?is)<article class="c-card-event--result"[^>]*>`)
	headlinePattern             = regexp.MustCompile(`(?is)c-card-event--result__headline[^>]*>\s*<a[^>]*>(.*?)</a>`)
	mainCardTsPattern           = regexp.MustCompile(`(?is)data-main-card-timestamp=["'](\d+)["']`)
	prelimsCardTsPattern        = regexp.MustCompile(`(?is)data-prelims-card-timestamp=["'](\d+)["']`)
	earlyCardTsPattern          = regexp.MustCompile(`(?is)data-early-card-timestamp=["'](\d+)["']`)
	athleteNicknamePattern      = regexp.MustCompile(`(?is)<p[^>]*class=["'][^"']*hero-profile__nickname[^"']*["'][^>]*>(.*?)</p>`)
	athleteDivisionTitlePattern = regexp.MustCompile(`(?is)<div[^>]*class=["'][^"']*hero-profile__division-title[^"']*["'][^>]*>(.*?)</div>`)
	athleteDivisionBodyPattern  = regexp.MustCompile(`(?is)<div[^>]*class=["'][^"']*hero-profile__division-body[^"']*["'][^>]*>(.*?)</div>`)
	heroStatNumberPattern       = regexp.MustCompile(`(?is)<p[^>]*class=["'][^"']*hero-profile__stat-numb[^"']*["'][^>]*>(.*?)</p>`)
	heroStatLabelPattern        = regexp.MustCompile(`(?is)<p[^>]*class=["'][^"']*hero-profile__stat-text[^"']*["'][^>]*>(.*?)</p>`)
	compareStatNumberPattern    = regexp.MustCompile(`(?is)<div[^>]*class=["'][^"']*c-stat-compare__number[^"']*["'][^>]*>(.*?)</div>`)
	compareStatLabelPattern     = regexp.MustCompile(`(?is)<div[^>]*class=["'][^"']*c-stat-compare__label[^"']*["'][^>]*>(.*?)</div>`)
	bioLabelPattern             = regexp.MustCompile(`(?is)<div[^>]*class=["'][^"']*c-bio__label[^"']*["'][^>]*>(.*?)</div>`)
	bioValuePattern             = regexp.MustCompile(`(?is)<div[^>]*class=["'][^"']*c-bio__text[^"']*["'][^>]*>(.*?)</div>`)
	recordWLDPattern            = regexp.MustCompile(`(?i)\b(\d{1,2}-\d{1,2}(?:-\d{1,2})?)\s*\(\s*W-L-D\s*\)`)
	recordLabelPattern          = regexp.MustCompile(`(?i)\brecord\b[^0-9]{0,20}(\d{1,2}-\d{1,2}(?:-\d{1,2})?)`)
	divisionWeightClassPattern  = regexp.MustCompile(`(?i)(women['’]s\s+strawweight|women['’]s\s+flyweight|women['’]s\s+bantamweight|women['’]s\s+featherweight|strawweight|flyweight|bantamweight|featherweight|lightweight|welterweight|middleweight|light\s+heavyweight|heavyweight|catchweight)\s+division`)
	athletePFPRankPattern       = regexp.MustCompile(`(?i)#\s*(\d+)\s*PFP`)
	athleteFightHistoryPattern  = regexp.MustCompile(`(?is)([A-Z][a-z]{2}\.\s+\d{1,2},\s+\d{4})\s+Round\s+(\d+)\s+Time\s+(\d{1,2}:\d{2})\s+Method\s+(.+?)(?:Watch Replay|Fight Card|Load More|$)`)
	fightResultTokenPattern     = regexp.MustCompile(`(?i)\b(no contest|win|loss|draw)\b`)
)

type Scraper interface {
	ListEventLinks(ctx context.Context, scheduleURL string) ([]EventLink, error)
	GetEventCard(ctx context.Context, eventURL string) (EventCard, error)
	ListAthleteLinks(ctx context.Context, athletesURL string) ([]string, error)
	GetAthleteProfile(ctx context.Context, athleteURL string) (AthleteProfile, error)
}

type HTTPClient struct {
	client *http.Client
}

func NewHTTPClient(client *http.Client) *HTTPClient {
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &HTTPClient{client: client}
}

func (c *HTTPClient) ListEventLinks(ctx context.Context, scheduleURL string) ([]EventLink, error) {
	html, err := c.fetchHTML(ctx, scheduleURL)
	if err != nil {
		return nil, err
	}
	base := detectBaseURL(scheduleURL)
	return parseEventLinksHTML(html, base), nil
}

func (c *HTTPClient) GetEventCard(ctx context.Context, eventURL string) (EventCard, error) {
	html, err := c.fetchHTML(ctx, eventURL)
	if err != nil {
		return EventCard{}, err
	}
	base := detectBaseURL(eventURL)
	return parseEventCardHTML(html, eventURL, base), nil
}

func (c *HTTPClient) ListAthleteLinks(ctx context.Context, athletesURL string) ([]string, error) {
	html, err := c.fetchHTML(ctx, athletesURL)
	if err != nil {
		return nil, err
	}
	base := detectBaseURL(athletesURL)
	return parseAthleteLinksHTML(html, base), nil
}

func (c *HTTPClient) GetAthleteProfile(ctx context.Context, athleteURL string) (AthleteProfile, error) {
	html, err := c.fetchHTML(ctx, athleteURL)
	if err != nil {
		return AthleteProfile{}, err
	}
	return parseAthleteProfileHTML(html, athleteURL), nil
}

func (c *HTTPClient) fetchHTML(ctx context.Context, targetURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; w-mma-bot/1.0)")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", errors.New("fetch failed with status " + resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 3<<20))
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func parseEventLinksHTML(html string, baseURL string) []EventLink {
	if items := parseEventLinksByCardBlocks(html, baseURL); len(items) > 0 {
		return items
	}

	matches := eventHrefPattern.FindAllStringSubmatch(html, -1)
	seen := map[string]struct{}{}
	items := make([]EventLink, 0, len(matches))
	for _, m := range matches {
		abs := toAbsURL(baseURL, m[1])
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		items = append(items, EventLink{
			Name:     eventNameFromURL(abs),
			URL:      abs,
			StartsAt: parseDateFromEventURL(abs),
		})
	}
	return items
}

func parseEventLinksByCardBlocks(rawHTML string, baseURL string) []EventLink {
	blockIndices := eventCardBlockPattern.FindAllStringIndex(rawHTML, -1)
	if len(blockIndices) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	items := make([]EventLink, 0, len(blockIndices))
	for i := range blockIndices {
		start := blockIndices[i][0]
		end := len(rawHTML)
		if i+1 < len(blockIndices) {
			end = blockIndices[i+1][0]
		}
		chunk := rawHTML[start:end]
		hrefMatches := eventHrefPattern.FindAllStringSubmatch(chunk, -1)
		if len(hrefMatches) == 0 || len(hrefMatches[0]) < 2 {
			continue
		}
		absURL := toAbsURL(baseURL, hrefMatches[0][1])
		if _, ok := seen[absURL]; ok {
			continue
		}
		seen[absURL] = struct{}{}

		name := eventNameFromURL(absURL)
		if headline := extractEventHeadline(chunk); headline != "" {
			name = headline
		}
		startsAt := extractEventStartsAtFromCardBlock(chunk)
		if startsAt.IsZero() {
			startsAt = parseDateFromEventURL(absURL)
		}

		items = append(items, EventLink{
			Name:     name,
			URL:      absURL,
			StartsAt: startsAt,
		})
	}
	return items
}

func extractEventHeadline(chunk string) string {
	m := headlinePattern.FindStringSubmatch(chunk)
	if len(m) < 2 {
		return ""
	}
	return cleanText(m[1])
}

func extractEventStartsAtFromCardBlock(chunk string) time.Time {
	mainTs := parseUnixTimestampFromChunk(chunk, mainCardTsPattern)
	prelimsTs := parseUnixTimestampFromChunk(chunk, prelimsCardTsPattern)
	earlyTs := parseUnixTimestampFromChunk(chunk, earlyCardTsPattern)
	return earliestNonZeroTime(mainTs, prelimsTs, earlyTs)
}

func earliestNonZeroTime(items ...time.Time) time.Time {
	earliest := time.Time{}
	for _, item := range items {
		if item.IsZero() {
			continue
		}
		if earliest.IsZero() || item.Before(earliest) {
			earliest = item
		}
	}
	return earliest
}

func parseUnixTimestampFromChunk(chunk string, pattern *regexp.Regexp) time.Time {
	m := pattern.FindStringSubmatch(chunk)
	if len(m) < 2 {
		return time.Time{}
	}
	ts := atoi(strings.TrimSpace(m[1]))
	if ts <= 0 {
		return time.Time{}
	}
	return time.Unix(int64(ts), 0).UTC()
}

func parseEventCardHTML(html string, eventURL string, baseURL string) EventCard {
	name := extractH1(html)
	if name == "" {
		name = extractTitle(html)
	}
	if name == "" {
		name = eventNameFromURL(eventURL)
	}

	startsAt := extractEventStartsAt(html)
	if startsAt.IsZero() {
		startsAt = parseDateFromEventURL(eventURL)
	}
	athleteLinks := parseAthleteLinkMatchesHTML(html, baseURL)
	boutResults := parseBoutResultMeta(html)
	sectionMarkers := parseSectionMarkers(html)
	bouts := make([]EventBout, 0, len(athleteLinks)/2)
	for idx := 0; idx+1 < len(athleteLinks); idx += 2 {
		red := athleteLinks[idx]
		blue := athleteLinks[idx+1]
		weightClass, redRank, blueRank := extractBoutMeta(html, red.Start, blue.End)
		boutResult := boutResultMeta{}
		resultIdx := idx / 2
		if resultIdx < len(boutResults) {
			boutResult = boutResults[resultIdx]
		}
		bouts = append(bouts, EventBout{
			RedName:     athleteNameFromURL(red.URL),
			RedURL:      red.URL,
			RedRank:     redRank,
			BlueName:    athleteNameFromURL(blue.URL),
			BlueURL:     blue.URL,
			BlueRank:    blueRank,
			WeightClass: weightClass,
			CardSegment: cardSegmentByOffset((red.Start+blue.End)/2, sectionMarkers),
			WinnerSide:  boutResult.WinnerSide,
			Result:      boutResult.Result,
			Method:      boutResult.Method,
			Round:       boutResult.Round,
			TimeSec:     boutResult.TimeSec,
		})
	}

	status := inferEventStatus(html, startsAt, bouts)

	return EventCard{
		Name:      name,
		URL:       eventURL,
		Status:    status,
		StartsAt:  startsAt,
		Venue:     "TBD",
		PosterURL: extractEventPosterURL(html, baseURL),
		Bouts:     bouts,
	}
}

func parseAthleteLinksHTML(html string, baseURL string) []string {
	matches := parseAthleteLinkMatchesHTML(html, baseURL)
	items := make([]string, 0, len(matches))
	for _, m := range matches {
		items = append(items, m.URL)
	}
	return items
}

type athleteLinkMatch struct {
	URL   string
	Start int
	End   int
}

type sectionMarker struct {
	Start   int
	Segment string
}

func parseAthleteLinkMatchesHTML(html string, baseURL string) []athleteLinkMatch {
	rawMatches := athleteHrefPattern.FindAllStringSubmatchIndex(html, -1)
	seen := map[string]struct{}{}
	items := make([]athleteLinkMatch, 0, len(rawMatches))
	for _, m := range rawMatches {
		if len(m) < 4 {
			continue
		}
		abs := toAbsURL(baseURL, html[m[2]:m[3]])
		if _, ok := seen[abs]; ok {
			continue
		}
		seen[abs] = struct{}{}
		items = append(items, athleteLinkMatch{
			URL:   abs,
			Start: m[0],
			End:   m[1],
		})
	}
	return items
}

func parseSectionMarkers(html string) []sectionMarker {
	markers := make([]sectionMarker, 0, 4)
	for _, m := range mainCardPattern.FindAllStringIndex(html, -1) {
		markers = append(markers, sectionMarker{Start: m[0], Segment: "main_card"})
	}
	for _, m := range prelimsPattern.FindAllStringIndex(html, -1) {
		markers = append(markers, sectionMarker{Start: m[0], Segment: "prelims"})
	}
	sort.Slice(markers, func(i, j int) bool {
		return markers[i].Start < markers[j].Start
	})
	return markers
}

func cardSegmentByOffset(offset int, markers []sectionMarker) string {
	segment := "prelims"
	bestStart := -1
	for _, marker := range markers {
		if marker.Start > offset {
			break
		}
		if marker.Start >= bestStart {
			bestStart = marker.Start
			segment = marker.Segment
		}
	}
	return segment
}

type boutResultMeta struct {
	WinnerSide string
	Result     string
	Method     string
	Round      int
	TimeSec    int
}

func parseBoutResultMeta(rawHTML string) []boutResultMeta {
	indices := fightBlockPattern.FindAllStringIndex(rawHTML, -1)
	if len(indices) == 0 {
		return nil
	}
	items := make([]boutResultMeta, 0, len(indices))
	for i := range indices {
		start := indices[i][0]
		end := len(rawHTML)
		if i+1 < len(indices) {
			end = indices[i+1][0]
		}
		chunk := rawHTML[start:end]
		method := extractFightResultText(chunk, methodTextPattern)
		round := atoi(extractFightResultText(chunk, roundTextPattern))
		timeText := extractFightResultText(chunk, timeTextPattern)
		items = append(items, boutResultMeta{
			WinnerSide: inferWinnerSide(chunk),
			Result:     composeBoutResult(method, round, timeText),
			Method:     method,
			Round:      round,
			TimeSec:    parseFightTimeToSeconds(timeText),
		})
	}
	return items
}

func extractFightResultText(chunk string, pattern *regexp.Regexp) string {
	m := pattern.FindStringSubmatch(chunk)
	if len(m) < 2 {
		return ""
	}
	return cleanText(m[1])
}

func inferWinnerSide(chunk string) string {
	red := detectCornerOutcome(chunk, "red")
	blue := detectCornerOutcome(chunk, "blue")

	if red == "win" && blue != "win" {
		return "red"
	}
	if blue == "win" && red != "win" {
		return "blue"
	}
	if red == "loss" && blue != "loss" {
		return "blue"
	}
	if blue == "loss" && red != "loss" {
		return "red"
	}
	if red == "draw" || blue == "draw" || red == "no-contest" || blue == "no-contest" || drawOutcomePattern.MatchString(chunk) {
		return ""
	}
	return ""
}

func detectCornerOutcome(chunk string, side string) string {
	lower := strings.ToLower(chunk)
	markers := []string{
		"c-listing-fight__corner-body--" + side,
		"c-listing-fight__corner--" + side,
	}
	otherSide := "blue"
	if side == "blue" {
		otherSide = "red"
	}
	otherMarkers := []string{
		"c-listing-fight__corner-body--" + otherSide,
		"c-listing-fight__corner--" + otherSide,
	}
	for _, marker := range markers {
		idx := strings.Index(lower, marker)
		if idx < 0 {
			continue
		}
		end := len(lower)
		searchStart := idx + len(marker)
		if searchStart < len(lower) {
			for _, other := range otherMarkers {
				next := strings.Index(lower[searchStart:], other)
				if next < 0 {
					continue
				}
				candidate := searchStart + next
				if candidate < end {
					end = candidate
				}
			}
		}
		maxEnd := idx + 1200
		if end > maxEnd {
			end = maxEnd
		}
		if end > len(lower) {
			end = len(lower)
		}
		segment := lower[idx:end]
		switch {
		case strings.Contains(segment, "c-listing-fight__outcome--win"):
			return "win"
		case strings.Contains(segment, "c-listing-fight__outcome--loss"):
			return "loss"
		case strings.Contains(segment, "c-listing-fight__outcome--draw"):
			return "draw"
		case strings.Contains(segment, "c-listing-fight__outcome--no-contest"),
			strings.Contains(segment, "c-listing-fight__outcome--nc"):
			return "no-contest"
		}
	}
	return ""
}

func composeBoutResult(method string, round int, clock string) string {
	parts := make([]string, 0, 3)
	if method != "" {
		parts = append(parts, method)
	}
	if round > 0 {
		parts = append(parts, "R"+strconv.Itoa(round))
	}
	if clock != "" {
		parts = append(parts, clock)
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func parseFightTimeToSeconds(raw string) int {
	parts := strings.Split(strings.TrimSpace(raw), ":")
	if len(parts) != 2 {
		return 0
	}
	min := atoi(parts[0])
	sec := atoi(parts[1])
	if sec < 0 || sec > 59 {
		return 0
	}
	return min*60 + sec
}

func inferEventStatus(rawHTML string, startsAt time.Time, bouts []EventBout) string {
	now := time.Now().UTC()
	if final, ok := extractEventLiveFinal(rawHTML); ok {
		if final {
			return "completed"
		}
		if startsAt.IsZero() {
			return "scheduled"
		}
		if startsAt.After(now) {
			return "scheduled"
		}
		if startsAt.Before(now.Add(-staleEventCompletionWindow)) {
			return "completed"
		}
		return "live"
	}
	resolved := 0
	for _, bout := range bouts {
		if boutOutcomeAvailable(bout) {
			resolved++
		}
	}
	if len(bouts) > 0 {
		if resolved == len(bouts) {
			return "completed"
		}
		if resolved > 0 {
			return "live"
		}
	}
	if startsAt.IsZero() {
		return "scheduled"
	}
	if startsAt.After(now) {
		return "scheduled"
	}
	if startsAt.Before(now.Add(-staleEventCompletionWindow)) {
		return "completed"
	}
	return "live"
}

func boutOutcomeAvailable(bout EventBout) bool {
	if bout.WinnerSide != "" {
		return true
	}
	if strings.TrimSpace(bout.Result) != "" {
		return true
	}
	if strings.TrimSpace(bout.Method) != "" {
		return true
	}
	return bout.Round > 0 || bout.TimeSec > 0
}

func extractEventLiveFinal(rawHTML string) (bool, bool) {
	m := eventLiveFinalPattern.FindStringSubmatch(rawHTML)
	if len(m) < 2 {
		return false, false
	}
	return strings.EqualFold(strings.TrimSpace(m[1]), "true"), true
}

func extractEventStartsAt(rawHTML string) time.Time {
	if mainCard := extractEarliestTimeByPattern(rawHTML, fightCardTimeMainPattern); !mainCard.IsZero() {
		return mainCard
	}
	if prelims := extractEarliestTimeByPattern(rawHTML, fightCardTimePrelimsPattern); !prelims.IsZero() {
		return prelims
	}
	if earlyPrelims := extractEarliestTimeByPattern(rawHTML, fightCardTimeEarlyPattern); !earlyPrelims.IsZero() {
		return earlyPrelims
	}
	return extractEarliestTimeByPattern(rawHTML, fightCardTimePattern)
}

func extractEarliestTimeByPattern(rawHTML string, pattern *regexp.Regexp) time.Time {
	matches := pattern.FindAllStringSubmatch(rawHTML, -1)
	earliest := time.Time{}
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		parsed := parseRFC3339Time(m[1])
		if parsed.IsZero() {
			continue
		}
		if earliest.IsZero() || parsed.Before(earliest) {
			earliest = parsed
		}
	}
	return earliest
}

func parseRFC3339Time(raw string) time.Time {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z0700",
		"2006-01-02T15:04:05.000Z0700",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.UTC()
		}
	}
	return time.Time{}
}

func extractBoutMeta(html string, start int, end int) (string, string, string) {
	windowStart := start - 140
	if windowStart < 0 {
		windowStart = 0
	}
	windowEnd := end + 140
	if windowEnd > len(html) {
		windowEnd = len(html)
	}
	chunk := cleanText(html[windowStart:windowEnd])

	weightClass := ""
	weightMatches := weightClassPattern.FindAllStringSubmatch(chunk, -1)
	if len(weightMatches) > 0 {
		last := weightMatches[len(weightMatches)-1]
		if len(last) > 1 {
			weightClass = normalizeWeightClass(last[1])
		}
	}

	ranks := rankingPattern.FindAllString(chunk, -1)
	normalizedRanks := make([]string, 0, 2)
	seen := map[string]struct{}{}
	for _, rank := range ranks {
		normalized := normalizeRanking(rank)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		normalizedRanks = append(normalizedRanks, normalized)
		if len(normalizedRanks) == 2 {
			break
		}
	}

	redRank := ""
	blueRank := ""
	if len(normalizedRanks) > 0 {
		redRank = normalizedRanks[0]
	}
	if len(normalizedRanks) > 1 {
		blueRank = normalizedRanks[1]
	}
	return weightClass, redRank, blueRank
}

func normalizeRanking(raw string) string {
	text := strings.ReplaceAll(strings.TrimSpace(raw), " ", "")
	if !strings.HasPrefix(text, "#") {
		return ""
	}
	if len(text) < 2 {
		return ""
	}
	return text
}

func normalizeWeightClass(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "women's strawweight", "women’s strawweight":
		return "Women's Strawweight"
	case "women's flyweight", "women’s flyweight":
		return "Women's Flyweight"
	case "women's bantamweight", "women’s bantamweight":
		return "Women's Bantamweight"
	case "women's featherweight", "women’s featherweight":
		return "Women's Featherweight"
	case "strawweight":
		return "Strawweight"
	case "flyweight":
		return "Flyweight"
	case "bantamweight":
		return "Bantamweight"
	case "featherweight":
		return "Featherweight"
	case "lightweight":
		return "Lightweight"
	case "welterweight":
		return "Welterweight"
	case "middleweight":
		return "Middleweight"
	case "light heavyweight":
		return "Light Heavyweight"
	case "heavyweight":
		return "Heavyweight"
	case "catchweight":
		return "Catchweight"
	default:
		return ""
	}
}

func parseAthleteProfileHTML(html string, athleteURL string) AthleteProfile {
	name := extractH1(html)
	if name == "" {
		name = extractTitle(html)
	}
	if name == "" {
		name = athleteNameFromURL(athleteURL)
	}

	nickname := extractByPattern(html, athleteNicknamePattern)
	nickname = strings.Trim(nickname, " \"'")
	weightClass := normalizeDivisionWeightClass(extractByPattern(html, athleteDivisionTitlePattern))
	weightClass = extractAthleteWeightClass(html, name, weightClass)

	divisionBody := extractByPattern(html, athleteDivisionBodyPattern)
	record := extractAthleteRecord(html, divisionBody)

	records := buildAthleteRecordMap(html, record)
	stats := buildAthleteStatMap(html)
	stats = mergeStringMap(stats, buildAthleteHeroMeta(html, name))
	bio := buildAthleteBioMap(html)
	stats = mergeStringMap(stats, bio)
	updates := buildAthleteFightHistory(html)
	country := inferAthleteCountry(bio)
	if country == "" {
		country = inferCountryByLegacySnippet(html)
	}

	return AthleteProfile{
		Name:        name,
		Nickname:    nickname,
		URL:         athleteURL,
		Country:     country,
		Record:      record,
		WeightClass: weightClass,
		AvatarURL:   extractMetaOGImage(html),
		Stats:       stats,
		Records:     records,
		Updates:     updates,
	}
}

func extractByPattern(raw string, pattern *regexp.Regexp) string {
	m := pattern.FindStringSubmatch(raw)
	if len(m) < 2 {
		return ""
	}
	return cleanText(m[1])
}

func extractAthleteRecord(rawHTML string, divisionBody string) string {
	if match := recordPattern.FindString(divisionBody); match != "" {
		return match
	}
	if m := recordWLDPattern.FindStringSubmatch(rawHTML); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	text := cleanText(rawHTML)
	if m := recordWLDPattern.FindStringSubmatch(text); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	if m := recordLabelPattern.FindStringSubmatch(text); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	lower := strings.ToLower(text)
	if idx := strings.Index(lower, "wins by"); idx > 0 {
		text = strings.TrimSpace(text[:idx])
	}
	if len(text) > 2200 {
		text = text[:2200]
	}
	return recordPattern.FindString(text)
}

func extractAthleteWeightClass(rawHTML string, athleteName string, existing string) string {
	if existing != "" {
		return existing
	}
	window := athleteHeroWindow(rawHTML, athleteName)
	if m := divisionWeightClassPattern.FindStringSubmatch(window); len(m) > 1 {
		return normalizeWeightClass(m[1])
	}
	full := cleanText(rawHTML)
	if m := divisionWeightClassPattern.FindStringSubmatch(full); len(m) > 1 {
		return normalizeWeightClass(m[1])
	}
	if strings.Contains(strings.ToLower(full), "light heavyweight") {
		return "Light Heavyweight"
	}
	return ""
}

func buildAthleteHeroMeta(rawHTML string, athleteName string) map[string]string {
	window := athleteHeroWindow(rawHTML, athleteName)
	meta := map[string]string{}
	if m := athletePFPRankPattern.FindStringSubmatch(window); len(m) > 1 {
		meta["PFP Rank"] = "#" + strings.TrimSpace(m[1])
	}
	lower := strings.ToLower(window)
	switch {
	case strings.Contains(lower, "active"):
		meta["Athlete Status"] = "Active"
	case strings.Contains(lower, "inactive"):
		meta["Athlete Status"] = "Inactive"
	case strings.Contains(lower, "retired"):
		meta["Athlete Status"] = "Retired"
	}
	if strings.Contains(lower, "title holder") {
		meta["Title Status"] = "Title Holder"
	}
	if len(meta) == 0 {
		full := strings.ToLower(cleanText(rawHTML))
		if m := athletePFPRankPattern.FindStringSubmatch(full); len(m) > 1 {
			meta["PFP Rank"] = "#" + strings.TrimSpace(m[1])
		}
		if strings.Contains(full, "title holder") {
			meta["Title Status"] = "Title Holder"
		}
	}
	if len(meta) == 0 {
		return nil
	}
	return meta
}

func athleteHeroWindow(rawHTML string, athleteName string) string {
	text := cleanText(rawHTML)
	if text == "" {
		return ""
	}
	needle := strings.ToLower(strings.TrimSpace(athleteName))
	if needle == "" {
		if len(text) > 1400 {
			return text[:1400]
		}
		return text
	}
	idx := strings.Index(strings.ToLower(text), needle)
	if idx < 0 {
		if len(text) > 1400 {
			return text[:1400]
		}
		return text
	}
	start := idx - 480
	if start < 0 {
		start = 0
	}
	end := idx + 260
	if end > len(text) {
		end = len(text)
	}
	return strings.TrimSpace(text[start:end])
}

func buildAthleteFightHistory(rawHTML string) []AthleteUpdate {
	text := cleanText(rawHTML)
	idx := strings.Index(strings.ToLower(text), "athlete record")
	if idx < 0 {
		return nil
	}
	chunk := strings.TrimSpace(text[idx:])
	if len(chunk) > 5000 {
		chunk = chunk[:5000]
	}
	matches := athleteFightHistoryPattern.FindAllStringSubmatchIndex(chunk, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	items := make([]AthleteUpdate, 0, len(matches))
	for _, match := range matches {
		if len(match) < 10 {
			continue
		}
		dateText := strings.TrimSpace(chunk[match[2]:match[3]])
		round := strings.TrimSpace(chunk[match[4]:match[5]])
		clock := strings.TrimSpace(chunk[match[6]:match[7]])
		methodRaw := strings.TrimSpace(chunk[match[8]:match[9]])
		date := parseAthleteFightDate(dateText)
		resultToken := detectFightResultAround(chunk, match[0])
		result := normalizeFightResult(resultToken)
		method := normalizeFightMethod(methodRaw)
		content := formatAthleteFightHistory(date, dateText, result, method, round, clock)
		if content == "" {
			continue
		}
		if _, exists := seen[content]; exists {
			continue
		}
		seen[content] = struct{}{}
		publishedAt := date
		if publishedAt.IsZero() {
			publishedAt = time.Now().UTC()
		}
		items = append(items, AthleteUpdate{
			Content:     content,
			PublishedAt: publishedAt,
		})
		if len(items) >= 12 {
			break
		}
	}
	if len(items) == 0 {
		return nil
	}
	return items
}

func detectFightResultAround(chunk string, dateStart int) string {
	if dateStart < 0 {
		return ""
	}
	windowStart := dateStart - 2400
	if windowStart < 0 {
		windowStart = 0
	}
	window := chunk[windowStart:dateStart]
	matches := fightResultTokenPattern.FindAllStringSubmatchIndex(window, -1)
	if len(matches) == 0 {
		return ""
	}
	last := matches[len(matches)-1]
	if len(last) < 4 {
		return ""
	}
	return strings.TrimSpace(window[last[2]:last[3]])
}

func parseAthleteFightDate(raw string) time.Time {
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Time{}
	}
	text = strings.ReplaceAll(text, ".", "")
	parsed, err := time.Parse("Jan 2, 2006", text)
	if err != nil {
		return time.Time{}
	}
	return parsed.UTC()
}

func normalizeFightResult(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "win":
		return "胜"
	case "loss":
		return "负"
	case "draw":
		return "平"
	case "no contest":
		return "无结果"
	default:
		return ""
	}
}

func normalizeFightMethod(raw string) string {
	method := strings.TrimSpace(spacePattern.ReplaceAllString(raw, " "))
	method = strings.Trim(method, " .,-")
	lower := strings.ToLower(method)
	switch {
	case strings.Contains(lower, "decision - unanimous"):
		return "一致判定"
	case strings.Contains(lower, "decision - split"):
		return "分歧判定"
	case strings.Contains(lower, "decision - majority"):
		return "多数判定"
	case strings.Contains(lower, "ko/tko"):
		return "KO/TKO终结"
	case strings.Contains(lower, "submission"):
		return "降服"
	case strings.Contains(lower, "tko"):
		return "TKO终结"
	}
	return method
}

func formatAthleteFightHistory(date time.Time, fallbackDate string, result string, method string, round string, clock string) string {
	dateText := strings.TrimSpace(fallbackDate)
	if !date.IsZero() {
		dateText = date.Format("2006-01-02")
	}
	if dateText == "" || round == "" || clock == "" {
		return ""
	}
	parts := []string{dateText}
	if result != "" {
		parts = append(parts, result)
	}
	if method != "" {
		parts = append(parts, method)
	}
	parts = append(parts, "第"+round+"回合 "+clock)
	return strings.Join(parts, " · ")
}

func buildAthleteRecordMap(rawHTML string, record string) map[string]string {
	result := map[string]string{}
	if record != "" {
		result["Professional Record"] = record
	}
	labels := heroStatLabelPattern.FindAllStringSubmatch(rawHTML, -1)
	values := heroStatNumberPattern.FindAllStringSubmatch(rawHTML, -1)
	limit := len(labels)
	if len(values) < limit {
		limit = len(values)
	}
	for i := 0; i < limit; i++ {
		label := cleanText(labels[i][1])
		value := cleanText(values[i][1])
		if label == "" || value == "" {
			continue
		}
		result[label] = value
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func buildAthleteStatMap(rawHTML string) map[string]string {
	result := map[string]string{}
	labels := compareStatLabelPattern.FindAllStringSubmatch(rawHTML, -1)
	values := compareStatNumberPattern.FindAllStringSubmatch(rawHTML, -1)
	limit := len(labels)
	if len(values) < limit {
		limit = len(values)
	}
	for i := 0; i < limit; i++ {
		label := cleanText(labels[i][1])
		value := cleanText(values[i][1])
		if label == "" || value == "" {
			continue
		}
		result[label] = value
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func buildAthleteBioMap(rawHTML string) map[string]string {
	result := map[string]string{}
	labels := bioLabelPattern.FindAllStringSubmatch(rawHTML, -1)
	values := bioValuePattern.FindAllStringSubmatch(rawHTML, -1)
	limit := len(labels)
	if len(values) < limit {
		limit = len(values)
	}
	for i := 0; i < limit; i++ {
		label := cleanText(labels[i][1])
		value := cleanText(values[i][1])
		if label == "" || value == "" {
			continue
		}
		result[label] = value
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func inferAthleteCountry(bio map[string]string) string {
	if len(bio) == 0 {
		return ""
	}
	for key, value := range bio {
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "fighting out of", "place of birth":
			return countryFromBioValue(value)
		}
	}
	return ""
}

func inferCountryByLegacySnippet(rawHTML string) string {
	idx := strings.Index(strings.ToLower(rawHTML), "fighting out of")
	if idx <= -1 {
		return ""
	}
	text := cleanText(rawHTML[idx:])
	parts := strings.Split(text, "Fighting Out Of")
	if len(parts) < 2 {
		return ""
	}
	candidate := strings.TrimSpace(parts[1])
	lower := strings.ToLower(candidate)
	if cut := strings.Index(lower, "record"); cut > 0 {
		candidate = strings.TrimSpace(candidate[:cut])
	}
	if cut := strings.Index(lower, "status"); cut > 0 {
		candidate = strings.TrimSpace(candidate[:cut])
	}
	return countryFromBioValue(candidate)
}

func countryFromBioValue(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if strings.Contains(value, ",") {
		parts := strings.Split(value, ",")
		return strings.TrimSpace(parts[len(parts)-1])
	}
	words := strings.Fields(value)
	if len(words) == 0 {
		return ""
	}
	return strings.TrimSpace(words[len(words)-1])
}

func normalizeDivisionWeightClass(raw string) string {
	text := strings.TrimSpace(raw)
	text = strings.TrimSuffix(text, "Division")
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	return normalizeWeightClass(text)
}

func detectBaseURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "https://www.ufc.com"
	}
	return parsed.Scheme + "://" + parsed.Host
}

func toAbsURL(baseURL string, path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")
}

func cleanText(raw string) string {
	text := tagPattern.ReplaceAllString(raw, " ")
	text = stdhtml.UnescapeString(text)
	text = strings.TrimSpace(text)
	return spacePattern.ReplaceAllString(text, " ")
}

func mergeStringMap(base map[string]string, additions map[string]string) map[string]string {
	if len(additions) == 0 {
		return base
	}
	if base == nil {
		base = map[string]string{}
	}
	for key, value := range additions {
		normalizedKey := strings.TrimSpace(key)
		normalizedValue := strings.TrimSpace(value)
		if normalizedKey == "" || normalizedValue == "" {
			continue
		}
		if _, exists := base[normalizedKey]; exists {
			continue
		}
		base[normalizedKey] = normalizedValue
	}
	if len(base) == 0 {
		return nil
	}
	return base
}

func extractH1(html string) string {
	m := h1Pattern.FindStringSubmatch(html)
	if len(m) < 2 {
		return ""
	}
	return cleanText(m[1])
}

func extractTitle(html string) string {
	m := titlePattern.FindStringSubmatch(html)
	if len(m) < 2 {
		return ""
	}
	title := cleanText(m[1])
	title = strings.TrimSuffix(title, "| UFC")
	title = strings.TrimSpace(title)
	return title
}

func extractMetaOGImage(html string) string {
	m := metaOGImagePattern.FindStringSubmatch(html)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func extractEventPosterURL(rawHTML string, baseURL string) string {
	if ogImage := extractMetaOGImage(rawHTML); ogImage != "" {
		return ogImage
	}

	m := eventHeroImagePattern.FindStringSubmatch(rawHTML)
	if len(m) < 2 {
		return ""
	}

	src := strings.TrimSpace(stdhtml.UnescapeString(m[1]))
	if src == "" {
		return ""
	}
	return toAbsURL(baseURL, src)
}

func athleteNameFromURL(raw string) string {
	parts := strings.Split(strings.Trim(raw, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	last := strings.ReplaceAll(parts[len(parts)-1], "-", " ")
	return strings.Title(last)
}

func eventNameFromURL(raw string) string {
	parts := strings.Split(strings.Trim(raw, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	last := strings.ReplaceAll(parts[len(parts)-1], "-", " ")
	return strings.Title(last)
}

func parseDateFromEventURL(raw string) time.Time {
	months := map[string]time.Month{
		"january":   time.January,
		"february":  time.February,
		"march":     time.March,
		"april":     time.April,
		"may":       time.May,
		"june":      time.June,
		"july":      time.July,
		"august":    time.August,
		"september": time.September,
		"october":   time.October,
		"november":  time.November,
		"december":  time.December,
	}

	parts := strings.Split(strings.Trim(raw, "/"), "-")
	if len(parts) < 3 {
		return time.Time{}
	}
	yearRaw := parts[len(parts)-1]
	dayRaw := parts[len(parts)-2]
	monthRaw := parts[len(parts)-3]

	month, ok := months[strings.ToLower(monthRaw)]
	if !ok {
		return time.Time{}
	}
	day := atoi(dayRaw)
	year := atoi(yearRaw)
	if day <= 0 || year <= 0 {
		return time.Time{}
	}
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func atoi(raw string) int {
	n := 0
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return 0
		}
		n = n*10 + int(ch-'0')
	}
	return n
}
