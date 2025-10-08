package config

import (
	"time"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Security SecurityConfig `yaml:"security"`
	Metrics  MetricsConfig  `yaml:"metrics"`
	Network  NetworkConfig  `yaml:"network"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
	Address         string        `yaml:"address"`
	Environment     string        `yaml:"environment"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
}

type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

type SecurityConfig struct {
	JWTSecret       string        `yaml:"jwt_secret"`
	TokenExpiry     time.Duration `yaml:"token_expiry"`
	RateLimit       int           `yaml:"rate_limit"`
	RateLimitWindow time.Duration `yaml:"rate_limit_window"`
	CORSOrigins     []string      `yaml:"cors_origins"`
}

type MetricsConfig struct {
	Enabled      bool          `yaml:"enabled"`
	Port         int           `yaml:"port"`
	Path         string        `yaml:"path"`
	PushGateway  string        `yaml:"push_gateway"`
	PushInterval time.Duration `yaml:"push_interval"`
}

type NetworkConfig struct {
	BootstrapNodes []string `yaml:"bootstrap_nodes"`
	PeerDiscovery  bool     `yaml:"peer_discovery"`
	MaxPeers       int      `yaml:"max_peers"`
	ListenAddress  string   `yaml:"listen_address"`
	ExternalIP     string   `yaml:"external_ip"`
}

type LoggingConfig struct {
	Level    string `yaml:"level"`
	Format   string `yaml:"format"`
	Output   string `yaml:"output"`
	FilePath string `yaml:"file_path"`
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Address:         ":8080",
			Environment:     "development",
			ShutdownTimeout: 30 * time.Second,
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
		},
		Database: DatabaseConfig{
			Host:    "localhost",
			Port:    5432,
			SSLMode: "disable",
		},
		Redis: RedisConfig{
			Address:  "localhost:6379",
			PoolSize: 100,
		},
		Security: SecurityConfig{
			TokenExpiry:     24 * time.Hour,
			RateLimit:       100,
			RateLimitWindow: time.Minute,
			CORSOrigins:     []string{"*"},
		},
		Metrics: MetricsConfig{
			Enabled:      true,
			Port:         9090,
			Path:         "/metrics",
			PushInterval: 60 * time.Second,
		},
		Network: NetworkConfig{
			BootstrapNodes: []string{
				"bootstrap1.relativistic-sdk.com:8080",
				"bootstrap2.relativistic-sdk.com:8080",
			},
			PeerDiscovery: true,
			MaxPeers:      50,
			ListenAddress: ":8080",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
	}
}
