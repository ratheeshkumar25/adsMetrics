package services

import (
	"github.com/ratheeshkumar25/adsmetrictracker/internal/model"
)

// ClickService handles click processing for Kafka consumers
type ClickService struct {
	AdsService *AdsService
}

// NewClickService creates a new ClickService
func NewClickService(adsService *AdsService) *ClickService {
	return &ClickService{
		AdsService: adsService,
	}
}

// ProcessClick processes a click event (implements the interface expected by Kafka consumer)
func (cs *ClickService) ProcessClick(click model.Clicks) error {
	return cs.AdsService.ProcessClick(click)
}
