package model

import "time"

type Clicks struct {
	ID            string    `gorm:"type:char(36);primaryKey;column:id" json:"id"`
	AdID          string    `gorm:"type:char(36);not null;column:ad_id" json:"ad_id"`
	Ad            Ad        `gorm:"foreignKey:AdID;references:ID"`                 // No column needed
	IP            string    `gorm:"type:varchar(45);not null;column:ip" json:"ip"` // Changed to varchar(45)
	VideoPlayTime int       `gorm:"not null;column:playback_time" json:"playback_time"`
	Timestamp     time.Time `gorm:"not null;column:timestamp" json:"timestamp"`
}
