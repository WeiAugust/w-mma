package ufc

import (
	"context"
	"errors"
	"io"
	"net/http"
	neturl "net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bajiaozhi/w-mma/backend/internal/source"
)

const staleEventCompletionWindow = 18 * time.Hour

var ErrUnsupportedSource = errors.New("unsupported source for ufc sync")

type SourceRepository interface {
	GetAny(ctx context.Context, sourceID int64) (source.DataSource, error)
	List(ctx context.Context, filter source.ListFilter) ([]source.DataSource, error)
}

type Store interface {
	UpsertEvent(ctx context.Context, item EventRecord) (int64, error)
	UpsertFighter(ctx context.Context, item FighterRecord) (int64, error)
	ReplaceEventBouts(ctx context.Context, eventID int64, bouts []BoutRecord) error
}

type Service struct {
	sourceRepo  SourceRepository
	store       Store
	scraper     Scraper
	imageMirror ImageMirror
	now         func() time.Time
}

type ServiceOption func(*Service)

func WithImageMirror(mirror ImageMirror) ServiceOption {
	return func(s *Service) {
		if mirror != nil {
			s.imageMirror = mirror
		}
	}
}

func NewService(sourceRepo SourceRepository, store Store, scraper Scraper, opts ...ServiceOption) *Service {
	svc := &Service{
		sourceRepo:  sourceRepo,
		store:       store,
		scraper:     scraper,
		imageMirror: passthroughImageMirror{},
		now:         time.Now,
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

func (s *Service) SyncSource(ctx context.Context, sourceID int64) (SyncResult, error) {
	src, err := s.sourceRepo.GetAny(ctx, sourceID)
	if err != nil {
		return SyncResult{}, err
	}
	if src.SourceURL == "" {
		return SyncResult{}, ErrUnsupportedSource
	}
	switch src.ParserKind {
	case "ufc_schedule":
		return s.syncScheduleSource(ctx, src)
	case "ufc_athletes":
		return s.syncAthletesSource(ctx, src)
	default:
		return SyncResult{}, ErrUnsupportedSource
	}
}

func (s *Service) syncScheduleSource(ctx context.Context, src source.DataSource) (SyncResult, error) {
	eventLinks, err := s.scraper.ListEventLinks(ctx, src.SourceURL)
	if err != nil {
		return SyncResult{}, err
	}

	result := SyncResult{}
	for _, eventLink := range eventLinks {
		card, err := s.scraper.GetEventCard(ctx, eventLink.URL)
		if err != nil {
			continue
		}
		startsAt := eventLink.StartsAt
		if startsAt.IsZero() {
			startsAt = card.StartsAt
		}
		if startsAt.IsZero() {
			startsAt = s.now()
		}
		status := normalizeEventStatus(card.Status, startsAt, s.now().UTC())
		posterURL := chooseNonEmpty(card.PosterURL, eventLink.PosterURL)
		posterURL = s.mirrorImageURL(ctx, posterURL)
		eventID, err := s.store.UpsertEvent(ctx, EventRecord{
			SourceID:    src.ID,
			Org:         "UFC",
			Name:        chooseNonEmpty(card.Name, eventLink.Name),
			Status:      status,
			StartsAt:    startsAt,
			Venue:       chooseNonEmpty(card.Venue, "TBD"),
			PosterURL:   posterURL,
			ExternalURL: eventLink.URL,
		})
		if err != nil {
			continue
		}
		result.Events++

		bouts := make([]BoutRecord, 0, len(card.Bouts))
		for _, bout := range card.Bouts {
			redProfile, err := s.scraper.GetAthleteProfile(ctx, bout.RedURL)
			if err != nil {
				continue
			}
			blueProfile, err := s.scraper.GetAthleteProfile(ctx, bout.BlueURL)
			if err != nil {
				continue
			}

			redID, err := s.store.UpsertFighter(ctx, FighterRecord{
				SourceID:    src.ID,
				Name:        chooseNonEmpty(redProfile.Name, bout.RedName),
				NameZH:      strings.TrimSpace(redProfile.NameZH),
				Nickname:    strings.TrimSpace(redProfile.Nickname),
				Country:     redProfile.Country,
				Record:      redProfile.Record,
				WeightClass: chooseNonEmpty(bout.WeightClass, redProfile.WeightClass),
				AvatarURL:   s.mirrorImageURL(ctx, redProfile.AvatarURL),
				ExternalURL: redProfile.URL,
				Stats:       redProfile.Stats,
				Records:     redProfile.Records,
				Updates:     redProfile.Updates,
			})
			if err != nil {
				continue
			}
			blueID, err := s.store.UpsertFighter(ctx, FighterRecord{
				SourceID:    src.ID,
				Name:        chooseNonEmpty(blueProfile.Name, bout.BlueName),
				NameZH:      strings.TrimSpace(blueProfile.NameZH),
				Nickname:    strings.TrimSpace(blueProfile.Nickname),
				Country:     blueProfile.Country,
				Record:      blueProfile.Record,
				WeightClass: chooseNonEmpty(bout.WeightClass, blueProfile.WeightClass),
				AvatarURL:   s.mirrorImageURL(ctx, blueProfile.AvatarURL),
				ExternalURL: blueProfile.URL,
				Stats:       blueProfile.Stats,
				Records:     blueProfile.Records,
				Updates:     blueProfile.Updates,
			})
			if err != nil {
				continue
			}
			result.Fighters += 2
			winnerID := int64(0)
			switch strings.ToLower(strings.TrimSpace(bout.WinnerSide)) {
			case "red":
				winnerID = redID
			case "blue":
				winnerID = blueID
			}
			bouts = append(bouts, BoutRecord{
				RedFighterID:  redID,
				BlueFighterID: blueID,
				CardSegment:   bout.CardSegment,
				WeightClass:   bout.WeightClass,
				RedRanking:    bout.RedRank,
				BlueRanking:   bout.BlueRank,
				Result:        bout.Result,
				WinnerID:      winnerID,
				Method:        bout.Method,
				Round:         bout.Round,
				TimeSec:       bout.TimeSec,
			})
		}

		if len(bouts) > 0 {
			if err := s.store.ReplaceEventBouts(ctx, eventID, bouts); err == nil {
				result.Bouts += len(bouts)
			}
		}
	}

	return result, nil
}

func (s *Service) syncAthletesSource(ctx context.Context, src source.DataSource) (SyncResult, error) {
	links, err := s.scraper.ListAthleteLinks(ctx, src.SourceURL)
	if err != nil {
		return SyncResult{}, err
	}
	result := SyncResult{}
	for _, link := range links {
		profile, err := s.scraper.GetAthleteProfile(ctx, link)
		if err != nil {
			continue
		}
		profile = s.enrichSingleAthleteProfile(ctx, link, profile)
		if _, err := s.store.UpsertFighter(ctx, FighterRecord{
			SourceID:    src.ID,
			Name:        chooseNonEmpty(profile.Name, athleteNameFromURL(link)),
			NameZH:      strings.TrimSpace(profile.NameZH),
			Nickname:    strings.TrimSpace(profile.Nickname),
			Country:     profile.Country,
			Record:      profile.Record,
			WeightClass: profile.WeightClass,
			AvatarURL:   s.mirrorImageURL(ctx, profile.AvatarURL),
			ExternalURL: chooseNonEmpty(profile.URL, link),
			Stats:       profile.Stats,
			Records:     profile.Records,
			Updates:     profile.Updates,
		}); err != nil {
			continue
		}
		result.Fighters++
	}
	return result, nil
}

func (s *Service) SyncSingleAthlete(ctx context.Context, sourceID int64, athleteURL string) (SyncResult, error) {
	src, err := s.sourceRepo.GetAny(ctx, sourceID)
	if err != nil {
		return SyncResult{}, err
	}
	if src.ParserKind != "ufc_athletes" && src.ParserKind != "ufc_schedule" {
		return SyncResult{}, ErrUnsupportedSource
	}
	athleteURL = strings.TrimSpace(athleteURL)
	if athleteURL == "" {
		return SyncResult{}, ErrUnsupportedSource
	}

	profile, err := s.scraper.GetAthleteProfile(ctx, athleteURL)
	if err != nil {
		return SyncResult{}, err
	}
	profile = s.enrichSingleAthleteProfile(ctx, athleteURL, profile)
	if _, err := s.store.UpsertFighter(ctx, FighterRecord{
		SourceID:    src.ID,
		Name:        chooseNonEmpty(profile.Name, athleteNameFromURL(athleteURL)),
		NameZH:      strings.TrimSpace(profile.NameZH),
		Nickname:    strings.TrimSpace(profile.Nickname),
		Country:     profile.Country,
		Record:      profile.Record,
		WeightClass: profile.WeightClass,
		AvatarURL:   s.mirrorImageURL(ctx, profile.AvatarURL),
		ExternalURL: chooseNonEmpty(profile.URL, athleteURL),
		Stats:       profile.Stats,
		Records:     profile.Records,
		Updates:     profile.Updates,
	}); err != nil {
		return SyncResult{}, err
	}
	return SyncResult{Fighters: 1}, nil
}

func (s *Service) enrichSingleAthleteProfile(ctx context.Context, athleteURL string, profile AthleteProfile) AthleteProfile {
	if !needsMirrorEnrichment(profile) {
		return profile
	}

	mirrorRaw := ""
	if raw, err := fetchAthleteMirrorText(ctx, athleteURL); err == nil {
		mirrorRaw = raw
	}

	if missingSingleAthleteHeroFields(profile) {
		if strings.TrimSpace(mirrorRaw) != "" {
			if profile.WeightClass == "" {
				if m := divisionWeightClassPattern.FindStringSubmatch(mirrorRaw); len(m) > 1 {
					profile.WeightClass = normalizeWeightClass(m[1])
				}
			}
			if profile.Stats == nil {
				profile.Stats = map[string]string{}
			}
			if profile.Stats["PFP Rank"] == "" {
				if m := athletePFPRankPattern.FindStringSubmatch(mirrorRaw); len(m) > 1 {
					profile.Stats["PFP Rank"] = "#" + strings.TrimSpace(m[1])
				}
			}
			lower := strings.ToLower(mirrorRaw)
			if profile.Stats["Title Status"] == "" && strings.Contains(lower, "title holder") {
				profile.Stats["Title Status"] = "Title Holder"
			}
			if profile.Stats["Athlete Status"] == "" && profile.Stats["Status"] == "" {
				if strings.Contains(lower, " active ") {
					profile.Stats["Athlete Status"] = "Active"
				}
			}
		}
	}
	profile.Updates = mergeAthleteUpdates(profile.Updates, collectAthleteMirrorUpdates(ctx, athleteURL, 8))
	profile.Updates = fillFightResultsFromNarrative(profile.Updates, mirrorRaw)
	return profile
}

func needsMirrorEnrichment(profile AthleteProfile) bool {
	if missingSingleAthleteHeroFields(profile) {
		return true
	}
	if len(profile.Updates) == 0 {
		return true
	}
	// A short visible list usually means only first-page fight history was parsed.
	if len(profile.Updates) < 8 {
		return true
	}
	for _, item := range profile.Updates {
		content := strings.TrimSpace(item.Content)
		if content == "" {
			return true
		}
		if !strings.Contains(content, " · 胜 · ") &&
			!strings.Contains(content, " · 负 · ") &&
			!strings.Contains(content, " · 平 · ") &&
			!strings.Contains(content, " · 无结果 · ") {
			return true
		}
	}
	return false
}

func missingSingleAthleteHeroFields(profile AthleteProfile) bool {
	if strings.TrimSpace(profile.WeightClass) == "" {
		return true
	}
	if profile.Stats == nil {
		return true
	}
	if strings.TrimSpace(profile.Stats["PFP Rank"]) == "" {
		return true
	}
	if strings.TrimSpace(profile.Stats["Title Status"]) == "" {
		return true
	}
	return false
}

func fetchAthleteMirrorText(ctx context.Context, athleteURL string) (string, error) {
	athleteURL = strings.TrimSpace(athleteURL)
	if athleteURL == "" {
		return "", errors.New("empty athlete url")
	}
	parsed, err := neturl.Parse(athleteURL)
	if err != nil || parsed.Host == "" {
		return "", errors.New("invalid athlete url")
	}
	normalized := athleteURL
	if parsed.Scheme == "" {
		normalized = "https://" + strings.TrimLeft(athleteURL, "/")
	}
	mirrorURL := "https://r.jina.ai/http://" + strings.TrimPrefix(strings.TrimPrefix(normalized, "https://"), "http://")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mirrorURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; w-mma-bot/1.0)")
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", errors.New("fetch mirror failed with status " + resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 3<<20))
	if err != nil {
		return "", err
	}
	return string(body), nil
}

var mirrorLoadMorePattern = regexp.MustCompile(`(?is)\[\s*Load More\s*\]\(([^)\s]+)`)
var mirrorFightHistoryPattern = regexp.MustCompile(`(?is)([A-Z][a-z]{2}\.\s+\d{1,2},\s+\d{4})\s+Round\s+(\d+)\s+Time\s+(\d{1,2}:\d{2})\s+Method\s+(.+?)(?:\[\s*Watch Replay|\[\s*Fight Card|$)`)

func collectAthleteMirrorUpdates(ctx context.Context, athleteURL string, maxPages int) []AthleteUpdate {
	return collectAthleteMirrorUpdatesWithFetcher(ctx, athleteURL, maxPages, fetchAthleteMirrorText)
}

func collectAthleteMirrorUpdatesWithFetcher(
	ctx context.Context,
	athleteURL string,
	maxPages int,
	fetch func(context.Context, string) (string, error),
) []AthleteUpdate {
	if maxPages <= 0 {
		maxPages = 1
	}
	baseURL := strings.TrimSpace(athleteURL)
	if baseURL == "" {
		return nil
	}
	collected := make([]AthleteUpdate, 0, 16)
	stagnantPages := 0

	for page := 0; page < maxPages; page++ {
		targetURL := athleteURLWithPage(baseURL, page)
		raw, err := fetch(ctx, targetURL)
		if err != nil || strings.TrimSpace(raw) == "" {
			break
		}
		before := len(collected)
		pageUpdates := buildAthleteFightHistoryFromMirror(raw)
		if len(pageUpdates) == 0 {
			pageUpdates = buildAthleteFightHistory(raw)
		}
		collected = mergeAthleteUpdates(collected, pageUpdates)
		after := len(collected)

		if after == before {
			stagnantPages++
		} else {
			stagnantPages = 0
		}
		if page > 0 && stagnantPages >= 2 {
			break
		}
	}

	if len(collected) == 0 {
		return nil
	}
	sort.SliceStable(collected, func(i, j int) bool {
		return collected[i].PublishedAt.After(collected[j].PublishedAt)
	})
	return collected
}

func athleteURLWithPage(raw string, page int) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	parsed, err := neturl.Parse(trimmed)
	if err != nil {
		return trimmed
	}
	query := parsed.Query()
	if page <= 0 {
		query.Del("page")
	} else {
		query.Set("page", strconv.Itoa(page))
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func extractMirrorLoadMoreURL(raw string, currentURL string) string {
	match := mirrorLoadMorePattern.FindStringSubmatch(raw)
	if len(match) < 2 {
		return ""
	}
	next := strings.TrimSpace(match[1])
	if next == "" {
		return ""
	}
	if strings.HasPrefix(next, "http://") || strings.HasPrefix(next, "https://") {
		return next
	}
	return toAbsURL(detectBaseURL(currentURL), next)
}

func mergeAthleteUpdates(base []AthleteUpdate, additions []AthleteUpdate) []AthleteUpdate {
	if len(additions) == 0 {
		return base
	}
	seen := map[string]struct{}{}
	result := make([]AthleteUpdate, 0, len(base)+len(additions))

	for _, item := range base {
		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}
		if _, exists := seen[content]; exists {
			continue
		}
		seen[content] = struct{}{}
		result = append(result, item)
	}
	for _, item := range additions {
		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}
		if _, exists := seen[content]; exists {
			continue
		}
		seen[content] = struct{}{}
		result = append(result, item)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

var ufcHistoryLinePattern = regexp.MustCompile(`(?im)\*\*UFC[^\n]*\*\*\s*\((\d{1,2}/\d{1,2}/\d{2})\)\s*([^\n]+)`)

func fillFightResultsFromNarrative(items []AthleteUpdate, mirrorRaw string) []AthleteUpdate {
	if len(items) == 0 || strings.TrimSpace(mirrorRaw) == "" {
		return items
	}
	resultByDate := parseNarrativeResultMap(mirrorRaw)
	if len(resultByDate) == 0 {
		return items
	}
	out := make([]AthleteUpdate, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}
		dateKey := ""
		if len(content) >= 10 {
			dateKey = content[:10]
		}
		result := resultByDate[dateKey]
		if result == "" {
			if _, exists := seen[content]; !exists {
				seen[content] = struct{}{}
				out = append(out, item)
			}
			continue
		}
		item.Content = applyNarrativeFightResult(content, dateKey, result)
		if strings.TrimSpace(item.Content) == "" {
			continue
		}
		if _, exists := seen[item.Content]; exists {
			continue
		}
		seen[item.Content] = struct{}{}
		out = append(out, item)
	}
	return out
}

func applyNarrativeFightResult(content string, dateKey string, result string) string {
	partsRaw := strings.Split(content, "·")
	if len(partsRaw) == 0 {
		return content
	}
	parts := make([]string, 0, len(partsRaw))
	for _, part := range partsRaw {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		parts = append(parts, trimmed)
	}
	if len(parts) < 2 {
		return content
	}
	if dateKey == "" {
		dateKey = parts[0]
	}
	if parts[0] != dateKey {
		return content
	}
	if isFightResultValue(parts[1]) {
		parts[1] = result
		return strings.Join(parts, " · ")
	}
	parts = append(parts[:1], append([]string{result}, parts[1:]...)...)
	return strings.Join(parts, " · ")
}

func isFightResultValue(value string) bool {
	switch strings.TrimSpace(value) {
	case "胜", "负", "平", "无结果":
		return true
	default:
		return false
	}
}

func parseNarrativeResultMap(raw string) map[string]string {
	idx := strings.Index(strings.ToLower(raw), "ufc history")
	if idx < 0 {
		return nil
	}
	chunk := raw[idx:]
	if len(chunk) > 80000 {
		chunk = chunk[:80000]
	}
	matches := ufcHistoryLinePattern.FindAllStringSubmatch(chunk, -1)
	if len(matches) == 0 {
		return nil
	}
	resultByDate := map[string]string{}
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		date := parseNarrativeDate(match[1])
		if date == "" {
			continue
		}
		desc := strings.ToLower(match[2])
		result := ""
		switch {
		case strings.Contains(desc, " was knocked out"), strings.Contains(desc, " lost "), strings.Contains(desc, "losing "):
			result = "负"
		case strings.Contains(desc, " won "), strings.Contains(desc, " stopped "), strings.Contains(desc, " knocked out "):
			result = "胜"
		}
		if result == "" {
			continue
		}
		resultByDate[date] = result
	}
	if len(resultByDate) == 0 {
		return nil
	}
	return resultByDate
}

