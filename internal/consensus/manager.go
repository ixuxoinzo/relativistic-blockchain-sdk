package consensus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/network"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type ConsensusManager struct {
	timingManager   *TimingManager
	offsetManager   *OffsetManager
	calculator      *ConsensusCalculator
	validator       *ConsensusValidator
	synchronizer    *Synchronizer
	topologyManager *network.TopologyManager
	logger          *zap.Logger
	mu              sync.RWMutex
	stopChan        chan struct{}
}

func NewConsensusManager(topology *network.TopologyManager, logger *zap.Logger) *ConsensusManager {
	timingManager := NewTimingManager(topology, logger)
	offsetManager := NewOffsetManager(timingManager, logger)
	calculator := NewConsensusCalculator(timingManager, offsetManager, logger)
	validator := NewConsensusValidator(timingManager, offsetManager, logger)
	synchronizer := NewSynchronizer(offsetManager, logger)

	return &ConsensusManager{
		timingManager:   timingManager,
		offsetManager:   offsetManager,
		calculator:      calculator,
		validator:       validator,
		synchronizer:    synchronizer,
		topologyManager: topology,
		logger:          logger,
		stopChan:        make(chan struct{}),
	}
}

func (cm *ConsensusManager) Start(ctx context.Context) error {
	cm.logger.Info("Starting Consensus Manager")

	go cm.backgroundOffsetCalculation(ctx)
	go cm.backgroundSynchronization(ctx)

	cm.logger.Info("Consensus Manager started successfully")
	return nil
}

func (cm *ConsensusManager) Stop() error {
	cm.logger.Info("Stopping Consensus Manager")

	close(cm.stopChan)
	cm.synchronizer.Stop()

	cm.logger.Info("Consensus Manager stopped successfully")
	return nil
}

func (cm *ConsensusManager) backgroundOffsetCalculation(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopChan:
			return
		case <-ticker.C:
			cm.calculateAllOffsets()
		}
	}
}

func (cm *ConsensusManager) backgroundSynchronization(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopChan:
			return
		case <-ticker.C:
			cm.synchronizer.SyncAllNodes()
		}
	}
}

func (cm *ConsensusManager) calculateAllOffsets() {
	nodes := cm.topologyManager.GetAllNodes()
	nodeIDs := make([]string, len(nodes))
	for i, node := range nodes {
		nodeIDs[i] = node.ID
	}

	results := cm.offsetManager.BatchCalculateOffsets(nodeIDs)

	cm.logger.Info("Background offset calculation completed",
		zap.Int("total_nodes", len(nodes)),
		zap.Int("successful", len(results)),
	)
}

func (cm *ConsensusManager) ValidateBlockConsensus(block *types.Block, validators []string) (*ConsensusResult, error) {
	if block == nil {
		return nil, fmt.Errorf("block cannot be nil")
	}

	if len(validators) == 0 {
		return nil, fmt.Errorf("validators list cannot be empty")
	}

	timingValidation, err := cm.validator.ValidateBlockTiming(block, validators)
	if err != nil {
		return nil, fmt.Errorf("timing validation failed: %w", err)
	}

	offsetValidation := cm.validator.ValidateBlockOffsets(block, validators)

	result := &ConsensusResult{
		BlockHash:        block.Hash,
		Valid:            timingValidation.Valid && offsetValidation.Valid,
		TimingValidation: timingValidation,
		OffsetValidation: offsetValidation,
		Validators:       validators,
		ValidatedAt:      time.Now().UTC(),
	}

	if !result.Valid {
		result.Reason = "Consensus validation failed"
		if !timingValidation.Valid {
			result.Reason += fmt.Sprintf(" (Timing: %s)", timingValidation.Reason)
		}
		if !offsetValidation.Valid {
			result.Reason += fmt.Sprintf(" (Offset: %s)", offsetValidation.Reason)
		}
	}

	cm.logger.Info("Block consensus validation",
		zap.String("block_hash", block.Hash),
		zap.Bool("valid", result.Valid),
		zap.Int("validators", len(validators)),
	)

	return result, nil
}

func (cm *ConsensusManager) CalculateOptimalTiming(validators []string) (*ConsensusTiming, error) {
	return cm.timingManager.CalculateConsensusTiming(validators)
}

func (cm *ConsensusManager) GetNodeOffset(nodeID string) (*NodeOffset, error) {
	return cm.offsetManager.GetNodeOffset(nodeID)
}

func (cm *ConsensusManager) AdjustTimestamp(timestamp time.Time, nodeID string) (time.Time, error) {
	return cm.offsetManager.AdjustTimestamp(timestamp, nodeID)
}

func (cm *ConsensusManager) GetGlobalOffset() time.Duration {
	return cm.offsetManager.GetGlobalOffset()
}

func (cm *ConsensusManager) SyncNodeTime(nodeID string) error {
	return cm.synchronizer.SyncNode(nodeID)
}

func (cm *ConsensusManager) GetConsensusStats() *ConsensusStats {
	offsetStats := cm.offsetManager.GetOffsetStats()
	timingCacheStats := cm.timingManager.GetCacheStats()

	stats := &ConsensusStats{
		OffsetStats:      offsetStats,
		TimingCacheStats: timingCacheStats,
		GlobalOffset:     cm.offsetManager.GetGlobalOffset(),
		Timestamp:        time.Now().UTC(),
	}

	nodes := cm.topologyManager.GetAllNodes()
	stats.TotalNodes = len(nodes)
	stats.ActiveNodes = len(cm.topologyManager.GetActiveNodes())

	return stats
}

type ConsensusStats struct {
	TotalNodes       int                    `json:"total_nodes"`
	ActiveNodes      int                    `json:"active_nodes"`
	GlobalOffset     time.Duration          `json:"global_offset"`
	OffsetStats      *OffsetStats           `json:"offset_stats"`
	TimingCacheStats map[string]interface{} `json:"timing_cache_stats"`
	Timestamp        time.Time              `json:"timestamp"`
}

func (cm *ConsensusManager) HealthCheck() *types.HealthStatus {
	nodes := cm.topologyManager.GetAllNodes()
	
	status := &types.HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   "1.0.0",
		NodeCount: len(nodes),
		Components: map[string]string{
			"timing_manager":   "healthy",
			"offset_manager":   "healthy",
			"consensus_calculator": "healthy",
			"consensus_validator":  "healthy",
			"synchronizer":     "healthy",
		},
	}

	if len(nodes) == 0 {
		status.Components["timing_manager"] = "degraded"
		status.Components["offset_manager"] = "degraded"
	}

	offsetStats := cm.offsetManager.GetOffsetStats()
	if offsetStats.TotalNodes == 0 {
		status.Components["offset_manager"] = "degraded"
	}

	return status
}

type ConsensusResult struct {
	BlockHash        string                  `json:"block_hash"`
	Valid            bool                    `json:"valid"`
	Reason           string                  `json:"reason,omitempty"`
	TimingValidation *TimingValidationResult `json:"timing_validation"`
	OffsetValidation *OffsetValidationResult `json:"offset_validation"`
	Validators       []string                `json:"validators"`
	ValidatedAt      time.Time               `json:"validated_at"`
}