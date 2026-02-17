package mysqlrepo

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/bajiaozhi/w-mma/backend/internal/fighter"
	"github.com/bajiaozhi/w-mma/backend/internal/model"
)

type FighterRepository struct {
	db *gorm.DB
}

func NewFighterRepository(db *gorm.DB) *FighterRepository {
	return &FighterRepository{db: db}
}

func (r *FighterRepository) SearchByName(ctx context.Context, q string) ([]fighter.Profile, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return []fighter.Profile{}, nil
	}

	var rows []model.Fighter
	if err := r.db.WithContext(ctx).
		Where("name LIKE ?", "%"+q+"%").
		Order("name ASC").
		Limit(20).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]fighter.Profile, 0, len(rows))
	for _, row := range rows {
		items = append(items, toProfile(row, nil))
	}
	return items, nil
}

func (r *FighterRepository) GetByID(ctx context.Context, fighterID int64) (fighter.Profile, error) {
	var row model.Fighter
	if err := r.db.WithContext(ctx).Where("id = ?", fighterID).Take(&row).Error; err != nil {
		return fighter.Profile{}, err
	}

	var updates []model.FighterUpdate
	if err := r.db.WithContext(ctx).
		Where("fighter_id = ?", fighterID).
		Order("published_at DESC").
		Limit(10).
		Find(&updates).Error; err != nil {
		return fighter.Profile{}, err
	}

	return toProfile(row, updates), nil
}

func toProfile(row model.Fighter, updates []model.FighterUpdate) fighter.Profile {
	country := ""
	if row.Country != nil {
		country = *row.Country
	}
	record := ""
	if row.Record != nil {
		record = *row.Record
	}

	list := make([]string, 0, len(updates))
	for _, update := range updates {
		list = append(list, update.Content)
	}

	return fighter.Profile{
		ID:      row.ID,
		Name:    row.Name,
		Country: country,
		Record:  record,
		Updates: list,
	}
}
