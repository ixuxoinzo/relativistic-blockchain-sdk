
package network

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type NetworkMonitor struct {
	topologyManager  *TopologyManager
	latencyMonitor   *LatencyMonitor
	discoveryService *DiscoveryService
	logger           *zap.Logger
	mu               sync.RWMutex
	alerts           map[string]*Alert
	stopChan         chan struct{}
	metrics          *NetworkMetrics
	httpClient       *http.Client
}

type Alert struct {
	ID           string
	Type         types.AlertType
	Severity     types.AlertSeverity
	Message      string
	NodeID       string
	Timestamp    time.Time
	Acknowledged bool
	Data         map[string]interface{}
}

type NetworkMetrics struct {
	TotalNodes       int
	ActiveNodes      int
	NetworkHealth    float64
	AverageLatency   time.Duration
	PeersDiscovered  int
	AlertsActive     int
	LastUpdated      time.Time
}

func NewNetworkMonitor(topology *TopologyManager, latency *LatencyMonitor, discovery *DiscoveryService, logger *zap.Logger) *NetworkMonitor {
	return &NetworkMonitor{
		topologyManager:  topology,
		latencyMonitor:   latency,
		discoveryService: discovery,
		logger:           logger,
		alerts:           make(map[string]*Alert),
		stopChan:         make(chan struct{}),
		metrics:          &NetworkMetrics{},
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (nm *NetworkMonitor) StartMonitoring(ctx context.Context) error {
	nm.logger.Info("Starting network monitor")
	go nm.monitoringLoop(ctx)
	go nm.alertProcessing(ctx)
	go nm.metricsCollection(ctx)
	return nil
}

func (nm *NetworkMonitor) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-nm.stopChan:
			return
		case <-ticker.C:
			nm.checkNodeHealth()
			nm.checkNetworkLatency()
			nm.checkNetworkPartitions()
			nm.checkDiscoveryHealth()
		}
	}
}

func (nm *NetworkMonitor) checkNodeHealth() {
	nodes := nm.topologyManager.GetAllNodes()
	currentTime := time.Now().UTC()
	staleThreshold := currentTime.Add(-5 * time.Minute)

	for _, node := range nodes {
		if node.LastSeen.Before(staleThreshold) && node.IsActive {
			isHealthy := nm.pingNode(node)
			if !isHealthy {
				nm.triggerAlert(Alert{
					ID:       fmt.Sprintf("node_down_%s_%d", node.ID, currentTime.Unix()),
					Type:     types.AlertTypeNodeDown,
					Severity: types.AlertSeverityMedium,
					Message:  fmt.Sprintf("Node %s is not responding", node.ID),
					NodeID:   node.ID,
					Timestamp: currentTime,
					Data: map[string]interface{}{
						"last_seen": node.LastSeen,
						"position":  node.Position,
						"address":   node.Address,
					},
				})
			}
		}
	}
}

func (nm *NetworkMonitor) pingNode(node *types.Node) bool {
	timeout := 5 * time.Second
	conn, err := net.DialTimeout("tcp", node.Address, timeout)
	if err != nil {
		nm.logger.Debug("Node ping failed",
			zap.String("node_id", node.ID),
			zap.String("address", node.Address),
			zap.Error(err),
		)
		return false
	}
	defer conn.Close()
	return true
}

func (nm *NetworkMonitor) checkNetworkLatency() {
	health := nm.latencyMonitor.GetNetworkHealth()
	if health.AverageLatency > time.Second {
		nm.triggerAlert(Alert{
			ID:        fmt.Sprintf("high_latency_%d", time.Now().Unix()),
			Type:      types.AlertTypeHighLatency,
			Severity:  types.AlertSeverityHigh,
			Message:   fmt.Sprintf("High network latency detected: %v", health.AverageLatency),
			Timestamp: time.Now().UTC(),
			Data: map[string]interface{}{
				"average_latency":     health.AverageLatency,
				"healthy_connections": health.HealthyConnections,
				"total_measurements":  health.TotalMeasurements,
			},
		})
	}
}

func (nm *NetworkMonitor) checkNetworkPartitions() {
	nodes := nm.topologyManager.GetActiveNodes()
	if len(nodes) < 2 {
		return
	}

	expectedConnections := len(nodes) * (len(nodes) - 1) / 2
	actualConnections := len(nm.latencyMonitor.GetAllMeasurements())
	connectionRatio := float64(actualConnections) / float64(expectedConnections)

	if connectionRatio < 0.3 {
		nm.triggerAlert(Alert{
			ID:        fmt.Sprintf("network_partition_%d", time.Now().Unix()),
			Type:      types.AlertTypeNetworkPartition,
			Severity:  types.AlertSeverityCritical,
			Message:   fmt.Sprintf("Possible network partition detected. Connection ratio: %.2f", connectionRatio),
			Timestamp: time.Now().UTC(),
			Data: map[string]interface{}{
				"expected_connections": expectedConnections,
				"actual_connections":   actualConnections,
				"connection_ratio":     connectionRatio,
			},
		})
	}
}

