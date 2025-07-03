package repo

import (
	"log"

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
