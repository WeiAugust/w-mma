package model

import "time"

type SummaryJob struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	SourceID   int64     `gorm:"column:source_id;not null"`
	TargetType string    `gorm:"column:target_type;size:32;not null"`
	TargetID   int64     `gorm:"column:target_id;not null"`
	Status     string    `gorm:"size:32;not null;index:idx_summary_jobs_status_created,priority:1"`
	Provider   *string   `gorm:"size:64"`
	ErrorMsg   *string   `gorm:"column:error_msg;type:text"`
	CreatedAt  time.Time `gorm:"not null;index:idx_summary_jobs_status_created,priority:2"`
	UpdatedAt  time.Time `gorm:"not null"`
}

func (SummaryJob) TableName() string {
	return "summary_jobs"
}
