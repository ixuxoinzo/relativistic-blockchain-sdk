package network

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type HealthMonitor struct {
	topologyManager  *TopologyManager
	latencyMonitor   *LatencyMonitor
	discoveryService *DiscoveryService
	peeringManager   *PeeringManager
	logger           *zap.Logger
	mu               sync.RWMutex
	healthStatus     *HealthStatus
	checkInterval    time.Duration
	stopChan         chan struct{}
}

type HealthStatus struct {
	OverallStatus string            `json:"overall_status"`
	Timestamp     time.Time         `json:"timestamp"`
	Components    map[string]string `json:"components"`
	Metrics       *HealthMetrics    `json:"metrics"`
	ActiveAlerts  int               `json:"active_alerts"`
	LastChecked   time.Time         `json:"last_checked"`
}

type HealthMetrics struct {
	TotalNodes        int     `json:"total_nodes"`
	ActiveNodes       int     `json:"active_nodes"`
	NetworkLatency    float64 `json:"network_latency_ms"`
	PacketLoss        float64 `json:"packet_loss"`
	ConnectionHealth  float64 `json:"connection_health"`
	PeersDiscovered   int     `json:"peers_discovered"`
	ActiveConnections int     `json:"active_connections"`
}

func NewHealthMonitor(topology *TopologyManager, latency *LatencyMonitor, discovery *DiscoveryService, peering *PeeringManager, logger *zap.Logger) *HealthMonitor {
	return &HealthMonitor{
		topologyManager:  topology,
		latencyMonitor:   latency,
		discoveryService: discovery,
		peeringManager:   peering,
		logger:           logger,
		healthStatus:     &HealthStatus{},
		checkInterval:    30 * time.Second,
		stopChan:         make(chan struct{}),
	}
}

func (hm *HealthMonitor) Start() {
	hm.logger.Info("Starting Health Monitor")

	go hm.healthCheckLoop()

	hm.logger.Info("Health Monitor started")
}

func (hm *HealthMonitor) healthCheckLoop() {
	ticker := time.NewTicker(hm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hm.stopChan:
			return
		case <-ticker.C:
			hm.performHealthCheck()
		}
	}
}

func (hm *HealthMonitor) performHealthCheck() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	status := &HealthStatus{
		Timestamp:  time.Now().UTC(),
		Components: make(map[string]string),
		Metrics:    &HealthMetrics{},
	}

	hm.checkTopologyHealth(status)
	hm.checkLatencyHealth(status)
	hm.checkDiscoveryHealth(status)
	hm.checkPeeringHealth(status)
	hm.checkNetworkConnectivity(status)

	hm.determineOverallStatus(status)

	hm.healthStatus = status

	hm.logger.Debug("Health check completed",
		zap.String("overall_status", status.OverallStatus),
		zap.Int("active_nodes", status.Metrics.ActiveNodes),
		zap.Float64("network_latency", status.Metrics.NetworkLatency),
	)
}

func (hm *HealthMonitor) checkTopologyHealth(status *HealthStatus) {
	nodes := hm.topologyManager.GetAllNodes()
	activeNodes := hm.topologyManager.GetActiveNodes()

	status.Metrics.TotalNodes = len(nodes)
	status.Metrics.ActiveNodes = len(activeNodes)

	if len(nodes) == 0 {
		status.Components["topology"] = "critical"
		return
	}

	activeRatio := float64(len(activeNodes)) / float64(len(nodes))
	if activeRatio > 0.8 {
		status.Components["topology"] = "healthy"
	} else if activeRatio > 0.5 {
		status.Components["topology"] = "degraded"
	} else {
		status.Components["topology"] = "critical"
	}
}

func (hm *HealthMonitor) checkLatencyHealth(status *HealthStatus) {
	networkHealth := hm.latencyMonitor.GetNetworkHealth()

	status.Metrics.NetworkLatency = float64(networkHealth.AverageLatency.Milliseconds())
	status.Metrics.PacketLoss = networkHealth.AveragePacketLoss
	status.Metrics.ConnectionHealth = networkHealth.ConnectionHealth

	if networkHealth.Status == "healthy" {
		status.Components["latency"] = "healthy"
	} else if networkHealth.Status == "degraded" {
		status.Components["latency"] = "degraded"
	} else {
		status.Components["latency"] = "critical"
	}
}

func (hm *HealthMonitor) checkDiscoveryHealth(status *HealthStatus) {
	stats := hm.discoveryService.GetDiscoveryStats()

	status.Metrics.PeersDiscovered = stats.TotalPeers

	connectedPeers := 0
	for _, count := range stats.StatusBreakdown {
		if count > 0 {
			connectedPeers += count
		}
	}

	if stats.TotalPeers == 0 {
		status.Components["discovery"] = "critical"
	} else if connectedPeers >= 3 {
		status.Components["discovery"] = "healthy"
	} else if connectedPeers >= 1 {
		status.Components["discovery"] = "degraded"
	} else {
		status.Components["discovery"] = "critical"
	}
}

