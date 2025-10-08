package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type PrometheusMetrics struct {
	collector *MetricsCollector
	logger    *zap.Logger
	mu        sync.RWMutex
	registry  map[string]*PrometheusMetric
}

type PrometheusMetric struct {
	Name   string
	Type   string
	Help   string
	Labels map[string]string
	Value  float64
}

func NewPrometheusMetrics(collector *MetricsCollector, logger *zap.Logger) *PrometheusMetrics {
	return &PrometheusMetrics{
		collector: collector,
		logger:    logger,
		registry:  make(map[string]*PrometheusMetric),
	}
}

func (pm *PrometheusMetrics) RegisterCounter(name, help string, labels map[string]string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.registry[name] = &PrometheusMetric{
		Name:   name,
		Type:   "counter",
		Help:   help,
		Labels: labels,
		Value:  0,
	}
}

func (pm *PrometheusMetrics) RegisterGauge(name, help string, labels map[string]string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.registry[name] = &PrometheusMetric{
		Name:   name,
		Type:   "gauge",
		Help:   help,
		Labels: labels,
		Value:  0,
	}
}

func (pm *PrometheusMetrics) RegisterHistogram(name, help string, labels map[string]string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.registry[name] = &PrometheusMetric{
		Name:   name,
		Type:   "histogram",
		Help:   help,
		Labels: labels,
		Value:  0,
	}
}

func (pm *PrometheusMetrics) SetCounterValue(name string, value float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if metric, exists := pm.registry[name]; exists {
		metric.Value = value
	}
}

func (pm *PrometheusMetrics) SetGaugeValue(name string, value float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if metric, exists := pm.registry[name]; exists {
		metric.Value = value
	}
}

func (pm *PrometheusMetrics) GenerateMetrics() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var output string

	for _, metric := range pm.registry {
		if metric.Help != "" {
			output += fmt.Sprintf("# HELP %s %s\n", metric.Name, metric.Help)
			output += fmt.Sprintf("# TYPE %s %s\n", metric.Name, metric.Type)
		}

		labels := ""
		if len(metric.Labels) > 0 {
			labels = "{"
			for k, v := range metric.Labels {
				labels += fmt.Sprintf("%s=\"%s\",", k, v)
			}
			labels = labels[:len(labels)-1] + "}"
		}

		output += fmt.Sprintf("%s%s %f\n", metric.Name, labels, metric.Value)
	}

	return output
}

func (pm *PrometheusMetrics) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	metrics := pm.GenerateMetrics()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.Write([]byte(metrics))
}

func (pm *PrometheusMetrics) UpdateFromCollector() {
	metrics := pm.collector.GetMetrics()

	for name, value := range metrics {
		switch v := value.(type) {
		case int64:
			pm.SetGaugeValue(name, float64(v))
		case float64:
			pm.SetGaugeValue(name, v)
		}
	}
}

func (pm *PrometheusMetrics) StartAutoUpdate(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		pm.UpdateFromCollector()
	}
}
