package model

import "time"

type DataSource struct {
	ID              int64      `gorm:"primaryKey;autoIncrement"`
	Name            string     `gorm:"size:128;not null"`
	SourceType      string     `gorm:"column:source_type;size:16;not null;index:idx_data_sources_type_enabled,priority:1"`
	Platform        string     `gorm:"size:64;not null"`
	AccountID       *string    `gorm:"column:account_id;size:128"`
	SourceURL       string     `gorm:"column:source_url;size:512;not null"`
	ParserKind      string     `gorm:"column:parser_kind;size:64;not null"`
	Enabled         bool       `gorm:"not null;index:idx_data_sources_type_enabled,priority:2"`
	IsBuiltin       bool       `gorm:"column:is_builtin;not null"`
	RightsDisplay   bool       `gorm:"column:rights_display;not null"`
	RightsPlayback  bool       `gorm:"column:rights_playback;not null"`
	RightsAISummary bool       `gorm:"column:rights_ai_summary;not null"`
	RightsExpiresAt *time.Time `gorm:"column:rights_expires_at"`
	RightsProofURL  *string    `gorm:"column:rights_proof_url;size:512"`
	LastFetchAt     *time.Time `gorm:"column:last_fetch_at"`
	LastFetchStatus *string    `gorm:"column:last_fetch_status;size:32"`
	LastFetchError  *string    `gorm:"column:last_fetch_error;size:1024"`
	DeletedAt       *time.Time `gorm:"column:deleted_at;index"`
	CreatedAt       time.Time  `gorm:"not null"`
	UpdatedAt       time.Time  `gorm:"not null"`
}

func (DataSource) TableName() string {
	return "data_sources"
}
