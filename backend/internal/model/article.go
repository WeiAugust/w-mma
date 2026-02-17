package model

import "time"

type Article struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	Title       string    `gorm:"size:255;not null"`
	Content     string    `gorm:"type:text;not null"`
	SourceURL   string    `gorm:"size:512;not null;uniqueIndex"`
	Status      string    `gorm:"size:16;not null"`
	PublishedAt time.Time `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

type PendingArticle struct {
	ID         int64  `gorm:"primaryKey;autoIncrement"`
	Title      string `gorm:"size:255;not null"`
	Summary    string `gorm:"type:text;not null"`
	SourceURL  string `gorm:"size:512;not null;uniqueIndex"`
	Status     string `gorm:"size:16;not null;index:idx_pending_status_created,priority:1"`
	ReviewerID *int64 `gorm:"column:reviewer_id"`
	ReviewedAt *time.Time
	CreatedAt  time.Time `gorm:"not null;index:idx_pending_status_created,priority:2"`
	UpdatedAt  time.Time `gorm:"not null"`
}

func (PendingArticle) TableName() string {
	return "pending_articles"
}
