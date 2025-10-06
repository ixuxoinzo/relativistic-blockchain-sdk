package consensus

import (
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type OffsetManager struct {
	timingManager   *TimingManager
	logger          *zap.Logger
	mu              sync.RWMutex
	nodeOffsets     map[string]*NodeOffset
	globalOffset    time.Duration
}

type NodeOffset struct {
	NodeID         string        `json:"node_id"`
	Offset         time.Duration `json:"offset"`
	Confidence     float64       `json:"confidence"`
	LastCalculated time.Time     `json:"last_calculated"`
	Measurements   int           `json:"measurements"`
	Region         string        `json:"region"`
}

type OffsetCalculation struct {
	SourceNode      string        `json:"source_node"`
	TargetNode      string        `json:"target_node"`
	CalculatedOffset time.Duration `json:"calculated_offset"`
	Distance        float64       `json:"distance_km"`
	PropagationDelay time.Duration `json:"propagation_delay"`
	Timestamp       time.Time     `json:"timestamp"`
}

func NewOffsetManager(timingManager *TimingManager, logger *zap.Logger) *OffsetManager {
	return &OffsetManager{
		timingManager: timingManager,
		logger:        logger,
		nodeOffsets:   make(map[string]*NodeOffset),
		globalOffset:  0,
	}
}

func (om *OffsetManager) CalculateNodeOffset(nodeID string, referenceNodes []string) (*NodeOffset, error) {
	if len(referenceNodes) == 0 {
		return nil, fmt.Errorf("reference nodes list cannot be empty")
	}

	node, err := om.getNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node %s: %w", nodeID, err)
	}

	var totalOffset time.Duration
	var totalWeight float64
	measurements := 0

	for _, refNodeID := range referenceNodes {
		if refNodeID == nodeID {
			continue
		}

		refNode, err := om.getNode(refNodeID)
		if err != nil {
			om.logger.Debug("Reference node not found, skipping",
				zap.String("ref_node", refNodeID),
				zap.Error(err),
			)
			continue
		}

		offset, confidence, err := om.calculateOffsetBetweenNodes(node, refNode)
		if err != nil {
			om.logger.Debug("Failed to calculate offset",
				zap.String("node", nodeID),
				zp.String("ref_node", refNodeID),
				zap.Error(err),
			)
			continue
		}

		totalOffset += time.Duration(float64(offset) * confidence)
		totalWeight += confidence
		measurements++
	}

	if measurements == 0 {
		return nil, fmt.Errorf("no valid offset measurements for node %s", nodeID)
	}

	averageOffset := time.Duration(float64(totalOffset) / totalWeight)
	overallConfidence := totalWeight / float64(measurements)

	offset := &NodeOffset{
		NodeID:         nodeID,
		Offset:         averageOffset,
		Confidence:     overallConfidence,
		LastCalculated: time.Now().UTC(),
		Measurements:   measurements,
		Region:         node.Metadata.Region,
	}

	om.mu.Lock()
	om.nodeOffsets[nodeID] = offset
	om.mu.Unlock()

	om.logger.Info("Node offset calculated",
		zap.String("node_id", nodeID),
		zap.Duration("offset", offset.Offset),
		zap.Float64("confidence", offset.Confidence),
		zap.Int("measurements", offset.Measurements),
	)

	return offset, nil
}

func (om *OffsetManager) calculateOffsetBetweenNodes(nodeA, nodeB *types.Node) (time.Duration, float64, error) {
	distance, err := om.calculateDistance(nodeA.Position, nodeB.Position)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to calculate distance: %w", err)
	}

	propagationDelay := distance / types.SpeedOfLight
	networkDelay := time.Duration(propagationDelay * types.NetworkFactor * float64(time.Second))

	offset := networkDelay / 2

	confidence := om.calculateConfidence(distance, propagationDelay)

	return offset, confidence, nil
}

func (om *OffsetManager) calculateDistance(pos1, pos2 types.Position) (float64, error) {
	lat1 := pos1.Latitude * math.Pi / 180
	lon1 := pos1.Longitude * math.Pi / 180
	lat2 := pos2.Latitude * math.Pi / 180
	lon2 := pos2.Longitude * math.Pi / 180

	dLat := lat2 - lat1
	dLon := lon2 - lon1

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := types.EarthRadius * c

	altDiff := math.Abs(pos1.Altitude - pos2.Altitude)
	totalDistance := math.Sqrt(math.Pow(distance, 2) + math.Pow(altDiff, 2))

	return totalDistance, nil
}

func (om *OffsetManager) calculateConfidence(distance, propagationDelay float64) float64 {
	maxDistance := 20000.0
	distanceConfidence := 1.0 - (distance / maxDistance)
	if distanceConfidence < 0.1 {
		distanceConfidence = 0.1
	}

	maxDelay := 0.5
	delayConfidence := 1.0 - (propagationDelay / maxDelay)
	if delayConfidence < 0.1 {
		delayConfidence = 0.1
	}

	return (distanceConfidence + delayConfidence) / 2
}

func (om *OffsetManager) getNode(nodeID string) (*types.Node, error) {
	return om.timingManager.topologyManager.GetNode(nodeID)
}

