// config/config.go

package config

import (
	"fmt"
	"log/slog" // <-- CHANGED: Import slog instead of log

	"github.com/spf13/viper"
)

// Config struct remains the same
type Config struct {
	ServerPort             string `mapstructure:"SERVER_PORT"`
	DatabaseURL            string `mapstructure:"DATABASE_URL"`
	APISecretKey           string `mapstructure:"API_SECRET_KEY"`
	MonitorDefaultInterval int    `mapstructure:"MONITOR_DEFAULT_INTERVAL_SECONDS"`
	MonitorCheckTimeoutSec int    `mapstructure:"MONITOR_CHECK_TIMEOUT_SECONDS"`
	SchedulerConcurrency   int    `mapstructure:"SCHEDULER_CONCURRENCY"`
}

// LoadConfig loads configuration from environment variables and/or config file
func LoadConfig() (config Config, err error) {
	viper.AddConfigPath("./")
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.SetEnvPrefix("GOLINKCHECK")
	viper.AutomaticEnv()

	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("DATABASE_URL", "sqlite:./golinkcheck.db")
	viper.SetDefault("MONITOR_DEFAULT_INTERVAL_SECONDS", 60)
	viper.SetDefault("MONITOR_CHECK_TIMEOUT_SECONDS", 10)
	viper.SetDefault("SCHEDULER_CONCURRENCY", 5)

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// <-- CHANGED: Use slog.Warn for notable but non-fatal events.
			slog.Warn("Config file not found, using defaults and environment variables.")
		} else {
			// This is a real error, so we return it.
			return
		}
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return
	}

	// Validation checks remain the same
	if config.APISecretKey == "" {
		return config, fmt.Errorf("API_SECRET_KEY is not set. The server will not start without it for security reasons")
	}
	if config.SchedulerConcurrency <= 0 {
		return config, fmt.Errorf("SCHEDULER_CONCURRENCY must be a positive integer")
	}
	if config.MonitorCheckTimeoutSec <= 0 {
		return config, fmt.Errorf("MONITOR_CHECK_TIMEOUT_SECONDS must be a positive integer")
	}

	// <-- CHANGED: Use a simple, secure log message. DO NOT log the config struct.
	slog.Info("Configuration loaded successfully")
	return
}