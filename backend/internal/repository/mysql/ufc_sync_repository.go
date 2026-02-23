package mysqlrepo

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

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
			fighterID, updateErr := r.updateFighter(ctx, row.ID, item)
			if updateErr != nil {
				return 0, updateErr
			}
			if err := r.upsertFighterUpdates(ctx, fighterID, item.Updates); err != nil {
				return 0, err
			}
			return fighterID, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, err
		}
	}

	err := query.Where("name = ?", item.Name).Take(&row).Error
	if err == nil {
		fighterID, updateErr := r.updateFighter(ctx, row.ID, item)
		if updateErr != nil {
			return 0, updateErr
		}
		if err := r.upsertFighterUpdates(ctx, fighterID, item.Updates); err != nil {
			return 0, err
		}
		return fighterID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	newRow := model.Fighter{
		SourceID:    ptrInt64OrNil(item.SourceID),
		Name:        item.Name,
		NameZH:      ptrString(item.NameZH),
		Nickname:    ptrString(item.Nickname),
		Country:     ptrString(item.Country),
		Record:      ptrString(item.Record),
		WeightClass: ptrString(item.WeightClass),
		StatsJSON:   ptrJSONString(item.Stats),
		RecordsJSON: ptrJSONString(item.Records),
		AvatarURL:   ptrString(item.AvatarURL),
		ExternalURL: ptrString(item.ExternalURL),
		IsManual:    false,
	}
	if err := query.Create(&newRow).Error; err != nil {
		return 0, err
	}
	if err := r.upsertFighterUpdates(ctx, newRow.ID, item.Updates); err != nil {
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
	if item.NameZH != "" {
		updates["name_zh"] = ptrString(item.NameZH)
	}
	if item.Nickname != "" {
		updates["nickname"] = ptrString(item.Nickname)
	}
	if value := ptrJSONString(item.Stats); value != nil {
		updates["stats_json"] = value
	}
	if value := ptrJSONString(item.Records); value != nil {
		updates["records_json"] = value
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

func ptrJSONString(value map[string]string) *string {
	if len(value) == 0 {
		return nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	raw := string(payload)
	return &raw
}

func (r *UFCSyncRepository) upsertFighterUpdates(ctx context.Context, fighterID int64, updates []ufc.AthleteUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	candidates := make([]model.FighterUpdate, 0, len(updates))
	seen := map[string]struct{}{}
	for _, item := range updates {
		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}
		if _, exists := seen[content]; exists {
			continue
		}
		seen[content] = struct{}{}
		publishedAt := item.PublishedAt.UTC()
		if publishedAt.IsZero() {
			publishedAt = time.Now().UTC()
		}
		candidates = append(candidates, model.FighterUpdate{
			FighterID:   fighterID,
			Content:     content,
			PublishedAt: publishedAt,
		})
	}
	if len(candidates) == 0 {
		return nil
	}

	var existing []model.FighterUpdate
	if err := r.db.WithContext(ctx).
		Where("fighter_id = ? AND content REGEXP ?", fighterID, "^[0-9]{4}-[0-9]{2}-[0-9]{2}").
		Find(&existing).Error; err != nil {
		return err
	}
	merged := mergeFighterUpdatesByDate(existing, candidates)
	if len(merged) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).
		Where("fighter_id = ? AND content REGEXP ?", fighterID, "^[0-9]{4}-[0-9]{2}-[0-9]{2}").
		Delete(&model.FighterUpdate{}).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(&merged).Error
}

func mergeFighterUpdatesByDate(existing []model.FighterUpdate, incoming []model.FighterUpdate) []model.FighterUpdate {
	byDate := map[string]model.FighterUpdate{}
	for _, item := range existing {
		key := fighterUpdateDateKey(item.Content)
		if key == "" {
			continue
		}
		current, exists := byDate[key]
		if !exists || item.PublishedAt.After(current.PublishedAt) {
			byDate[key] = item
		}
	}
	for _, item := range incoming {
		key := fighterUpdateDateKey(item.Content)
		if key == "" {
			continue
		}
		byDate[key] = item
	}
	if len(byDate) == 0 {
		return nil
	}
	out := make([]model.FighterUpdate, 0, len(byDate))
	for _, item := range byDate {
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].PublishedAt.Equal(out[j].PublishedAt) {
			return out[i].Content > out[j].Content
		}
		return out[i].PublishedAt.After(out[j].PublishedAt)
	})
	return out
}

func fighterUpdateDateKey(content string) string {
	text := strings.TrimSpace(content)
	if len(text) < 10 {
		return ""
	}
	candidate := text[:10]
	if _, err := time.Parse("2006-01-02", candidate); err != nil {
		return ""
	}
	return candidate
}
