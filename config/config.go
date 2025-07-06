package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	HttpHost         string `mapstructure:"HTTP_HOST"`
	HttpPort         string `mapstructure:"HTTP_PORT"`
	LogFile          string `mapstructure:"LOG_FILE"`
	NATSURL          string `mapstructure:"NATS_URL"`
	PostgresHost     string `mapstructure:"POSTGRES_HOST"`
	PostgresPort     string `mapstructure:"POSTGRES_PORT"`
	PostgresUser     string `mapstructure:"POSTGRES_USER"`
	PostgresPassword string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresDBName   string `mapstructure:"POSTGRES_DB"`
	RedisHost        string `mapstructure:"REDIS_HOST"`
	RedisPort        string `mapstructure:"REDIS_PORT"`
	RedisPassword    string `mapstructure:"REDIS_PASSWORD"`
	RedisDB          int    `mapstructure:"REDIS_DB"`
}

func NewConfig() *Config {
	viper.AutomaticEnv()

	// Load .env file support
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read .env file: %v", err)
	}

	viper.SetDefault("HTTP_PORT", "8080")
	viper.SetDefault("LOG_FILE", "app.log")

	config := &Config{
		HttpHost:         viper.GetString("HTTP_HOST"),
		HttpPort:         viper.GetString("HTTP_PORT"),
		LogFile:          viper.GetString("LOG_FILE"),
		NATSURL:          viper.GetString("NATS_URL"),
		PostgresHost:     viper.GetString("POSTGRES_HOST"),
		PostgresPort:     viper.GetString("POSTGRES_PORT"),
		PostgresUser:     viper.GetString("POSTGRES_USER"),
		PostgresPassword: viper.GetString("POSTGRES_PASSWORD"),
		PostgresDBName:   viper.GetString("POSTGRES_DB"),
		RedisHost:        viper.GetString("REDIS_HOST"),
		RedisPort:        viper.GetString("REDIS_PORT"),
		RedisPassword:    viper.GetString("REDIS_PASSWORD"),
		RedisDB:          viper.GetInt("REDIS_DB"),
	}

	config.Validate()
	return config
}

func (c *Config) Validate() {
	var missing []string

	if c.HttpHost == "" {
		missing = append(missing, "HTTP_HOST")
	}
	if c.HttpPort == "" {
		missing = append(missing, "HTTP_PORT")
	}
	if c.LogFile == "" {
		missing = append(missing, "LOG_FILE")
	}
	if c.NATSURL == "" {
		missing = append(missing, "NATS_URL")
	}
	if c.PostgresHost == "" {
		missing = append(missing, "POSTGRES_HOST")
	}
	if c.PostgresPort == "" {
		missing = append(missing, "POSTGRES_PORT")
	}
	if c.PostgresUser == "" {
		missing = append(missing, "POSTGRES_USER")
	}
	if c.PostgresPassword == "" {
		missing = append(missing, "POSTGRES_PASSWORD")
	}
	if c.PostgresDBName == "" {
		missing = append(missing, "POSTGRES_DB")
	}
	if c.RedisHost == "" {
		missing = append(missing, "REDIS_HOST")
	}
	if c.RedisPort == "" {
		missing = append(missing, "REDIS_PORT")
	}

	if len(missing) > 0 {
		log.Println("Missing required configuration values:")
		for _, key := range missing {
			fmt.Println(" -", key)
		}
		panic("configuration validation failed")
	}
}
