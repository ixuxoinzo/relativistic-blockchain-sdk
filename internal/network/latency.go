package network

import (
	"context"
	"fmt"
	"math"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type LatencyMonitor struct {
	topologyManager *TopologyManager
	logger          *zap.Logger
	measurements    map[string]*LatencyMeasurement
	mu              sync.RWMutex
	stopChan        chan struct{}
}

type LatencyMeasurement struct {
	SourceNode   string
	TargetNode   string
	Theoretical  time.Duration
	Actual       time.Duration
	Jitter       time.Duration
	PacketLoss   float64
	LastMeasured time.Time
	Measurements int
	Average      time.Duration
	StdDev       time.Duration
}

func NewLatencyMonitor(topology *TopologyManager, logger *zap.Logger) *LatencyMonitor {
	return &LatencyMonitor{
		topologyManager: topology,
		logger:          logger,
		measurements:    make(map[string]*LatencyMeasurement),
		stopChan:        make(chan struct{}),
	}
}

func (lm *LatencyMonitor) StartMonitoring(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-lm.stopChan:
			return
		case <-ticker.C:
			lm.performMeasurements()
		}
	}
}

func (lm *LatencyMonitor) performMeasurements() {
	nodes := lm.topologyManager.GetActiveNodes()
	if len(nodes) < 2 {
		return
	}

	for i := 0; i < min(5, len(nodes)); i++ {
		for j := i + 1; j < min(i+3, len(nodes)); j++ {
			lm.measureLatency(nodes[i], nodes[j])
		}
	}
}

func (lm *LatencyMonitor) measureLatency(nodeA, nodeB *types.Node) {
	key := fmt.Sprintf("%s-%s", nodeA.ID, nodeB.ID)

	latency, jitter, packetLoss, err := lm.pingNode(nodeB.Address)
	if err != nil {
		lm.logger.Debug("Latency measurement failed",
			zap.String("target", nodeB.Address),
			zap.Error(err),
		)
		return
	}

	theoretical := lm.calculateTheoreticalLatency(nodeA, nodeB)

	measurement := &LatencyMeasurement{
		SourceNode:   nodeA.ID,
		TargetNode:   nodeB.ID,
		Theoretical:  theoretical,
		Actual:       latency,
		Jitter:       jitter,
		PacketLoss:   packetLoss,
		LastMeasured: time.Now().UTC(),
		Measurements: 1,
		Average:      latency,
	}

	lm.mu.Lock()
	if existing, exists := lm.measurements[key]; exists {
		existing.Measurements++
		existing.Actual = latency
		existing.Jitter = jitter
		existing.PacketLoss = packetLoss
		existing.LastMeasured = time.Now().UTC()
		existing.Average = time.Duration(
			(float64(existing.Average)*float64(existing.Measurements-1) + float64(latency)) / float64(existing.Measurements),
		)
	} else {
		lm.measurements[key] = measurement
	}
	lm.mu.Unlock()

	lm.logger.Debug("Latency measurement completed",
		zap.String("source", nodeA.ID),
		zap.String("target", nodeB.ID),
		zap.Duration("theoretical", theoretical),
		zap.Duration("actual", latency),
		zap.Duration("jitter", jitter),
		zap.Float64("packet_loss", packetLoss),
	)
}

func (lm *LatencyMonitor) pingNode(address string) (time.Duration, time.Duration, float64, error) {
	const attempts = 5
	var latencies []time.Duration
	successful := 0

	for i := 0; i < attempts; i++ {
		start := time.Now()

		conn, err := net.DialTimeout("tcp", address, 5*time.Second)
		if err != nil {
			continue
		}
		conn.Close()

		latency := time.Since(start)
		latencies = append(latencies, latency)
		successful++

		time.Sleep(100 * time.Millisecond)
	}

	if successful == 0 {
		return 0, 0, 1.0, fmt.Errorf("all ping attempts failed")
	}

	totalLatency := time.Duration(0)
	for _, lat := range latencies {
		totalLatency += lat
	}
	averageLatency := totalLatency / time.Duration(successful)

	var variance time.Duration
	for _, lat := range latencies {
		diff := lat - averageLatency
		variance += diff * diff
	}
	variance /= time.Duration(successful)
	jitter := time.Duration(math.Sqrt(float64(variance)))

	packetLoss := 1.0 - float64(successful)/float64(attempts)

	return averageLatency, jitter, packetLoss, nil
}

func (lm *LatencyMonitor) calculateTheoreticalLatency(nodeA, nodeB *types.Node) time.Duration {
	distance := lm.calculateDistance(nodeA.Position, nodeB.Position)
	lightDelay := distance / types.SpeedOfLight
	return time.Duration(lightDelay * 1.5 * float64(time.Second))
}

func (lm *LatencyMonitor) calculateDistance(pos1, pos2 types.Position) float64 {
	latDiff := pos1.Latitude - pos2.Latitude
	lonDiff := pos1.Longitude - pos2.Longitude
	return math.Sqrt(latDiff*latDiff+lonDiff*lonDiff) * 111000
}

func (lm *LatencyMonitor) GetMeasurement(source, target string) *LatencyMeasurement {
	key := fmt.Sprintf("%s-%s", source, target)

	lm.mu.RLock()
	defer lm.mu.RUnlock()

	return lm.measurements[key]
}

func (lm *LatencyMonitor) GetAllMeasurements() map[string]*LatencyMeasurement {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	result := make(map[string]*LatencyMeasurement)
	for k, v := range lm.measurements {
		result[k] = v
	}
	return result
}

func (lm *LatencyMonitor) GetNetworkHealth() *NetworkHealth {
	measurements := lm.GetAllMeasurements()

	health := &NetworkHealth{
		TotalMeasurements: len(measurements),
		Timestamp:         time.Now().UTC(),
	}

	if len(measurements) == 0 {
		health.Status = "unknown"
		return health
	}

	var totalLatency time.Duration
	var totalJitter time.Duration
	var totalPacketLoss float64
	healthyConnections := 0

	for _, measurement := range measurements {
		totalLatency += measurement.Actual
		totalJitter += measurement.Jitter
		totalPacketLoss += measurement.PacketLoss

		if measurement.PacketLoss < 0.05 && measurement.Actual < time.Second {
			healthyConnections++
		}
	}

	health.AverageLatency = totalLatency / time.Duration(len(measurements))
	health.AverageJitter = totalJitter / time.Duration(len(measurements))
	health.AveragePacketLoss = totalPacketLoss / float64(len(measurements))
	health.HealthyConnections = healthyConnections
	health.ConnectionHealth = float64(healthyConnections) / float64(len(measurements))

	if health.ConnectionHealth > 0.8 {
		health.Status = "healthy"
	} else if health.ConnectionHealth > 0.5 {
		health.Status = "degraded"
	} else {
		health.Status = "unhealthy"
	}

	return health
}

type NetworkHealth struct {
	Status             string
	TotalMeasurements  int
	AverageLatency     time.Duration
	AverageJitter      time.Duration
	AveragePacketLoss  float64
	HealthyConnections int
	ConnectionHealth   float64
	Timestamp          time.Time
}

func (lm *LatencyMonitor) Stop() {
	close(lm.stopChan)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
