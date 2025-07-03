package model

import (
	"time"

	"gorm.io/gorm"
)

type Ad struct {
	ID          string         `gorm:"type:char(36);primaryKey;column:id" json:"id"`
	ImageURL    string         `gorm:"type:varchar(2048);not null;column:image_url" json:"image_url"` // URLs need more space
	TargetURL   string         `gorm:"type:varchar(2048);not null;column:target_url" json:"target_url"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at"`
	Clicks      []Clicks       `gorm:"foreignKey:AdID"` // No column needed (relationship)
	TotalClicks int            `gorm:"column:total_clicks;not null;default:0" json:"total_clicks"`
}
