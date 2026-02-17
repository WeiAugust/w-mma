package mysqlrepo

import (
	"context"

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

func (r *SourceRepository) List(ctx context.Context) ([]source.DataSource, error) {
	var rows []model.DataSource
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]source.DataSource, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapSourceRow(row))
	}
	return items, nil
}

func (r *SourceRepository) Get(ctx context.Context, sourceID int64) (source.DataSource, error) {
	var row model.DataSource
	if err := r.db.WithContext(ctx).Where("id = ?", sourceID).Take(&row).Error; err != nil {
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
	return r.db.WithContext(ctx).Model(&model.DataSource{}).Where("id = ?", sourceID).Updates(updates).Error
}

func (r *SourceRepository) SetEnabled(ctx context.Context, sourceID int64, enabled bool) error {
	return r.db.WithContext(ctx).Model(&model.DataSource{}).Where("id = ?", sourceID).Update("enabled", enabled).Error
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
		RightsDisplay:   row.RightsDisplay,
		RightsPlayback:  row.RightsPlayback,
		RightsAISummary: row.RightsAISummary,
		RightsExpiresAt: row.RightsExpiresAt,
	}
	if row.AccountID != nil {
		item.AccountID = *row.AccountID
	}
	if row.RightsProofURL != nil {
		item.RightsProofURL = *row.RightsProofURL
	}
	return item
}
