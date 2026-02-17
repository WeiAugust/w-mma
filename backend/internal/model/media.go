package model

import "time"

type MediaAsset struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	OwnerType string    `gorm:"column:owner_type;size:16;not null;index:idx_media_owner_sort,priority:1"`
	OwnerID   int64     `gorm:"column:owner_id;not null;index:idx_media_owner_sort,priority:2"`
	MediaType string    `gorm:"column:media_type;size:16;not null"`
	URL       string    `gorm:"column:url;size:512;not null"`
	CoverURL  *string   `gorm:"column:cover_url;size:512"`
	Title     *string   `gorm:"column:title;size:255"`
	SortNo    int       `gorm:"column:sort_no;not null;index:idx_media_owner_sort,priority:3"`
	CreatedAt time.Time `gorm:"not null"`
}

func (MediaAsset) TableName() string {
	return "media_assets"
}
