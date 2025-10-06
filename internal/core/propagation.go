package core

import (
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type PropagationManager struct {
	engine  *RelativisticEngine
	logger  *zap.Logger
	mu      sync.RWMutex
	history map[string]*PropagationHistory
}

type PropagationHistory struct {
	SourceNode      string
	TargetNode      string
	CalculatedDelay time.Duration
	ActualDelay     time.Duration
	Distance        float64
	Timestamp       time.Time
	Success         bool
	Error           string
}

func NewPropagationManager(engine *RelativisticEngine, logger *zap.Logger) *PropagationManager {
	return &PropagationManager{
		engine:  engine,
		logger:  logger,
		history: make(map[string]*PropagationHistory),
	}
}

func (pm *PropagationManager) CalculatePropagationPath(source string, targets []string) (map[string]*types.PropagationResult, error) {
	results := make(map[string]*types.PropagationResult)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errCh := make(chan error, len(targets))

	sourceNode, err := pm.engine.topologyManager.GetNode(source)
	if err != nil {
		return nil, fmt.Errorf("source node not found: %w", err)
	}

	for _, target := range targets {
		wg.Add(1)
		go func(targetID string) {
			defer wg.Done()

			targetNode, err := pm.engine.topologyManager.GetNode(targetID)
			if err != nil {
				errCh <- fmt.Errorf("target node %s not found: %w", targetID, err)
				return
			}

			delay, err := pm.engine.CalculatePropagationDelay(sourceNode, targetNode)
			if err != nil {
				errCh <- fmt.Errorf("failed to calculate delay to %s: %w", targetID, err)
				return
			}

			distance, _ := pm.calculateDistance(sourceNode.Position, targetNode.Position)

			result := &types.PropagationResult{
				SourceNode:       source,
				TargetNode:       targetID,
				TheoreticalDelay: delay,
				ActualDelay:      0,
				Distance:         distance / 1000,
				Success:          true,
				Timestamp:        time.Now().UTC(),
			}

			mu.Lock()
			results[targetID] = result
			mu.Unlock()

			pm.recordPropagation(source, targetID, delay, 0, distance, true, "")
		}(target)
	}

	wg.Wait()
	close(errCh)

	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("propagation calculation completed with %d errors", len(errors))
	}

	pm.logger.Info("Propagation path calculation completed",
		zap.String("source", source),
		zap.Int("targets", len(targets)),
		zap.Int("successful", len(results)),
		zap.Int("errors", len(errors)),
	)

	return results, nil
}

func (pm *PropagationManager) CalculateOptimalPropagationPath(source string, targets []string) ([]string, time.Duration, error) {
	if len(targets) == 0 {
		return nil, 0, fmt.Errorf("no targets provided")
	}

	delays := make(map[string]time.Duration)
	for _, target := range targets {
		sourceNode, err := pm.engine.topologyManager.GetNode(source)
		if err != nil {
			continue
		}
		targetNode, err := pm.engine.topologyManager.GetNode(target)
		if err != nil {
			continue
		}

		delay, err := pm.engine.CalculatePropagationDelay(sourceNode, targetNode)
		if err != nil {
			continue
		}
		delays[target] = delay
	}

	sortedTargets := make([]string, 0, len(delays))
	for target := range delays {
		sortedTargets = append(sortedTargets, target)
	}

	for i := 0; i < len(sortedTargets)-1; i++ {
		for j := i + 1; j < len(sortedTargets); j++ {
			if delays[sortedTargets[i]] > delays[sortedTargets[j]] {
				sortedTargets[i], sortedTargets[j] = sortedTargets[j], sortedTargets[i]
			}
		}
	}

	totalDelay := time.Duration(0)
	if len(sortedTargets) > 0 {
		totalDelay = delays[sortedTargets[len(sortedTargets)-1]]
	}

	pm.logger.Debug("Optimal propagation path calculated",
		zap.String("source", source),
		zap.Strings("optimal_path", sortedTargets),
		zap.Duration("total_delay", totalDelay),
	)

	return sortedTargets, totalDelay, nil
}

func (pm *PropagationManager) recordPropagation(source, target string, calculated, actual time.Duration, distance float64, success bool, errorMsg string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	key := fmt.Sprintf("%s->%s:%d", source, target, time.Now().Unix())
	pm.history[key] = &PropagationHistory{
		SourceNode:      source,
		TargetNode:      target,
		CalculatedDelay: calculated,
		ActualDelay:     actual,
		Distance:        distance,
		Timestamp:       time.Now().UTC(),
		Success:         success,
		Error:           errorMsg,
	}

	if len(pm.history) > 1000 {
		pm.cleanupHistory()
	}
}

func (pm *PropagationManager) cleanupHistory() {
	keysToDelete := make([]string, 0, len(pm.history)-1000)
	for key := range pm.history {
		if len(keysToDelete) >= len(pm.history)-1000 {
			break
		}
		keysToDelete = append(keysToDelete, key)
	}

	for _, key := range keysToDelete {
		delete(pm.history, key)
	}
}

func (pm *PropagationManager) GetPropagationHistory(source, target string, limit int) []*PropagationHistory {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var history []*PropagationHistory
	count := 0

	for key, entry := range pm.history {
		if entry.SourceNode == source && entry.TargetNode == target {
			history = append(history, entry)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}

	return history
}

func (pm *PropagationManager) calculateDistance(pos1, pos2 types.Position) (float64, error) {
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

func (pm *PropagationManager) GetPropagationStats(source, target string) *PropagationStats {
	history := pm.GetPropagationHistory(source, target, 0)
	if len(history) == 0 {
		return nil
	}

	stats := &PropagationStats{
		TotalCalculations: len(history),
		SourceNode:       source,
		TargetNode:       target,
	}

	var totalDelay time.Duration
	successCount := 0

	for _, entry := range history {
		totalDelay += entry.CalculatedDelay
		if entry.Success {
			successCount++
		}
	}

	stats.SuccessRate = float64(successCount) / float64(len(history))
	stats.AverageDelay = totalDelay / time.Duration(len(history))

	return stats
}

type PropagationStats struct {
	TotalCalculations int
	SuccessRate       float64
	AverageDelay      time.Duration
	SourceNode        string
	TargetNode        string
}