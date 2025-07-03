package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/model"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/logger"
)

const (
	consumerGroupID = "ad-clicks-group"
	topicName       = "ad-clicks"
	maxRetries      = 5
	retryDelay      = 2 * time.Second
)

type KafkaService struct {
	producer      sarama.SyncProducer
	consumerGroup sarama.ConsumerGroup
	log           *logger.Logger
	config        *sarama.Config
	brokers       []string
}

func NewKafkaService(brokers []string, log *logger.Logger) (*KafkaService, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = maxRetries
	config.Producer.Retry.Backoff = retryDelay

	// Add timeout settings
	config.Net.DialTimeout = 10 * time.Second
	config.Net.ReadTimeout = 10 * time.Second
	config.Net.WriteTimeout = 10 * time.Second

	// Try to connect with retries
	var producer sarama.SyncProducer
	var err error

	for i := 0; i < maxRetries; i++ {
		log.Logger.Info("Attempting to connect to Kafka brokers: %v (attempt %d/%d)", brokers, i+1, maxRetries)
		producer, err = sarama.NewSyncProducer(brokers, config)
		if err == nil {
			log.Logger.Info("Successfully connected to Kafka")
			break
		}
		log.Logger.Errorf("Failed to connect to Kafka (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryDelay)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create producer after %d attempts: %v", maxRetries, err)
	}

	return &KafkaService{
		producer: producer,
		config:   config,
		brokers:  brokers,
		log:      log,
	}, nil
}

// ClickConsumerHandler implements sarama.ConsumerGroupHandler
type ClickConsumerHandler struct {
	AdsService *AdsService
	log        *logger.Logger
	workerID   int
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *ClickConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	h.log.Logger.Infof("Worker %d: Consumer group setup", h.workerID)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *ClickConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.log.Logger.Infof("Worker %d: Consumer group cleanup", h.workerID)
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (h *ClickConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var click model.Clicks
		if err := json.Unmarshal(message.Value, &click); err != nil {
			h.log.Logger.Errorf("Worker %d: Failed to unmarshal click: %v", h.workerID, err)
			continue
		}

		if err := h.AdsService.ProcessClick(click); err != nil {
			h.log.Logger.Errorf("Worker %d: Failed to process click: %v", h.workerID, err)
			continue
		}

		session.MarkMessage(message, "")
	}
	return nil
}

func (s *KafkaService) StartConsumer(clickService *ClickService, numWorkers int) error {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	group, err := sarama.NewConsumerGroup(s.brokers, consumerGroupID, config)
	if err != nil {
		return fmt.Errorf("failed to create consumer group: %v", err)
	}
	s.consumerGroup = group

	// Create topic if it doesn't exist
	if err := s.CreateTopic(); err != nil {
		s.log.Logger.Warnf("Failed to create topic: %v", err)
	}

	// Start multiple workers
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			handler := &ClickConsumerHandler{
				clickService: clickService,
				log:          s.log,
				workerID:     workerID,
			}
			for {
				err := group.Consume(context.Background(), []string{topicName}, handler)
				if err != nil {
					s.log.Logger.Errorf("Worker %d: Error from consumer: %v", workerID, err)
				}
				// Check if context was cancelled, indicating shutdown
				if err == sarama.ErrClosedConsumerGroup {
					return
				}
				// Wait before retrying
				time.Sleep(time.Second * 5)
			}
		}(i)
	}

	return nil
}

func (s *KafkaService) Close() error {
	if err := s.producer.Close(); err != nil {
		return fmt.Errorf("failed to close producer: %v", err)
	}
	if s.consumerGroup != nil {
		if err := s.consumerGroup.Close(); err != nil {
			return fmt.Errorf("failed to close consumer group: %v", err)
		}
	}
	return nil
}

func (s *KafkaService) CreateTopic() error {
	admin, err := sarama.NewClusterAdmin(s.brokers, s.config)
	if err != nil {
		return fmt.Errorf("failed to create admin client: %v", err)
	}
	defer admin.Close()
	err = admin.CreateTopic(topicName, &sarama.TopicDetail{
		NumPartitions:     3,
		ReplicationFactor: 1,
	}, false)
	if err != nil {
		return fmt.Errorf("failed to create topic %v", err)
	}
	return nil
}
