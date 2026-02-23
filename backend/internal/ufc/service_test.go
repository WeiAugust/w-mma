package ufc

import (
	"context"
	"testing"
	"time"

	"github.com/bajiaozhi/w-mma/backend/internal/source"
)

type fakeSourceRepo struct {
	item source.DataSource
	err  error
}

func (f *fakeSourceRepo) GetAny(_ context.Context, _ int64) (source.DataSource, error) {
	return f.item, f.err
}

func (f *fakeSourceRepo) List(_ context.Context, _ source.ListFilter) ([]source.DataSource, error) {
	return []source.DataSource{f.item}, nil
}

type fakeScraper struct{}

func (f fakeScraper) ListEventLinks(context.Context, string) ([]EventLink, error) {
	return []EventLink{
		{
			Name:     "UFC Fight Night: A vs B",
			URL:      "https://www.ufc.com/event/ufc-fight-night-a-b",
			StartsAt: time.Date(2026, 2, 21, 0, 0, 0, 0, time.UTC),
		},
	}, nil
}

func (f fakeScraper) GetEventCard(context.Context, string) (EventCard, error) {
	return EventCard{
		Name:     "UFC Fight Night: A vs B",
		Status:   "upcoming",
		StartsAt: time.Date(2026, 2, 21, 0, 0, 0, 0, time.UTC),
		Bouts: []EventBout{
			{
				RedName: "Fighter A", RedURL: "https://www.ufc.com/athlete/fighter-a", RedRank: "#6",
				BlueName: "Fighter B", BlueURL: "https://www.ufc.com/athlete/fighter-b", BlueRank: "#9",
				CardSegment: "main_card", WeightClass: "Flyweight",
				WinnerSide: "red", Result: "KO/TKO R2 1:40", Method: "KO/TKO", Round: 2, TimeSec: 100,
			},
		},
	}, nil
}

func (f fakeScraper) ListAthleteLinks(context.Context, string) ([]string, error) {
	return []string{
		"https://www.ufc.com/athlete/fighter-a",
		"https://www.ufc.com/athlete/fighter-b",
	}, nil
}

func (f fakeScraper) GetAthleteProfile(_ context.Context, url string) (AthleteProfile, error) {
	if url == "https://www.ufc.com/athlete/fighter-a" {
		return AthleteProfile{Name: "Fighter A", URL: url, Country: "USA", Record: "10-1-0"}, nil
	}
	return AthleteProfile{Name: "Fighter B", URL: url, Country: "Brazil", Record: "15-3-0"}, nil
}

type trackingScraper struct {
	fakeScraper
	listAthletesCalled bool
}

func (s *trackingScraper) ListAthleteLinks(context.Context, string) ([]string, error) {
	s.listAthletesCalled = true
	return []string{}, nil
}

type mismatchedStartsAtScraper struct {
	fakeScraper
}

func (s mismatchedStartsAtScraper) ListEventLinks(context.Context, string) ([]EventLink, error) {
	return []EventLink{
		{
			Name:     "UFC 326",
			URL:      "https://www.ufc.com/event/ufc-326",
			StartsAt: time.Date(2026, 3, 7, 22, 0, 0, 0, time.UTC),
		},
	}, nil
}

func (s mismatchedStartsAtScraper) GetEventCard(context.Context, string) (EventCard, error) {
	card, _ := s.fakeScraper.GetEventCard(context.Background(), "")
	card.StartsAt = time.Date(2026, 3, 7, 17, 0, 0, 0, time.UTC)
	return card, nil
}

type athletesOnlyScraper struct {
	listEventsCalled bool
}

func (s *athletesOnlyScraper) ListEventLinks(context.Context, string) ([]EventLink, error) {
	s.listEventsCalled = true
	return nil, nil
}

func (s *athletesOnlyScraper) GetEventCard(context.Context, string) (EventCard, error) {
	return EventCard{}, nil
}

