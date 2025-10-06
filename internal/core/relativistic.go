package core

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.comcom/ixuxoinzo/relativistic-blockchain-sdk/internal/network"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/utils"
)

type RelativisticEngine struct {
	topologyManager *network.TopologyManager
	latencyMonitor  *network.LatencyMonitor
	logger          *zap.Logger
	config          *EngineConfig
	cache           *sync.Map
	mu              sync.RWMutex
	metrics         *types.EngineMetrics
}

type EngineConfig struct {
	SpeedOfLight          float64
	NetworkFactor         float64
	ConsensusSafetyFactor float64
	MaxAcceptableDelay    time.Duration
	CacheTTL              time.Duration
	EnableMonitoring      bool
	ValidationThreshold   float64
}

func NewRelativisticEngine(topology *network.TopologyManager, latency *network.LatencyMonitor, logger *zap.Logger) *RelativisticEngine {
	return &RelativisticEngine{
		topologyManager: topology,
		latencyMonitor:  latency,
		logger:          logger,
		config: &EngineConfig{
			SpeedOfLight:          types.SpeedOfLight,
			NetworkFactor:         types.NetworkFactor,
			ConsensusSafetyFactor: types.ConsensusSafetyFactor,
			MaxAcceptableDelay:    time.Second * time.Duration(types.MaxAcceptableDelay),
			CacheTTL:              5 * time.Minute,
			EnableMonitoring:      true,
			ValidationThreshold:   0.8,
		},
		cache:   &sync.Map{},
		metrics: &types.EngineMetrics{},
	}
}

func (e *RelativisticEngine) CalculatePropagationDelay(nodeA, nodeB *types.Node) (time.Duration, error) {
	startTime := time.Now()
	e.metrics.Mu.Lock()
	e.metrics.CalculationsTotal++
	e.metrics.Mu.Unlock()

	defer func() {
		e.logger.Debug("Propagation delay calculation completed",
			zap.Duration("duration", time.Since(startTime)),
			zap.String("node_a", nodeA.ID),
			zap.String("node_b", nodeB.ID),
		)
	}()

	cacheKey := fmt.Sprintf("delay:%s:%s", nodeA.ID, nodeB.ID)
	
	if cached, found := e.cache.Load(cacheKey); found {
		e.metrics.Mu.Lock()
		e.metrics.CacheHits++
		e.metrics.Mu.Unlock()
		return cached.(time.Duration), nil
	}

	e.metrics.Mu.Lock()
	e.metrics.CacheMisses++
	e.metrics.Mu.Unlock()

	distance, err := e.calculateGreatCircleDistance(nodeA.Position, nodeB.Position)
	if err != nil {
		e.metrics.Mu.Lock()
		e.metrics.ErrorsTotal++
		e.metrics.Mu.Unlock()
		return 0, fmt.Errorf("failed to calculate distance: %w", err)
	}

	lightDelay := distance / e.config.SpeedOfLight
	realisticDelay := lightDelay * e.config.NetworkFactor
	result := time.Duration(realisticDelay * float64(time.Second))
	
	e.cache.Store(cacheKey, result)
	time.AfterFunc(e.config.CacheTTL, func() {
		e.cache.Delete(cacheKey)
	})

	e.logger.Debug("Calculated propagation delay",
		zap.String("node_a", nodeA.ID),
		zap.String("node_b", nodeB.ID),
		zap.Float64("distance_km", distance/1000),
		zap.Duration("delay", result),
	)

	return result, nil
}

