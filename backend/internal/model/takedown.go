package model

import "time"

type RightsTakedown struct {
	ID          int64      `gorm:"primaryKey;autoIncrement"`
	TargetType  string     `gorm:"column:target_type;size:16;not null;index:idx_takedowns_target_status,priority:1"`
	TargetID    int64      `gorm:"column:target_id;not null;index:idx_takedowns_target_status,priority:2"`
	Reason      string     `gorm:"type:text;not null"`
	Complainant *string    `gorm:"size:128"`
	EvidenceURL *string    `gorm:"column:evidence_url;size:512"`
	Status      string     `gorm:"size:16;not null;index:idx_takedowns_target_status,priority:3"`
	Action      *string    `gorm:"size:16"`
	CreatedAt   time.Time  `gorm:"not null"`
	ResolvedAt  *time.Time `gorm:"column:resolved_at"`
}

func (RightsTakedown) TableName() string {
	return "rights_takedowns"
}
