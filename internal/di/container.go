package di

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ratheeshkumar25/adsmetrictracker/config"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/db"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/handlers"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/handlers/routes"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/model"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/repo"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/services"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/breaker"
	"github.com/ratheeshkumar25/adsmetrictracker/pkg/logger"
)

// Init initializes the HTTP server and related components following the new DI format
func Init() (*gin.Engine, error) {
	// Initialize configuration
	cfg := config.NewConfig()

	// Initialize logger
	logger, err := logger.NewLogger(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Connect to the database
	database, err := db.NewDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize circuit breaker
	circuitBreaker := breaker.NewCircuitBreaker(5, 10*time.Second, "ads-metric-circuit-breaker")

	// Create a new repository instance for Ads
	adsRepo := repo.NewAdsRepository(database.PostgresDB)

	// Initialize NATS Service
	natsService, err := services.NewNATSService(cfg.NATSURL, logger)
	if err != nil {
		logger.Logger.Warnf("NATS service initialization failed: %v", err)
		// Continue without NATS - application will work in direct mode
	}

	// Create a new service instance for Ads (use case layer)
	adsService := services.NewAdsService(adsRepo, logger, natsService, circuitBreaker)

	// Create a new handler instance for Ads (delivery layer)
	adsHandler := handlers.NewHandler(adsService, logger)

	// Create new routes and pass in the handler
	router := routes.NewRouter(adsHandler)

	// Setup routes and return the initialized server
	server := router.SetupRoutes(logger)

	// Start background services
	adsService.StartBatchProcessor()
	if natsService != nil {
		clickService := services.NewClickService(adsService)
		if err := natsService.StartConsumer(clickService, 5); err != nil {
			logger.Logger.Errorf("Failed to start NATS consumer: %v", err)
		}
	}

	return server, nil
}

// Container holds all dependencies
type Container struct {
	Config   *config.Config
	Logger   *logger.Logger
	Database *db.Database

	// Repositories
	AdsRepo    repo.AdsRepoInt
	ClicksRepo repo.AdsRepoInt

	// Services
	AdsService  services.AdsServiceInt
	NATSService *services.NATSService

	// Circuit Breaker
	CircuitBreaker *breaker.CircuitBreaker

	// HTTP Components
	Handler *handlers.Handler
	Router  *routes.Router
}

// AdsRepoInterface defines the interface for ads repository
type AdsRepoInterface interface {
	GetAllAds() ([]model.Ad, error)
	GetAdByID(id string) (*model.Ad, error)
	CreateAd(ad *model.Ad) error
	UpdateAd(ad *model.Ad) error
	DeleteAd(id string) error
	IncrementClickCount(adID string) error
}

// ClicksRepoInterface defines the interface for clicks repository
type ClicksRepoInterface interface {
	SaveClick(click *model.Clicks) error
	SaveBatchClicks(clicks []model.Clicks) error
	GetClickCountByTimeFrame(adID string, start, end time.Time) (int, error)
	GetClickCountByIP(adID string, ip string) (int, error)
	GetRecentClicks(adID string, minutes int) ([]model.Clicks, error)
}

// AdsServiceInterface defines the interface for ads service
type AdsServiceInterface interface {
	GetAllAds() ([]model.Ad, error)
	GetAdByID(id string) (*model.Ad, error)
	CreateAd(ad *model.Ad) error
	UpdateAd(ad *model.Ad) error
	DeleteAd(id string) error
	ProcessClick(click model.Clicks) error
	GetAnalytics(adID string) (*services.AnalyticsResponse, error)
	PublishClick(click model.Clicks) error
}

// NewContainer creates and initializes a new DI container
func NewContainer() (*Container, error) {
	container := &Container{}

	if err := container.initConfig(); err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	if err := container.initLogger(); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	if err := container.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := container.initCircuitBreaker(); err != nil {
		return nil, fmt.Errorf("failed to initialize circuit breaker: %w", err)
	}

	if err := container.initRepositories(); err != nil {
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	if err := container.initServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	if err := container.initHandlers(); err != nil {
		return nil, fmt.Errorf("failed to initialize handlers: %w", err)
	}

	return container, nil
}

func (c *Container) initConfig() error {
	c.Config = config.NewConfig()
	return nil
}

func (c *Container) initLogger() error {
	logger, err := logger.NewLogger(c.Config)
	if err != nil {
		return err
	}
	c.Logger = logger
	return nil
}

func (c *Container) initDatabase() error {
	database, err := db.NewDatabase(c.Config)
	if err != nil {
		return err
	}
	c.Database = database
	return nil
}

func (c *Container) initCircuitBreaker() error {
	c.CircuitBreaker = breaker.NewCircuitBreaker(5, 10*time.Second, "ads-metric-circuit-breaker")
	return nil
}

func (c *Container) initRepositories() error {
	// Initialize Ads Repository
	c.AdsRepo = repo.NewAdsRepository(c.Database.PostgresDB)

	// Initialize Clicks Repository
	//ClicksRepo = repo.NewClicksRepository(c.Database.PostgresDB)

	return nil
}

func (c *Container) initServices() error {
	// Initialize NATS Service
	natsService, err := services.NewNATSService(c.Config.NATSURL, c.Logger)
	if err != nil {
		c.Logger.Logger.Warnf("NATS service initialization failed: %v", err)
		// Continue without NATS - application will work in direct mode
	}
	c.NATSService = natsService

	// Initialize Ads Service
	c.AdsService = services.NewAdsService(
		c.AdsRepo.(*repo.AdsRepository),
		c.Logger,
		c.NATSService,
		c.CircuitBreaker,
	)

	// Start background services
	if err := c.startBackgroundServices(); err != nil {
		return err
	}

	return nil
}

func (c *Container) initHandlers() error {
	// Initialize Handler
	c.Handler = handlers.NewHandler(c.AdsService.(*services.AdsService), c.Logger)

	// Initialize Router
	c.Router = routes.NewRouter(c.Handler)

	return nil
}

func (c *Container) startBackgroundServices() error {
	// Start batch processor
	if adsService, ok := c.AdsService.(*services.AdsService); ok {
		adsService.StartBatchProcessor()
	}

	// Start NATS consumer if available
	if c.NATSService != nil {
		clickService := services.NewClickService(c.AdsService.(*services.AdsService))
		if err := c.NATSService.StartConsumer(clickService, 5); err != nil {
			c.Logger.Logger.Errorf("Failed to start NATS consumer: %v", err)
			// Not a fatal error - continue without NATS consumer
		}
	}

	return nil
}

// Cleanup gracefully shuts down all services
func (c *Container) Cleanup(ctx context.Context) error {
	c.Logger.Logger.Info("Starting graceful shutdown...")

	// Close NATS connections
	if c.NATSService != nil {
		if err := c.NATSService.Close(); err != nil {
			c.Logger.Logger.Errorf("Failed to close NATS service: %v", err)
		}
	}

	// Close database connections
	if c.Database != nil {
		if err := c.Database.Close(); err != nil {
			c.Logger.Logger.Errorf("Failed to close database: %v", err)
		}
	}

	// Close logger
	// No Close method for logger, nothing to do here

	c.Logger.Logger.Info("Graceful shutdown completed")
	return nil
}

// HealthCheck performs health checks on all components
func (c *Container) HealthCheck() error {
	// Check database health
	if c.Database != nil {
		if err := c.Database.Health(); err != nil {
			return fmt.Errorf("database health check failed: %w", err)
		}
	}

	// Add more health checks as needed
	return nil
}
