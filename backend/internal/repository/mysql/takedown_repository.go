package mysqlrepo

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/bajiaozhi/w-mma/backend/internal/model"
	"github.com/bajiaozhi/w-mma/backend/internal/takedown"
)

type TakedownRepository struct {
	db *gorm.DB
}

func NewTakedownRepository(db *gorm.DB) *TakedownRepository {
	return &TakedownRepository{db: db}
}

func (r *TakedownRepository) Create(ctx context.Context, input takedown.CreateInput) (takedown.Ticket, error) {
	row := model.RightsTakedown{
		TargetType: input.TargetType,
		TargetID:   input.TargetID,
		Reason:     input.Reason,
		Status:     "open",
	}
	if input.Complainant != "" {
		row.Complainant = &input.Complainant
	}
	if input.EvidenceURL != "" {
		row.EvidenceURL = &input.EvidenceURL
	}

	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return takedown.Ticket{}, err
	}
	return mapTakedownRow(row), nil
}

func (r *TakedownRepository) Get(ctx context.Context, ticketID int64) (takedown.Ticket, error) {
	var row model.RightsTakedown
	if err := r.db.WithContext(ctx).Where("id = ?", ticketID).Take(&row).Error; err != nil {
		return takedown.Ticket{}, err
	}
	return mapTakedownRow(row), nil
}

func (r *TakedownRepository) Resolve(ctx context.Context, ticketID int64, action string) error {
	resolvedAt := time.Now()
	return r.db.WithContext(ctx).Model(&model.RightsTakedown{}).
		Where("id = ?", ticketID).
		Updates(map[string]any{
			"status":      "resolved",
			"action":      action,
			"resolved_at": resolvedAt,
		}).Error
}

func mapTakedownRow(row model.RightsTakedown) takedown.Ticket {
	item := takedown.Ticket{
		ID:         row.ID,
		TargetType: row.TargetType,
		TargetID:   row.TargetID,
		Reason:     row.Reason,
		Status:     row.Status,
	}
	if row.Complainant != nil {
		item.Complainant = *row.Complainant
	}
	if row.EvidenceURL != nil {
		item.EvidenceURL = *row.EvidenceURL
	}
	if row.Action != nil {
		item.Action = *row.Action
	}
	return item
}
