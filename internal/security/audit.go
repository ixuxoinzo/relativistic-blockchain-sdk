package security

import (
	"encoding/json"
	"sync"
	"time"

	"go.uber.org/zap"
)

type AuditLogger struct {
	logger *zap.Logger
	mu     sync.RWMutex
	events []*AuditEvent
}

type AuditEvent struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	UserID    string                 `json:"user_id"`
	Resource  string                 `json:"resource"`
	Timestamp time.Time              `json:"timestamp"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	Details   map[string]interface{} `json:"details"`
	Status    string                 `json:"status"`
	Error     string                 `json:"error,omitempty"`
}

func NewAuditLogger(logger *zap.Logger) *AuditLogger {
	return &AuditLogger{
		logger: logger,
		events: make([]*AuditEvent, 0),
	}
}

func (al *AuditLogger) LogEvent(event *AuditEvent) {
	al.mu.Lock()
	defer al.mu.Unlock()

	al.events = append(al.events, event)

	fields := []zap.Field{
		zap.String("action", event.Action),
		zap.String("user_id", event.UserID),
		zap.String("resource", event.Resource),
		zap.String("status", event.Status),
		zap.String("ip_address", event.IPAddress),
	}

	if event.Error != "" {
		fields = append(fields, zap.String("error", event.Error))
	}

	if len(event.Details) > 0 {
		detailsJSON, _ := json.Marshal(event.Details)
		fields = append(fields, zap.String("details", string(detailsJSON)))
	}

	al.logger.Info("Audit event", fields...)
}

func (al *AuditLogger) GetEvents(filter *AuditFilter) []*AuditEvent {
	al.mu.RLock()
	defer al.mu.RUnlock()

	var filteredEvents []*AuditEvent

	for _, event := range al.events {
		if filter.Matches(event) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents
}

func (al *AuditLogger) CleanupOldEvents(maxAge time.Duration) {
	al.mu.Lock()
	defer al.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	var validEvents []*AuditEvent

	for _, event := range al.events {
		if event.Timestamp.After(cutoff) {
			validEvents = append(validEvents, event)
		}
	}

	al.events = validEvents
}

type AuditFilter struct {
	Action   string
	UserID   string
	Resource string
	Status   string
	From     time.Time
	To       time.Time
}

func (af *AuditFilter) Matches(event *AuditEvent) bool {
	if af.Action != "" && event.Action != af.Action {
		return false
	}

	if af.UserID != "" && event.UserID != af.UserID {
		return false
	}

	if af.Resource != "" && event.Resource != af.Resource {
		return false
	}

	if af.Status != "" && event.Status != af.Status {
		return false
	}

	if !af.From.IsZero() && event.Timestamp.Before(af.From) {
		return false
	}

	if !af.To.IsZero() && event.Timestamp.After(af.To) {
		return false
	}

	return true
}
