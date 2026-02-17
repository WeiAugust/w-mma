package mysqlrepo

import (
	"context"

	"gorm.io/gorm"

	"github.com/bajiaozhi/w-mma/backend/internal/media"
	"github.com/bajiaozhi/w-mma/backend/internal/model"
)

type MediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) Attach(ctx context.Context, input media.AttachInput) (media.Asset, error) {
	row := model.MediaAsset{
		OwnerType: input.OwnerType,
		OwnerID:   input.OwnerID,
		MediaType: input.MediaType,
		URL:       input.URL,
		SortNo:    input.SortNo,
	}
	if input.CoverURL != "" {
		row.CoverURL = &input.CoverURL
	}
	if input.Title != "" {
		row.Title = &input.Title
	}

	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return media.Asset{}, err
	}
	return mapMediaRow(row), nil
}

func (r *MediaRepository) ListByOwner(ctx context.Context, ownerType string, ownerID int64) ([]media.Asset, error) {
	var rows []model.MediaAsset
	if err := r.db.WithContext(ctx).
		Where("owner_type = ?", ownerType).
		Where("owner_id = ?", ownerID).
		Order("sort_no ASC").
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]media.Asset, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapMediaRow(row))
	}
	return items, nil
}

func mapMediaRow(row model.MediaAsset) media.Asset {
	item := media.Asset{
		ID:        row.ID,
		OwnerType: row.OwnerType,
		OwnerID:   row.OwnerID,
		MediaType: row.MediaType,
		URL:       row.URL,
		SortNo:    row.SortNo,
	}
	if row.CoverURL != nil {
		item.CoverURL = *row.CoverURL
	}
	if row.Title != nil {
		item.Title = *row.Title
	}
	return item
}
