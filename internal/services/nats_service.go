package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/model"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/logger"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/metrics"
)

const (
	subjectName = "ad.clicks"
	queueGroup  = "ad-clicks-workers"
	maxRetries  = 5
	retryDelay  = 2 * time.Second
)

type NATSService struct {
	conn    *nats.Conn
	log     *logger.Logger
	natsURL string
	subs    []*nats.Subscription
}

// NewNATSService creates a new NATS service instance
func NewNATSService(natsURL string, log *logger.Logger) (*NATSService, error) {
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	// Connection options with retry and timeout
	opts := []nats.Option{
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(5),
		nats.Timeout(10 * time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Logger.Warnf("NATS disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Logger.Info("NATS reconnected")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Logger.Info("NATS connection closed")
		}),
	}

	var conn *nats.Conn
	var err error

	// Try to connect with retries
	for i := 0; i < maxRetries; i++ {
		log.Logger.Infof("Attempting to connect to NATS server: %s (attempt %d/%d)", natsURL, i+1, maxRetries)
		conn, err = nats.Connect(natsURL, opts...)
		if err == nil {
			log.Logger.Info("Successfully connected to NATS")
			break
		}
		log.Logger.Errorf("Failed to connect to NATS (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryDelay)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS after %d attempts: %v", maxRetries, err)
	}

	return &NATSService{
		conn:    conn,
		log:     log,
		natsURL: natsURL,
		subs:    make([]*nats.Subscription, 0),
	}, nil
}

// PublishClick publishes a click event to NATS
func (s *NATSService) PublishClick(click model.Clicks) error {
	data, err := json.Marshal(click)
	if err != nil {
		metrics.RecordError("marshal_click_error", "nats_service")
		return fmt.Errorf("failed to marshal click: %w", err)
	}

	err = s.conn.Publish(subjectName, data)
	if err != nil {
		metrics.RecordError("nats_publish_error", "nats_service")
		return fmt.Errorf("failed to publish click to NATS: %w", err)
	}

	s.log.Logger.Debugf("Click published to NATS subject: %s", subjectName)
	return nil
}

// StartConsumer starts NATS consumers for processing click events
func (s *NATSService) StartConsumer(clickService *ClickService, numWorkers int) error {
	// Create multiple queue subscribers for load balancing
	for i := 0; i < numWorkers; i++ {
		sub, err := s.conn.QueueSubscribe(subjectName, queueGroup, func(msg *nats.Msg) {
			var click model.Clicks
			if err := json.Unmarshal(msg.Data, &click); err != nil {
				s.log.Logger.Errorf("Failed to unmarshal click: %v", err)
				return
			}

			if err := clickService.ProcessClick(click); err != nil {
				s.log.Logger.Errorf("Failed to process click: %v", err)
				// Implement retry logic if needed
				return
			}

			s.log.Logger.Debugf("Successfully processed click for ad: %s", click.AdID)
		})

		if err != nil {
			return fmt.Errorf("failed to subscribe to NATS subject: %v", err)
		}

		s.subs = append(s.subs, sub)
		s.log.Logger.Infof("Started NATS consumer worker %d", i+1)
	}

	s.log.Logger.Infof("Started %d NATS consumer workers", numWorkers)
	return nil
}

// Close gracefully closes the NATS connection and unsubscribes
func (s *NATSService) Close() error {
	// Unsubscribe from all subscriptions
	for _, sub := range s.subs {
		if err := sub.Unsubscribe(); err != nil {
			s.log.Logger.Errorf("Failed to unsubscribe: %v", err)
		}
	}

	// Close the connection
	if s.conn != nil {
		s.conn.Close()
	}

	s.log.Logger.Info("NATS service closed")
	return nil
}

// IsConnected checks if NATS connection is active
func (s *NATSService) IsConnected() bool {
	return s.conn != nil && s.conn.IsConnected()
}

// GetStats returns NATS connection statistics
func (s *NATSService) GetStats() nats.Statistics {
	if s.conn != nil {
		return s.conn.Stats()
	}
	return nats.Statistics{}
}

// Health checks the health of the NATS service
func (s *NATSService) Health() error {
	if !s.IsConnected() {
		return fmt.Errorf("NATS connection is not active")
	}
	return nil
}