func parseNarrativeDate(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}
	parsed, err := time.Parse("1/2/06", text)
	if err != nil {
		return ""
	}
	return parsed.UTC().Format("2006-01-02")
}

func buildAthleteFightHistoryFromMirror(raw string) []AthleteUpdate {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil
	}
	idx := strings.Index(strings.ToLower(text), "athlete record")
	if idx < 0 {
		return nil
	}
	chunk := strings.TrimSpace(text[idx:])
	if len(chunk) > 80000 {
		chunk = chunk[:80000]
	}

	matches := mirrorFightHistoryPattern.FindAllStringSubmatchIndex(chunk, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := map[string]struct{}{}
	items := make([]AthleteUpdate, 0, len(matches))
	for _, match := range matches {
		if len(match) < 10 {
			continue
		}
		dateStart := match[2]
		resultToken := detectFightResultAroundMirror(chunk, dateStart)
		result := normalizeFightResult(resultToken)
		dateText := strings.TrimSpace(chunk[match[2]:match[3]])
		round := strings.TrimSpace(chunk[match[4]:match[5]])
		clock := strings.TrimSpace(chunk[match[6]:match[7]])
		method := normalizeFightMethod(chunk[match[8]:match[9]])
		date := parseAthleteFightDate(dateText)
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
	}
	if len(items) == 0 {
		return nil
	}
	return items
}

