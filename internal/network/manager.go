package network

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type NetworkManager struct {
	topologyManager  *TopologyManager
	latencyMonitor   *LatencyMonitor
	discoveryService *DiscoveryService
	networkMonitor   *NetworkMonitor
	logger           *zap.Logger
	mu               sync.RWMutex
	stopChan         chan struct{}
}

func NewNetworkManager(redisAddr string, logger *zap.Logger) (*NetworkManager, error) {
	topology, err := NewTopologyManager(redisAddr, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create topology manager: %w", err)
	}

	latencyMonitor := NewLatencyMonitor(topology, logger)
	discoveryService := NewDiscoveryService(topology, logger)
	networkMonitor := NewNetworkMonitor(topology, latencyMonitor, discoveryService, logger)

	return &NetworkManager{
		topologyManager:  topology,
		latencyMonitor:   latencyMonitor,
		discoveryService: discoveryService,
		networkMonitor:   networkMonitor,
		logger:           logger,
		stopChan:         make(chan struct{}),
	}, nil
}

func (nm *NetworkManager) Start(ctx context.Context) error {
	nm.logger.Info("Starting Network Manager")

	if err := nm.discoveryService.StartDiscovery(ctx); err != nil {
		return fmt.Errorf("failed to start discovery service: %w", err)
	}

	go nm.latencyMonitor.StartMonitoring(ctx)
	go nm.networkMonitor.StartMonitoring(ctx)

	nm.logger.Info("Network Manager started successfully")
	return nil
}

func (nm *NetworkManager) Stop() error {
	nm.logger.Info("Stopping Network Manager")

	close(nm.stopChan)
	nm.discoveryService.Stop()
	nm.latencyMonitor.Stop()
	nm.networkMonitor.Stop()
	nm.topologyManager.Close()

	nm.logger.Info("Network Manager stopped successfully")
	return nil
}

func (nm *NetworkManager) AddNode(node *types.Node) error {
	return nm.topologyManager.AddNode(node)
}

func (nm *NetworkManager) GetNode(nodeID string) (*types.Node, error) {
	return nm.topologyManager.GetNode(nodeID)
}

func (nm *NetworkManager) RemoveNode(nodeID string) error {
	return nm.topologyManager.RemoveNode(nodeID)
}

func (nm *NetworkManager) UpdateNodePosition(nodeID string, position types.Position) error {
	return nm.topologyManager.UpdateNodePosition(nodeID, position)
}

func (nm *NetworkManager) GetAllNodes() []*types.Node {
	return nm.topologyManager.GetAllNodes()
}

func (nm *NetworkManager) GetActiveNodes() []*types.Node {
	return nm.topologyManager.GetActiveNodes()
}

func (nm *NetworkManager) GetNodesByRegion(region string) []*types.Node {
	return nm.topologyManager.GetNodesByRegion(region)
}

func (nm *NetworkManager) GetNetworkHealth() *NetworkHealth {
	return nm.latencyMonitor.GetNetworkHealth()
}

func (nm *NetworkManager) GetNetworkStatus() *NetworkStatus {
	return nm.networkMonitor.GetNetworkStatus()
}

func (nm *NetworkManager) GetDiscoveryStats() *DiscoveryStats {
	return nm.discoveryService.GetDiscoveryStats()
}

func (nm *NetworkManager) GetAlerts(includeAcknowledged bool) []*Alert {
	return nm.networkMonitor.GetAlerts(includeAcknowledged)
}

func (nm *NetworkManager) AcknowledgeAlert(alertID string) error {
	return nm.networkMonitor.AcknowledgeAlert(alertID)
}

func (nm *NetworkManager) GetLatencyMeasurement(source, target string) *LatencyMeasurement {
	return nm.latencyMonitor.GetMeasurement(source, target)
}

func (nm *NetworkManager) GetAllLatencyMeasurements() map[string]*LatencyMeasurement {
	return nm.latencyMonitor.GetAllMeasurements()
}

