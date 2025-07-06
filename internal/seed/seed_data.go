package seed

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/model"
	"gorm.io/gorm"
)

type Seeder struct {
	db *gorm.DB
}

func NewSeeder(db *gorm.DB) *Seeder {
	return &Seeder{db: db}
}

func (s *Seeder) SeedAll() error {
	log.Println("Starting data seeding...")

	// Check if data already exists
	var adCount int64
	if err := s.db.Model(&model.Ad{}).Count(&adCount).Error; err != nil {
		return fmt.Errorf("failed to count existing ads: %w", err)
	}

	if adCount > 0 {
		log.Printf("Data already exists (%d ads found), skipping seed", adCount)
		return nil
	}

	// Seed ads
	if err := s.seedAds(); err != nil {
		return fmt.Errorf("failed to seed ads: %w", err)
	}

	// Seed sample clicks
	if err := s.seedSampleClicks(); err != nil {
		return fmt.Errorf("failed to seed sample clicks: %w", err)
	}

	log.Println("Data seeding completed successfully")
	return nil
}

func (s *Seeder) seedAds() error {
	ads := []model.Ad{
		// Technology Ads
		{
			ID:          "tech-001",
			ImageURL:    "https://images.unsplash.com/photo-1518770660439-4636190af475?w=800",
			TargetURL:   "https://example.com/tech/smartphone",
			TotalClicks: 0,
		},
		{
			ID:          "tech-002",
			ImageURL:    "https://images.unsplash.com/photo-1496181133206-80ce9b88a853?w=800",
			TargetURL:   "https://example.com/tech/laptop",
			TotalClicks: 0,
		},
		{
			ID:          "tech-003",
			ImageURL:    "https://images.unsplash.com/photo-1583394838336-acd977736f90?w=800",
			TargetURL:   "https://example.com/tech/headphones",
			TotalClicks: 0,
		},

		// Fashion Ads
		{
			ID:          "fashion-001",
			ImageURL:    "https://images.unsplash.com/photo-1445205170230-053b83016050?w=800",
			TargetURL:   "https://example.com/fashion/summer-collection",
			TotalClicks: 0,
		},
		{
			ID:          "fashion-002",
			ImageURL:    "https://images.unsplash.com/photo-1556905055-8f358a7a47b2?w=800",
			TargetURL:   "https://example.com/fashion/shoes",
			TotalClicks: 0,
		},
		{
			ID:          "fashion-003",
			ImageURL:    "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=800",
			TargetURL:   "https://example.com/fashion/sneakers",
			TotalClicks: 0,
		},

		// Food & Beverage Ads
		{
			ID:          "food-001",
			ImageURL:    "https://images.unsplash.com/photo-1565299624946-b28f40a0ca4b?w=800",
			TargetURL:   "https://example.com/food/pizza-delivery",
			TotalClicks: 0,
		},
		{
			ID:          "food-002",
			ImageURL:    "https://images.unsplash.com/photo-1551024506-0bccd828d307?w=800",
			TargetURL:   "https://example.com/food/burger-joint",
			TotalClicks: 0,
		},
		{
			ID:          "food-003",
			ImageURL:    "https://images.unsplash.com/photo-1551024709-8f23befc6f87?w=800",
			TargetURL:   "https://example.com/food/healthy-meals",
			TotalClicks: 0,
		},

		// Travel Ads
		{
			ID:          "travel-001",
			ImageURL:    "https://images.unsplash.com/photo-1488646953014-85cb44e25828?w=800",
			TargetURL:   "https://example.com/travel/beach-vacation",
			TotalClicks: 0,
		},
		{
			ID:          "travel-002",
			ImageURL:    "https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=800",
			TargetURL:   "https://example.com/travel/mountain-adventure",
			TotalClicks: 0,
		},
		{
			ID:          "travel-003",
			ImageURL:    "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=800",
			TargetURL:   "https://example.com/travel/city-tours",
			TotalClicks: 0,
		},

		// Fitness Ads
		{
			ID:          "fitness-001",
			ImageURL:    "https://images.unsplash.com/photo-1571019613454-1cb2f99b2d8b?w=800",
			TargetURL:   "https://example.com/fitness/gym-membership",
			TotalClicks: 0,
		},
		{
			ID:          "fitness-002",
			ImageURL:    "https://images.unsplash.com/photo-1544367567-0f2fcb009e0b?w=800",
			TargetURL:   "https://example.com/fitness/yoga-classes",
			TotalClicks: 0,
		},
		{
			ID:          "fitness-003",
			ImageURL:    "https://images.unsplash.com/photo-1517836357463-d25dfeac3438?w=800",
			TargetURL:   "https://example.com/fitness/home-workout",
			TotalClicks: 0,
		},

		// Finance Ads
		{
			ID:          "finance-001",
			ImageURL:    "https://images.unsplash.com/photo-1554224155-6726b3ff858f?w=800",
			TargetURL:   "https://example.com/finance/credit-card",
			TotalClicks: 0,
		},
		{
			ID:          "finance-002",
			ImageURL:    "https://images.unsplash.com/photo-1579621970563-ebec7560ff3e?w=800",
			TargetURL:   "https://example.com/finance/investment-app",
			TotalClicks: 0,
		},
		{
			ID:          "finance-003",
			ImageURL:    "https://images.unsplash.com/photo-1559526324-593bc073d938?w=800",
			TargetURL:   "https://example.com/finance/loan-services",
			TotalClicks: 0,
		},

		// Education Ads
		{
			ID:          "edu-001",
			ImageURL:    "https://images.unsplash.com/photo-1513475382585-d06e58bcb0e0?w=800",
			TargetURL:   "https://example.com/education/online-courses",
			TotalClicks: 0,
		},
		{
			ID:          "edu-002",
			ImageURL:    "https://images.unsplash.com/photo-1524178232363-1fb2b075b655?w=800",
			TargetURL:   "https://example.com/education/programming-bootcamp",
			TotalClicks: 0,
		},
		{
			ID:          "edu-003",
			ImageURL:    "https://images.unsplash.com/photo-1427504494785-3a9ca7044f45?w=800",
			TargetURL:   "https://example.com/education/language-learning",
			TotalClicks: 0,
		},

		// Entertainment Ads
		{
			ID:          "entertainment-001",
			ImageURL:    "https://images.unsplash.com/photo-1489599849927-2ee91cede3ba?w=800",
			TargetURL:   "https://example.com/entertainment/streaming-service",
			TotalClicks: 0,
		},
		{
			ID:          "entertainment-002",
			ImageURL:    "https://images.unsplash.com/photo-1594909122845-11baa439b7bf?w=800",
			TargetURL:   "https://example.com/entertainment/gaming-console",
			TotalClicks: 0,
		},
		{
			ID:          "entertainment-003",
			ImageURL:    "https://images.unsplash.com/photo-1493225457124-a3eb161ffa5f?w=800",
			TargetURL:   "https://example.com/entertainment/music-festival",
			TotalClicks: 0,
		},
	}

	// Insert ads in batches
	batchSize := 10
	for i := 0; i < len(ads); i += batchSize {
		end := i + batchSize
		if end > len(ads) {
			end = len(ads)
		}

		batch := ads[i:end]
		if err := s.db.Create(&batch).Error; err != nil {
			return fmt.Errorf("failed to create ads batch: %w", err)
		}
	}

	log.Printf("Successfully seeded %d ads", len(ads))
	return nil
}