func (s *athletesOnlyScraper) ListAthleteLinks(context.Context, string) ([]string, error) {
	return []string{
		"https://www.ufc.com/athlete/fighter-a",
		"https://www.ufc.com/athlete/fighter-b",
	}, nil
}

func (s *athletesOnlyScraper) GetAthleteProfile(_ context.Context, url string) (AthleteProfile, error) {
	if url == "https://www.ufc.com/athlete/fighter-a" {
		return AthleteProfile{
			Name:        "Fighter A",
			URL:         url,
			Nickname:    "Alpha",
			Country:     "USA",
			Record:      "11-2-0",
			WeightClass: "Flyweight",
			Stats: map[string]string{
				"Sig. Str. Landed": "3.01",
			},
			Records: map[string]string{
				"Wins by Knockout": "4",
			},
		}, nil
	}
	return AthleteProfile{
		Name:        "Fighter B",
		URL:         url,
		Nickname:    "Bravo",
		Country:     "Brazil",
		Record:      "14-1-0",
		WeightClass: "Flyweight",
	}, nil
}

type fakeStore struct {
	events    int
	fighters  int
	bouts     int
	lastBouts []BoutRecord
	lastEvent EventRecord
	lastFight FighterRecord
	allFights []FighterRecord
}

func (f *fakeStore) UpsertEvent(_ context.Context, item EventRecord) (int64, error) {
	f.events++
	f.lastEvent = item
	return int64(100 + f.events), nil
}

func (f *fakeStore) UpsertFighter(_ context.Context, item FighterRecord) (int64, error) {
	f.fighters++
	f.lastFight = item
	f.allFights = append(f.allFights, item)
	return int64(200 + f.fighters), nil
}

func (f *fakeStore) ReplaceEventBouts(_ context.Context, _ int64, bouts []BoutRecord) error {
	f.bouts += len(bouts)
	f.lastBouts = bouts
	return nil
}

func TestSyncSource_PersistsEventsBoutsAndFighters(t *testing.T) {
	svc := NewService(
		&fakeSourceRepo{
			item: source.DataSource{
				ID:         1,
				SourceType: "schedule",
				Platform:   "ufc",
				SourceURL:  "https://www.ufc.com/events",
				ParserKind: "ufc_schedule",
				Enabled:    true,
			},
		},
		&fakeStore{},
		fakeScraper{},
	)
	svc.now = func() time.Time {
		return time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	}

	result, err := svc.SyncSource(context.Background(), 1)
	if err != nil {
		t.Fatalf("sync source: %v", err)
	}
	if result.Events == 0 {
		t.Fatalf("expected events to be synced")
	}
	if svc.store.(*fakeStore).lastEvent.Status != "scheduled" {
		t.Fatalf("expected event status normalized to scheduled, got %q", svc.store.(*fakeStore).lastEvent.Status)
	}
	if result.Bouts == 0 {
		t.Fatalf("expected bouts to be synced")
	}
	if result.Fighters == 0 {
		t.Fatalf("expected fighters to be synced")
	}
	if len(svc.store.(*fakeStore).lastBouts) == 0 {
		t.Fatalf("expected persisted bouts")
	}
	bout := svc.store.(*fakeStore).lastBouts[0]
	if bout.CardSegment == "" {
		t.Fatalf("expected card segment to be persisted")
	}
	if bout.WeightClass == "" {
		t.Fatalf("expected weight class to be persisted")
	}
	if bout.RedRanking != "#6" || bout.BlueRanking != "#9" {
		t.Fatalf("expected rankings to be persisted, got red=%q blue=%q", bout.RedRanking, bout.BlueRanking)
	}
	if bout.WinnerID <= 0 {
		t.Fatalf("expected winner id to be persisted")
	}
	if bout.Method != "KO/TKO" || bout.Round != 2 || bout.TimeSec != 100 {
		t.Fatalf("expected result meta to be persisted, got method=%q round=%d timeSec=%d", bout.Method, bout.Round, bout.TimeSec)
	}
	if bout.Result == "" {
		t.Fatalf("expected result text to be persisted")
	}
}

