package consensus

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type ConsensusValidator struct {
	timingManager *TimingManager
	offsetManager *OffsetManager
	logger        *zap.Logger
}

type OffsetValidationResult struct {
	Valid    bool          `json:"valid"`
	Reason   string        `json:"reason"`
	NodeID   string        `json:"node_id"`
	Offset   time.Duration `json:"offset"`
	Expected time.Duration `json:"expected_offset"`
}

type TimingValidationResult struct {
    Valid  bool   `json:"valid"`
    Reason string `json:"reason,omitempty"`
}

func NewConsensusValidator(timingManager *TimingManager, offsetManager *OffsetManager, logger *zap.Logger) *ConsensusValidator {
	return &ConsensusValidator{
		timingManager: timingManager,
		offsetManager: offsetManager,
		logger:        logger,
	}
}

func (cv *ConsensusValidator) ValidateBlockTiming(block *types.Block, validators []string) (*TimingValidationResult, error) {
    isValid, err := cv.timingManager.ValidateBlockTiming(block.Timestamp, block.ProposedBy, validators)
    if err != nil {
        return nil, err
    }
    return &TimingValidationResult{Valid: isValid}, nil
}

func (cv *ConsensusValidator) ValidateBlockOffsets(block *types.Block, validators []string) *OffsetValidationResult {
	result := &OffsetValidationResult{
		Valid:  true,
		NodeID: block.ProposedBy,
	}

	expectedOffset, err := cv.offsetManager.GetNodeOffset(block.ProposedBy)
	if err != nil {
		result.Valid = false
		result.Reason = fmt.Sprintf("Failed to get offset for node %s: %v", block.ProposedBy, err)
		return result
	}

	adjustedTimestamp, err := cv.offsetManager.AdjustTimestamp(block.Timestamp, block.ProposedBy)
	if err != nil {
		result.Valid = false
		result.Reason = fmt.Sprintf("Failed to adjust timestamp: %v", err)
		return result
	}

	now := time.Now().UTC()
	timeDiff := now.Sub(adjustedTimestamp)

	maxAcceptable := expectedOffset.Offset * 2
	if timeDiff > maxAcceptable {
		result.Valid = false
		result.Reason = fmt.Sprintf("Adjusted timestamp difference too large: %v > %v", timeDiff, maxAcceptable)
	}

	result.Offset = expectedOffset.Offset
	result.Expected = expectedOffset.Offset

	cv.logger.Debug("Block offset validation",
		zap.String("node_id", block.ProposedBy),
		zap.Duration("offset", result.Offset),
		zap.Duration("time_diff", timeDiff),
		zap.Bool("valid", result.Valid),
	)

	return result
}

func (cv *ConsensusValidator) ValidateVoteTiming(vote *types.Vote, validators []string) (bool, string) {
	timing, err := cv.timingManager.CalculateConsensusTiming(validators)
	if err != nil {
		return false, fmt.Sprintf("Failed to calculate timing: %v", err)
	}

	now := time.Now().UTC()
	timeDiff := now.Sub(vote.Timestamp)

	maxAcceptable := timing.MaxPropagation + timing.SafetyMargin

	if timeDiff > maxAcceptable {
		return false, fmt.Sprintf("Vote timestamp too old: %v > %v", timeDiff, maxAcceptable)
	}

	return true, ""
}

func (cv *ConsensusValidator) ValidateProposalTiming(proposal *types.Proposal, validators []string) (bool, string) {
	timing, err := cv.timingManager.CalculateConsensusTiming(validators)
	if err != nil {
		return false, fmt.Sprintf("Failed to calculate timing: %v", err)
	}

	now := time.Now().UTC()
	timeDiff := now.Sub(proposal.Timestamp)

	maxAcceptable := timing.MaxPropagation * 2

	if timeDiff > maxAcceptable {
		return false, fmt.Sprintf("Proposal timestamp too old: %v > %v", timeDiff, maxAcceptable)
	}

	return true, ""
}

func (cv *ConsensusValidator) BatchValidateBlocks(blocks []*types.Block, validators []string) []*ConsensusResult {
	results := make([]*ConsensusResult, len(blocks))

	for i, block := range blocks {
		result, err := cv.ValidateBlockConsensus(block, validators)
		if err != nil {
			results[i] = &ConsensusResult{
				BlockHash: block.Hash,
				Valid:     false,
				Reason:    fmt.Sprintf("Validation error: %v", err),
				ValidatedAt: time.Now().UTC(),
			}
		} else {
			results[i] = result
		}
	}

	cv.logger.Info("Batch block validation completed",
		zap.Int("total_blocks", len(blocks)),
		zap.Int("validators", len(validators)),
	)

	return results
}

func (cv *ConsensusValidator) ValidateBlockConsensus(block *types.Block, validators []string) (*ConsensusResult, error) {
	timingResult, err := cv.ValidateBlockTiming(block, validators)
	if err != nil {
		return nil, err
	}

	offsetResult := cv.ValidateBlockOffsets(block, validators)

	result := &ConsensusResult{
		BlockHash:        block.Hash,
		Valid:            timingResult.Valid && offsetResult.Valid,
		TimingValidation: timingResult,
		OffsetValidation: offsetResult,
		Validators:       validators,
		ValidatedAt:      time.Now().UTC(),
	}

	if !result.Valid {
		result.Reason = "Consensus validation failed"
		if !timingResult.Valid {
			result.Reason += fmt.Sprintf(" (Timing: %s)", timingResult.Reason)
		}
		if !offsetResult.Valid {
			result.Reason += fmt.Sprintf(" (Offset: %s)", offsetResult.Reason)
		}
	}

	return result, nil
}

func (cv *ConsensusValidator) CheckConsensusHealth(validators []string) *ConsensusHealth {
	health := &ConsensusHealth{
		Timestamp: time.Now().UTC(),
		Validators: validators,
	}

	timing, err := cv.timingManager.CalculateConsensusTiming(validators)
	if err != nil {
		health.Status = "degraded"
		health.Issues = append(health.Issues, fmt.Sprintf("Timing calculation failed: %v", err))
	} else {
		health.BlockTime = timing.BlockTime
		health.MaxPropagation = timing.MaxPropagation
	}

	offsetStats := cv.offsetManager.GetOffsetStats()
	health.OffsetStats = offsetStats

	if offsetStats.TotalNodes < len(validators) {
		health.Status = "degraded"
		health.Issues = append(health.Issues, "Not all validators have offset data")
	}

	if offsetStats.AverageConfidence < 0.7 {
		health.Status = "degraded"
		health.Issues = append(health.Issues, "Low confidence in offset calculations")
	}

	if health.Status == "" {
		health.Status = "healthy"
	}

	return health
}

type ConsensusHealth struct {
	Status        string         `json:"status"`
	BlockTime     time.Duration  `json:"block_time,omitempty"`
	MaxPropagation time.Duration `json:"max_propagation,omitempty"`
	OffsetStats   *OffsetStats   `json:"offset_stats"`
	Validators    []string       `json:"validators"`
	Issues        []string       `json:"issues,omitempty"`
	Timestamp     time.Time      `json:"timestamp"`
}