func (nm *NetworkMonitor) checkDiscoveryHealth() {
	stats := nm.discoveryService.GetDiscoveryStats()
	if stats.TotalPeers == 0 {
		nm.triggerAlert(Alert{
			ID:        fmt.Sprintf("discovery_issue_%d", time.Now().Unix()),
			Type:      types.AlertTypeDiscoveryIssue,
			Severity:  types.AlertSeverityHigh,
			Message:   "No peers discovered - discovery service may be failing",
			Timestamp: time.Now().UTC(),
		})
	}
}

func (nm *NetworkMonitor) triggerAlert(alert Alert) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	for _, existingAlert := range nm.alerts {
		if existingAlert.Type == alert.Type &&
			existingAlert.NodeID == alert.NodeID &&
			!existingAlert.Acknowledged &&
			time.Since(existingAlert.Timestamp) < 10*time.Minute {
			return
		}
	}
	nm.alerts[alert.ID] = &alert
	nm.logger.Warn("Network alert triggered",
		zap.String("alert_id", alert.ID),
		zap.String("type", string(alert.Type)),
		zap.String("severity", string(alert.Severity)),
		zap.String("message", alert.Message),
		zap.String("node_id", alert.NodeID),
	)
}

func (nm *NetworkMonitor) alertProcessing(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-nm.stopChan:
			return
		case <-ticker.C:
			nm.cleanupOldAlerts()
			nm.escalateAlerts()
		}
	}
}

func (nm *NetworkMonitor) cleanupOldAlerts() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	cutoffTime := time.Now().Add(-24 * time.Hour)
	cleanedCount := 0
	for alertID, alert := range nm.alerts {
		if alert.Timestamp.Before(cutoffTime) {
			delete(nm.alerts, alertID)
			cleanedCount++
		}
	}
	if cleanedCount > 0 {
		nm.logger.Debug("Cleaned up old alerts", zap.Int("count", cleanedCount))
	}
}

func (nm *NetworkMonitor) escalateAlerts() {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	for _, alert := range nm.alerts {
		if !alert.Acknowledged {
			age := time.Since(alert.Timestamp)
			if age > 30*time.Minute && alert.Severity != types.AlertSeverityCritical {
				nm.logger.Warn("Alert escalation needed",
					zap.String("alert_id", alert.ID),
					zap.String("type", string(alert.Type)),
					zap.Duration("age", age),
				)
			}
		}
	}
}

func (nm *NetworkMonitor) metricsCollection(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-nm.stopChan:
			return
		case <-ticker.C:
			nm.collectMetrics()
		}
	}
}

func (nm *NetworkMonitor) collectMetrics() {
	nodes := nm.topologyManager.GetAllNodes()
	health := nm.latencyMonitor.GetNetworkHealth()
	discoveryStats := nm.discoveryService.GetDiscoveryStats()

	nm.mu.Lock()
	nm.metrics.TotalNodes = len(nodes)
	nm.metrics.ActiveNodes = len(nm.topologyManager.GetActiveNodes())
	nm.metrics.NetworkHealth = health.ConnectionHealth
	nm.metrics.AverageLatency = health.AverageLatency
	nm.metrics.PeersDiscovered = discoveryStats.TotalPeers
	nm.metrics.AlertsActive = len(nm.getActiveAlerts())
	nm.metrics.LastUpdated = time.Now().UTC()
	nm.mu.Unlock()
}

func (nm *NetworkMonitor) GetAlerts(includeAcknowledged bool) []*Alert {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	var alerts []*Alert
	for _, alert := range nm.alerts {
		if includeAcknowledged || !alert.Acknowledged {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

func (nm *NetworkMonitor) getActiveAlerts() []*Alert {
	return nm.GetAlerts(false)
}

func (nm *NetworkMonitor) AcknowledgeAlert(alertID string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	alert, exists := nm.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert %s not found", alertID)
	}
	alert.Acknowledged = true
	nm.logger.Info("Alert acknowledged", zap.String("alert_id", alertID))
	return nil
}

func (nm *NetworkMonitor) GetMetrics() *NetworkMetrics {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	return nm.metrics
}

func (nm *NetworkMonitor) GetNetworkStatus() *NetworkStatus {
	metrics := nm.GetMetrics()
	alerts := nm.getActiveAlerts()

	status := &NetworkStatus{
		Metrics:      metrics,
		ActiveAlerts: alerts,
		Timestamp:    time.Now().UTC(),
	}

	if len(alerts) == 0 {
		status.OverallStatus = "healthy"
	} else {
		for _, alert := range alerts {
			if alert.Severity == types.AlertSeverityCritical {
				status.OverallStatus = "critical"
				return status
			}
		}
		status.OverallStatus = "degraded"
	}
	return status
}

type NetworkStatus struct {
	OverallStatus string
	Metrics       *NetworkMetrics
	ActiveAlerts  []*Alert
	Timestamp     time.Time
}

func (nm *NetworkMonitor) Stop() {
	close(nm.stopChan)
	nm.logger.Info("Network monitor stopped")
}