func TestSyncSource_PrefersEventLinkStartsAt(t *testing.T) {
	store := &fakeStore{}
	svc := NewService(
		&fakeSourceRepo{
			item: source.DataSource{
				ID:         1,
				SourceType: "schedule",
				Platform:   "ufc",
				SourceURL:  "https://www.ufc.com/events",
				ParserKind: "ufc_schedule",
				Enabled:    true,
			},
		},
		store,
		mismatchedStartsAtScraper{},
	)
	svc.now = func() time.Time {
		return time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	}

	if _, err := svc.SyncSource(context.Background(), 1); err != nil {
		t.Fatalf("sync source: %v", err)
	}
	expected := time.Date(2026, 3, 7, 22, 0, 0, 0, time.UTC)
	if !store.lastEvent.StartsAt.Equal(expected) {
		t.Fatalf("expected event link starts_at %s, got %s", expected.Format(time.RFC3339), store.lastEvent.StartsAt.Format(time.RFC3339))
	}
}

func TestSyncSource_DoesNotCrawlAthleteDirectory(t *testing.T) {
	scraper := &trackingScraper{}
	svc := NewService(
		&fakeSourceRepo{
			item: source.DataSource{
				ID:         1,
				SourceType: "schedule",
				Platform:   "ufc",
				SourceURL:  "https://www.ufc.com/events",
				ParserKind: "ufc_schedule",
				Enabled:    true,
			},
		},
		&fakeStore{},
		scraper,
	)

	if _, err := svc.SyncSource(context.Background(), 1); err != nil {
		t.Fatalf("sync source: %v", err)
	}
	if scraper.listAthletesCalled {
		t.Fatalf("expected sync source to skip athlete directory crawl")
	}
}

func TestSyncSource_AthleteParserSyncsFightersOnly(t *testing.T) {
	store := &fakeStore{}
	scraper := &athletesOnlyScraper{}
	svc := NewService(
		&fakeSourceRepo{
			item: source.DataSource{
				ID:         2,
				SourceType: "fighter",
				Platform:   "ufc",
				SourceURL:  "https://www.ufc.com/athletes",
				ParserKind: "ufc_athletes",
				Enabled:    true,
			},
		},
		store,
		scraper,
	)

	result, err := svc.SyncSource(context.Background(), 2)
	if err != nil {
		t.Fatalf("sync source: %v", err)
	}
	if result.Events != 0 || result.Bouts != 0 {
		t.Fatalf("expected athlete sync to skip events and bouts, got %+v", result)
	}
	if result.Fighters != 2 {
		t.Fatalf("expected 2 fighters synced, got %d", result.Fighters)
	}
	if scraper.listEventsCalled {
		t.Fatalf("expected athlete parser to skip event listing")
	}
	if store.lastFight.Nickname == "" {
		t.Fatalf("expected nickname to be persisted for fighter records")
	}
	hasStats := false
	for _, fight := range store.allFights {
		if fight.Stats["Sig. Str. Landed"] != "" {
			hasStats = true
			break
		}
	}
	if !hasStats {
		t.Fatalf("expected stats to be persisted for fighter records")
	}
}

type fakeImageMirror struct{}

func (f fakeImageMirror) MirrorImage(_ context.Context, rawURL string) (string, error) {
	switch rawURL {
	case "https://cdn.ufc.test/poster.jpg":
		return "http://localhost:8080/media-cache/ufc/poster.jpg", nil
	case "https://cdn.ufc.test/fighter-a.jpg":
		return "http://localhost:8080/media-cache/ufc/fighter-a.jpg", nil
	case "https://cdn.ufc.test/fighter-b.jpg":
		return "http://localhost:8080/media-cache/ufc/fighter-b.jpg", nil
	default:
		return "", nil
	}
}

type fakeScraperWithMedia struct{}

