package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type ValidationEngine struct {
	relativisticEngine *RelativisticEngine
	logger             *zap.Logger
	validationHistory  map[string]*ValidationRecord
	mu                 sync.RWMutex
}

type ValidationRecord struct {
	BlockHash     string
	Timestamp     time.Time
	NodePosition  types.Position
	OriginNode    string
	Valid         bool
	Confidence    float64
	ExpectedDelay time.Duration
	ActualDiff    time.Duration
	Reason        string
	ValidatedAt   time.Time
}

func NewValidationEngine(relativisticEngine *RelativisticEngine, logger *zap.Logger) *ValidationEngine {
	return &ValidationEngine{
		relativisticEngine: relativisticEngine,
		logger:             logger,
		validationHistory:  make(map[string]*ValidationRecord),
	}
}

func (ve *ValidationEngine) ValidateBlockTimestamp(ctx context.Context, block *types.Block, originNode string) (*types.ValidationResult, error) {
	if block == nil {
		return nil, fmt.Errorf("block cannot be nil")
	}

	valid, result := ve.relativisticEngine.ValidateTimestamp(ctx, block.Timestamp, block.NodePosition, originNode)

	ve.recordValidation(block.Hash, block.Timestamp, block.NodePosition, originNode, valid, result)

	ve.logger.Info("Block timestamp validation",
		zap.String("block_hash", block.Hash),
		zap.String("origin_node", originNode),
		zap.Bool("valid", valid),
		zap.Float64("confidence", result.Confidence),
		zap.Duration("time_diff", result.ActualDiff),
	)

	return result, nil
}

func (ve *ValidationEngine) ValidateTransactionTimestamp(ctx context.Context, tx *types.Transaction, originNode string) (*types.ValidationResult, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction cannot be nil")
	}

	valid, result := ve.relativisticEngine.ValidateTimestamp(ctx, tx.Timestamp, tx.NodePosition, originNode)

	ve.recordValidation(tx.Hash, tx.Timestamp, tx.NodePosition, originNode, valid, result)

	ve.logger.Debug("Transaction timestamp validation",
		zap.String("tx_hash", tx.Hash),
		zap.String("origin_node", originNode),
		zap.Bool("valid", valid),
		zap.Float64("confidence", result.Confidence),
	)

	return result, nil
}

func (ve *ValidationEngine) BatchValidateTimestamps(ctx context.Context, items []*types.ValidatableItem, originNode string) ([]*types.ValidationResult, error) {
	results := make([]*types.ValidationResult, len(items))
	var wg sync.WaitGroup
	var mu sync.Mutex
	errCh := make(chan error, len(items))

	for i, item := range items {
		wg.Add(1)
		go func(index int, validatableItem *types.ValidatableItem) {
			defer wg.Done()

			var result *types.ValidationResult
			var err error

			switch validatableItem.Type {
			case types.ItemTypeBlock:
				result, err = ve.ValidateBlockTimestamp(ctx, validatableItem.Block, originNode)
			case types.ItemTypeTransaction:
				result, err = ve.ValidateTransactionTimestamp(ctx, validatableItem.Transaction, originNode)
			default:
				err = fmt.Errorf("unknown validatable item type: %s", validatableItem.Type)
			}

			if err != nil {
				errCh <- fmt.Errorf("validation failed for item %d: %w", index, err)
				return
			}

			mu.Lock()
			results[index] = result
			mu.Unlock()
		}(i, item)
	}

	wg.Wait()
	close(errCh)

	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("batch validation completed with %d errors", len(errors))
	}

	ve.logger.Info("Batch timestamp validation completed",
		zap.Int("total_items", len(items)),
		zap.Int("successful", len(results)),
		zap.Int("errors", len(errors)),
	)

	return results, nil
}

