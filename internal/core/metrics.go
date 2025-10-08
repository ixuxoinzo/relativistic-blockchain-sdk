package core

import (
	"sync"
	"time"

	"go.uber.org/zap"
        "github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"

)

func NewMetricsCollector(engine *RelativisticEngine, logger *zap.Logger) *MetricsCollector {
	return &MetricsCollector{
		engine:             engine,
		logger:             logger,
		metricsHistory: make([]*types.EngineMetrics, 0),
		maxHistorySize:     1000,
		collectionInterval: time.Minute,
	}
}

func (mc *MetricsCollector) StartCollection() {
	ticker := time.NewTicker(mc.collectionInterval)
	defer ticker.Stop()

	for range ticker.C {
		mc.collectMetrics()
	}
}

func (mc *MetricsCollector) collectMetrics() {
	metrics := mc.engine.GetEngineMetrics()

	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metricsHistory = append(mc.metricsHistory, metrics)

	if len(mc.metricsHistory) > mc.maxHistorySize {
		mc.metricsHistory = mc.metricsHistory[1:]
	}

	mc.logger.Debug("Collected engine metrics",
		zap.Int64("calculations", metrics.CalculationsTotal),
		zap.Int64("validations", metrics.ValidationsTotal),
		zap.Int64("cache_hits", metrics.CacheHits),
		zap.Int64("cache_misses", metrics.CacheMisses),
		zap.Int64("errors", metrics.ErrorsTotal),
	)
}

func (mc *MetricsCollector) GetMetricsHistory(limit int) []*types.EngineMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if limit <= 0 || limit > len(mc.metricsHistory) {
		limit = len(mc.metricsHistory)
	}

	start := len(mc.metricsHistory) - limit
	if start < 0 {
		start = 0
	}

	return mc.metricsHistory[start:]
}

func (mc *MetricsCollector) GetMetricsSummary() *MetricsSummary {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if len(mc.metricsHistory) == 0 {
		return &MetricsSummary{}
	}

	summary := &MetricsSummary{
		CollectionPeriod: time.Since(mc.metricsHistory[0].StartTime),
		DataPoints:       len(mc.metricsHistory),
	}

	for _, metrics := range mc.metricsHistory {
		summary.TotalCalculations += metrics.CalculationsTotal
		summary.TotalValidations += metrics.ValidationsTotal
		summary.TotalCacheHits += metrics.CacheHits
		summary.TotalCacheMisses += metrics.CacheMisses
		summary.TotalErrors += metrics.ErrorsTotal
	}

	minutes := summary.CollectionPeriod.Minutes()
	if minutes > 0 {
		summary.CalculationsPerMinute = float64(summary.TotalCalculations) / minutes
		summary.ValidationsPerMinute = float64(summary.TotalValidations) / minutes
	}

	totalCacheAccess := summary.TotalCacheHits + summary.TotalCacheMisses
	if totalCacheAccess > 0 {
		summary.CacheHitRate = float64(summary.TotalCacheHits) / float64(totalCacheAccess)
	}

	return summary
}

type MetricsSummary struct {
	TotalCalculations     int64
	TotalValidations      int64
	TotalCacheHits        int64
	TotalCacheMisses      int64
	TotalErrors           int64
	CalculationsPerMinute float64
	ValidationsPerMinute  float64
	CacheHitRate          float64
	CollectionPeriod      time.Duration
	DataPoints            int
}

type EngineMetrics struct {
	CalculationsTotal int64
	ValidationsTotal  int64
	CacheHits         int64
	CacheMisses       int64
	ErrorsTotal       int64
	CollectionTime    time.Time
}

type MetricsCollector struct {
    engine             *RelativisticEngine
    logger             *zap.Logger
    mu                 sync.RWMutex
    metricsHistory     []*types.EngineMetrics 
    maxHistorySize     int
    collectionInterval time.Duration
}

