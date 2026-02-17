package model

import "time"

type Fighter struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	Name        string    `gorm:"size:128;not null;index:idx_fighters_name"`
	Nickname    *string   `gorm:"size:128"`
	Country     *string   `gorm:"size:64"`
	WeightClass *string   `gorm:"size:64"`
	Record      *string   `gorm:"size:64"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
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
