package mysqlrepo

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/bajiaozhi/w-mma/backend/internal/model"
	"github.com/bajiaozhi/w-mma/backend/internal/ufc"
)

type UFCSyncRepository struct {
	db *gorm.DB
}

func NewUFCSyncRepository(db *gorm.DB) *UFCSyncRepository {
	return &UFCSyncRepository{db: db}
}

func (r *UFCSyncRepository) UpsertEvent(ctx context.Context, item ufc.EventRecord) (int64, error) {
	var row model.Event
	query := r.db.WithContext(ctx)

	if item.ExternalURL != "" {
		err := query.Where("external_url = ?", item.ExternalURL).Take(&row).Error
		if err == nil {
			return r.updateEvent(ctx, row.ID, item)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, err
		}
	}

	err := query.
		Where("org = ? AND name = ? AND starts_at = ?", item.Org, item.Name, item.StartsAt).
		Take(&row).Error
	if err == nil {
		return r.updateEvent(ctx, row.ID, item)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	newRow := model.Event{
		SourceID:    ptrInt64OrNil(item.SourceID),
		Org:         item.Org,
		Name:        item.Name,
		Status:      item.Status,
		StartsAt:    item.StartsAt,
		Venue:       item.Venue,
		PosterURL:   ptrString(item.PosterURL),
		ExternalURL: ptrString(item.ExternalURL),
	}
	if err := query.Create(&newRow).Error; err != nil {
		return 0, err
	}
	return newRow.ID, nil
}

func (r *UFCSyncRepository) updateEvent(ctx context.Context, eventID int64, item ufc.EventRecord) (int64, error) {
	updates := map[string]any{
		"source_id":    ptrInt64OrNil(item.SourceID),
		"status":       item.Status,
		"venue":        item.Venue,
		"poster_url":   ptrString(item.PosterURL),
		"external_url": ptrString(item.ExternalURL),
	}
	if item.Name != "" {
		updates["name"] = item.Name
	}
	if !item.StartsAt.IsZero() {
		updates["starts_at"] = item.StartsAt
	}
	if err := r.db.WithContext(ctx).Model(&model.Event{}).Where("id = ?", eventID).Updates(updates).Error; err != nil {
		return 0, err
	}
	return eventID, nil
}

func (r *UFCSyncRepository) UpsertFighter(ctx context.Context, item ufc.FighterRecord) (int64, error) {
	var row model.Fighter
	query := r.db.WithContext(ctx)

	if item.ExternalURL != "" {
		err := query.Where("external_url = ?", item.ExternalURL).Take(&row).Error
		if err == nil {
			return r.updateFighter(ctx, row.ID, item)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, err
		}
	}

	err := query.Where("name = ?", item.Name).Take(&row).Error
	if err == nil {
		return r.updateFighter(ctx, row.ID, item)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	newRow := model.Fighter{
		SourceID:    ptrInt64OrNil(item.SourceID),
		Name:        item.Name,
		Country:     ptrString(item.Country),
		Record:      ptrString(item.Record),
		WeightClass: ptrString(item.WeightClass),
		AvatarURL:   ptrString(item.AvatarURL),
		ExternalURL: ptrString(item.ExternalURL),
		IsManual:    false,
	}
	if err := query.Create(&newRow).Error; err != nil {
		return 0, err
	}
	return newRow.ID, nil
}

func (r *UFCSyncRepository) updateFighter(ctx context.Context, fighterID int64, item ufc.FighterRecord) (int64, error) {
	updates := map[string]any{
		"source_id":    ptrInt64OrNil(item.SourceID),
		"country":      ptrString(item.Country),
		"record":       ptrString(item.Record),
		"weight_class": ptrString(item.WeightClass),
		"avatar_url":   ptrString(item.AvatarURL),
		"external_url": ptrString(item.ExternalURL),
		"is_manual":    false,
	}
	if item.Name != "" {
		updates["name"] = item.Name
	}
	if err := r.db.WithContext(ctx).Model(&model.Fighter{}).Where("id = ?", fighterID).Updates(updates).Error; err != nil {
		return 0, err
	}
	return fighterID, nil
}

func (r *UFCSyncRepository) ReplaceEventBouts(ctx context.Context, eventID int64, bouts []ufc.BoutRecord) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("event_id = ?", eventID).Delete(&model.Bout{}).Error; err != nil {
			return err
		}
		for idx, bout := range bouts {
			row := model.Bout{
				EventID:         eventID,
				RedFighterID:    bout.RedFighterID,
				BlueFighterID:   bout.BlueFighterID,
				SequenceNo:      idx + 1,
				CardSegment:     ptrString(bout.CardSegment),
				WeightClass:     ptrString(bout.WeightClass),
				RedRanking:      ptrString(bout.RedRanking),
				BlueRanking:     ptrString(bout.BlueRanking),
				Result:          ptrString(bout.Result),
				WinnerFighterID: ptrInt64OrNil(bout.WinnerID),
				Method:          ptrString(bout.Method),
				Round:           ptrIntOrNil(bout.Round),
				TimeSec:         ptrIntOrNil(bout.TimeSec),
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func ptrIntOrNil(value int) *int {
	if value <= 0 {
		return nil
	}
	copy := value
	return &copy
}
