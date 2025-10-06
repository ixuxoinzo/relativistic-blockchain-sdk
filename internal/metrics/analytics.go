package metrics

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type AnalyticsEngine struct {
	collector *MetricsCollector
	logger    *zap.Logger
	mu        sync.RWMutex
	analytics map[string]*Analytic
}

type Analytic struct {
	Name      string
	Data      []float64
	Window    time.Duration
	MaxPoints int
	Summary   *AnalyticSummary
}

type AnalyticSummary struct {
	Count   int
	Sum     float64
	Average float64
	Min     float64
	Max     float64
	StdDev  float64
}

func NewAnalyticsEngine(collector *MetricsCollector, logger *zap.Logger) *AnalyticsEngine {
	return &AnalyticsEngine{
		collector: collector,
		logger:    logger,
		analytics: make(map[string]*Analytic),
	}
}

func (ae *AnalyticsEngine) RegisterAnalytic(name string, window time.Duration, maxPoints int) {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	
	ae.analytics[name] = &Analytic{
		Name:      name,
		Data:      make([]float64, 0),
		Window:    window,
		MaxPoints: maxPoints,
		Summary:   &AnalyticSummary{},
	}
}

func (ae *AnalyticsEngine) AddDataPoint(name string, value float64) {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	
	analytic, exists := ae.analytics[name]
	if !exists {
		return
	}
	
	analytic.Data = append(analytic.Data, value)
	
	if len(analytic.Data) > analytic.MaxPoints {
		analytic.Data = analytic.Data[1:]
	}
	
	ae.updateSummary(analytic)
}

func (ae *AnalyticsEngine) updateSummary(analytic *Analytic) {
	if len(analytic.Data) == 0 {
		return
	}
	
	summary := &AnalyticSummary{
		Count: len(analytic.Data),
		Min:   analytic.Data[0],
		Max:   analytic.Data[0],
	}
	
	for _, value := range analytic.Data {
		summary.Sum += value
		if value < summary.Min {
			summary.Min = value
		}
		if value > summary.Max {
			summary.Max = value
		}
	}
	
	summary.Average = summary.Sum / float64(summary.Count)
	
	var variance float64
	for _, value := range analytic.Data {
		diff := value - summary.Average
		variance += diff * diff
	}
	variance /= float64(summary.Count)
	summary.StdDev = math.Sqrt(variance)
	
	analytic.Summary = summary
}

func (ae *AnalyticsEngine) GetAnalytic(name string) *Analytic {
	ae.mu.RLock()
	defer ae.mu.RUnlock()
	
	return ae.analytics[name]
}

func (ae *AnalyticsEngine) GetAnalyticSummary(name string) *AnalyticSummary {
	analytic := ae.GetAnalytic(name)
	if analytic == nil {
		return nil
	}
	return analytic.Summary
}

func (ae *AnalyticsEngine) CalculateTrend(name string) string {
	analytic := ae.GetAnalytic(name)
	if analytic == nil || len(analytic.Data) < 2 {
		return "unknown"
	}
	
	recent := analytic.Data[len(analytic.Data)-1]
	previous := analytic.Data[len(analytic.Data)-2]
	
	if recent > previous {
		return "increasing"
	} else if recent < previous {
		return "decreasing"
	}
	return "stable"
}

func (ae *AnalyticsEngine) StartCollection() {
	go ae.collectMetrics()
}

func (ae *AnalyticsEngine) collectMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		metrics := ae.collector.GetMetrics()
		
		for name, value := range metrics {
			switch v := value.(type) {
			case int64:
				ae.AddDataPoint(name, float64(v))
			case float64:
				ae.AddDataPoint(name, v)
			}
		}
	}
}
