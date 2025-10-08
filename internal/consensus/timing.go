package consensus
import (
        "fmt"
        "math"
        "sync"
        "time"
        "go.uber.org/zap"
        "github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/network"
        "github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)
type TimingManager struct {
        topologyManager *network.TopologyManager
        logger          *zap.Logger
        mu              sync.RWMutex
        timingCache     map[string]*ConsensusTiming
}
type ConsensusTiming struct {
        BlockTime      time.Duration `json:"block_time"`
        MaxPropagation time.Duration `json:"max_propagation_delay"`
        SafetyMargin   time.Duration `json:"safety_margin"`
        OptimalOffset  time.Duration `json:"optimal_offset"`
        ValidatorCount int           `json:"validator_count"`
        CalculatedAt   time.Time     `json:"calculated_at"`
}
type TimingValidationResult struct {
<<<<<<< HEAD
	Valid           bool          `json:"valid"`
	Reason          string        `json:"reason"`
	TimeDiff        time.Duration `json:"time_diff"`
	MaxAcceptable   time.Duration `json:"max_acceptable"`
	ProposedBy      string        `json:"proposed_by"`
	ValidationError error         `json:"validation_error,omitempty"`
}


func NewTimingManager(topology *network.TopologyManager, logger *zap.Logger) *TimingManager {
	return &TimingManager{
		topologyManager: topology,
		logger:          logger,
		timingCache:     make(map[string]*ConsensusTiming),
	}
=======
        Valid           bool          `json:"valid"`
        Reason          string        `json:"reason"`
        TimeDiff        time.Duration `json:"time_diff"`
        MaxAcceptable   time.Duration `json:"max_acceptable"`
        ProposedBy      string        `json:"proposed_by"`
        ValidationError error         `json:"validation_error,omitempty"`
}
func NewTimingManager(topology *network.TopologyManager, logger *zap.Logger) *TimingManager {
        return &TimingManager{
                topologyManager: topology,
                logger:          logger,
                timingCache:     make(map[string]*ConsensusTiming),
        }
}

func (tm *TimingManager) GetNodeOffset(nodeID string) (time.Duration, error) {
    return 0, fmt.Errorf("node offset feature not yet implemented for node: %s", nodeID)
}

func (tm *TimingManager) GetAllOffsets() map[string]time.Duration {
    return map[string]time.Duration{}
}

func (tm *TimingManager) ValidateBlockConsensus(block *types.Block, votes []*types.Vote) error {
    return fmt.Errorf("block consensus validation feature not yet implemented")
}

func (tm *TimingManager) CheckConsensusHealth() (map[string]interface{}, error) {
    return map[string]interface{}{"status": "unimplemented"}, nil
>>>>>>> 8d063fc (FIX: Implement all missing core method signatures and synchronize API usage.)
}

func (t *TimingValidationResult) Error() string {
	if t.ValidationError != nil {
		return t.ValidationError.Error()
	}
	return t.Reason
}

func (tm *TimingManager) CalculateConsensusTiming(validatorNodes []string) (*ConsensusTiming, error) {
        if len(validatorNodes) == 0 {
                return nil, fmt.Errorf("validator nodes list cannot be empty")
        }
        cacheKey := tm.generateCacheKey(validatorNodes)
        tm.mu.RLock()
        if cached, exists := tm.timingCache[cacheKey]; exists {
                if time.Since(cached.CalculatedAt) < 5*time.Minute {
                        tm.mu.RUnlock()
                        return cached, nil
                }
        }
        tm.mu.RUnlock()
        timing, err := tm.calculateTiming(validatorNodes)
        if err != nil {
                return nil, err
        }
        tm.mu.Lock()
        tm.timingCache[cacheKey] = timing
        tm.mu.Unlock()
        return timing, nil
}
func (tm *TimingManager) calculateTiming(validatorNodes []string) (*ConsensusTiming, error) {
        nodes := make([]*types.Node, 0, len(validatorNodes))
        for _, nodeID := range validatorNodes {
                node, err := tm.topologyManager.GetNode(nodeID)
                if err != nil {
                        tm.logger.Warn("Validator node not found, skipping",
                                zap.String("node_id", nodeID),
                                zap.Error(err),
                        )
                        continue
                }
                nodes = append(nodes, node)
        }
        if len(nodes) < 2 {
                return nil, fmt.Errorf("insufficient validator nodes: %d", len(nodes))
        }
        maxDelay := tm.calculateMaxPropagationDelay(nodes)
        timing := &ConsensusTiming{
                MaxPropagation: maxDelay,
                SafetyMargin:   time.Duration(float64(maxDelay) * types.ConsensusSafetyFactor),
                ValidatorCount: len(nodes),
                CalculatedAt:   time.Now().UTC(),
        }
        timing.BlockTime = tm.calculateOptimalBlockTime(timing.MaxPropagation, timing.SafetyMargin)
        timing.OptimalOffset = tm.calculateOptimalOffset(timing.MaxPropagation)
        tm.logger.Info("Consensus timing calculated",
                zap.Duration("block_time", timing.BlockTime),
                zap.Duration("max_propagation", timing.MaxPropagation),
                zap.Duration("safety_margin", timing.SafetyMargin),
                zap.Duration("optimal_offset", timing.OptimalOffset),
                zap.Int("validator_count", timing.ValidatorCount),
        )
        return timing, nil
}
func (tm *TimingManager) calculateMaxPropagationDelay(nodes []*types.Node) time.Duration {
        maxDelay := time.Duration(0)
        for i := 0; i < len(nodes); i++ {
                for j := i + 1; j < len(nodes); j++ {
                        distance, err := tm.calculateDistance(nodes[i].Position, nodes[j].Position)
                        if err != nil {
                                continue
                        }
                        lightDelay := distance / types.SpeedOfLight
                        networkDelay := time.Duration(lightDelay * types.NetworkFactor * float64(time.Second))
                        if networkDelay > maxDelay {
                                maxDelay = networkDelay
                        }
                }
        }
        if maxDelay < 100*time.Millisecond {
                maxDelay = 100 * time.Millisecond
        }
        return maxDelay
}
func (tm *TimingManager) calculateDistance(pos1, pos2 types.Position) (float64, error) {
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
func (tm *TimingManager) calculateOptimalBlockTime(maxPropagation, safetyMargin time.Duration) time.Duration {
        blockTime := maxPropagation + safetyMargin
        minBlockTime := 2 * time.Second
        if blockTime < minBlockTime {
                blockTime = minBlockTime
        }
        maxBlockTime := 10 * time.Minute
        if blockTime > maxBlockTime {
                blockTime = maxBlockTime
        }
        return blockTime
}
func (tm *TimingManager) calculateOptimalOffset(maxPropagation time.Duration) time.Duration {
        return maxPropagation / 2
}
func (tm *TimingManager) generateCacheKey(validatorNodes []string) string {
        return fmt.Sprintf("timing:%v", validatorNodes)
}
func (tm *TimingManager) GetTimingForValidators(validatorNodes []string) (*ConsensusTiming, error) {
        return tm.CalculateConsensusTiming(validatorNodes)
}
func (tm *TimingManager) ValidateBlockTiming(blockTimestamp time.Time, proposedBy string, validators []string) (bool, *TimingValidationResult) {
        timing, err := tm.CalculateConsensusTiming(validators)
        if err != nil {
                return false, &TimingValidationResult{
                        Valid:  false,
                        Reason: fmt.Sprintf("Failed to calculate consensus timing: %v", err),
                        ValidationError:  err,
                }
        }
        now := time.Now().UTC()
        timeDiff := now.Sub(blockTimestamp)
        maxAcceptable := timing.MaxPropagation + timing.SafetyMargin
        valid := timeDiff <= maxAcceptable
        result := &TimingValidationResult{
                Valid:         valid,
                Reason:        fmt.Sprintf("Time difference: %v, Max acceptable: %v", timeDiff, maxAcceptable),
                TimeDiff:      timeDiff,
                MaxAcceptable: maxAcceptable,
                ProposedBy:    proposedBy,
        }
        if !valid {
                result.Reason = fmt.Sprintf("Block timestamp too old. Difference: %v, Max: %v", timeDiff, maxAcceptable)
        }
        tm.logger.Debug("Block timing validation",
                zap.String("proposed_by", proposedBy),
                zap.Time("block_timestamp", blockTimestamp),
                zap.Duration("time_diff", timeDiff),
                zap.Duration("max_acceptable", maxAcceptable),
                zap.Bool("valid", valid),
        )
        return valid, result
}
func (tm *TimingManager) ClearCache() {
        tm.mu.Lock()
        defer tm.mu.Unlock()
        tm.timingCache = make(map[string]*ConsensusTiming)
        tm.logger.Info("Timing cache cleared")
}
func (tm *TimingManager) GetCacheStats() map[string]interface{} {
        tm.mu.RLock()
        defer tm.mu.RUnlock()
        stats := make(map[string]interface{})
        stats["cache_size"] = len(tm.timingCache)
        oldest := time.Now()
        newest := time.Time{}
        for _, timing := range tm.timingCache {
                if timing.CalculatedAt.Before(oldest) {
                        oldest = timing.CalculatedAt
                }
                if timing.CalculatedAt.After(newest) {
                        newest = timing.CalculatedAt
                }
        }
        stats["oldest_entry"] = oldest
        stats["newest_entry"] = newest
        stats["cache_age"] = time.Since(oldest)
        return stats
}
func (t *TimingValidationResult) Error() string {
        if t.ValidationError != nil {
                return t.ValidationError.Error()
        }
        return t.Reason
}
func (tm *TimingManager) GetConsensusStats() (interface{}, error) {
    return nil, nil
}

func (tm *TimingManager) RecalculateAllOffsets() error {
    return nil
}

func (tm *TimingManager) SyncAllNodes() error {
    return nil
}
