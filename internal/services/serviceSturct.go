package services

import (
	"sync"
	"time"

	"github.com/ratheeshkumar25/adsmetrictracker/internal/model"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/repo"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/breaker"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/logger"
)

type AdsServiceInt interface {
	GetAdsAllAds() ([]model.Ad, error)
	ProcessClick(click model.Clicks) error
	ProcessBatch() error
	RecordClick(click model.Clicks) error
	UpdateCounter(click model.Clicks)
	GetClickCount(adID string) (int64, error)
	AdsExists(adID string) (bool, error)
	ParseTimeFrame(timeFrame string) (time.Duration, error)
	GetClickCountByTimeFrame(adID string, timeFrame string) (int64, error)
	PublishClick(click model.Clicks) error
}

type CounterEntry struct {
	ClickCount int64
	LastUpdate time.Time
}

type AdsService struct {
	adsRepo *repo.AdsRepository
	log     *logger.Logger
	nats    *NATSService
	cb      *breaker.CircuitBreaker
	// In-memory counters for better performance
	counters     map[string]*CounterEntry
	counterMutex sync.RWMutex

	// Batch processing
	currentBatch []model.Clicks
	batchMutex   sync.Mutex

	// Deduplication tracking
	processedIDs sync.Map
}
