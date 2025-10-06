package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

type ConfigValidator struct {
	errors []string
}

func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		errors: make([]string, 0),
	}
}

func (cv *ConfigValidator) Validate(config *Config) error {
	cv.validateServerConfig(&config.Server)
	cv.validateDatabaseConfig(&config.Database)
	cv.validateRedisConfig(&config.Redis)
	cv.validateSecurityConfig(&config.Security)
	cv.validateMetricsConfig(&config.Metrics)
	cv.validateNetworkConfig(&config.Network)
	cv.validateLoggingConfig(&config.Logging)

	if len(cv.errors) > 0 {
		return fmt.Errorf("config validation failed: %s", strings.Join(cv.errors, "; "))
	}

	return nil
}

func (cv *ConfigValidator) validateServerConfig(config *ServerConfig) {
	if config.Address == "" {
		cv.addError("server address is required")
	}

	if config.Environment == "" {
		cv.addError("server environment is required")
	}

	validEnvironments := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}
	if !validEnvironments[config.Environment] {
		cv.addError("invalid server environment: " + config.Environment)
	}

	if config.ShutdownTimeout <= 0 {
		cv.addError("shutdown timeout must be positive")
	}

	if config.ReadTimeout <= 0 {
		cv.addError("read timeout must be positive")
	}

	if config.WriteTimeout <= 0 {
		cv.addError("write timeout must be positive")
	}
}

func (cv *ConfigValidator) validateDatabaseConfig(config *DatabaseConfig) {
	if config.Host == "" {
		cv.addError("database host is required")
	}

	if config.Port <= 0 || config.Port > 65535 {
		cv.addError("database port must be between 1 and 65535")
	}

	if config.Name == "" {
		cv.addError("database name is required")
	}

	validSSLMode := map[string]bool{
		"disable": true,
		"require": true,
		"verify-ca": true,
		"verify-full": true,
	}
	if !validSSLMode[config.SSLMode] {
		cv.addError("invalid database SSL mode: " + config.SSLMode)
	}
}

func (cv *ConfigValidator) validateRedisConfig(config *RedisConfig) {
	if config.Address == "" {
		cv.addError("redis address is required")
	}

	if config.PoolSize <= 0 {
		cv.addError("redis pool size must be positive")
	}

	if config.DB < 0 || config.DB > 15 {
		cv.addError("redis database must be between 0 and 15")
	}
}

func (cv *ConfigValidator) validateSecurityConfig(config *SecurityConfig) {
	if config.JWTSecret == "" {
		cv.addError("JWT secret is required")
	}

	if len(config.JWTSecret) < 32 {
		cv.addError("JWT secret must be at least 32 characters")
	}

	if config.TokenExpiry <= 0 {
		cv.addError("token expiry must be positive")
	}

	if config.RateLimit <= 0 {
		cv.addError("rate limit must be positive")
	}

	if config.RateLimitWindow <= 0 {
		cv.addError("rate limit window must be positive")
	}

	if len(config.CORSOrigins) == 0 {
		cv.addError("at least one CORS origin is required")
	}
}

func (cv *ConfigValidator) validateMetricsConfig(config *MetricsConfig) {
	if config.Port <= 0 || config.Port > 65535 {
		cv.addError("metrics port must be between 1 and 65535")
	}

	if config.Path == "" {
		cv.addError("metrics path is required")
	}

	if config.PushInterval <= 0 {
		cv.addError("metrics push interval must be positive")
	}

	if config.PushGateway != "" {
		if _, err := url.Parse(config.PushGateway); err != nil {
			cv.addError("invalid metrics push gateway URL")
		}
	}
}

func (cv *ConfigValidator) validateNetworkConfig(config *NetworkConfig) {
	if config.ListenAddress == "" {
		cv.addError("network listen address is required")
	}

	if config.MaxPeers <= 0 {
		cv.addError("max peers must be positive")
	}

	for _, node := range config.BootstrapNodes {
		if _, _, err := net.SplitHostPort(node); err != nil {
			cv.addError("invalid bootstrap node address: " + node)
		}
	}
}

func (cv *ConfigValidator) validateLoggingConfig(config *LoggingConfig) {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	if !validLevels[config.Level] {
		cv.addError("invalid log level: " + config.Level)
	}

	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}
	if !validFormats[config.Format] {
		cv.addError("invalid log format: " + config.Format)
	}

	validOutputs := map[string]bool{
		"stdout": true,
		"stderr": true,
		"file":   true,
	}
	if !validOutputs[config.Output] {
		cv.addError("invalid log output: " + config.Output)
	}

	if config.Output == "file" && config.FilePath == "" {
		cv.addError("file path is required when logging to file")
	}
}

func (cv *ConfigValidator) addError(message string) {
	cv.errors = append(cv.errors, message)
}

func Validate(config *Config) error {
	validator := NewConfigValidator()
	return validator.Validate(config)
}