func (hm *HealthMonitor) checkPeeringHealth(status *HealthStatus) {
	activeConns := hm.peeringManager.GetActiveConnections()
	status.Metrics.ActiveConnections = len(activeConns)

	if len(activeConns) >= 5 {
		status.Components["peering"] = "healthy"
	} else if len(activeConns) >= 2 {
		status.Components["peering"] = "degraded"
	} else {
		status.Components["peering"] = "critical"
	}
}

func (hm *HealthMonitor) checkNetworkConnectivity(status *HealthStatus) {
	connectivityReport := hm.validateConnectivity()

	if connectivityReport.ConnectivityPercentage > 0.8 {
		status.Components["connectivity"] = "healthy"
	} else if connectivityReport.ConnectivityPercentage > 0.5 {
		status.Components["connectivity"] = "degraded"
	} else {
		status.Components["connectivity"] = "critical"
	}
}

func (hm *HealthMonitor) validateConnectivity() *ConnectivityReport {
	nodes := hm.topologyManager.GetActiveNodes()
	if len(nodes) < 2 {
		return &ConnectivityReport{
			TotalNodes:             len(nodes),
			ConnectedPairs:         0,
			ConnectivityPercentage: 0.0,
		}
	}

	connectedPairs := 0
	measurements := hm.latencyMonitor.GetAllMeasurements()

	for _, measurement := range measurements {
		if measurement.Actual > 0 && measurement.PacketLoss < 0.1 {
			connectedPairs++
		}
	}

	totalPossiblePairs := len(nodes) * (len(nodes) - 1) / 2
	connectivityPercentage := 0.0
	if totalPossiblePairs > 0 {
		connectivityPercentage = float64(connectedPairs) / float64(totalPossiblePairs)
	}

	return &ConnectivityReport{
		TotalNodes:             len(nodes),
		ConnectedPairs:         connectedPairs,
		ConnectivityPercentage: connectivityPercentage,
	}
}

func (hm *HealthMonitor) determineOverallStatus(status *HealthStatus) {
	criticalCount := 0
	degradedCount := 0

	for _, componentStatus := range status.Components {
		switch componentStatus {
		case "critical":
			criticalCount++
		case "degraded":
			degradedCount++
		}
	}

	if criticalCount > 0 {
		status.OverallStatus = "critical"
	} else if degradedCount > 0 {
		status.OverallStatus = "degraded"
	} else {
		status.OverallStatus = "healthy"
	}
}

func (hm *HealthMonitor) GetHealthStatus() *HealthStatus {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	return hm.healthStatus
}

func (hm *HealthMonitor) GetTypesHealthStatus() *types.HealthStatus {
	healthStatus := hm.GetHealthStatus()

	return &types.HealthStatus{
		Status:    healthStatus.OverallStatus,
		Timestamp: healthStatus.Timestamp,
		Version:   "1.0.0",
		NodeCount: healthStatus.Metrics.TotalNodes,
		Components: map[string]string{
			"topology":  healthStatus.Components["topology"],
			"latency":   healthStatus.Components["latency"],
			"discovery": healthStatus.Components["discovery"],
			"peering":   healthStatus.Components["peering"],
		},
	}
}

func (hm *HealthMonitor) DeepHealthCheck() *DetailedHealthReport {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	report := &DetailedHealthReport{
		Timestamp:       time.Now().UTC(),
		ComponentChecks: make([]*ComponentHealth, 0),
		Recommendations: make([]string, 0),
	}

	report.ComponentChecks = append(report.ComponentChecks, hm.checkTopologyDetailed())
	report.ComponentChecks = append(report.ComponentChecks, hm.checkLatencyDetailed())
	report.ComponentChecks = append(report.ComponentChecks, hm.checkDiscoveryDetailed())
	report.ComponentChecks = append(report.ComponentChecks, hm.checkPeeringDetailed())

	hm.generateRecommendations(report)

	hm.determineOverallHealth(report)

	return report
}

func (hm *HealthMonitor) checkTopologyDetailed() *ComponentHealth {
	nodes := hm.topologyManager.GetAllNodes()
	activeNodes := hm.topologyManager.GetActiveNodes()

	health := &ComponentHealth{
		Name:    "topology",
		Status:  "healthy",
		Metrics: make(map[string]interface{}),
	}

	health.Metrics["total_nodes"] = len(nodes)
	health.Metrics["active_nodes"] = len(activeNodes)
	health.Metrics["active_ratio"] = float64(len(activeNodes)) / float64(len(nodes))

	if len(nodes) == 0 {
		health.Status = "critical"
		health.Message = "No nodes registered in topology"
	} else if len(activeNodes) < len(nodes)/2 {
		health.Status = "degraded"
		health.Message = "Less than 50% of nodes are active"
	}

	return health
}

