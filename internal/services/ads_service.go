package services

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/model"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/repo"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/breaker"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/logger"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/metrics"
)

func NewAdsService(adsRepo *repo.AdsRepository, log *logger.Logger, nats *NATSService, cb *breaker.CircuitBreaker) *AdsService {
	return &AdsService{
		adsRepo:      adsRepo,
		log:          log,
		nats:         nats,
		cb:           cb,
		counters:     make(map[string]*CounterEntry),
		currentBatch: make([]model.Clicks, 0),
		processedIDs: sync.Map{},
	}
}

func (s *AdsService) GetAdsAllAds() ([]model.Ad, error) {
	start := time.Now()

	ads, err := s.adsRepo.FetchAdsAll()
	if err != nil {
		metrics.RecordError("fetch_ads_error", "ads_service")
		s.log.Logger.Errorf("Failed to fetch ads: %v", err)
		return nil, fmt.Errorf("failed to fetch ads: %w", err)
	}

	metrics.RecordDatabaseOperation("select", "success", time.Since(start).Seconds())
	return ads, nil
}

func (s *AdsService) ProcessClick(click model.Clicks) error {
	start := time.Now()

	// Validate ad exists
	exists, err := s.AdsExists(click.AdID)
	if err != nil {
		metrics.RecordError("ads_exists_check_error", "ads_service")
		return fmt.Errorf("failed to check if ad exists: %w", err)
	}
	if !exists {
		metrics.RecordError("ad_not_found", "ads_service")
		return fmt.Errorf("ad not found: %s", click.AdID)
	}

	// Check for duplicate processing
	clickKey := fmt.Sprintf("%s-%s-%d", click.AdID, click.IP, click.Timestamp.Unix())
	if _, exists := s.processedIDs.LoadOrStore(clickKey, true); exists {
		s.log.Logger.Warnf("Duplicate click detected: %s", clickKey)
		return nil
	}

	// Generate ID if not present
	if click.ID == "" {
		click.ID = uuid.New().String()
	}

	// Record click
	if err := s.RecordClick(click); err != nil {
		metrics.RecordError("record_click_error", "ads_service")
		return fmt.Errorf("failed to record click: %w", err)
	}

	// Update counter
	s.UpdateCounter(click)

	// Update ad total clicks
	if err := s.adsRepo.UpdateAdTotalClicks(click.AdID, 1); err != nil {
		s.log.Logger.Errorf("Failed to update ad total clicks: %v", err)
	}

	metrics.RecordClick(click.AdID, time.Since(start).Seconds())
	return nil
}

func (s *AdsService) RecordClick(click model.Clicks) error {
	// Use circuit breaker for database operations
	return s.cb.Call(func() error {
		s.batchMutex.Lock()
		defer s.batchMutex.Unlock()

		s.currentBatch = append(s.currentBatch, click)

		// Process batch if it reaches threshold
		if len(s.currentBatch) >= 100 {
			return s.processBatchInternal()
		}

		return nil
	})
}

func (s *AdsService) ProcessBatch() error {
	s.batchMutex.Lock()
	defer s.batchMutex.Unlock()

	if len(s.currentBatch) == 0 {
		return nil
	}

	return s.processBatchInternal()
}

func (s *AdsService) processBatchInternal() error {
	if len(s.currentBatch) == 0 {
		return nil
	}

	start := time.Now()

	err := s.adsRepo.SaveBatchAds(s.currentBatch)
	if err != nil {
		metrics.RecordError("batch_save_error", "ads_service")
		return fmt.Errorf("failed to save batch: %w", err)
	}

	s.log.Logger.Infof("Processed batch of %d clicks", len(s.currentBatch))
	s.currentBatch = s.currentBatch[:0] // Clear the batch

	metrics.RecordDatabaseOperation("batch_insert", "success", time.Since(start).Seconds())
	return nil
}

func (s *AdsService) UpdateCounter(click model.Clicks) {
	s.counterMutex.Lock()
	defer s.counterMutex.Unlock()

	entry, exists := s.counters[click.AdID]
	if !exists {
		entry = &CounterEntry{
			ClickCount: 0,
			LastUpdate: time.Now(),
		}
		s.counters[click.AdID] = entry
	}

	entry.ClickCount++
	entry.LastUpdate = time.Now()
}

func (s *AdsService) GetClickCount(adID string) (int64, error) {
	s.counterMutex.RLock()
	defer s.counterMutex.RUnlock()

	if entry, exists := s.counters[adID]; exists {
		return entry.ClickCount, nil
	}

	// Fall back to database
	count, err := s.adsRepo.GetAdsTotalClicks(adID)
	return int64(count), err
}

