package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	Network  NetworkConfig  `mapstructure:"network"`
	Security SecurityConfig `mapstructure:"security"`
}

type ServerConfig struct {
	Address         string        `mapstructure:"address"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	EnableCORS      bool          `mapstructure:"enable_cors"`
}

type RedisConfig struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MetricsConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Port       int    `mapstructure:"port"`
	Path       string `mapstructure:"path"`
}

type NetworkConfig struct {
	PropagationFactor float64       `mapstructure:"propagation_factor"`
	SafetyFactor      float64       `mapstructure:"safety_factor"`
	MaxAcceptableDelay time.Duration `mapstructure:"max_acceptable_delay"`
	MonitoringInterval time.Duration `mapstructure:"monitoring_interval"`
}

type SecurityConfig struct {
	EnableTimestampValidation bool    `mapstructure:"enable_timestamp_validation"`
	MaxConfidenceThreshold    float64 `mapstructure:"max_confidence_threshold"`
	EnableNodeAuth            bool    `mapstructure:"enable_node_auth"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/relativistic-sdk/")

	// Set defaults
	setDefaults()

	// Read config
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	viper.SetDefault("server.address", ":8080")
	viper.SetDefault("server.shutdown_timeout", "30s")
	viper.SetDefault("server.enable_cors", true)

	viper.SetDefault("redis.address", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.port", 9090)
	viper.SetDefault("metrics.path", "/metrics")

	viper.SetDefault("network.propagation_factor", 1.5)
	viper.SetDefault("network.safety_factor", 2.0)
	viper.SetDefault("network.max_acceptable_delay", "30s")
	viper.SetDefault("network.monitoring_interval", "60s")

	viper.SetDefault("security.enable_timestamp_validation", true)
	viper.SetDefault("security.max_confidence_threshold", 0.8)
	viper.SetDefault("security.enable_node_auth", false)
}
