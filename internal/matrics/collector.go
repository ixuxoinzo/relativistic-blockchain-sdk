package metrics

import (
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

type MetricsCollector struct {
	logger     *zap.Logger
	mu         sync.RWMutex
	metrics    map[string]interface{}
	counters   map[string]int64
	gauges     map[string]float64
	histograms map[string]*Histogram
	startTime  time.Time

	// Prometheus metrics
	totalNodes           prometheus.Gauge
	activeNodes          prometheus.Gauge
	networkCoverage      prometheus.Gauge
	propagationDelay     prometheus.Histogram
	propagationCalculations prometheus.Counter
	validationSuccess    prometheus.Counter
	validationFailure    prometheus.Counter
	validationLatency    prometheus.Histogram
	requestsTotal        prometheus.Counter
	requestDuration      prometheus.Histogram
	errorCount           prometheus.Counter
}

type Histogram struct {
	values []float64
	sum    float64
	count  int64
}

func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	collector := &MetricsCollector{
		logger:     logger,
		metrics:    make(map[string]interface{}),
		counters:   make(map[string]int64),
		gauges:     make(map[string]float64),
		histograms: make(map[string]*Histogram),
		startTime:  time.Now(),
	}

	// Initialize Prometheus metrics
	collector.totalNodes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "relativistic_nodes_total",
		Help: "Total number of registered nodes",
	})
	collector.activeNodes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "relativistic_nodes_active",
		Help: "Number of active nodes",
	})
	collector.networkCoverage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "relativistic_network_coverage_ratio",
		Help: "Network coverage ratio (0-1)",
	})
	collector.propagationDelay = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "relativistic_propagation_delay_ms",
		Help:    "Propagation delay distribution in milliseconds",
		Buckets: prometheus.LinearBuckets(0, 50, 20),
	})
	collector.propagationCalculations = promauto.NewCounter(prometheus.CounterOpts{
		Name: "relativistic_propagation_calculations_total",
		Help: "Total number of propagation calculations",
	})
	collector.validationSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "relativistic_validation_success_total",
		Help: "Total successful validations",
	})
	collector.validationFailure = promauto.NewCounter(prometheus.CounterOpts{
		Name: "relativistic_validation_failure_total",
		Help: "Total failed validations",
	})
	collector.validationLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "relativistic_validation_latency_ms",
		Help:    "Validation latency distribution in milliseconds",
		Buckets: prometheus.LinearBuckets(0, 10, 20),
	})
	collector.requestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "relativistic_requests_total",
		Help: "Total number of HTTP requests",
	})
	collector.requestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "relativistic_request_duration_ms",
		Help:    "HTTP request duration distribution in milliseconds",
		Buckets: prometheus.LinearBuckets(0, 100, 20),
	})
	collector.errorCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "relativistic_errors_total",
		Help: "Total number of errors",
	})

	return collector
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
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	labels := map[string]string{"type": "memory"}
	mc.SetGauge("go_mem_alloc", float64(stats.Alloc), labels)
	mc.SetGauge("go_mem_sys", float64(stats.Sys), labels)
	mc.SetGauge("go_mem_heap_alloc", float64(stats.HeapAlloc), labels)
	mc.SetGauge("go_goroutines", float64(runtime.NumGoroutine()), nil)
	mc.SetGauge("go_gc_cycles", float64(stats.NumGC), nil)
}

// Prometheus-specific methods
func (mc *MetricsCollector) RecordNodeCount(total, active int, coverage float64) {
	mc.totalNodes.Set(float64(total))
	mc.activeNodes.Set(float64(active))
	mc.networkCoverage.Set(coverage)
}

func (mc *MetricsCollector) RecordPropagation(delay time.Duration) {
	mc.propagationDelay.Observe(float64(delay.Milliseconds()))
	mc.propagationCalculations.Inc()
}

func (mc *MetricsCollector) RecordValidation(success bool, latency time.Duration) {
	if success {
		mc.validationSuccess.Inc()
	} else {
		mc.validationFailure.Inc()
	}
	mc.validationLatency.Observe(float64(latency.Milliseconds()))
}

func (mc *MetricsCollector) RecordRequest(duration time.Duration) {
	mc.requestsTotal.Inc()
	mc.requestDuration.Observe(float64(duration.Milliseconds()))
}

func (mc *MetricsCollector) RecordError() {
	mc.errorCount.Inc()
}

// GetPrometheusMetrics returns a snapshot of current Prometheus metrics
func (mc *MetricsCollector) GetPrometheusMetrics() map[string]float64 {
	return map[string]float64{
		"nodes_total":       getGaugeValue(mc.totalNodes),
		"nodes_active":      getGaugeValue(mc.activeNodes),
		"network_coverage":  getGaugeValue(mc.networkCoverage),
		"requests_total":    getCounterValue(mc.requestsTotal),
		"errors_total":      getCounterValue(mc.errorCount),
		"validations_total": getCounterValue(mc.validationSuccess) + getCounterValue(mc.validationFailure),
	}
}

// Helper functions to get metric values
func getGaugeValue(gauge prometheus.Gauge) float64 {
	metric := &dto.Metric{}
	if err := gauge.Write(metric); err == nil && metric.Gauge != nil {
		return metric.Gauge.GetValue()
	}
	return 0
}

func getCounterValue(counter prometheus.Counter) float64 {
	metric := &dto.Metric{}
	if err := counter.Write(metric); err == nil && metric.Counter != nil {
		return metric.Counter.GetValue()
	}
	return 0
}