func (om *OffsetManager) GetNodeOffset(nodeID string) (*NodeOffset, error) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	offset, exists := om.nodeOffsets[nodeID]
	if !exists {
		return nil, fmt.Errorf("offset not found for node %s", nodeID)
	}

	if time.Since(offset.LastCalculated) > 30*time.Minute {
		return nil, fmt.Errorf("offset for node %s is stale", nodeID)
	}

	return offset, nil
}

func (om *OffsetManager) CalculateGlobalOffset() time.Duration {
	om.mu.RLock()
	defer om.mu.RUnlock()

	if len(om.nodeOffsets) == 0 {
		return 0
	}

	var totalWeightedOffset time.Duration
	var totalWeight float64

	for _, offset := range om.nodeOffsets {
		if offset.Confidence > 0.5 {
			totalWeightedOffset += time.Duration(float64(offset.Offset) * offset.Confidence)
			totalWeight += offset.Confidence
		}
	}

	if totalWeight == 0 {
		return 0
	}

	globalOffset := time.Duration(float64(totalWeightedOffset) / totalWeight)
	om.globalOffset = globalOffset

	om.logger.Info("Global offset calculated",
		zap.Duration("offset", globalOffset),
		zap.Int("nodes_considered", len(om.nodeOffsets)),
	)

	return globalOffset
}

func (om *OffsetManager) GetGlobalOffset() time.Duration {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.globalOffset
}

func (om *OffsetManager) AdjustTimestamp(timestamp time.Time, nodeID string) (time.Time, error) {
	offset, err := om.GetNodeOffset(nodeID)
	if err != nil {
		return timestamp, fmt.Errorf("failed to get offset for node %s: %w", nodeID, err)
	}

	adjusted := timestamp.Add(offset.Offset)
	return adjusted, nil
}

func (om *OffsetManager) BatchCalculateOffsets(nodeIDs []string) map[string]*NodeOffset {
	results := make(map[string]*NodeOffset)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, nodeID := range nodeIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			allNodes := om.getAllNodeIDs()
			offset, err := om.CalculateNodeOffset(id, allNodes)
			if err != nil {
				om.logger.Warn("Failed to calculate offset for node",
					zap.String("node_id", id),
					zap.Error(err),
				)
				return
			}

			mu.Lock()
			results[id] = offset
			mu.Unlock()
		}(nodeID)
	}

	wg.Wait()

	om.logger.Info("Batch offset calculation completed",
		zap.Int("total_nodes", len(nodeIDs)),
		zap.Int("successful", len(results)),
	)

	return results
}

func (om *OffsetManager) getAllNodeIDs() []string {
	nodes := om.timingManager.topologyManager.GetAllNodes()
	nodeIDs := make([]string, len(nodes))
	for i, node := range nodes {
		nodeIDs[i] = node.ID
	}
	return nodeIDs
}

func (om *OffsetManager) GetOffsetStats() *OffsetStats {
	om.mu.RLock()
	defer om.mu.RUnlock()

	stats := &OffsetStats{
		TotalNodes:      len(om.nodeOffsets),
		Timestamp:       time.Now().UTC(),
		RegionBreakdown: make(map[string]int),
		ConfidenceStats: &ConfidenceStats{},
	}

	var totalConfidence float64
	var minOffset, maxOffset time.Duration

	for _, offset := range om.nodeOffsets {
		stats.RegionBreakdown[offset.Region]++
		totalConfidence += offset.Confidence

		if offset.Offset < minOffset || minOffset == 0 {
			minOffset = offset.Offset
		}
		if offset.Offset > maxOffset {
			maxOffset = offset.Offset
		}

		switch {
		case offset.Confidence >= 0.8:
			stats.ConfidenceStats.High++
		case offset.Confidence >= 0.5:
			stats.ConfidenceStats.Medium++
		default:
			stats.ConfidenceStats.Low++
		}
	}

	if stats.TotalNodes > 0 {
		stats.AverageConfidence = totalConfidence / float64(stats.TotalNodes)
		stats.MinOffset = minOffset
		stats.MaxOffset = maxOffset
	}

	return stats
}

type OffsetStats struct {
	TotalNodes       int                `json:"total_nodes"`
	AverageConfidence float64           `json:"average_confidence"`
	MinOffset        time.Duration      `json:"min_offset"`
	MaxOffset        time.Duration      `json:"max_offset"`
	RegionBreakdown  map[string]int     `json:"region_breakdown"`
	ConfidenceStats  *ConfidenceStats   `json:"confidence_stats"`
	Timestamp        time.Time          `json:"timestamp"`
}

type ConfidenceStats struct {
	High   int `json:"high"`
	Medium int `json:"medium"`
	Low    int `json:"low"`
}

func (om *OffsetManager) ClearStaleOffsets() int {
	om.mu.Lock()
	defer om.mu.Unlock()

	staleThreshold := time.Now().Add(-30 * time.Minute)
	removedCount := 0

	for nodeID, offset := range om.nodeOffsets {
		if offset.LastCalculated.Before(staleThreshold) {
			delete(om.nodeOffsets, nodeID)
			removedCount++
		}
	}

	if removedCount > 0 {
		om.logger.Info("Cleared stale offsets",
			zap.Int("count", removedCount),
		)
	}

	return removedCount
}