func (nm *NetworkManager) GetPeer(peerID string) *Peer {
	return nm.discoveryService.GetPeer(peerID)
}

func (nm *NetworkManager) GetActivePeers() []*Peer {
	return nm.discoveryService.GetActivePeers()
}

func (nm *NetworkManager) AddPeer(peer *Peer) {
	nm.discoveryService.AddPeer(peer)
}

func (nm *NetworkManager) RemovePeer(peerID string) {
	nm.discoveryService.RemovePeer(peerID)
}

func (nm *NetworkManager) GetNetworkMetrics() *NetworkMetrics {
	return nm.networkMonitor.GetMetrics()
}

func (nm *NetworkManager) ValidateNetworkConnectivity() *ConnectivityReport {
	nodes := nm.GetActiveNodes()
	report := &ConnectivityReport{
		Timestamp:       time.Now().UTC(),
		TotalNodes:      len(nodes),
		ConnectivityMap: make(map[string]map[string]bool),
	}

	for i, source := range nodes {
		report.ConnectivityMap[source.ID] = make(map[string]bool)
		
		for j, target := range nodes {
			if i == j {
				continue
			}

			measurement := nm.GetLatencyMeasurement(source.ID, target.ID)
			isConnected := measurement != nil && measurement.Actual > 0
			
			report.ConnectivityMap[source.ID][target.ID] = isConnected
			
			if isConnected {
				report.ConnectedPairs++
			}
		}
	}

	totalPossiblePairs := len(nodes) * (len(nodes) - 1)
	if totalPossiblePairs > 0 {
		report.ConnectivityPercentage = float64(report.ConnectedPairs) / float64(totalPossiblePairs)
	}

	return report
}

type ConnectivityReport struct {
	Timestamp            time.Time
	TotalNodes           int
	ConnectedPairs       int
	ConnectivityPercentage float64
	ConnectivityMap      map[string]map[string]bool
}

func (nm *NetworkManager) GetTopologyGraph() *TopologyGraph {
	nodes := nm.GetAllNodes()
	graph := &TopologyGraph{
		Nodes: make([]*TopologyNode, len(nodes)),
		Links: make([]*TopologyLink, 0),
		Timestamp: time.Now().UTC(),
	}

	for i, node := range nodes {
		graph.Nodes[i] = &TopologyNode{
			ID:       node.ID,
			Position: node.Position,
			Region:   node.Metadata.Region,
			Status:   "active",
		}
	}

	measurements := nm.GetAllLatencyMeasurements()
	for key, measurement := range measurements {
		graph.Links = append(graph.Links, &TopologyLink{
			Source: measurement.SourceNode,
			Target: measurement.TargetNode,
			Latency: measurement.Actual,
			Jitter:  measurement.Jitter,
		})
	}

	return graph
}

type TopologyGraph struct {
	Nodes     []*TopologyNode
	Links     []*TopologyLink
	Timestamp time.Time
}

type TopologyNode struct {
	ID       string
	Position types.Position
	Region   string
	Status   string
}

type TopologyLink struct {
	Source string
	Target string
	Latency time.Duration
	Jitter  time.Duration
}

func (nm *NetworkManager) HealthCheck() *types.HealthStatus {
	nodes := nm.GetAllNodes()
	status := &types.HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   "1.0.0",
		NodeCount: len(nodes),
		Components: map[string]string{
			"topology_manager":  "healthy",
			"latency_monitor":   "healthy",
			"discovery_service": "healthy",
			"network_monitor":   "healthy",
		},
	}

	if len(nodes) == 0 {
		status.Components["topology_manager"] = "degraded"
	}

	alerts := nm.GetAlerts(false)
	for _, alert := range alerts {
		if alert.Severity == SeverityCritical {
			status.Status = "degraded"
			break
		}
	}

	return status
}