package network

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type EventManager struct {
	logger        *zap.Logger
	mu            sync.RWMutex
	subscribers   map[string]map[EventType][]EventCallback
	eventHistory  []*EventRecord
	maxHistory    int
	webhookConfig *WebhookConfig
}

type EventType string

const (
	EventNodeRegistered   EventType = "node_registered"
	EventNodeDeregistered EventType = "node_deregistered"
	EventNodeUpdated      EventType = "node_updated"
	EventPeerConnected    EventType = "peer_connected"
	EventPeerDisconnected EventType = "peer_disconnected"
	EventNetworkAlert     EventType = "network_alert"
	EventLatencySpike     EventType = "latency_spike"
	EventPartitionDetected EventType = "partition_detected"
)

type EventCallback func(*Event)

type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Severity  EventSeverity          `json:"severity"`
}

type EventSeverity string

const (
	SeverityInfo     EventSeverity = "info"
	SeverityWarning  EventSeverity = "warning"
	SeverityError    EventSeverity = "error"
	SeverityCritical EventSeverity = "critical"
)

type EventRecord struct {
	Event     *Event     `json:"event"`
	Processed bool       `json:"processed"`
	Error     string     `json:"error,omitempty"`
}

type WebhookConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Timeout time.Duration     `json:"timeout"`
	Enabled bool              `json:"enabled"`
}

func NewEventManager(logger *zap.Logger) *EventManager {
	return &EventManager{
		logger:      logger,
		subscribers: make(map[string]map[EventType][]EventCallback),
		eventHistory: make([]*EventRecord, 0),
		maxHistory:  10000,
		webhookConfig: &WebhookConfig{
			Enabled: false,
			Timeout: 10 * time.Second,
		},
	}
}

func (em *EventManager) Subscribe(component string, eventType EventType, callback EventCallback) string {
	em.mu.Lock()
	defer em.mu.Unlock()

	if _, exists := em.subscribers[component]; !exists {
		em.subscribers[component] = make(map[EventType][]EventCallback)
	}

	if _, exists := em.subscribers[component][eventType]; !exists {
		em.subscribers[component][eventType] = make([]EventCallback, 0)
	}

	em.subscribers[component][eventType] = append(em.subscribers[component][eventType], callback)

	subscriptionID := fmt.Sprintf("%s-%s-%d", component, eventType, time.Now().UnixNano())

	em.logger.Debug("Event subscription created",
		zap.String("component", component),
		zap.String("event_type", string(eventType)),
		zap.String("subscription_id", subscriptionID),
	)

	return subscriptionID
}

func (em *EventManager) Unsubscribe(component string, eventType EventType, callback EventCallback) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if componentSubs, exists := em.subscribers[component]; exists {
		if callbacks, exists := componentSubs[eventType]; exists {
			for i, cb := range callbacks {
				if &cb == &callback {
					em.subscribers[component][eventType] = append(callbacks[:i], callbacks[i+1:]...)
					em.logger.Debug("Event subscription removed",
						zap.String("component", component),
						zap.String("event_type", string(eventType)),
					)
					break
				}
			}
		}
	}
}

func (em *EventManager) EmitEvent(eventType EventType, source string, data map[string]interface{}, severity EventSeverity) {
	event := &Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Source:    source,
		Data:      data,
		Severity:  severity,
	}

	em.recordEvent(event)

	em.notifySubscribers(event)

	if em.webhookConfig.Enabled {
		go em.sendWebhook(event)
	}

	em.logger.Info("Event emitted",
		zap.String("event_id", event.ID),
		zap.String("event_type", string(eventType)),
		zap.String("source", source),
		zap.String("severity", string(severity)),
	)
}

func (em *EventManager) recordEvent(event *Event) {
	em.mu.Lock()
	defer em.mu.Unlock()

	record := &EventRecord{
		Event:     event,
		Processed: false,
	}

	em.eventHistory = append(em.eventHistory, record)

	if len(em.eventHistory) > em.maxHistory {
		em.eventHistory = em.eventHistory[1:]
	}
}

func (em *EventManager) notifySubscribers(event *Event) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	for component, eventSubs := range em.subscribers {
		if callbacks, exists := eventSubs[event.Type]; exists {
			for _, callback := range callbacks {
				go func(cb EventCallback) {
					defer func() {
						if r := recover(); r != nil {
							em.logger.Error("Event callback panicked",
								zap.String("component", component),
								zap.String("event_type", string(event.Type)),
								zap.Any("panic", r),
							)
						}
					}()
					cb(event)
				}(callback)
			}
		}
	}
}

