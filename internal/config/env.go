package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type EnvLoader struct {
	prefix string
}

func NewEnvLoader(prefix string) *EnvLoader {
	return &EnvLoader{prefix: prefix}
}

func (el *EnvLoader) Load(config *Config) error {
	el.loadServerConfig(config)
	el.loadDatabaseConfig(config)
	el.loadRedisConfig(config)
	el.loadSecurityConfig(config)
	el.loadMetricsConfig(config)
	el.loadNetworkConfig(config)
	el.loadLoggingConfig(config)

	return nil
}

func (el *EnvLoader) loadServerConfig(config *Config) {
	if addr := el.getEnv("SERVER_ADDRESS"); addr != "" {
		config.Server.Address = addr
	}

	if env := el.getEnv("ENVIRONMENT"); env != "" {
		config.Server.Environment = env
	}

	if timeout := el.getEnv("SHUTDOWN_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil {
			config.Server.ShutdownTimeout = duration
		}
	}

	if timeout := el.getEnv("READ_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil {
			config.Server.ReadTimeout = duration
		}
	}

	if timeout := el.getEnv("WRITE_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil {
			config.Server.WriteTimeout = duration
		}
	}
}

func (el *EnvLoader) loadDatabaseConfig(config *Config) {
	if host := el.getEnv("DB_HOST"); host != "" {
		config.Database.Host = host
	}

	if port := el.getEnv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Database.Port = p
		}
	}

	if user := el.getEnv("DB_USERNAME"); user != "" {
		config.Database.Username = user
	}

	if pass := el.getEnv("DB_PASSWORD"); pass != "" {
		config.Database.Password = pass
	}

	if name := el.getEnv("DB_NAME"); name != "" {
		config.Database.Name = name
	}

	if ssl := el.getEnv("DB_SSL_MODE"); ssl != "" {
		config.Database.SSLMode = ssl
	}
}

func (el *EnvLoader) loadRedisConfig(config *Config) {
	if addr := el.getEnv("REDIS_ADDRESS"); addr != "" {
		config.Redis.Address = addr
	}

	if pass := el.getEnv("REDIS_PASSWORD"); pass != "" {
		config.Redis.Password = pass
	}

	if db := el.getEnv("REDIS_DB"); db != "" {
		if d, err := strconv.Atoi(db); err == nil {
			config.Redis.DB = d
		}
	}

	if size := el.getEnv("REDIS_POOL_SIZE"); size != "" {
		if s, err := strconv.Atoi(size); err == nil {
			config.Redis.PoolSize = s
		}
	}
}

func (el *EnvLoader) loadSecurityConfig(config *Config) {
	if secret := el.getEnv("JWT_SECRET"); secret != "" {
		config.Security.JWTSecret = secret
	}

	if expiry := el.getEnv("TOKEN_EXPIRY"); expiry != "" {
		if duration, err := time.ParseDuration(expiry); err == nil {
			config.Security.TokenExpiry = duration
		}
	}

	if limit := el.getEnv("RATE_LIMIT"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			config.Security.RateLimit = l
		}
	}

	if window := el.getEnv("RATE_LIMIT_WINDOW"); window != "" {
		if duration, err := time.ParseDuration(window); err == nil {
			config.Security.RateLimitWindow = duration
		}
	}

	if origins := el.getEnv("CORS_ORIGINS"); origins != "" {
		config.Security.CORSOrigins = strings.Split(origins, ",")
	}
}

func (el *EnvLoader) loadMetricsConfig(config *Config) {
	if enabled := el.getEnv("METRICS_ENABLED"); enabled != "" {
		config.Metrics.Enabled = strings.ToLower(enabled) == "true"
	}

	if port := el.getEnv("METRICS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Metrics.Port = p
		}
	}

	if path := el.getEnv("METRICS_PATH"); path != "" {
		config.Metrics.Path = path
	}

	if gateway := el.getEnv("METRICS_PUSH_GATEWAY"); gateway != "" {
		config.Metrics.PushGateway = gateway
	}

	if interval := el.getEnv("METRICS_PUSH_INTERVAL"); interval != "" {
		if duration, err := time.ParseDuration(interval); err == nil {
			config.Metrics.PushInterval = duration
		}
	}
}

func (el *EnvLoader) loadNetworkConfig(config *Config) {
	if nodes := el.getEnv("BOOTSTRAP_NODES"); nodes != "" {
		config.Network.BootstrapNodes = strings.Split(nodes, ",")
	}

	if discovery := el.getEnv("PEER_DISCOVERY"); discovery != "" {
		config.Network.PeerDiscovery = strings.ToLower(discovery) == "true"
	}

	if maxPeers := el.getEnv("MAX_PEERS"); maxPeers != "" {
		if mp, err := strconv.Atoi(maxPeers); err == nil {
			config.Network.MaxPeers = mp
		}
	}

	if addr := el.getEnv("LISTEN_ADDRESS"); addr != "" {
		config.Network.ListenAddress = addr
	}

	if ip := el.getEnv("EXTERNAL_IP"); ip != "" {
		config.Network.ExternalIP = ip
	}
}

func (el *EnvLoader) loadLoggingConfig(config *Config) {
	if level := el.getEnv("LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}

	if format := el.getEnv("LOG_FORMAT"); format != "" {
		config.Logging.Format = format
	}

	if output := el.getEnv("LOG_OUTPUT"); output != "" {
		config.Logging.Output = output
	}

	if path := el.getEnv("LOG_FILE_PATH"); path != "" {
		config.Logging.FilePath = path
	}
}

func (el *EnvLoader) getEnv(key string) string {
	fullKey := el.prefix + "_" + key
	return os.Getenv(fullKey)
}

func LoadFromEnv(config *Config) error {
	loader := NewEnvLoader("RELATIVISTIC")
	return loader.Load(config)
}
