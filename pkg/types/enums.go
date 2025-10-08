package types

import (
	"sync"
	"time"
)

type NodeStatus string

const (
	NodeStatusActive    NodeStatus = "active"
	NodeStatusInactive  NodeStatus = "inactive"
	NodeStatusPending   NodeStatus = "pending"
	NodeStatusSuspended NodeStatus = "suspended"
)

type ValidationStatus string

const (
	ValidationStatusValid   ValidationStatus = "valid"
	ValidationStatusInvalid ValidationStatus = "invalid"
	ValidationStatusPending ValidationStatus = "pending"
	ValidationStatusExpired ValidationStatus = "expired"
)

type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityCritical AlertSeverity = "critical"
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
)

type AlertType string

const (
	AlertTypeNodeDown         AlertType = "node_down"
	AlertTypeHighLatency      AlertType = "high_latency"
	AlertTypeNetworkIssue     AlertType = "network_issue"
	AlertTypeSecurity         AlertType = "security"
	AlertTypePerformance      AlertType = "performance"
	AlertTypeCapacity         AlertType = "capacity"
	AlertTypeDiscoveryIssue   AlertType = "discovery_issue"
	AlertTypeNetworkPartition AlertType = "network_partition"
)

type EventType string

const (
	EventTypeNodeRegistered   EventType = "node_registered"
	EventTypeNodeUpdated      EventType = "node_updated"
	EventTypeNodeRemoved      EventType = "node_removed"
	EventTypeAlertTriggered   EventType = "alert_triggered"
	EventTypeAlertResolved    EventType = "alert_resolved"
	EventTypeConsensusChange  EventType = "consensus_change"
	EventTypeNetworkPartition EventType = "network_partition"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
)

type CacheType string

const (
	CacheTypeMemory CacheType = "memory"
	CacheTypeRedis  CacheType = "redis"
)

type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypeSQLite     DatabaseType = "sqlite"
)

type AuthMethod string

const (
	AuthMethodJWT    AuthMethod = "jwt"
	AuthMethodAPIKey AuthMethod = "api_key"
	AuthMethodOAuth  AuthMethod = "oauth"
)

type NetworkProtocol string

const (
	NetworkProtocolTCP       NetworkProtocol = "tcp"
	NetworkProtocolUDP       NetworkProtocol = "udp"
	NetworkProtocolTLS       NetworkProtocol = "tls"
	NetworkProtocolWebSocket NetworkProtocol = "websocket"
)

type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

type SyncStatus string

const (
	SyncStatusPending   SyncStatus = "pending"
	SyncStatusSyncing   SyncStatus = "syncing"
	SyncStatusSynced    SyncStatus = "synced"
	SyncStatusFailed    SyncStatus = "failed"
	SyncStatusOutOfSync SyncStatus = "out_of_sync"
)

type BroadcastType string

const (
	BroadcastTypeAll      BroadcastType = "all"
	BroadcastTypeRegion   BroadcastType = "region"
	BroadcastTypeSpecific BroadcastType = "specific"
)

type EngineMetrics struct {
    StartTime time.Time     `json:"start_time"`  
    NodeCount int           `json:"node_count"`   
    Status    string        `json:"status"`      
    CalculationsTotal int64 `json:"calculations_total"`
    ValidationsTotal  int64 `json:"validations_total"`
    CacheHits         int64 `json:"cache_hits"`
    CacheMisses       int64 `json:"cache_misses"`
    ErrorsTotal       int64 `json:"errors_total"`
    CPUUsage          float64       `json:"cpu_usage"`
    MemoryUsage       float64       `json:"memory_usage"`
    NetworkIO         float64       `json:"network_io"`
    BlockRate         float64       `json:"block_rate"`
    Latency           float64       `json:"latency"`
    Throughput        float64       `json:"throughput"`
    Uptime            time.Duration `json:"uptime"`
    ActiveConnections int           `json:"active_connections"`
    PeersCount        int           `json:"peers_count"`
    QueueSize         int           `json:"queue_size"`
    Mu                sync.RWMutex  `json:"-"`
}
