package mysqlrepo

import (
	"context"
	"fmt"

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
		card.Bouts = append(card.Bouts, event.Bout{
			ID:            b.ID,
			RedFighterID:  b.RedFighterID,
			BlueFighterID: b.BlueFighterID,
			Result:        result,
			WinnerID:      winnerID,
		})
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
