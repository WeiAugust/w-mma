package mysqlrepo

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/bajiaozhi/w-mma/backend/internal/event"
	"github.com/bajiaozhi/w-mma/backend/internal/model"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) GetEventCard(ctx context.Context, eventID int64) (event.Card, error) {
	var row model.Event
	if err := r.db.WithContext(ctx).Where("id = ?", eventID).Take(&row).Error; err != nil {
		return event.Card{}, err
	}

	var bouts []model.Bout
	if err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Order("sequence_no ASC").Find(&bouts).Error; err != nil {
		return event.Card{}, err
	}

	card := event.Card{
		ID:            row.ID,
		Org:           row.Org,
		Name:          row.Name,
		Status:        row.Status,
		PosterURL:     ptrStringValue(row.PosterURL),
		PromoVideoURL: ptrStringValue(row.PromoVideoURL),
		Bouts:         make([]event.Bout, 0, len(bouts)),
		MainCard:      make([]event.BoutDetail, 0),
		Prelims:       make([]event.BoutDetail, 0),
	}

	fighterIDs := make([]int64, 0, len(bouts)*2)
	for _, b := range bouts {
		fighterIDs = append(fighterIDs, b.RedFighterID, b.BlueFighterID)
	}
	fighterByID := map[int64]model.Fighter{}
	if len(fighterIDs) > 0 {
		var fighters []model.Fighter
		if err := r.db.WithContext(ctx).Where("id IN ?", fighterIDs).Find(&fighters).Error; err != nil {
			return event.Card{}, err
		}
		for _, fighter := range fighters {
			fighterByID[fighter.ID] = fighter
		}
	}

	for _, b := range bouts {
		winnerID := int64(0)
		if b.WinnerFighterID != nil {
			winnerID = *b.WinnerFighterID
		}
		result := ""
		if b.Result != nil {
			result = *b.Result
		}
		method := ""
		if b.Method != nil {
			method = *b.Method
		}
		round := 0
		if b.Round != nil {
			round = *b.Round
		}
		timeSec := 0
		if b.TimeSec != nil {
			timeSec = *b.TimeSec
		}
		card.Bouts = append(card.Bouts, event.Bout{
			ID:            b.ID,
			RedFighterID:  b.RedFighterID,
			BlueFighterID: b.BlueFighterID,
			CardSegment:   ptrStringValue(b.CardSegment),
			WeightClass:   ptrStringValue(b.WeightClass),
			RedRanking:    ptrStringValue(b.RedRanking),
			BlueRanking:   ptrStringValue(b.BlueRanking),
			Result:        result,
			WinnerID:      winnerID,
			Method:        method,
			Round:         round,
			TimeSec:       timeSec,
		})

		redProfile := fighterByID[b.RedFighterID]
		blueProfile := fighterByID[b.BlueFighterID]
		weightClass := ptrStringValue(b.WeightClass)
		if weightClass == "" {
			weightClass = chooseNonEmpty(ptrStringValue(redProfile.WeightClass), ptrStringValue(blueProfile.WeightClass))
		}
		detail := event.BoutDetail{
			ID:          b.ID,
			CardSegment: ptrStringValue(b.CardSegment),
			WeightClass: weightClass,
			Result:      result,
			WinnerID:    winnerID,
			Method:      method,
			Round:       round,
			TimeSec:     timeSec,
			RedFighter: event.FighterSnapshot{
				ID:          b.RedFighterID,
				Name:        redProfile.Name,
				Country:     ptrStringValue(redProfile.Country),
				Rank:        ptrStringValue(b.RedRanking),
				WeightClass: chooseNonEmpty(weightClass, ptrStringValue(redProfile.WeightClass)),
				AvatarURL:   ptrStringValue(redProfile.AvatarURL),
			},
			BlueFighter: event.FighterSnapshot{
				ID:          b.BlueFighterID,
				Name:        blueProfile.Name,
				Country:     ptrStringValue(blueProfile.Country),
				Rank:        ptrStringValue(b.BlueRanking),
				WeightClass: chooseNonEmpty(weightClass, ptrStringValue(blueProfile.WeightClass)),
				AvatarURL:   ptrStringValue(blueProfile.AvatarURL),
			},
		}
		switch strings.ToLower(ptrStringValue(b.CardSegment)) {
		case "main_card":
			card.MainCard = append(card.MainCard, detail)
		case "prelims":
			card.Prelims = append(card.Prelims, detail)
		default:
			// Legacy data without card segment defaults to prelims.
			card.Prelims = append(card.Prelims, detail)
		}
	}

	if len(card.MainCard) == 0 && len(card.Prelims) > 0 {
		mainCount := 5
		if len(card.Prelims) < mainCount {
			mainCount = len(card.Prelims)
		}
		card.MainCard = append(card.MainCard, card.Prelims[:mainCount]...)
		card.Prelims = card.Prelims[mainCount:]
	}
	return card, nil
}

func (r *EventRepository) ListEvents(ctx context.Context) ([]event.EventSummary, error) {
	var rows []model.Event
	if err := r.db.WithContext(ctx).Order("starts_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]event.EventSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, event.EventSummary{
			ID:            row.ID,
			Org:           row.Org,
			Name:          row.Name,
			Status:        row.Status,
			StartsAt:      row.StartsAt.UTC().Format("2006-01-02T15:04:05Z"),
			PosterURL:     ptrStringValue(row.PosterURL),
			PromoVideoURL: ptrStringValue(row.PromoVideoURL),
		})
	}
	return items, nil
}

func ptrStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func chooseNonEmpty(value string, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func (r *EventRepository) UpdateEvent(ctx context.Context, eventID int64, input event.UpdateEventInput) error {
	updates := map[string]any{}
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Status != "" {
		updates["status"] = input.Status
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.Event{}).Where("id = ?", eventID).Updates(updates).Error
}

func (r *EventRepository) UpsertBoutResult(ctx context.Context, eventID int64, boutID int64, winnerID int64, result string) error {
	winner := winnerID
	res := result
	updates := map[string]any{
		"winner_fighter_id": &winner,
		"result":            &res,
	}
	ret := r.db.WithContext(ctx).Model(&model.Bout{}).
		Where("event_id = ?", eventID).
		Where("id = ?", boutID).
		Updates(updates)
	if ret.Error != nil {
		return ret.Error
	}
	if ret.RowsAffected == 0 {
		return fmt.Errorf("bout not found for event=%d bout=%d", eventID, boutID)
	}
	return nil
}