func (f fakeScraperWithMedia) ListEventLinks(context.Context, string) ([]EventLink, error) {
	return []EventLink{
		{Name: "UFC Fight Night: A vs B", URL: "https://www.ufc.com/event/ufc-fight-night-a-b", StartsAt: time.Date(2026, 2, 21, 0, 0, 0, 0, time.UTC)},
	}, nil
}

func (f fakeScraperWithMedia) GetEventCard(context.Context, string) (EventCard, error) {
	return EventCard{
		Name:      "UFC Fight Night: A vs B",
		Status:    "scheduled",
		StartsAt:  time.Date(2026, 2, 21, 0, 0, 0, 0, time.UTC),
		PosterURL: "https://cdn.ufc.test/poster.jpg",
		Bouts: []EventBout{
			{RedName: "Fighter A", RedURL: "https://www.ufc.com/athlete/fighter-a", BlueName: "Fighter B", BlueURL: "https://www.ufc.com/athlete/fighter-b"},
		},
	}, nil
}

func (f fakeScraperWithMedia) ListAthleteLinks(context.Context, string) ([]string, error) {
	return []string{}, nil
}
func (f fakeScraperWithMedia) GetAthleteProfile(_ context.Context, rawURL string) (AthleteProfile, error) {
	if rawURL == "https://www.ufc.com/athlete/fighter-a" {
		return AthleteProfile{Name: "Fighter A", URL: rawURL, AvatarURL: "https://cdn.ufc.test/fighter-a.jpg"}, nil
	}
	return AthleteProfile{Name: "Fighter B", URL: rawURL, AvatarURL: "https://cdn.ufc.test/fighter-b.jpg"}, nil
}

func TestSyncSource_MirrorsPosterAndAvatarURLs(t *testing.T) {
	store := &fakeStore{}
	svc := NewService(
		&fakeSourceRepo{
			item: source.DataSource{
				ID:         1,
				SourceType: "schedule",
				Platform:   "ufc",
				SourceURL:  "https://www.ufc.com/events",
				ParserKind: "ufc_schedule",
				Enabled:    true,
			},
		},
		store,
		fakeScraperWithMedia{},
		WithImageMirror(fakeImageMirror{}),
	)

	if _, err := svc.SyncSource(context.Background(), 1); err != nil {
		t.Fatalf("sync source: %v", err)
	}
	if store.lastEvent.PosterURL != "http://localhost:8080/media-cache/ufc/poster.jpg" {
		t.Fatalf("expected mirrored poster url, got %q", store.lastEvent.PosterURL)
	}
	if store.lastFight.AvatarURL != "http://localhost:8080/media-cache/ufc/fighter-b.jpg" {
		t.Fatalf("expected mirrored fighter avatar url, got %q", store.lastFight.AvatarURL)
	}
}

func TestNormalizeEventStatus_StartedUpcomingBecomesLive(t *testing.T) {
	now := time.Date(2026, 2, 23, 14, 0, 0, 0, time.UTC)
	startedAt := now.Add(-5 * time.Minute)
	status := normalizeEventStatus("upcoming", startedAt, now)
	if status != "live" {
		t.Fatalf("expected live for started upcoming event, got %q", status)
	}
}

func TestNormalizeEventStatus_UnknownStartedBecomesLive(t *testing.T) {
	now := time.Date(2026, 2, 23, 14, 0, 0, 0, time.UTC)
	startedAt := now.Add(-30 * time.Minute)
	status := normalizeEventStatus("", startedAt, now)
	if status != "live" {
		t.Fatalf("expected live for started unknown event, got %q", status)
	}
}

func TestNormalizeEventStatus_StartedLongAgoBecomesCompleted(t *testing.T) {
	now := time.Date(2026, 2, 23, 14, 0, 0, 0, time.UTC)
	startedAt := now.Add(-48 * time.Hour)
	status := normalizeEventStatus("live", startedAt, now)
	if status != "completed" {
		t.Fatalf("expected completed for stale live event, got %q", status)
	}
}