func (e *RelativisticEngine) calculateGreatCircleDistance(pos1, pos2 types.Position) (float64, error) {
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

func (e *RelativisticEngine) ValidateTimestamp(ctx context.Context, blockTimestamp time.Time, nodePosition types.Position, originNode string) (bool, *types.ValidationResult) {
	startTime := time.Now()
	e.metrics.Mu.Lock()
	e.metrics.ValidationsTotal++
	e.metrics.Mu.Unlock()

	defer func() {
		e.logger.Debug("Timestamp validation completed",
			zap.Duration("duration", time.Since(startTime)),
			zap.String("origin_node", originNode),
		)
	}()

	currentNode, err := e.topologyManager.GetNode(originNode)
	if err != nil {
		e.metrics.Mu.Lock()
		e.metrics.ErrorsTotal++
		e.metrics.Mu.Unlock()
		return false, &types.ValidationResult{
			Valid:      false,
			Reason:     fmt.Sprintf("Node not found: %s", originNode),
			Confidence: 0.0,
			ErrorCode:  types.ErrNodeNotFound,
		}
	}

	expectedDelay, err := e.CalculatePropagationDelay(currentNode, &types.Node{Position: nodePosition})
	if err != nil {
		e.metrics.Mu.Lock()
		e.metrics.ErrorsTotal++
		e.metrics.Mu.Unlock()
		return false, &types.ValidationResult{
			Valid:      false,
			Reason:     fmt.Sprintf("Delay calculation failed: %v", err),
			Confidence: 0.0,
			ErrorCode:  types.ErrCalculationFailed,
		}
	}

	now := time.Now().UTC()
	timeDiff := now.Sub(blockTimestamp.UTC())
	absTimeDiff := utils.AbsDuration(timeDiff)

	maxAcceptable := expectedDelay + e.config.MaxAcceptableDelay

	valid := absTimeDiff <= maxAcceptable
	
	confidence := 1.0 - (float64(absTimeDiff) / float64(maxAcceptable))
	if confidence < 0 {
		confidence = 0.0
	}

	if confidence < e.config.ValidationThreshold {
		valid = false
	}

	validationLevel := zap.DebugLevel
	if !valid {
		validationLevel = zap.WarnLevel
	}

	e.logger.Log(validationLevel, "Timestamp validation",
		zap.String("origin_node", originNode),
		zap.Time("block_timestamp", blockTimestamp),
		zap.Time("current_time", now),
		zap.Duration("time_diff", timeDiff),
		zap.Duration("expected_delay", expectedDelay),
		zap.Duration("max_acceptable", maxAcceptable),
		zap.Bool("valid", valid),
		zap.Float64("confidence", confidence),
	)

	return valid, &types.ValidationResult{
		Valid:          valid,
		Reason:         fmt.Sprintf("Time difference: %v, Max acceptable: %v", timeDiff, maxAcceptable),
		Confidence:     confidence,
		ExpectedDelay:  expectedDelay,
		ActualDiff:     timeDiff,
		Threshold:      e.config.ValidationThreshold,
		ErrorCode:      "",
		ValidatedAt:    now,
	}
}

func (e *RelativisticEngine) CalculateInterplanetaryDelay(planetA, planetB string) (time.Duration, error) {
	cacheKey := fmt.Sprintf("interplanetary:%s:%s", planetA, planetB)
	
	if cached, found := e.cache.Load(cacheKey); found {
		e.metrics.Mu.Lock()
		e.metrics.CacheHits++
		e.metrics.Mu.Unlock()
		return cached.(time.Duration), nil
	}

	e.metrics.Mu.Lock()
	e.metrics.CacheMisses++
	e.metrics.Mu.Unlock()

	distance, exists := types.PlanetaryDistances[planetA+"-"+planetB]
	if !exists {
		e.metrics.Mu.Lock()
		e.metrics.ErrorsTotal++
		e.metrics.Mu.Unlock()
		return 0, fmt.Errorf("planetary distance not found for %s-%s", planetA, planetB)
	}

	distanceMeters := distance * 1000
	delaySeconds := distanceMeters / e.config.SpeedOfLight
	result := time.Duration(delaySeconds * float64(time.Second))

	e.cache.Store(cacheKey, result)
	time.AfterFunc(e.config.CacheTTL, func() {
		e.cache.Delete(cacheKey)
	})

	e.logger.Info("Calculated interplanetary delay",
		zap.String("planet_a", planetA),
		zap.String("planet_b", planetB),
		zap.Float64("distance_km", distance),
		zap.Duration("delay", result),
	)

	return result, nil
}

func (e *RelativisticEngine) BatchCalculateDelays(nodes []*types.Node) (map[string]time.Duration, error) {
	results := make(map[string]time.Duration)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errCh := make(chan error, len(nodes)*len(nodes))

	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			wg.Add(1)
			go func(nodeA, nodeB *types.Node) {
				defer wg.Done()
				
				delay, err := e.CalculatePropagationDelay(nodeA, nodeB)
				if err != nil {
					errCh <- fmt.Errorf("failed to calculate delay between %s and %s: %w", nodeA.ID, nodeB.ID, err)
					return
				}

				key := fmt.Sprintf("%s-%s", nodeA.ID, nodeB.ID)
				mu.Lock()
				results[key] = delay
				mu.Unlock()
			}(nodes[i], nodes[j])
		}
	}

	wg.Wait()
	close(errCh)

	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		e.metrics.Mu.Lock()
		e.metrics.ErrorsTotal += int64(len(errors))
		e.metrics.Mu.Unlock()
		return results, fmt.Errorf("batch calculation completed with %d errors: %v", len(errors), errors)
	}

	return results, nil
}

