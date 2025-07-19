
package config

import (
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10" // 
	"github.com/spf13/viper"
)

// Add 'validate' tags to the struct fields.
type Config struct {
	ServerPort   string `mapstructure:"SERVER_PORT" validate:"required,numeric"`
	DatabaseURL  string `mapstructure:"DATABASE_URL" validate:"required"`
	APISecretKey string `mapstructure:"API_SECRET_KEY" validate:"required,min=16"` // Example: require a minimum length

	MonitorDefaultInterval int `mapstructure:"MONITOR_DEFAULT_INTERVAL_SECONDS" validate:"required,gt=0"`
	MonitorCheckTimeoutSec int `mapstructure:"MONITOR_CHECK_TIMEOUT_SECONDS" validate:"required,gt=0"`
	SchedulerConcurrency   int `mapstructure:"SCHEDULER_CONCURRENCY" validate:"required,gt=0"`
}

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
			slog.Warn("Config file not found, using defaults and environment variables.")
		} else {
			return
		}
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return
	}

	
	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		// The error message from the validator is very informative.
		return config, fmt.Errorf("configuration validation failed: %w", err)
	}
	

	slog.Info("Configuration loaded successfully")
	return
}