package metrics

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type MetricsExporter struct {
	collector *MetricsCollector
	logger    *zap.Logger
}

func NewMetricsExporter(collector *MetricsCollector, logger *zap.Logger) *MetricsExporter {
	return &MetricsExporter{
		collector: collector,
		logger:    logger,
	}
}

func (me *MetricsExporter) ExportMetrics() map[string]interface{} {
	metrics := me.collector.GetMetrics()
	
	// Add timestamp
	metrics["timestamp"] = time.Now().UTC()
	metrics["version"] = "1.0.0"
	
	return metrics
}

func (me *MetricsExporter) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	metrics := me.ExportMetrics()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (me *MetricsExporter) PrometheusHandler(w http.ResponseWriter, r *http.Request) {
	metrics := me.collector.GetMetrics()
	
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	
	for key, value := range metrics {
		switch v := value.(type) {
		case int64:
			me.writePrometheusMetric(w, key, float64(v))
		case float64:
			me.writePrometheusMetric(w, key, v)
		}
	}
}

func (me *MetricsExporter) writePrometheusMetric(w http.ResponseWriter, name string, value float64) {
	// Simple Prometheus format output
	// In production, use proper Prometheus client library
	metricLine := fmt.Sprintf("%s %f\n", name, value)
	w.Write([]byte(metricLine))
}

type MetricsReporter struct {
	exporters []MetricsExporter
	logger    *zap.Logger
}

func NewMetricsReporter(logger *zap.Logger) *MetricsReporter {
	return &MetricsReporter{
		exporters: make([]MetricsExporter, 0),
		logger:    logger,
	}
}

func (mr *MetricsReporter) AddExporter(exporter MetricsExporter) {
	mr.exporters = append(mr.exporters, exporter)
}

func (mr *MetricsReporter) Report() {
	metrics := make(map[string]interface{})
	
	for _, exporter := range mr.exporters {
		exporterMetrics := exporter.ExportMetrics()
		for k, v := range exporterMetrics {
			metrics[k] = v
		}
	}
	
	mr.logger.Info("Metrics reported", 
		zap.Int("metric_count", len(metrics)),
		zap.Time("timestamp", time.Now().UTC()),
	)
}
