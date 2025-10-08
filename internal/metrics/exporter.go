package metrics

import (
	"encoding/json"
	"fmt"
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
	metricLine := fmt.Sprintf("%s %f\n", name, value)
	w.Write([]byte(metricLine))
}
