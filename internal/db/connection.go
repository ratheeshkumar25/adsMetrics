package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ratheeshkumar25/adsmetrictracker/config"
	"github.com/ratheeshkumar25/adsmetrictracker/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	PostgresDB *gorm.DB
	RedisDB    *redis.Client
}

type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxIdleConns int
	MaxOpenConns int
	MaxLifetime  time.Duration
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	// PostgreSQL connection
	db, err := initPostgreSQL(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL: %w", err)
	}

	// Redis connection
	rdb, err := initRedis(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Auto-migrate models
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &Database{
		PostgresDB: db,
		RedisDB:    rdb,
	}, nil
}

func initPostgreSQL(cfg *config.Config) (*gorm.DB, error) {
	dbConfig := DatabaseConfig{
		Host:         cfg.PostgresHost,
		Port:         cfg.PostgresPort,
		User:         cfg.PostgresUser,
		Password:     cfg.PostgresPassword,
		DBName:       cfg.PostgresDBName,
		SSLMode:      "disable",
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		MaxLifetime:  time.Hour,
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.DBName,
		dbConfig.SSLMode,
	)

	// Configure GORM logger
	gormLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent, // Silent in production
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(dbConfig.MaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	log.Println("PostgreSQL connection established successfully")
	return db, nil
}

func initRedis(cfg *config.Config) (*redis.Client, error) {
	redisConfig := RedisConfig{
		Host:     cfg.RedisHost,
		Port:     cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	addr := fmt.Sprintf("%s:%s", redisConfig.Host, redisConfig.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     redisConfig.Password,
		DB:           redisConfig.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis connection established successfully")
	return rdb, nil
}

func runMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	if err := db.AutoMigrate(&model.Ad{}, &model.Clicks{}); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create indexes for better performance
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func createIndexes(db *gorm.DB) error {
	// Index on clicks.ad_id for faster analytics queries
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_clicks_ad_id ON clicks(ad_id)").Error; err != nil {
		return err
	}

	// Index on clicks.timestamp for time-based queries
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_clicks_timestamp ON clicks(timestamp)").Error; err != nil {
		return err
	}

	// Composite index on ad_id and timestamp for analytics
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_clicks_ad_timestamp ON clicks(ad_id, timestamp)").Error; err != nil {
		return err
	}

	// Index on clicks.ip for duplicate detection
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_clicks_ip ON clicks(ip)").Error; err != nil {
		return err
	}

	log.Println("Database indexes created successfully")
	return nil
}

func (d *Database) Close() error {
	// Close PostgreSQL connection
	if d.PostgresDB != nil {
		sqlDB, err := d.PostgresDB.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying sql.DB: %w", err)
		}
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close PostgreSQL connection: %w", err)
		}
	}

	// Close Redis connection
	if d.RedisDB != nil {
		if err := d.RedisDB.Close(); err != nil {
			return fmt.Errorf("failed to close Redis connection: %w", err)
		}
	}

	log.Println("Database connections closed successfully")
	return nil
}

func (d *Database) Health() error {
	// Check PostgreSQL health
	if d.PostgresDB != nil {
		sqlDB, err := d.PostgresDB.DB()
		if err != nil {
			return fmt.Errorf("PostgreSQL health check failed: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := sqlDB.PingContext(ctx); err != nil {
			return fmt.Errorf("PostgreSQL ping failed: %w", err)
		}
	}

	// Check Redis health
	if d.RedisDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if _, err := d.RedisDB.Ping(ctx).Result(); err != nil {
			return fmt.Errorf("Redis ping failed: %w", err)
		}
	}

	return nil
}
