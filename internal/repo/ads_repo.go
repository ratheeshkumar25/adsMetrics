package repo

import "github.com/ratheeshkumar25/adsmetrictracker/internal/model"

func (r *AdsRepository) FetchAdsAll() ([]model.Ad, error) {
	var ads []model.Ad
	if err := r.DB.Preload("Clicks").Find(&ads).Error; err != nil {
		return nil, err
	}
	return ads, nil
}

func (r *AdsRepository) CountAds() (int, error) {
	var count int64
	if err := r.DB.Model(&model.Ad{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}
