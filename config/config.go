package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	HttpHost      string   `mapstructure:"HTTP_HOST"`
	HttpPort      string   `mapstructure:"HTTP_PORT"`
	LogFile       string   `mapstructure:"LOG_FILE"`
	KafkaBrokers  []string `mapstructure:"KAFKA_BROKER"` // comma-separated
	MysqlHost     string   `mapstructure:"MYSQL_HOST"`
	MysqlPort     string   `mapstructure:"MYSQL_PORT"`
	MysqlUser     string   `mapstructure:"MYSQL_USER"`
	MysqlPassword string   `mapstructure:"MYSQL_PASSWORD"`
	MysqlDBName   string   `mapstructure:"MYSQL_DB"`
}

func NewConfig() *Config {
	viper.AutomaticEnv()

	// Optional file support
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	_ = viper.ReadInConfig()

	viper.SetDefault("HTTP_PORT", "8080")
	viper.SetDefault("LOG_FILE", "app.log")

	config := &Config{
		HttpHost:      viper.GetString("HTTP_HOST"),
		HttpPort:      viper.GetString("HTTP_PORT"),
		LogFile:       viper.GetString("LOG_FILE"),
		KafkaBrokers:  strings.Split(viper.GetString("KAFKA_BROKER"), ","),
		MysqlHost:     viper.GetString("MYSQL_HOST"),
		MysqlPort:     viper.GetString("MYSQL_PORT"),
		MysqlUser:     viper.GetString("MYSQL_USER"),
		MysqlPassword: viper.GetString("MYSQL_PASSWORD"),
		MysqlDBName:   viper.GetString("MYSQL_DB"),
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
	if len(c.KafkaBrokers) == 0 || c.KafkaBrokers[0] == "" {
		missing = append(missing, "KAFKA_BROKER")
	}
	if c.MysqlHost == "" {
		missing = append(missing, "MYSQL_HOST")
	}
	if c.MysqlPort == "" {
		missing = append(missing, "MYSQL_PORT")
	}
	if c.MysqlUser == "" {
		missing = append(missing, "MYSQL_USER")
	}
	if c.MysqlPassword == "" {
		missing = append(missing, "MYSQL_PASSWORD")
	}
	if c.MysqlDBName == "" {
		missing = append(missing, "MYSQL_DB")
	}

	if len(missing) > 0 {
		log.Println("Missing required configuration values:")
		for _, key := range missing {
			fmt.Println(" -", key)
		}
		panic("configuration validation failed")
	}
}
