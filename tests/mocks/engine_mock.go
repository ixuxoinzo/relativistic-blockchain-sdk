package mocks

import (
	"context"
	"time"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type EngineMock struct {
	propagationDelays map[string]time.Duration
	validationResults map[string]bool
}

func NewEngineMock() *EngineMock {
	return &EngineMock{
		propagationDelays: make(map[string]time.Duration),
		validationResults: make(map[string]bool),
	}
}

func (em *EngineMock) CalculatePropagationDelay(nodeA, nodeB *types.Node) (time.Duration, error) {
	key := nodeA.ID + "-" + nodeB.ID
	if delay, exists := em.propagationDelays[key]; exists {
		return delay, nil
	}
	return 100 * time.Millisecond, nil
}

func (em *EngineMock) ValidateTimestamp(ctx context.Context, timestamp time.Time, position types.Position, originNode string) (bool, *types.ValidationResult) {
	key := originNode + "-" + timestamp.String()
	if valid, exists := em.validationResults[key]; exists {
		return valid, &types.ValidationResult{
			Valid:      valid,
			Confidence: 0.95,
			Reason:     "mock validation",
		}
	}
	return true, &types.ValidationResult{
		Valid:      true,
		Confidence: 0.95,
		Reason:     "mock validation",
	}
}

func (em *EngineMock) CalculateInterplanetaryDelay(planetA, planetB string) (time.Duration, error) {
	return 300 * time.Millisecond, nil
}

func (em *EngineMock) BatchCalculateDelays(nodes []*types.Node) (map[string]time.Duration, error) {
	results := make(map[string]time.Duration)
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			key := nodes[i].ID + "-" + nodes[j].ID
			results[key] = 100 * time.Millisecond
		}
	}
	return results, nil
}

func (em *EngineMock) SetPropagationDelay(source, target string, delay time.Duration) {
	em.propagationDelays[source+"-"+target] = delay
}

func (em *EngineMock) SetValidationResult(originNode string, timestamp time.Time, valid bool) {
	key := originNode + "-" + timestamp.String()
	em.validationResults[key] = valid
}