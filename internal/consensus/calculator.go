package consensus

import (
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type ConsensusCalculator struct {
	timingManager *TimingManager
	offsetManager *OffsetManager
	logger        *zap.Logger
	mu            sync.RWMutex
	cache         map[string]*CalculationResult
}

type CalculationResult struct {
	OptimalBlockTime  time.Duration   `json:"optimal_block_time"`
	MaxPropagation    time.Duration   `json:"max_propagation_delay"`
	SafetyMargin      time.Duration   `json:"safety_margin"`
	NodeOffsets       map[string]time.Duration `json:"node_offsets"`
	Confidence        float64         `json:"confidence"`
	CalculatedAt      time.Time       `json:"calculated_at"`
	ValidatorCount    int             `json:"validator_count"`
}

func NewConsensusCalculator(timingManager *TimingManager, offsetManager *OffsetManager, logger *zap.Logger) *ConsensusCalculator {
	return &ConsensusCalculator{
		timingManager: timingManager,
		offsetManager: offsetManager,
		logger:        logger,
		cache:         make(map[string]*CalculationResult),
	}
}

func (cc *ConsensusCalculator) CalculateConsensusParameters(validatorNodes []string) (*CalculationResult, error) {
	if len(validatorNodes) == 0 {
		return nil, fmt.Errorf("validator nodes list cannot be empty")
	}

	cacheKey := cc.generateCacheKey(validatorNodes)

	cc.mu.RLock()
	if cached, exists := cc.cache[cacheKey]; exists {
		if time.Since(cached.CalculatedAt) < 2*time.Minute {
			cc.mu.RUnlock()
			return cached, nil
		}
	}
	cc.mu.RUnlock()

	result, err := cc.performCalculation(validatorNodes)
	if err != nil {
		return nil, err
	}

	cc.mu.Lock()
	cc.cache[cacheKey] = result
	cc.mu.Unlock()

	return result, nil
}

func (cc *ConsensusCalculator) performCalculation(validatorNodes []string) (*CalculationResult, error) {
	timing, err := cc.timingManager.CalculateConsensusTiming(validatorNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate timing: %w", err)
	}

	nodeOffsets := make(map[string]time.Duration)
	totalConfidence := 0.0
	validOffsets := 0

	for _, nodeID := range validatorNodes {
		offset, err := cc.offsetManager.GetNodeOffset(nodeID)
		if err != nil {
			cc.logger.Debug("Failed to get offset for node, using zero offset",
				zap.String("node_id", nodeID),
				zap.Error(err),
			)
			nodeOffsets[nodeID] = 0
			continue
		}

		nodeOffsets[nodeID] = offset.Offset
		totalConfidence += offset.Confidence
		validOffsets++
	}

	overallConfidence := 0.0
	if validOffsets > 0 {
		overallConfidence = totalConfidence / float64(validOffsets)
	}

	result := &CalculationResult{
		OptimalBlockTime: timing.BlockTime,
		MaxPropagation:   timing.MaxPropagation,
		SafetyMargin:     timing.SafetyMargin,
		NodeOffsets:      nodeOffsets,
		Confidence:       overallConfidence,
		CalculatedAt:     time.Now().UTC(),
		ValidatorCount:   len(validatorNodes),
	}

	cc.logger.Info("Consensus parameters calculated",
		zap.Duration("block_time", result.OptimalBlockTime),
		zap.Duration("max_propagation", result.MaxPropagation),
		zap.Float64("confidence", result.Confidence),
		zap.Int("validators", result.ValidatorCount),
	)

	return result, nil
}

func (cc *ConsensusCalculator) CalculateOptimalBlockInterval(validatorNodes []string, networkLoad float64) (time.Duration, error) {
	result, err := cc.CalculateConsensusParameters(validatorNodes)
	if err != nil {
		return 0, err
	}

	baseInterval := result.OptimalBlockTime

	loadFactor := 1.0
	if networkLoad > 0.8 {
		loadFactor = 0.7
	} else if networkLoad > 0.5 {
		loadFactor = 0.85
	}

	adjustedInterval := time.Duration(float64(baseInterval) * loadFactor)

	minInterval := 1 * time.Second
	if adjustedInterval < minInterval {
		adjustedInterval = minInterval
	}

	cc.logger.Debug("Optimal block interval calculated",
		zap.Duration("base_interval", baseInterval),
		zap.Duration("adjusted_interval", adjustedInterval),
		zap.Float64("network_load", networkLoad),
		zap.Float64("load_factor", loadFactor),
	)

	return adjustedInterval, nil
}

func (cc *ConsensusCalculator) CalculateVotingWindow(validatorNodes []string) (time.Duration, error) {
	result, err := cc.CalculateConsensusParameters(validatorNodes)
	if err != nil {
		return 0, err
	}

	votingWindow := result.MaxPropagation + result.SafetyMargin/2

	minWindow := 2 * time.Second
	if votingWindow < minWindow {
		votingWindow = minWindow
	}

	maxWindow := 30 * time.Second
	if votingWindow > maxWindow {
		votingWindow = maxWindow
	}

	return votingWindow, nil
}

func (cc *ConsensusCalculator) CalculateTimeoutParameters(validatorNodes []string) (*TimeoutParameters, error) {
	result, err := cc.CalculateConsensusParameters(validatorNodes)
	if err != nil {
		return nil, err
	}

	params := &TimeoutParameters{
		ProposalTimeout:    result.MaxPropagation * 2,
		VoteTimeout:        result.MaxPropagation + result.SafetyMargin,
		CommitTimeout:      result.MaxPropagation * 3,
		ViewChangeTimeout:  result.MaxPropagation * 5,
	}

	cc.applyTimeoutLimits(params)

	return params, nil
}

func (cc *ConsensusCalculator) applyTimeoutLimits(params *TimeoutParameters) {
	minTimeout := 2 * time.Second
	maxTimeout := 60 * time.Second

	if params.ProposalTimeout < minTimeout {
		params.ProposalTimeout = minTimeout
	} else if params.ProposalTimeout > maxTimeout {
		params.ProposalTimeout = maxTimeout
	}

	if params.VoteTimeout < minTimeout {
		params.VoteTimeout = minTimeout
	} else if params.VoteTimeout > maxTimeout {
		params.VoteTimeout = maxTimeout
	}

	if params.CommitTimeout < minTimeout {
		params.CommitTimeout = minTimeout
	} else if params.CommitTimeout > maxTimeout {
		params.CommitTimeout = maxTimeout
	}

	if params.ViewChangeTimeout < minTimeout {
		params.ViewChangeTimeout = minTimeout
	} else if params.ViewChangeTimeout > maxTimeout {
		params.ViewChangeTimeout = maxTimeout
	}
}

func (cc *ConsensusCalculator) CalculateFaultTolerance(validatorNodes []string) (*FaultTolerance, error) {
	n := len(validatorNodes)
	if n == 0 {
		return nil, fmt.Errorf("no validator nodes")
	}

	tolerance := &FaultTolerance{
		TotalNodes:    n,
		ByzantineFaults: (n - 1) / 3,
		CrashFaults:   (n - 1) / 2,
		QuorumSize:    (2 * n) / 3,
	}

	if tolerance.ByzantineFaults < 0 {
		tolerance.ByzantineFaults = 0
	}
	if tolerance.CrashFaults < 0 {
		tolerance.CrashFaults = 0
	}
	if tolerance.QuorumSize < 1 {
		tolerance.QuorumSize = 1
	}

	tolerance.ByzantineTolerance = float64(tolerance.ByzantineFaults) / float64(n)
	tolerance.CrashTolerance = float64(tolerance.CrashFaults) / float64(n)

	return tolerance, nil
}

func (cc *ConsensusCalculator) generateCacheKey(validatorNodes []string) string {
	return fmt.Sprintf("calc:%v", validatorNodes)
}

func (cc *ConsensusCalculator) ClearCache() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.cache = make(map[string]*CalculationResult)
	cc.logger.Info("Calculator cache cleared")
}

type TimeoutParameters struct {
	ProposalTimeout   time.Duration `json:"proposal_timeout"`
	VoteTimeout       time.Duration `json:"vote_timeout"`
	CommitTimeout     time.Duration `json:"commit_timeout"`
	ViewChangeTimeout time.Duration `json:"view_change_timeout"`
}

type FaultTolerance struct {
	TotalNodes         int     `json:"total_nodes"`
	ByzantineFaults    int     `json:"byzantine_faults"`
	CrashFaults        int     `json:"crash_faults"`
	QuorumSize         int     `json:"quorum_size"`
	ByzantineTolerance float64 `json:"byzantine_tolerance"`
	CrashTolerance     float64 `json:"crash_tolerance"`
}