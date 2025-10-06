package metrics

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type MetricsMonitor struct {
	collector  *MetricsCollector
	exporters  []MetricsExporter
	logger     *zap.Logger
	mu         sync.RWMutex
	alerts     map[string]*Alert
	thresholds map[string]float64
}

type Alert struct {
	ID        string
	Metric    string
	Threshold float64
	Current   float64
	Severity  string
	Message   string
	Timestamp time.Time
	Active    bool
}

func NewMetricsMonitor(collector *MetricsCollector, logger *zap.Logger) *MetricsMonitor {
	return &MetricsMonitor{
		collector:  collector,
		exporters:  make([]MetricsExporter, 0),
		logger:     logger,
		alerts:     make(map[string]*Alert),
		thresholds: make(map[string]float64),
	}
}

func (mm *MetricsMonitor) AddExporter(exporter MetricsExporter) {
	mm.exporters = append(mm.exporters, exporter)
}

func (mm *MetricsMonitor) SetThreshold(metric string, threshold float64) {
	mm.thresholds[metric] = threshold
}

func (mm *MetricsMonitor) CheckThresholds() {
	metrics := mm.collector.GetMetrics()
	
	for metric, threshold := range mm.thresholds {
		if value, exists := metrics[metric]; exists {
			var current float64
			switch v := value.(type) {
			case int64:
				current = float64(v)
			case float64:
				current = v
			default:
				continue
			}
			
			if current > threshold {
				mm.triggerAlert(metric, current, threshold)
			} else {
				mm.resolveAlert(metric)
			}
		}
	}
}

func (mm *MetricsMonitor) triggerAlert(metric string, current, threshold float64) {
	alertID := metric + "_threshold_exceeded"
	
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	if alert, exists := mm.alerts[alertID]; exists {
		alert.Current = current
		alert.Timestamp = time.Now().UTC()
	} else {
		mm.alerts[alertID] = &Alert{
			ID:        alertID,
			Metric:    metric,
			Threshold: threshold,
			Current:   current,
			Severity:  "warning",
			Message:   fmt.Sprintf("Metric %s exceeded threshold: %f > %f", metric, current, threshold),
			Timestamp: time.Now().UTC(),
			Active:    true,
		}
		
		mm.logger.Warn("Metric threshold exceeded",
			zap.String("metric", metric),
			zap.Float64("current", current),
			zap.Float64("threshold", threshold),
		)
	}
}

func (mm *MetricsMonitor) resolveAlert(metric string) {
	alertID := metric + "_threshold_exceeded"
	
	mm.mu.Lock()
	defer mm.mu.Unlock()
	
	if alert, exists := mm.alerts[alertID]; exists && alert.Active {
		alert.Active = false
		mm.logger.Info("Metric alert resolved",
			zap.String("metric", metric),
		)
	}
}

func (mm *MetricsMonitor) GetActiveAlerts() []*Alert {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	
	var activeAlerts []*Alert
	for _, alert := range mm.alerts {
		if alert.Active {
			activeAlerts = append(activeAlerts, alert)
		}
	}
	return activeAlerts
}

func (mm *MetricsMonitor) StartMonitoring(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for range ticker.C {
		mm.CheckThresholds()
		mm.ReportMetrics()
	}
}

func (mm *MetricsMonitor) ReportMetrics() {
	for _, exporter := range mm.exporters {
		exporter.ExportMetrics()
	}
}