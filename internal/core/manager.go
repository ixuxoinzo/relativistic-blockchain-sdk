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

type CoreManager struct {
	engine            *Engine
	calculator        *Calculator
	topologyManager   *network.TopologyManager
	logger            *zap.Logger
	mu                sync.RWMutex
	shutdownChan      chan struct{}
	backgroundWorkers *sync.WaitGroup
}

func NewCoreManager(topology *network.TopologyManager, latency *network.LatencyMonitor, logger *zap.Logger) *CoreManager {
	engine := NewEngine(topology, latency, logger)
	calculator := NewCalculator(logger)

	return &CoreManager{
		engine:            engine,
		calculator:        calculator,
		topologyManager:   topology,
		logger:            logger,
		shutdownChan:      make(chan struct{}),
		backgroundWorkers: &sync.WaitGroup{},
	}
}

func (cm *CoreManager) Start(ctx context.Context) error {
	cm.logger.Info("Starting Core Manager")

	cm.backgroundWorkers.Add(2)
	go cm.metricsCollector(ctx)
	go cm.cacheCleaner(ctx)

	cm.logger.Info("Core Manager started successfully")
	return nil
}

func (cm *CoreManager) Stop() error {
	cm.logger.Info("Stopping Core Manager")

	close(cm.shutdownChan)
	cm.backgroundWorkers.Wait()

	cm.engine.Shutdown()
	cm.logger.Info("Core Manager stopped successfully")
	return nil
}

func (cm *CoreManager) metricsCollector(ctx context.Context) {
	defer cm.backgroundWorkers.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.shutdownChan:
			return
		case <-ticker.C:
			cm.collectMetrics()
		}
	}
}

func (cm *CoreManager) cacheCleaner(ctx context.Context) {
	defer cm.backgroundWorkers.Done()

	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.shutdownChan:
			return
		case <-ticker.C:
			cm.engine.ClearCache()
		}
	}
}

func (cm *CoreManager) collectMetrics() {
	metrics := cm.engine.GetNetworkMetrics()

	cm.logger.Info("Collected network metrics",
		zap.Int("total_nodes", metrics.TotalNodes),
		zap.Int("active_nodes", metrics.ActiveNodes),
		zap.Float64("network_coverage", metrics.NetworkCoverage),
		zap.Duration("average_delay", metrics.AverageDelay),
		zap.Int64("calculations", metrics.EngineCalculations),
		zap.Int64("validations", metrics.EngineValidations),
	)
}

func (cm *CoreManager) CalculateNetworkPropagation(source string, targets []string) (map[string]*types.PropagationResult, error) {
	return cm.engine.CalculateNodePropagation(source, targets)
}

func (cm *CoreManager) ValidateBlock(ctx context.Context, block *types.Block, originNode string) (*types.ValidationResult, error) {
	return cm.engine.ValidateBlockTimestamp(ctx, block, originNode)
}

func (cm *CoreManager) ValidateTransaction(ctx context.Context, tx *types.Transaction, originNode string) (*types.ValidationResult, error) {
	return cm.engine.ValidateTransactionTimestamp(ctx, tx, originNode)
}

func (cm *CoreManager) CalculateInterplanetary(planetA, planetB string) (time.Duration, error) {
	return cm.engine.CalculateInterplanetaryDelay(planetA, planetB)
}

func (cm *CoreManager) GetNetworkHealth() *types.HealthStatus {
	return cm.engine.HealthCheck(context.Background())
}

func (cm *CoreManager) GetPropagationAnalysis(source, target string) *PropagationAnalysis {
	stats := cm.engine.GetPropagationStats(source, target)
	if stats == nil {
		return nil
	}

	history := cm.engine.propagationManager.GetPropagationHistory(source, target, 100)

	analysis := &PropagationAnalysis{
		Stats:   stats,
		History: history,
	}

	if len(history) >= 2 {
		analysis.Trend = cm.calculateTrend(history)
	}

	return analysis
}

func (cm *CoreManager) calculateTrend(history []*PropagationHistory) string {
	if len(history) < 2 {
		return "stable"
	}

	recentCount := min(5, len(history))
	olderCount := min(5, len(history)-recentCount)

	recentAvg := cm.calculateAverageDelay(history[:recentCount])
	olderAvg := cm.calculateAverageDelay(history[recentCount : recentCount+olderCount])

	threshold := time.Millisecond * 10
	if recentAvg > olderAvg+threshold {
		return "increasing"
	} else if recentAvg < olderAvg-threshold {
		return "decreasing"
	}
	return "stable"
}

func (cm *CoreManager) calculateAverageDelay(history []*PropagationHistory) time.Duration {
	if len(history) == 0 {
		return 0
	}

	var total time.Duration
	for _, entry := range history {
		total += entry.CalculatedDelay
	}
	return total / time.Duration(len(history))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type PropagationAnalysis struct {
	Stats   *PropagationStats
	History []*PropagationHistory
	Trend   string
}

func (cm *CoreManager) GetValidationInsights(originNode string, period time.Duration) *ValidationInsights {
	since := time.Now().Add(-period)
	stats := cm.engine.GetValidationStats(originNode, since)
	anomalies := cm.engine.DetectValidationAnomalies(since)

	insights := &ValidationInsights{
		Stats:          stats,
		Anomalies:      anomalies,
		AnalysisPeriod: period,
		GeneratedAt:    time.Now().UTC(),
	}

	if stats != nil && stats.TotalValidations > 0 {
		insights.SuccessRate = float64(stats.Successful) / float64(stats.TotalValidations)
	}

	return insights
}

type ValidationInsights struct {
	Stats          *ValidationStats
	Anomalies      []*ValidationAnomaly
	SuccessRate    float64
	AnalysisPeriod time.Duration
	GeneratedAt    time.Time
}

func (cm *CoreManager) BatchCalculateOptimalPaths(sources []string, targets []string) (map[string]OptimalPathResult, error) {
	results := make(map[string]OptimalPathResult)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errCh := make(chan error, len(sources))

	for _, source := range sources {
		wg.Add(1)
		go func(src string) {
			defer wg.Done()

			path, totalDelay, err := cm.engine.propagationManager.CalculateOptimalPropagationPath(src, targets)
			if err != nil {
				errCh <- fmt.Errorf("failed to calculate optimal path for %s: %w", src, err)
				return
			}

			result := OptimalPathResult{
				Source:      src,
				OptimalPath: path,
				TotalDelay:  totalDelay,
				TargetCount: len(path),
			}

			mu.Lock()
			results[src] = result
			mu.Unlock()
		}(source)
	}

	wg.Wait()
	close(errCh)

	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("batch calculation completed with %d errors", len(errors))
	}

	return results, nil
}

type OptimalPathResult struct {
	Source      string
	OptimalPath []string
	TotalDelay  time.Duration
	TargetCount int
}