func detectFightResultAroundMirror(chunk string, dateStart int) string {
	if dateStart < 0 {
		return ""
	}
	windowStart := dateStart - 260
	if windowStart < 0 {
		windowStart = 0
	}
	windowEnd := dateStart + 420
	if windowEnd > len(chunk) {
		windowEnd = len(chunk)
	}
	window := chunk[windowStart:windowEnd]
	matches := fightResultTokenPattern.FindAllStringSubmatchIndex(window, -1)
	if len(matches) == 0 {
		return ""
	}
	target := dateStart - windowStart
	bestText := ""
	bestDistance := len(window) + 1
	for _, item := range matches {
		if len(item) < 4 {
			continue
		}
		tokenStart := item[2]
		tokenEnd := item[3]
		if tokenStart < 0 || tokenEnd > len(window) || tokenStart >= tokenEnd {
			continue
		}
		distance := tokenStart - target
		if distance < 0 {
			distance = -distance
		}
		if distance >= bestDistance {
			continue
		}
		bestDistance = distance
		bestText = strings.TrimSpace(window[tokenStart:tokenEnd])
	}
	return bestText
}

func (s *Service) SyncEnabledSources(ctx context.Context) (SyncResult, error) {
	enabled := true
	items, err := s.sourceRepo.List(ctx, source.ListFilter{
		Platform: "ufc",
		Enabled:  &enabled,
	})
	if err != nil {
		return SyncResult{}, err
	}

	total := SyncResult{}
	for _, item := range items {
		result, err := s.SyncSource(ctx, item.ID)
		if err != nil {
			continue
		}
		total.Events += result.Events
		total.Bouts += result.Bouts
		total.Fighters += result.Fighters
	}
	return total, nil
}

func (s *Service) mirrorImageURL(ctx context.Context, rawURL string) string {
	if strings.TrimSpace(rawURL) == "" {
		return ""
	}
	mirroredURL, err := s.imageMirror.MirrorImage(ctx, rawURL)
	if err != nil {
		// Mirror failures should not keep third-party URLs in miniapp payloads.
		return ""
	}
	return strings.TrimSpace(mirroredURL)
}

func chooseNonEmpty(value string, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func normalizeEventStatus(raw string, startsAt time.Time, now time.Time) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "completed", "final":
		return "completed"
	case "live":
		if !startsAt.IsZero() && startsAt.Before(now.Add(-staleEventCompletionWindow)) {
			return "completed"
		}
		return "live"
	case "scheduled", "upcoming":
		if startsAt.IsZero() || startsAt.After(now) {
			return "scheduled"
		}
		if startsAt.Before(now.Add(-staleEventCompletionWindow)) {
			return "completed"
		}
		return "live"
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