func (hm *HealthMonitor) checkLatencyDetailed() *ComponentHealth {
	networkHealth := hm.latencyMonitor.GetNetworkHealth()

	health := &ComponentHealth{
		Name:    "latency",
		Status:  "healthy",
		Metrics: make(map[string]interface{}),
	}

	health.Metrics["average_latency_ms"] = networkHealth.AverageLatency.Milliseconds()
	health.Metrics["average_packet_loss"] = networkHealth.AveragePacketLoss
	health.Metrics["connection_health"] = networkHealth.ConnectionHealth
	health.Metrics["healthy_connections"] = networkHealth.HealthyConnections

	if networkHealth.AverageLatency > time.Second {
		health.Status = "critical"
		health.Message = "Network latency exceeds 1 second"
	} else if networkHealth.AverageLatency > 500*time.Millisecond {
		health.Status = "degraded"
		health.Message = "Network latency is high"
	}

	if networkHealth.AveragePacketLoss > 0.1 {
		health.Status = "critical"
		health.Message = "Packet loss exceeds 10%"
	}

	return health
}

func (hm *HealthMonitor) checkDiscoveryDetailed() *ComponentHealth {
	stats := hm.discoveryService.GetDiscoveryStats()

	health := &ComponentHealth{
		Name:    "discovery",
		Status:  "healthy",
		Metrics: make(map[string]interface{}),
	}

	health.Metrics["total_peers"] = stats.TotalPeers
	health.Metrics["status_breakdown"] = stats.StatusBreakdown
	health.Metrics["region_breakdown"] = stats.RegionBreakdown

	if stats.TotalPeers == 0 {
		health.Status = "critical"
		health.Message = "No peers discovered"
	} else if stats.StatusBreakdown[PeerConnected] < 3 {
		health.Status = "degraded"
		health.Message = "Insufficient connected peers"
	}

	return health
}

func (hm *HealthMonitor) checkPeeringDetailed() *ComponentHealth {
	activeConns := hm.peeringManager.GetActiveConnections()
	stats := hm.peeringManager.GetPeeringStats()

	health := &ComponentHealth{
		Name:    "peering",
		Status:  "healthy",
		Metrics: make(map[string]interface{}),
	}

	health.Metrics["active_connections"] = len(activeConns)
	health.Metrics["status_breakdown"] = stats.StatusBreakdown
	health.Metrics["total_messages_sent"] = stats.TotalMessagesSent
	health.Metrics["total_messages_received"] = stats.TotalMessagesReceived

	if len(activeConns) == 0 {
		health.Status = "critical"
		health.Message = "No active peer connections"
	} else if len(activeConns) < 3 {
		health.Status = "degraded"
		health.Message = "Insufficient peer connections"
	}

	return health
}

func (hm *HealthMonitor) generateRecommendations(report *DetailedHealthReport) {
	for _, component := range report.ComponentChecks {
		switch component.Status {
		case "critical":
			switch component.Name {
			case "topology":
				report.Recommendations = append(report.Recommendations,
					"Add more nodes to the network topology",
					"Check node registration process")
			case "latency":
				report.Recommendations = append(report.Recommendations,
					"Investigate network infrastructure",
					"Check firewall and routing configurations")
			case "discovery":
				report.Recommendations = append(report.Recommendations,
					"Verify DNS seed configuration",
					"Check bootstrap node connectivity")
			case "peering":
				report.Recommendations = append(report.Recommendations,
					"Review peer connection settings",
					"Check network port accessibility")
			}
		case "degraded":
			switch component.Name {
			case "topology":
				report.Recommendations = append(report.Recommendations,
					"Monitor node health and restart inactive nodes")
			case "latency":
				report.Recommendations = append(report.Recommendations,
					"Optimize network routes",
					"Consider geographic distribution of nodes")
			}
		}
	}
}

func (hm *HealthMonitor) determineOverallHealth(report *DetailedHealthReport) {
	criticalCount := 0
	degradedCount := 0

	for _, component := range report.ComponentChecks {
		switch component.Status {
		case "critical":
			criticalCount++
		case "degraded":
			degradedCount++
		}
	}

	if criticalCount > 0 {
		report.OverallStatus = "critical"
	} else if degradedCount > 0 {
		report.OverallStatus = "degraded"
	} else {
		report.OverallStatus = "healthy"
	}
}

func (hm *HealthMonitor) HTTPHealthHandler(w http.ResponseWriter, r *http.Request) {
	healthStatus := hm.GetHealthStatus()

	w.Header().Set("Content-Type", "application/json")

	if healthStatus.OverallStatus == "critical" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else if healthStatus.OverallStatus == "degraded" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(healthStatus)
}

func (hm *HealthMonitor) Stop() {
	close(hm.stopChan)
	hm.logger.Info("Health Monitor stopped")
}

type DetailedHealthReport struct {
	Timestamp       time.Time          `json:"timestamp"`
	OverallStatus   string             `json:"overall_status"`
	ComponentChecks []*ComponentHealth `json:"component_checks"`
	Recommendations []string           `json:"recommendations"`
}

type ComponentHealth struct {
	Name    string                 `json:"name"`
	Status  string                 `json:"status"`
	Message string                 `json:"message,omitempty"`
	Metrics map[string]interface{} `json:"metrics"`
}