func (s *Seeder) seedSampleClicks() error {
	// Get all ads
	var ads []model.Ad
	if err := s.db.Find(&ads).Error; err != nil {
		return fmt.Errorf("failed to fetch ads: %w", err)
	}

	if len(ads) == 0 {
		return fmt.Errorf("no ads found to seed clicks for")
	}

	// Generate sample clicks for the last 30 days
	now := time.Now()
	startDate := now.AddDate(0, 0, -30) // 30 days ago

	var allClicks []model.Clicks
	rand.Seed(time.Now().UnixNano())

	// Generate different patterns for different ads
	for _, ad := range ads {
		clicks := s.generateClicksForAd(ad.ID, startDate, now)
		allClicks = append(allClicks, clicks...)
	}

	// Insert clicks in batches
	batchSize := 500
	for i := 0; i < len(allClicks); i += batchSize {
		end := i + batchSize
		if end > len(allClicks) {
			end = len(allClicks)
		}

		batch := allClicks[i:end]
		if err := s.db.Create(&batch).Error; err != nil {
			return fmt.Errorf("failed to create clicks batch: %w", err)
		}
	}

	// Update total clicks count for each ad
	for _, ad := range ads {
		var count int64
		if err := s.db.Model(&model.Clicks{}).Where("ad_id = ?", ad.ID).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to count clicks for ad %s: %w", ad.ID, err)
		}

		if err := s.db.Model(&ad).Update("total_clicks", count).Error; err != nil {
			return fmt.Errorf("failed to update total clicks for ad %s: %w", ad.ID, err)
		}
	}

	log.Printf("Successfully seeded %d sample clicks", len(allClicks))
	return nil
}

