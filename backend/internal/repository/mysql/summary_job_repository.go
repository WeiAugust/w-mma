package mysqlrepo

import (
	"context"

	"gorm.io/gorm"

	"github.com/bajiaozhi/w-mma/backend/internal/model"
	"github.com/bajiaozhi/w-mma/backend/internal/summary"
)

type SummaryJobRepository struct {
	db *gorm.DB
}

func NewSummaryJobRepository(db *gorm.DB) *SummaryJobRepository {
	return &SummaryJobRepository{db: db}
}

func (r *SummaryJobRepository) Create(ctx context.Context, input summary.CreateInput) (summary.Job, error) {
	row := model.SummaryJob{
		SourceID:   input.SourceID,
		TargetType: input.TargetType,
		TargetID:   input.TargetID,
		Status:     input.Status,
	}
	if input.Provider != "" {
		row.Provider = &input.Provider
	}
	if input.ErrorMsg != "" {
		row.ErrorMsg = &input.ErrorMsg
	}

	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return summary.Job{}, err
	}
	return mapSummaryJobRow(row), nil
}

func (r *SummaryJobRepository) List(ctx context.Context) ([]summary.Job, error) {
	var rows []model.SummaryJob
	if err := r.db.WithContext(ctx).Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]summary.Job, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapSummaryJobRow(row))
	}
	return items, nil
}

func (r *SummaryJobRepository) Get(ctx context.Context, jobID int64) (summary.Job, error) {
	var row model.SummaryJob
	if err := r.db.WithContext(ctx).Where("id = ?", jobID).Take(&row).Error; err != nil {
		return summary.Job{}, err
	}
	return mapSummaryJobRow(row), nil
}

func (r *SummaryJobRepository) UpdateStatus(ctx context.Context, jobID int64, status string, errorMsg string) error {
	updates := map[string]any{
		"status": status,
	}
	if errorMsg == "" {
		updates["error_msg"] = nil
	} else {
		updates["error_msg"] = errorMsg
	}
	return r.db.WithContext(ctx).Model(&model.SummaryJob{}).Where("id = ?", jobID).Updates(updates).Error
}

func mapSummaryJobRow(row model.SummaryJob) summary.Job {
	item := summary.Job{
		ID:         row.ID,
		SourceID:   row.SourceID,
		TargetType: row.TargetType,
		TargetID:   row.TargetID,
		Status:     row.Status,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
	if row.Provider != nil {
		item.Provider = *row.Provider
	}
	if row.ErrorMsg != nil {
		item.ErrorMsg = *row.ErrorMsg
	}
	return item
}
