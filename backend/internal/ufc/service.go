package ufc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/bajiaozhi/w-mma/backend/internal/source"
)

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
	if src.ParserKind != "ufc_schedule" || src.SourceURL == "" {
		return SyncResult{}, ErrUnsupportedSource
	}

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
		startsAt := card.StartsAt
		if startsAt.IsZero() {
			startsAt = eventLink.StartsAt
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
				Country:     redProfile.Country,
				Record:      redProfile.Record,
				WeightClass: bout.WeightClass,
				AvatarURL:   s.mirrorImageURL(ctx, redProfile.AvatarURL),
				ExternalURL: redProfile.URL,
			})
			if err != nil {
				continue
			}
			blueID, err := s.store.UpsertFighter(ctx, FighterRecord{
				SourceID:    src.ID,
				Name:        chooseNonEmpty(blueProfile.Name, bout.BlueName),
				Country:     blueProfile.Country,
				Record:      blueProfile.Record,
				WeightClass: bout.WeightClass,
				AvatarURL:   s.mirrorImageURL(ctx, blueProfile.AvatarURL),
				ExternalURL: blueProfile.URL,
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

func (s *Service) SyncEnabledSources(ctx context.Context) (SyncResult, error) {
	enabled := true
	items, err := s.sourceRepo.List(ctx, source.ListFilter{
		SourceType: "schedule",
		Platform:   "ufc",
		Enabled:    &enabled,
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
		return "live"
	case "scheduled", "upcoming":
		return "scheduled"
	}
	if startsAt.IsZero() {
		return "scheduled"
	}
	if startsAt.After(now) {
		return "scheduled"
	}
	return "completed"
}