func (s *AdsService) AdsExists(adID string) (bool, error) {
	return s.adsRepo.AdsExists(adID)
}

func (s *AdsService) ParseTimeFrame(timeFrame string) (time.Duration, error) {
	timeFrame = strings.ToLower(strings.TrimSpace(timeFrame))

	switch timeFrame {
	case "1m", "1min", "1minute":
		return time.Minute, nil
	case "5m", "5min", "5minutes":
		return 5 * time.Minute, nil
	case "15m", "15min", "15minutes":
		return 15 * time.Minute, nil
	case "30m", "30min", "30minutes":
		return 30 * time.Minute, nil
	case "1h", "1hour":
		return time.Hour, nil
	case "24h", "1day", "1d":
		return 24 * time.Hour, nil
	default:
		// Try to parse as duration string (e.g., "10m", "2h")
		if duration, err := time.ParseDuration(timeFrame); err == nil {
			return duration, nil
		}

		// Try to parse as minutes
		if minutes, err := strconv.Atoi(timeFrame); err == nil {
			return time.Duration(minutes) * time.Minute, nil
		}

		return 0, fmt.Errorf("invalid time frame: %s", timeFrame)
	}
}

func (s *AdsService) GetClickCountByTimeFrame(adID string, timeFrame string) (int64, error) {
	duration, err := s.ParseTimeFrame(timeFrame)
	if err != nil {
		return 0, err
	}

	end := time.Now()
	start := end.Add(-duration)

	count, err := s.adsRepo.GetClickCountByTimeFrame(adID, start, end)
	return int64(count), err
}

func (s *AdsService) PublishClick(click model.Clicks) error {
	// If NATS is not available, process directly
	if s.nats == nil {
		s.log.Logger.Debug("NATS not available, processing click directly")
		return s.ProcessClick(click)
	}

	// Publish to NATS
	err := s.nats.PublishClick(click)
	if err != nil {
		metrics.RecordError("nats_publish_error", "ads_service")
		// Fallback to direct processing if NATS fails
		s.log.Logger.Warnf("Failed to publish to NATS, processing directly: %v", err)
		return s.ProcessClick(click)
	}

	s.log.Logger.Debugf("Click published to NATS for ad: %s", click.AdID)
	return nil
}

// GetAnalytics returns comprehensive analytics for an ad
func (s *AdsService) GetAnalytics(adID string) (*AnalyticsResponse, error) {
	// Check if ad exists
	exists, err := s.AdsExists(adID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("ad not found: %s", adID)
	}

	// Get total clicks
	totalClicks, err := s.adsRepo.GetAdsTotalClicks(adID)
	if err != nil {
		return nil, err
	}

	// Get clicks for different time frames
	last1m, _ := s.GetClickCountByTimeFrame(adID, "1m")
	last5m, _ := s.GetClickCountByTimeFrame(adID, "5m")
	last15m, _ := s.GetClickCountByTimeFrame(adID, "15m")
	last1h, _ := s.GetClickCountByTimeFrame(adID, "1h")
	last24h, _ := s.GetClickCountByTimeFrame(adID, "24h")

	// Calculate CTR (assuming some impression data would be available)
	// For now, we'll use a placeholder calculation
	ctr := float64(0)
	if totalClicks > 0 {
		// Assuming 1000 impressions per click as placeholder
		ctr = float64(totalClicks) / (float64(totalClicks) * 1000) * 100
	}

	return &AnalyticsResponse{
		AdID:        adID,
		TotalClicks: int64(totalClicks),
		CTR:         ctr,
		TimeFrames: map[string]int64{
			"last_1_minute":   last1m,
			"last_5_minutes":  last5m,
			"last_15_minutes": last15m,
			"last_1_hour":     last1h,
			"last_24_hours":   last24h,
		},
		Timestamp: time.Now(),
	}, nil
}

// StartBatchProcessor starts a background goroutine to process batches periodically
func (s *AdsService) StartBatchProcessor() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := s.ProcessBatch(); err != nil {
				s.log.Logger.Errorf("Failed to process batch: %v", err)
			}
		}
	}()
}

type AnalyticsResponse struct {
	AdID        string           `json:"ad_id"`
	TotalClicks int64            `json:"total_clicks"`
	CTR         float64          `json:"ctr"`
	TimeFrames  map[string]int64 `json:"time_frames"`
	Timestamp   time.Time        `json:"timestamp"`
}
