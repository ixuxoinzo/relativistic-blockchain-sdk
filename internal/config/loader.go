package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type ConfigLoader struct {
	viper *viper.Viper
}

func NewConfigLoader() *ConfigLoader {
	v := viper.New()
	return &ConfigLoader{viper: v}
}

func (cl *ConfigLoader) Load(configPath string) (*Config, error) {
	if configPath != "" {
		cl.viper.SetConfigFile(configPath)
	} else {
		cl.setupDefaultPaths()
	}

	cl.setupEnvironment()
	cl.setupDefaults()

	if err := cl.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var config Config
	if err := cl.viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cl.validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func (cl *ConfigLoader) setupDefaultPaths() {
	cl.viper.SetConfigName("config")
	cl.viper.SetConfigType("yaml")
	cl.viper.AddConfigPath(".")
	cl.viper.AddConfigPath("./configs")
	cl.viper.AddConfigPath("/etc/relativistic-sdk/")
	cl.viper.AddConfigPath("$HOME/.relativistic-sdk")
}

func (cl *ConfigLoader) setupEnvironment() {
	cl.viper.SetEnvPrefix("RELATIVISTIC")
	cl.viper.AutomaticEnv()

	cl.viper.BindEnv("server.address", "RELATIVISTIC_SERVER_ADDRESS")
	cl.viper.BindEnv("server.environment", "RELATIVISTIC_ENVIRONMENT")
	cl.viper.BindEnv("database.host", "RELATIVISTIC_DB_HOST")
	cl.viper.BindEnv("database.port", "RELATIVISTIC_DB_PORT")
	cl.viper.BindEnv("database.username", "RELATIVISTIC_DB_USERNAME")
	cl.viper.BindEnv("database.password", "RELATIVISTIC_DB_PASSWORD")
	cl.viper.BindEnv("redis.address", "RELATIVISTIC_REDIS_ADDRESS")
	cl.viper.BindEnv("security.jwt_secret", "RELATIVISTIC_JWT_SECRET")
	cl.viper.BindEnv("metrics.enabled", "RELATIVISTIC_METRICS_ENABLED")
}

func (cl *ConfigLoader) setupDefaults() {
	defaultConfig := DefaultConfig()

	cl.viper.SetDefault("server.address", defaultConfig.Server.Address)
	cl.viper.SetDefault("server.environment", defaultConfig.Server.Environment)
	cl.viper.SetDefault("server.shutdown_timeout", defaultConfig.Server.ShutdownTimeout)
	cl.viper.SetDefault("server.read_timeout", defaultConfig.Server.ReadTimeout)
	cl.viper.SetDefault("server.write_timeout", defaultConfig.Server.WriteTimeout)

	cl.viper.SetDefault("database.host", defaultConfig.Database.Host)
	cl.viper.SetDefault("database.port", defaultConfig.Database.Port)
	cl.viper.SetDefault("database.ssl_mode", defaultConfig.Database.SSLMode)

	cl.viper.SetDefault("redis.address", defaultConfig.Redis.Address)
	cl.viper.SetDefault("redis.pool_size", defaultConfig.Redis.PoolSize)

	cl.viper.SetDefault("security.token_expiry", defaultConfig.Security.TokenExpiry)
	cl.viper.SetDefault("security.rate_limit", defaultConfig.Security.RateLimit)
	cl.viper.SetDefault("security.rate_limit_window", defaultConfig.Security.RateLimitWindow)
	cl.viper.SetDefault("security.cors_origins", defaultConfig.Security.CORSOrigins)

	cl.viper.SetDefault("metrics.enabled", defaultConfig.Metrics.Enabled)
	cl.viper.SetDefault("metrics.port", defaultConfig.Metrics.Port)
	cl.viper.SetDefault("metrics.path", defaultConfig.Metrics.Path)
	cl.viper.SetDefault("metrics.push_interval", defaultConfig.Metrics.PushInterval)

	cl.viper.SetDefault("network.bootstrap_nodes", defaultConfig.Network.BootstrapNodes)
	cl.viper.SetDefault("network.peer_discovery", defaultConfig.Network.PeerDiscovery)
	cl.viper.SetDefault("network.max_peers", defaultConfig.Network.MaxPeers)
	cl.viper.SetDefault("network.listen_address", defaultConfig.Network.ListenAddress)

	cl.viper.SetDefault("logging.level", defaultConfig.Logging.Level)
	cl.viper.SetDefault("logging.format", defaultConfig.Logging.Format)
	cl.viper.SetDefault("logging.output", defaultConfig.Logging.Output)
}

func (cl *ConfigLoader) validateConfig(config *Config) error {
	if config.Server.Address == "" {
		return fmt.Errorf("server address is required")
	}

	if config.Security.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if len(config.Security.JWTSecret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters")
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Redis.Address == "" {
		return fmt.Errorf("redis address is required")
	}

	return nil
}

func Load() (*Config, error) {
	loader := NewConfigLoader()

	configPath := os.Getenv("RELATIVISTIC_CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	return loader.Load(configPath)
}

func LoadForEnvironment(environment string) (*Config, error) {
	loader := NewConfigLoader()

	var configPath string
	switch environment {
	case "production":
		configPath = "configs/config.prod.yaml"
	case "staging":
		configPath = "configs/config.staging.yaml"
	case "development":
		configPath = "configs/config.dev.yaml"
	default:
		configPath = "configs/config.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "config.yaml"
	}

	return loader.Load(configPath)
}
