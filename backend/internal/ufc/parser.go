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
	eventHrefPattern      = regexp.MustCompile(`href=["']((?:https?://[^"'#?]+)?/event/[^"'#?]+)["']`)
	athleteHrefPattern    = regexp.MustCompile(`href=["']((?:https?://[^"'#?]+)?/athlete/[^"'#?]+)["']`)
	h1Pattern             = regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`)
	titlePattern          = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	tagPattern            = regexp.MustCompile(`(?is)<[^>]+>`)
	spacePattern          = regexp.MustCompile(`\s+`)
	metaOGImagePattern    = regexp.MustCompile(`(?is)<meta[^>]+property=["']og:image["'][^>]+content=["']([^"']+)["']`)
	eventHeroImagePattern = regexp.MustCompile(`(?is)<img[^>]+src=["']([^"']*/images/styles/background_image[^"']+)["']`)
	recordPattern         = regexp.MustCompile(`\b\d{1,2}-\d{1,2}(?:-\d{1,2})?\b`)
	mainCardPattern       = regexp.MustCompile(`(?i)\bmain\s+card\b`)
	prelimsPattern        = regexp.MustCompile(`(?i)\bprelims\b|\bearly\s+prelims\b`)
	weightClassPattern    = regexp.MustCompile(`(?i)(women['’]s\s+strawweight|women['’]s\s+flyweight|women['’]s\s+bantamweight|women['’]s\s+featherweight|strawweight|flyweight|bantamweight|featherweight|lightweight|welterweight|middleweight|light\s+heavyweight|heavyweight|catchweight)\s+bout`)
	rankingPattern        = regexp.MustCompile(`#\s*\d+`)
	fightCardTimePattern  = regexp.MustCompile(`(?is)field--name-fight-card-time[^>]*>.*?<time[^>]+datetime=["']([^"']+)["']`)
	eventLiveFinalPattern = regexp.MustCompile(`(?is)"eventLiveStats"\s*:\s*\{[^{}]*"final"\s*:\s*(true|false)`)
	fightBlockPattern     = regexp.MustCompile(`(?is)<div class="c-listing-fight"[^>]*>`)
	roundTextPattern      = regexp.MustCompile(`(?is)c-listing-fight__result-text\s+round[^>]*>([^<]*)<`)
	timeTextPattern       = regexp.MustCompile(`(?is)c-listing-fight__result-text\s+time[^>]*>([^<]*)<`)
	methodTextPattern     = regexp.MustCompile(`(?is)c-listing-fight__result-text\s+method[^>]*>([^<]*)<`)
	redWinnerPattern      = regexp.MustCompile(`(?is)c-listing-fight__corner-body--red.*?c-listing-fight__outcome--win`)
	blueWinnerPattern     = regexp.MustCompile(`(?is)c-listing-fight__corner-body--blue.*?c-listing-fight__outcome--win`)
	eventCardBlockPattern = regexp.MustCompile(`(?is)<article class="c-card-event--result"[^>]*>`)
	headlinePattern       = regexp.MustCompile(`(?is)c-card-event--result__headline[^>]*>\s*<a[^>]*>(.*?)</a>`)
	mainCardTsPattern     = regexp.MustCompile(`(?is)data-main-card-timestamp=["'](\d+)["']`)
	prelimsCardTsPattern  = regexp.MustCompile(`(?is)data-prelims-card-timestamp=["'](\d+)["']`)
	earlyCardTsPattern    = regexp.MustCompile(`(?is)data-early-card-timestamp=["'](\d+)["']`)
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
	if !mainTs.IsZero() {
		return mainTs
	}
	prelimsTs := parseUnixTimestampFromChunk(chunk, prelimsCardTsPattern)
	if !prelimsTs.IsZero() {
		return prelimsTs
	}
	return parseUnixTimestampFromChunk(chunk, earlyCardTsPattern)
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
	redWin := redWinnerPattern.MatchString(chunk)
	blueWin := blueWinnerPattern.MatchString(chunk)
	if redWin && !blueWin {
		return "red"
	}
	if blueWin && !redWin {
		return "blue"
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
	if final, ok := extractEventLiveFinal(rawHTML); ok {
		if final {
			return "completed"
		}
		if !startsAt.IsZero() && startsAt.Before(time.Now().UTC().Add(-6*time.Hour)) {
			return "completed"
		}
		return "scheduled"
	}
	for _, bout := range bouts {
		if bout.WinnerSide != "" || bout.Result != "" || bout.Method != "" || bout.Round > 0 || bout.TimeSec > 0 {
			return "completed"
		}
	}
	if startsAt.IsZero() {
		return "scheduled"
	}
	if startsAt.After(time.Now().UTC()) {
		return "scheduled"
	}
	return "completed"
}

func extractEventLiveFinal(rawHTML string) (bool, bool) {
	m := eventLiveFinalPattern.FindStringSubmatch(rawHTML)
	if len(m) < 2 {
		return false, false
	}
	return strings.EqualFold(strings.TrimSpace(m[1]), "true"), true
}

func extractEventStartsAt(rawHTML string) time.Time {
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

	record := recordPattern.FindString(html)
	country := ""
	if idx := strings.Index(strings.ToLower(html), "fighting out of"); idx > -1 {
		snippet := html[idx:]
		text := cleanText(snippet)
		parts := strings.Split(text, "Fighting Out Of")
		if len(parts) > 1 {
			country = strings.TrimSpace(strings.Fields(parts[1])[0])
		}
	}
	if country == "" {
		country = "Unknown"
	}

	return AthleteProfile{
		Name:      name,
		URL:       athleteURL,
		Country:   country,
		Record:    record,
		AvatarURL: extractMetaOGImage(html),
	}
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
	text = strings.TrimSpace(text)
	return spacePattern.ReplaceAllString(text, " ")
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
