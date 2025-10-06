package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/network"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type Engine struct {
	relativisticEngine *RelativisticEngine
	propagationManager *PropagationManager
	validationEngine   *ValidationEngine
	topologyManager    *network.TopologyManager
	logger             *zap.Logger
	mu                 sync.RWMutex
}

func NewEngine(topology *network.TopologyManager, latency *network.LatencyMonitor, logger *zap.Logger) *Engine {
	relativisticEngine := NewRelativisticEngine(topology, latency, logger)
	propagationManager := NewPropagationManager(relativisticEngine, logger)
	validationEngine := NewValidationEngine(relativisticEngine, logger)

	return &Engine{
		relativisticEngine: relativisticEngine,
		propagationManager: propagationManager,
		validationEngine:   validationEngine,
		topologyManager:    topology,
		logger:             logger,
	}
}

func (e *Engine) CalculateNodePropagation(source string, targets []string) (map[string]*types.PropagationResult, error) {
	return e.propagationManager.CalculatePropagationPath(source, targets)
}

func (e *Engine) ValidateBlockTimestamp(ctx context.Context, block *types.Block, originNode string) (*types.ValidationResult, error) {
	return e.validationEngine.ValidateBlockTimestamp(ctx, block, originNode)
}

func (e *Engine) ValidateTransactionTimestamp(ctx context.Context, tx *types.Transaction, originNode string) (*types.ValidationResult, error) {
	return e.validationEngine.ValidateTransactionTimestamp(ctx, tx, originNode)
}

func (e *Engine) CalculateInterplanetaryDelay(planetA, planetB string) (time.Duration, error) {
	return e.relativisticEngine.CalculateInterplanetaryDelay(planetA, planetB)
}

func (e *Engine) GetNetworkMetrics() *types.NetworkMetrics {
	return e.relativisticEngine.GetNetworkMetrics()
}

func (e *Engine) GetPropagationStats(source, target string) *PropagationStats {
	return e.propagationManager.GetPropagationStats(source, target)
}

func (e *Engine) GetValidationStats(originNode string, since time.Time) *ValidationStats {
	return e.validationEngine.GetValidationStats(originNode, since)
}

func (e *Engine) DetectValidationAnomalies(since time.Time) []*ValidationAnomaly {
	return e.validationEngine.DetectAnomalies(since)
}

func (e *Engine) BatchCalculateNodeDelays(nodes []*types.Node) (map[string]time.Duration, error) {
	return e.relativisticEngine.BatchCalculateDelays(nodes)
}

func (e *Engine) ClearCache() {
	e.relativisticEngine.ClearCache()
	e.logger.Info("Engine cache cleared")
}

func (e *Engine) HealthCheck(ctx context.Context) *types.HealthStatus {
	nodes := e.topologyManager.GetAllNodes()
	
	status := &types.HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   "1.0.0",
		NodeCount: len(nodes),
		Uptime:    "0s",
		Components: map[string]string{
			"relativistic_engine":  "healthy",
			"propagation_manager": "healthy",
			"validation_engine":   "healthy",
			"topology_manager":    "healthy",
		},
	}

	testNode := &types.Node{
		ID: "test",
		Position: types.Position{
			Latitude:  0,
			Longitude: 0,
			Altitude:  0,
		},
	}

	_, err := e.relativisticEngine.CalculatePropagationDelay(testNode, testNode)
	if err != nil {
		status.Status = "degraded"
		status.Components["relativistic_engine"] = "degraded"
	}

	return status
}

func (e *Engine) Shutdown() {
	e.logger.Info("Shutting down engine components...")
	e.logger.Info("Engine shutdown completed")
}