func (ve *ValidationEngine) recordValidation(hash string, timestamp time.Time, position types.Position, origin string, valid bool, result *types.ValidationResult) {
	ve.mu.Lock()
	defer ve.mu.Unlock()

	ve.validationHistory[hash] = &ValidationRecord{
		BlockHash:     hash,
		Timestamp:     timestamp,
		NodePosition:  position,
		OriginNode:    origin,
		Valid:         valid,
		Confidence:    result.Confidence,
		ExpectedDelay: result.ExpectedDelay,
		ActualDiff:    result.ActualDiff,
		Reason:        result.Reason,
		ValidatedAt:   time.Now().UTC(),
	}

	if len(ve.validationHistory) > 10000 {
		ve.cleanupValidationHistory()
	}
}

func (ve *ValidationEngine) cleanupValidationHistory() {
	keysToDelete := make([]string, 0, len(ve.validationHistory)-10000)
	for key := range ve.validationHistory {
		if len(keysToDelete) >= len(ve.validationHistory)-10000 {
			break
		}
		keysToDelete = append(keysToDelete, key)
	}

	for _, key := range keysToDelete {
		delete(ve.validationHistory, key)
	}

	ve.logger.Debug("Cleaned up validation history",
		zap.Int("removed", len(keysToDelete)),
		zap.Int("remaining", len(ve.validationHistory)),
	)
}

func (ve *ValidationEngine) GetValidationHistory(hash string) *ValidationRecord {
	ve.mu.RLock()
	defer ve.mu.RUnlock()

	return ve.validationHistory[hash]
}

func (ve *ValidationEngine) GetValidationStats(originNode string, since time.Time) *ValidationStats {
	ve.mu.RLock()
	defer ve.mu.RUnlock()

	stats := &ValidationStats{
		TotalValidations: 0,
		Successful:       0,
		Failed:           0,
		AverageConfidence: 0.0,
		StartTime:        since,
		EndTime:          time.Now().UTC(),
	}

	var totalConfidence float64

	for _, record := range ve.validationHistory {
		if record.OriginNode == originNode && record.ValidatedAt.After(since) {
			stats.TotalValidations++
			totalConfidence += record.Confidence

			if record.Valid {
				stats.Successful++
			} else {
				stats.Failed++
			}
		}
	}

	if stats.TotalValidations > 0 {
		stats.AverageConfidence = totalConfidence / float64(stats.TotalValidations)
	}

	return stats
}

type ValidationStats struct {
	TotalValidations int
	Successful       int
	Failed           int
	AverageConfidence float64
	StartTime        time.Time
	EndTime          time.Time
}

func (ve *ValidationEngine) DetectAnomalies(since time.Time) []*ValidationAnomaly {
	ve.mu.RLock()
	defer ve.mu.RUnlock()

	var anomalies []*ValidationAnomaly
	now := time.Now().UTC()

	for hash, record := range ve.validationHistory {
		if record.ValidatedAt.Before(since) {
			continue
		}

		if record.Confidence < 0.5 {
			anomalies = append(anomalies, &ValidationAnomaly{
				Type:        "LowConfidence",
				Hash:        hash,
				Confidence:  record.Confidence,
				Timestamp:   record.ValidatedAt,
				Description: fmt.Sprintf("Validation confidence too low: %.2f", record.Confidence),
			})
		}

		if record.ActualDiff > time.Hour {
			anomalies = append(anomalies, &ValidationAnomaly{
				Type:        "LargeTimeDifference",
				Hash:        hash,
				TimeDiff:    record.ActualDiff,
				Timestamp:   record.ValidatedAt,
				Description: fmt.Sprintf("Unusually large time difference: %v", record.ActualDiff),
			})
		}
	}

	ve.logger.Info("Anomaly detection completed",
		zap.Int("anomalies_found", len(anomalies)),
		zap.Time("since", since),
	)

	return anomalies
}

type ValidationAnomaly struct {
	Type        string
	Hash        string
	Confidence  float64
	TimeDiff    time.Duration
	Timestamp   time.Time
	Description string
}