func (s *Seeder) generateClicksForAd(adID string, startDate, endDate time.Time) []model.Clicks {
	var clicks []model.Clicks

	// Different click patterns based on ad category
	var dailyRange, hourlyVariance int
	switch {
	case adID[:4] == "tech":
		dailyRange = 50 + rand.Intn(100) // 50-150 clicks per day
		hourlyVariance = 20
	case adID[:7] == "fashion":
		dailyRange = 30 + rand.Intn(80) // 30-110 clicks per day
		hourlyVariance = 15
	case adID[:4] == "food":
		dailyRange = 80 + rand.Intn(120) // 80-200 clicks per day
		hourlyVariance = 30
	case adID[:6] == "travel":
		dailyRange = 20 + rand.Intn(60) // 20-80 clicks per day
		hourlyVariance = 10
	default:
		dailyRange = 40 + rand.Intn(80) // 40-120 clicks per day
		hourlyVariance = 25
	}

	// Generate clicks for each day
	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		dailyClicks := dailyRange + rand.Intn(hourlyVariance) - hourlyVariance/2

		// Distribute clicks throughout the day with peak hours
		for i := 0; i < dailyClicks; i++ {
			hour := s.getRandomHourWithPeaks()
			minute := rand.Intn(60)
			second := rand.Intn(60)

			clickTime := time.Date(d.Year(), d.Month(), d.Day(), hour, minute, second, 0, time.UTC)

			click := model.Clicks{
				ID:            uuid.New().String(),
				AdID:          adID,
				IP:            s.generateRandomIP(),
				VideoPlayTime: rand.Intn(300) + 5, // 5-305 seconds
				Timestamp:     clickTime,
			}

			clicks = append(clicks, click)
		}
	}

	return clicks
}

func (s *Seeder) getRandomHourWithPeaks() int {
	// Peak hours: 9-11 AM, 2-4 PM, 7-9 PM
	peakHours := []int{9, 10, 11, 14, 15, 16, 19, 20, 21}

	if rand.Float32() < 0.6 { // 60% chance for peak hours
		return peakHours[rand.Intn(len(peakHours))]
	}

	// Return any hour for off-peak
	return rand.Intn(24)
}

func (s *Seeder) generateRandomIP() string {
	// Generate realistic IP addresses (avoiding private ranges mostly)
	ranges := []string{
		"203.%d.%d.%d", // Asia-Pacific
		"185.%d.%d.%d", // Europe
		"104.%d.%d.%d", // North America
		"190.%d.%d.%d", // South America
	}

	pattern := ranges[rand.Intn(len(ranges))]
	return fmt.Sprintf(pattern,
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256))
}

// SeedTestData creates minimal test data for development
func (s *Seeder) SeedTestData() error {
	// Check if test data already exists
	var count int64
	if err := s.db.Model(&model.Ad{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		log.Printf("Test data already exists (%d ads), skipping", count)
		return nil
	}

	testAds := []model.Ad{
		{
			ID:          "test-001",
			ImageURL:    "https://via.placeholder.com/800x400/FF5733/FFFFFF?text=Test+Ad+1",
			TargetURL:   "https://example.com/test1",
			TotalClicks: 0,
		},
		{
			ID:          "test-002",
			ImageURL:    "https://via.placeholder.com/800x400/33FF57/FFFFFF?text=Test+Ad+2",
			TargetURL:   "https://example.com/test2",
			TotalClicks: 0,
		},
		{
			ID:          "test-003",
			ImageURL:    "https://via.placeholder.com/800x400/3357FF/FFFFFF?text=Test+Ad+3",
			TargetURL:   "https://example.com/test3",
			TotalClicks: 0,
		},
	}

	if err := s.db.Create(&testAds).Error; err != nil {
		return fmt.Errorf("failed to create test ads: %w", err)
	}

	log.Printf("Successfully seeded %d test ads", len(testAds))
	return nil
}