func (em *EventManager) sendWebhook(event *Event) {
	if !em.webhookConfig.Enabled || em.webhookConfig.URL == "" {
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		em.logger.Error("Failed to marshal event for webhook",
			zap.String("event_id", event.ID),
			zap.Error(err),
		)
		return
	}

	client := &http.Client{
		Timeout: em.webhookConfig.Timeout,
	}

	req, err := http.NewRequest("POST", em.webhookConfig.URL, bytes.NewBuffer(payload))
	if err != nil {
		em.logger.Error("Failed to create webhook request",
			zap.String("event_id", event.ID),
			zap.Error(err),
		)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range em.webhookConfig.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		em.logger.Error("Failed to send webhook",
			zap.String("event_id", event.ID),
			zap.Error(err),
		)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		em.logger.Warn("Webhook returned error status",
			zap.String("event_id", event.ID),
			zap.Int("status_code", resp.StatusCode),
		)
	} else {
		em.logger.Debug("Webhook sent successfully",
			zap.String("event_id", event.ID),
			zap.Int("status_code", resp.StatusCode),
		)
	}
}

func (em *EventManager) GetEventHistory(eventType EventType, limit int) []*Event {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var events []*Event
	count := 0

	for i := len(em.eventHistory) - 1; i >= 0 && count < limit; i-- {
		record := em.eventHistory[i]
		if eventType == "" || record.Event.Type == eventType {
			events = append([]*Event{record.Event}, events...)
			count++
		}
	}

	return events
}

func (em *EventManager) GetEventStatistics(since time.Time) *EventStatistics {
	em.mu.RLock()
	defer em.mu.RUnlock()

	stats := &EventStatistics{
		StartTime: since,
		EndTime:   time.Now().UTC(),
		Counts:    make(map[EventType]int),
		Severities: make(map[EventSeverity]int),
	}

	for _, record := range em.eventHistory {
		if record.Event.Timestamp.After(since) {
			stats.TotalEvents++
			stats.Counts[record.Event.Type]++
			stats.Severities[record.Event.Severity]++

			if !record.Processed {
				stats.UnprocessedEvents++
			}
		}
	}

	return stats
}

type EventStatistics struct {
	TotalEvents      int                    `json:"total_events"`
	UnprocessedEvents int                   `json:"unprocessed_events"`
	Counts          map[EventType]int      `json:"counts_by_type"`
	Severities      map[EventSeverity]int  `json:"counts_by_severity"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
}

func (em *EventManager) ConfigureWebhook(config *WebhookConfig) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.webhookConfig = config

	em.logger.Info("Webhook configuration updated",
		zap.Bool("enabled", config.Enabled),
		zap.String("url", config.URL),
		zap.Duration("timeout", config.Timeout),
	)
}

func (em *EventManager) CreateNodeEventHandlers(topology *TopologyManager) {
	topologyEventCh := topology.GetEventChannel()

	go func() {
		for event := range topologyEventCh {
			var eventType EventType
			var severity EventSeverity

			switch event.Type {
			case NodeAdded:
				eventType = EventNodeRegistered
				severity = SeverityInfo
			case NodeRemoved:
				eventType = EventNodeDeregistered
				severity = SeverityWarning
			case NodeUpdated:
				eventType = EventNodeUpdated
				severity = SeverityInfo
			}

			data := map[string]interface{}{
				"node_id": event.Node.ID,
				"position": event.Node.Position,
				"address":  event.Node.Address,
				"region":   event.Node.Metadata.Region,
			}

			em.EmitEvent(eventType, "topology_manager", data, severity)
		}
	}()
}

func (em *EventManager) CreateNetworkEventHandlers(monitor *NetworkMonitor) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				alerts := monitor.GetAlerts(false)
				for _, alert := range alerts {
					data := map[string]interface{}{
						"alert_id":   alert.ID,
						"message":    alert.Message,
						"node_id":    alert.NodeID,
						"severity":   string(alert.Severity),
						"alert_data": alert.Data,
					}

					em.EmitEvent(EventNetworkAlert, "network_monitor", data, em.mapAlertSeverity(alert.Severity))
				}
			}
		}
	}()
}

func (em *EventManager) mapAlertSeverity(alertSeverity AlertSeverity) EventSeverity {
	switch alertSeverity {
	case SeverityLow:
		return SeverityInfo
	case SeverityMedium:
		return SeverityWarning
	case SeverityHigh:
		return SeverityError
	case SeverityCritical:
		return SeverityCritical
	default:
		return SeverityInfo
	}
}

func (em *EventManager) GetSubscriberCount() map[string]int {
	em.mu.RLock()
	defer em.mu.RUnlock()

	counts := make(map[string]int)

	for component, eventSubs := range em.subscribers {
		total := 0
		for _, callbacks := range eventSubs {
			total += len(callbacks)
		}
		counts[component] = total
	}

	return counts
}

func (em *EventManager) ClearHistory() {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.eventHistory = make([]*EventRecord, 0)
	em.logger.Info("Event history cleared")
}

func (em *EventManager) Stop() {
	em.logger.Info("Event Manager stopped")
}