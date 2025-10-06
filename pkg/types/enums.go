package types

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
)

type AlertType string

const (
	AlertTypeNodeDown       AlertType = "node_down"
	AlertTypeHighLatency    AlertType = "high_latency"
	AlertTypeNetworkIssue   AlertType = "network_issue"
	AlertTypeSecurity       AlertType = "security"
	AlertTypePerformance    AlertType = "performance"
	AlertTypeCapacity       AlertType = "capacity"
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
	NetworkProtocolTCP  NetworkProtocol = "tcp"
	NetworkProtocolUDP  NetworkProtocol = "udp"
	NetworkProtocolTLS  NetworkProtocol = "tls"
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