package model

import "time"

type Event struct {
	ID            int64     `gorm:"primaryKey;autoIncrement"`
	SourceID      *int64    `gorm:"column:source_id"`
	Org           string    `gorm:"size:32;not null"`
	Name          string    `gorm:"size:255;not null"`
	Status        string    `gorm:"size:16;not null;index:idx_events_status_starts,priority:1"`
	StartsAt      time.Time `gorm:"not null;index:idx_events_status_starts,priority:2"`
	Venue         string    `gorm:"size:255;not null"`
	PosterURL     *string   `gorm:"size:512"`
	PromoVideoURL *string   `gorm:"column:promo_video_url;size:512"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

type Bout struct {
	ID              int64 `gorm:"primaryKey;autoIncrement"`
	EventID         int64 `gorm:"not null;uniqueIndex:uk_bouts_event_sequence,priority:1"`
	RedFighterID    int64 `gorm:"not null"`
	BlueFighterID   int64 `gorm:"not null"`
	SequenceNo      int   `gorm:"not null;uniqueIndex:uk_bouts_event_sequence,priority:2"`
	Result          *string
	WinnerFighterID *int64
	Method          *string
	Round           *int
	TimeSec         *int
	CreatedAt       time.Time `gorm:"not null"`
	UpdatedAt       time.Time `gorm:"not null"`
}
