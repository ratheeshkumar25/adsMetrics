package services

import (
	"sync"
	"time"

	"github.com/ratheeshkumar25/adsmetrictracker/internal/model"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/repo"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/breaker"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/logger"
)

const (
	consumerGroupID = "ad-clicks-group"
	topicName       = "ad-clicks"
	maxRetries      = 5
	retryDelay      = 2 * time.Second
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

type AdsService struct {
	adsRepo      *repo.AdsRepository
	log          *logger.Logger
	kafka        *KafkaService
	cb           *breaker.CircuitBreaker
	counters     map[string]*CounterEntry
	currentBatch []model.Clicks
	processedIDs sync.Map
	batchMutex   sync.Mutex
	counterMutex sync.RWMutex
}

type CounterEntry struct {
	ClickCount int64
	LastUpdate time.Time
}
