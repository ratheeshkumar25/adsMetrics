package repo

import (
	"time"

	"github.com/ratheeshkumar25/adsmetrictracker/internal/model"
	"gorm.io/gorm"
)

type AdsRepoInt interface {
	FetchAdsAll() ([]model.Ad, error)
	CountAds() (int, error)
	SaveBatchAds(clicks []model.Clicks) error
	UpdateAdTotalClicks(adID string, increment int) error
	GetAdsTotalClicks(adID string) (int, error)
	GetClickCountByTimeFrame(adID string, start, end time.Time) (int, error)
	AdsExists(adID string) (bool, error)
	GetClickCountByIP(adID string, ip string) (int, error)
}
type AdsRepository struct {
	DB *gorm.DB
}

func NewAdsRepository(db *gorm.DB) *AdsRepository {
	return &AdsRepository{DB: db}
}
