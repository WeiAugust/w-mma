package mysqlrepo

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/bajiaozhi/w-mma/backend/internal/model"
	"github.com/bajiaozhi/w-mma/backend/internal/source"
)

type SourceRepository struct {
	db *gorm.DB
}

func NewSourceRepository(db *gorm.DB) *SourceRepository {
	return &SourceRepository{db: db}
}

func (r *SourceRepository) Create(ctx context.Context, input source.CreateInput) (source.DataSource, error) {
	row := model.DataSource{
		Name:            input.Name,
		SourceType:      input.SourceType,
		Platform:        input.Platform,
		SourceURL:       input.SourceURL,
		ParserKind:      input.ParserKind,
		Enabled:         input.Enabled,
		IsBuiltin:       input.IsBuiltin,
		RightsDisplay:   input.RightsDisplay,
		RightsPlayback:  input.RightsPlayback,
		RightsAISummary: input.RightsAISummary,
	}
	if input.AccountID != "" {
		row.AccountID = &input.AccountID
	}
	if !input.RightsExpiresAt.IsZero() {
		expires := input.RightsExpiresAt
		row.RightsExpiresAt = &expires
	}
	if input.RightsProofURL != "" {
		row.RightsProofURL = &input.RightsProofURL
	}

	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return source.DataSource{}, err
	}
	return mapSourceRow(row), nil
}

func (r *SourceRepository) List(ctx context.Context, filter source.ListFilter) ([]source.DataSource, error) {
	query := r.db.WithContext(ctx).Model(&model.DataSource{})
	if !filter.IncludeDeleted {
		query = query.Where("deleted_at IS NULL")
	}
	if filter.SourceType != "" {
		query = query.Where("source_type = ?", filter.SourceType)
	}
	if filter.Platform != "" {
		query = query.Where("platform = ?", filter.Platform)
	}
	if filter.Enabled != nil {
		query = query.Where("enabled = ?", *filter.Enabled)
	}
	if filter.IsBuiltin != nil {
		query = query.Where("is_builtin = ?", *filter.IsBuiltin)
	}

	var rows []model.DataSource
	if err := query.Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]source.DataSource, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapSourceRow(row))
	}
	return items, nil
}

func (r *SourceRepository) Get(ctx context.Context, sourceID int64, includeDeleted bool) (source.DataSource, error) {
	query := r.db.WithContext(ctx).Where("id = ?", sourceID)
	if !includeDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	var row model.DataSource
	if err := query.Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return source.DataSource{}, source.ErrSourceNotFound
		}
		return source.DataSource{}, err
	}
	return mapSourceRow(row), nil
}

func (r *SourceRepository) Update(ctx context.Context, sourceID int64, input source.UpdateInput) error {
	updates := map[string]any{}
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Platform != "" {
		updates["platform"] = input.Platform
	}
	if input.AccountID != nil {
		updates["account_id"] = input.AccountID
	}
	if input.SourceURL != "" {
		updates["source_url"] = input.SourceURL
	}
	if input.ParserKind != "" {
		updates["parser_kind"] = input.ParserKind
	}
	if input.RightsDisplay != nil {
		updates["rights_display"] = *input.RightsDisplay
	}
	if input.RightsPlayback != nil {
		updates["rights_playback"] = *input.RightsPlayback
	}
	if input.RightsAISummary != nil {
		updates["rights_ai_summary"] = *input.RightsAISummary
	}
	if input.RightsExpiresAt != nil {
		updates["rights_expires_at"] = input.RightsExpiresAt
	}
	if input.RightsProofURL != nil {
		updates["rights_proof_url"] = input.RightsProofURL
	}
	if len(updates) == 0 {
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&model.DataSource{}).
		Where("id = ?", sourceID).
		Where("deleted_at IS NULL").
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return source.ErrSourceNotFound
	}
	return nil
}

func (r *SourceRepository) SetEnabled(ctx context.Context, sourceID int64, enabled bool) error {
	result := r.db.WithContext(ctx).
		Model(&model.DataSource{}).
		Where("id = ?", sourceID).
		Where("deleted_at IS NULL").
		Update("enabled", enabled)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return source.ErrSourceNotFound
	}
	return nil
}

func (r *SourceRepository) SoftDelete(ctx context.Context, sourceID int64) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&model.DataSource{}).
		Where("id = ?", sourceID).
		Where("deleted_at IS NULL").
		Update("deleted_at", &now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return source.ErrSourceNotFound
	}
	return nil
}

func (r *SourceRepository) Restore(ctx context.Context, sourceID int64) error {
	result := r.db.WithContext(ctx).
		Model(&model.DataSource{}).
		Where("id = ?", sourceID).
		Update("deleted_at", nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return source.ErrSourceNotFound
	}
	return nil
}

func mapSourceRow(row model.DataSource) source.DataSource {
	item := source.DataSource{
		ID:              row.ID,
		Name:            row.Name,
		SourceType:      row.SourceType,
		Platform:        row.Platform,
		SourceURL:       row.SourceURL,
		ParserKind:      row.ParserKind,
		Enabled:         row.Enabled,
		IsBuiltin:       row.IsBuiltin,
		RightsDisplay:   row.RightsDisplay,
		RightsPlayback:  row.RightsPlayback,
		RightsAISummary: row.RightsAISummary,
		RightsExpiresAt: row.RightsExpiresAt,
		LastFetchAt:     row.LastFetchAt,
		DeletedAt:       row.DeletedAt,
	}
	if row.AccountID != nil {
		item.AccountID = *row.AccountID
	}
	if row.RightsProofURL != nil {
		item.RightsProofURL = *row.RightsProofURL
	}
	if row.LastFetchStatus != nil {
		item.LastFetchStatus = *row.LastFetchStatus
	}
	if row.LastFetchError != nil {
		item.LastFetchError = *row.LastFetchError
	}
	return item
}