func (e *RelativisticEngine) GetNetworkMetrics() *types.NetworkMetrics {
	nodes := e.topologyManager.GetAllNodes()
	
	metrics := &types.NetworkMetrics{
		TotalNodes:      len(nodes),
		ActiveNodes:     e.getActiveNodeCount(nodes),
		NetworkCoverage: e.calculateNetworkCoverage(nodes),
		Regions:         e.getRegionDistribution(nodes),
		CalculatedAt:    time.Now().UTC(),
	}

	if len(nodes) >= 2 {
		delays, err := e.BatchCalculateDelays(nodes)
		if err == nil {
			metrics.AverageDelay = e.calculateAverageDelay(delays)
			metrics.MaxDelay = e.calculateMaxDelay(delays)
			metrics.MinDelay = e.calculateMinDelay(delays)
		}
	}

	e.metrics.Mu.RLock()
	metrics.EngineCalculations = e.metrics.CalculationsTotal
	metrics.EngineValidations = e.metrics.ValidationsTotal
	metrics.CacheHits = e.metrics.CacheHits
	metrics.CacheMisses = e.metrics.CacheMisses
	metrics.EngineErrors = e.metrics.ErrorsTotal
	e.metrics.Mu.RUnlock()

	return metrics
}

func (e *RelativisticEngine) getActiveNodeCount(nodes []*types.Node) int {
	count := 0
	for _, node := range nodes {
		if node.IsActive {
			count++
		}
	}
	return count
}

func (e *RelativisticEngine) calculateNetworkCoverage(nodes []*types.Node) float64 {
	if len(nodes) == 0 {
		return 0.0
	}

	regionCount := make(map[string]bool)
	for _, node := range nodes {
		if node.Metadata.Region != "" {
			regionCount[node.Metadata.Region] = true
		}
	}

	return float64(len(regionCount)) / float64(len(types.Regions))
}

func (e *RelativisticEngine) getRegionDistribution(nodes []*types.Node) map[string]int {
	distribution := make(map[string]int)
	for _, node := range nodes {
		distribution[node.Metadata.Region]++
	}
	return distribution
}

func (e *RelativisticEngine) calculateAverageDelay(delays map[string]time.Duration) time.Duration {
	if len(delays) == 0 {
		return 0
	}

	var total time.Duration
	for _, delay := range delays {
		total += delay
	}
	return total / time.Duration(len(delays))
}

func (e *RelativisticEngine) calculateMaxDelay(delays map[string]time.Duration) time.Duration {
	if len(delays) == 0 {
		return 0
	}

	var max time.Duration
	for _, delay := range delays {
		if delay > max {
			max = delay
		}
	}
	return max
}

func (e *RelativisticEngine) calculateMinDelay(delays map[string]time.Duration) time.Duration {
	if len(delays) == 0 {
		return 0
	}

	min := time.Hour * 24
	for _, delay := range delays {
		if delay < min {
			min = delay
		}
	}
	return min
}

func (e *RelativisticEngine) ClearCache() {
	e.cache.Range(func(key, value interface{}) bool {
		e.cache.Delete(key)
		return true
	})
	e.logger.Info("Cache cleared")
}

func (e *RelativisticEngine) GetEngineMetrics() *types.EngineMetrics {
	e.metrics.Mu.RLock()
	defer e.metrics.Mu.RUnlock()
	
	metricsCopy := &types.EngineMetrics{
		CalculationsTotal: e.metrics.CalculationsTotal,
		ValidationsTotal:  e.metrics.ValidationsTotal,
		CacheHits:         e.metrics.CacheHits,
		CacheMisses:       e.metrics.CacheMisses,
		ErrorsTotal:       e.metrics.ErrorsTotal,
	}
	return metricsCopy
}
