package repo

import (
	"log"
	"time"

	"github.com/ratheeshkumar25/adsmetrictracker/internal/model"
	"gorm.io/gorm"
)

func (r *AdsRepository) SaveBatchAds(clicks []model.Clicks) error {
	if err := r.DB.CreateInBatches(&clicks, 500).Error; err != nil {
		log.Printf("Failed to save click event: %v", err)
		return err
	}
	return nil
}

func (r *AdsRepository) UpdateAdTotalClicks(adID string, increment int) error {
	result := r.DB.Model(&model.Ad{}).
		Where("id = ?", adID).
		UpdateColumn("total_clicks", gorm.Expr("total_clicks + ?", increment))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *AdsRepository) GetAdsTotalClicks(adID string) (int, error) {
	var ad model.Ad
	if err := r.DB.Select("total_clicks").Where("id = ?", adID).First(&ad).Error; err != nil {
		return 0, err
	}
	return ad.TotalClicks, nil
}

func (r *AdsRepository) GetClickCountByTimeFrame(adID string, start, end time.Time) (int, error) {
	var count int64
	err := r.DB.Model(&model.Clicks{}).
		Where("ad_id = ? AND timestamp BETWEEN ? AND ?", adID, start, end).
		Count(&count).Error
	return int(count), err
}

func (r *AdsRepository) AdsExists(adID string) (bool, error) {
	var count int64
	err := r.DB.Model(&model.Ad{}).Where("id = ?", adID).Count(&count).Error
	return count > 0, err
}

func (r *AdsRepository) GetClickCountByIP(adID string, ip string) (int, error) {
	var count int64
	err := r.DB.Model(&model.Clicks{}).
		Where("ad_id = ? AND ip = ?", adID, ip).
		Count(&count).Error
	return int(count), err
}

// SaveClick saves a single click event
func (r *AdsRepository) SaveClick(click *model.Clicks) error {
	return r.DB.Create(click).Error
}

// GetRecentClicks gets recent clicks for analytics
func (r *AdsRepository) GetRecentClicks(adID string, minutes int) ([]model.Clicks, error) {
	var clicks []model.Clicks
	since := time.Now().Add(-time.Duration(minutes) * time.Minute)
	err := r.DB.Where("ad_id = ? AND timestamp > ?", adID, since).Find(&clicks).Error
	return clicks, err
}
