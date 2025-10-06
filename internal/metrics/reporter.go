package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type MetricsReporter struct {
	collector  *MetricsCollector
	monitor    *MetricsMonitor
	analytics  *AnalyticsEngine
	prometheus *PrometheusMetrics
	logger     *zap.Logger
	mu         sync.RWMutex
	config     *ReporterConfig
}

type ReporterConfig struct {
	ReportInterval time.Duration
	Endpoint       string
	APIKey         string
	Enabled        bool
}

func NewMetricsReporter(collector *MetricsCollector, logger *zap.Logger) *MetricsReporter {
	return &MetricsReporter{
		collector: collector,
		logger:    logger,
		config: &ReporterConfig{
			ReportInterval: 60 * time.Second,
			Enabled:        true,
		},
	}
}

func (mr *MetricsReporter) SetMonitor(monitor *MetricsMonitor) {
	mr.monitor = monitor
}

func (mr *MetricsReporter) SetAnalytics(analytics *AnalyticsEngine) {
	mr.analytics = analytics
}

func (mr *MetricsReporter) SetPrometheus(prometheus *PrometheusMetrics) {
	mr.prometheus = prometheus
}

func (mr *MetricsReporter) SetConfig(config *ReporterConfig) {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.config = config
}

func (mr *MetricsReporter) StartReporting() {
	if !mr.config.Enabled {
		return
	}
	
	go mr.reportLoop()
}

func (mr *MetricsReporter) reportLoop() {
	ticker := time.NewTicker(mr.config.ReportInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		mr.reportMetrics()
	}
}

func (mr *MetricsReporter) reportMetrics() {
	metrics := mr.collector.GetMetrics()
	
	if mr.config.Endpoint != "" {
		mr.sendToEndpoint(metrics)
	}
	
	if mr.prometheus != nil {
		mr.prometheus.UpdateFromCollector()
	}
	
	mr.logger.Info("Metrics reported",
		zap.Int("metric_count", len(metrics)),
		zap.Time("timestamp", time.Now().UTC()),
	)
}

func (mr *MetricsReporter) sendToEndpoint(metrics map[string]interface{}) {
	payload := map[string]interface{}{
		"timestamp": time.Now().UTC(),
		"metrics":   metrics,
		"version":   "1.0.0",
	}
	
	if mr.analytics != nil {
		payload["analytics"] = mr.getAnalyticsSummary()
	}
	
	if mr.monitor != nil {
		payload["alerts"] = mr.monitor.GetActiveAlerts()
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		mr.logger.Error("Failed to marshal metrics", zap.Error(err))
		return
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", mr.config.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		mr.logger.Error("Failed to create metrics request", zap.Error(err))
		return
	}
	
	req.Header.Set("Content-Type", "application/json")
	if mr.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+mr.config.APIKey)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		mr.logger.Error("Failed to send metrics", zap.Error(err))
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		mr.logger.Warn("Metrics endpoint returned error",
			zap.Int("status_code", resp.StatusCode),
		)
	}
}

func (mr *MetricsReporter) getAnalyticsSummary() map[string]interface{} {
	if mr.analytics == nil {
		return nil
	}
	
	summary := make(map[string]interface{})
	
	mr.analytics.mu.RLock()
	defer mr.analytics.mu.RUnlock()
	
	for name, analytic := range mr.analytics.analytics {
		if analytic.Summary != nil {
			summary[name] = map[string]interface{}{
				"count":   analytic.Summary.Count,
				"average": analytic.Summary.Average,
				"min":     analytic.Summary.Min,
				"max":     analytic.Summary.Max,
				"std_dev": analytic.Summary.StdDev,
				"trend":   mr.analytics.CalculateTrend(name),
			}
		}
	}
	
	return summary
}

func (mr *MetricsReporter) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	metrics := mr.collector.GetMetrics()
	response := map[string]interface{}{
		"timestamp": time.Now().UTC(),
		"metrics":   metrics,
	}
	
	if mr.analytics != nil {
		response["analytics"] = mr.getAnalyticsSummary()
	}
	
	if mr.monitor != nil {
		response["alerts"] = mr.monitor.GetActiveAlerts()
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (mr *MetricsReporter) PrometheusHandler(w http.ResponseWriter, r *http.Request) {
	if mr.prometheus != nil {
		mr.prometheus.HTTPHandler(w, r)
	} else {
		http.Error(w, "Prometheus metrics not enabled", http.StatusNotFound)
	}
}
