package metrics

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type MetricsCollector struct {
	logger        *zap.Logger
	mu            sync.RWMutex
	metrics       map[string]interface{}
	counters      map[string]int64
	gauges        map[string]float64
	histograms    map[string]*Histogram
	startTime     time.Time
}

type Histogram struct {
	values []float64
	sum    float64
	count  int64
}

func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	return &MetricsCollector{
		logger:     logger,
		metrics:    make(map[string]interface{}),
		counters:   make(map[string]int64),
		gauges:     make(map[string]float64),
		histograms: make(map[string]*Histogram),
		startTime:  time.Now(),
	}
}

func (mc *MetricsCollector) IncrementCounter(name string, labels map[string]string) {
	key := mc.buildKey(name, labels)
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.counters[key]++
}

func (mc *MetricsCollector) SetGauge(name string, value float64, labels map[string]string) {
	key := mc.buildKey(name, labels)
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.gauges[key] = value
}

func (mc *MetricsCollector) ObserveHistogram(name string, value float64, labels map[string]string) {
	key := mc.buildKey(name, labels)
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	hist, exists := mc.histograms[key]
	if !exists {
		hist = &Histogram{
			values: make([]float64, 0),
		}
		mc.histograms[key] = hist
	}
	
	hist.values = append(hist.values, value)
	hist.sum += value
	hist.count++
}

func (mc *MetricsCollector) buildKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	
	// Simple key building for demonstration
	// In production, use proper metric key formatting
	return name
}

func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	metrics := make(map[string]interface{})
	
	// Add counters
	for key, value := range mc.counters {
		metrics[key] = value
	}
	
	// Add gauges
	for key, value := range mc.gauges {
		metrics[key] = value
	}
	
	// Add histogram summaries
	for key, hist := range mc.histograms {
		if hist.count > 0 {
			metrics[key+"_count"] = hist.count
			metrics[key+"_sum"] = hist.sum
			metrics[key+"_avg"] = hist.sum / float64(hist.count)
		}
	}
	
	// Add system metrics
	metrics["uptime_seconds"] = time.Since(mc.startTime).Seconds()
	
	return metrics
}

func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.counters = make(map[string]int64)
	mc.gauges = make(map[string]float64)
	mc.histograms = make(map[string]*Histogram)
}

func (mc *MetricsCollector) StartCollection() {
	go mc.collectSystemMetrics()
}

func (mc *MetricsCollector) collectSystemMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		mc.collectRuntimeMetrics()
	}
}

func (mc *MetricsCollector) collectRuntimeMetrics() {
	// Collect Go runtime metrics
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	
	labels := map[string]string{"type": "memory"}
	mc.SetGauge("go_mem_alloc", float64(stats.Alloc), labels)
	mc.SetGauge("go_mem_sys", float64(stats.Sys), labels)
	mc.SetGauge("go_mem_heap_alloc", float64(stats.HeapAlloc), labels)
	mc.SetGauge("go_goroutines", float64(runtime.NumGoroutine()), nil)
	mc.SetGauge("go_gc_cycles", float64(stats.NumGC), nil)
}