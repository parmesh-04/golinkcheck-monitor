// config/config.go
package config

import (
	"fmt"
	"log"
	

	"github.com/spf13/viper"
)

// Config holds all application configurations
type Config struct {
	ServerPort             string `mapstructure:"SERVER_PORT"`
	DatabaseURL            string `mapstructure:"DATABASE_URL"`
	MonitorDefaultInterval int    `mapstructure:"MONITOR_DEFAULT_INTERVAL_SECONDS"`
	MonitorCheckTimeoutSec int    `mapstructure:"MONITOR_CHECK_TIMEOUT_SECONDS"`
	SchedulerConcurrency   int    `mapstructure:"SCHEDULER_CONCURRENCY"`
}

// LoadConfig loads configuration from environment variables and/or config file
func LoadConfig() (config Config, err error) {
	viper.AddConfigPath("./")              // Set config file search path
	viper.SetConfigName("app")             // Set config file name (without extension)
	viper.SetConfigType("env")             // Set config file type

	viper.SetEnvPrefix("GOLINKCHECK")      // Set environment variable prefix
	viper.AutomaticEnv()                   // Enable automatic environment variable binding

	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("DATABASE_URL", "sqlite:./golinkcheck.db")
	viper.SetDefault("MONITOR_DEFAULT_INTERVAL_SECONDS", 60)
	viper.SetDefault("MONITOR_CHECK_TIMEOUT_SECONDS", 10)
	viper.SetDefault("SCHEDULER_CONCURRENCY", 5)

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found, using defaults and environment variables.")
		} else {
			return
		}
	}

	err = viper.Unmarshal(&config)         // Populate config struct with values
	if err != nil {
		return
	}

	if config.SchedulerConcurrency <= 0 {
		return config, fmt.Errorf("SCHEDULER_CONCURRENCY must be a positive integer")
	}
	if config.MonitorCheckTimeoutSec <= 0 {
		return config, fmt.Errorf("MONITOR_CHECK_TIMEOUT_SECONDS must be a positive integer")
	}

	log.Printf("Configuration loaded: %+v\n", config)
	return
}
