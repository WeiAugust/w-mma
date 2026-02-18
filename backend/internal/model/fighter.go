package model

import "time"

type Fighter struct {
	ID            int64     `gorm:"primaryKey;autoIncrement"`
	SourceID      *int64    `gorm:"column:source_id"`
	Name          string    `gorm:"size:128;not null;index:idx_fighters_name"`
	Nickname      *string   `gorm:"size:128"`
	Country       *string   `gorm:"size:64"`
	WeightClass   *string   `gorm:"size:64"`
	Record        *string   `gorm:"size:64"`
	AvatarURL     *string   `gorm:"column:avatar_url;size:512"`
	IntroVideoURL *string   `gorm:"column:intro_video_url;size:512"`
	ExternalURL   *string   `gorm:"column:external_url;size:512;index:idx_fighters_external_url"`
	IsManual      bool      `gorm:"column:is_manual;not null"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

type FighterUpdate struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	FighterID   int64     `gorm:"not null;index:idx_fighter_updates_fighter_published,priority:1"`
	Content     string    `gorm:"type:text;not null"`
	PublishedAt time.Time `gorm:"not null;index:idx_fighter_updates_fighter_published,priority:2"`
	CreatedAt   time.Time `gorm:"not null"`
}

func (FighterUpdate) TableName() string {
	return "fighter_updates"
}
