package api
import (
        "net/http"
        "runtime"
        "time"
        "github.com/gin-gonic/gin"
        "go.uber.org/zap"
       // "github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)
type HealthChecker interface {
        HealthCheck() *HealthStatus
        DeepHealthCheck() *DetailedHealthStatus
}
type HealthStatus struct {
        Status     string            `json:"status"`
        Timestamp  time.Time         `json:"timestamp"`
        Version    string            `json:"version"`
        Uptime     string            `json:"uptime"`
        Components map[string]string `json:"components"`
        System     *SystemInfo       `json:"system,omitempty"`
}
type SystemInfo struct {
        GoVersion    string       `json:"go_version"`
        NumCPU       int          `json:"num_cpu"`
        NumGoroutine int          `json:"num_goroutine"`
        Memory       *MemoryStats `json:"memory"`
}
type MemoryStats struct {
        Alloc      uint64 `json:"alloc"`
        TotalAlloc uint64 `json:"total_alloc"`
        Sys        uint64 `json:"sys"`
        NumGC      uint32 `json:"num_gc"`
}
type HealthMonitor struct {
        startTime time.Time
        checkers  map[string]HealthChecker
        logger    *zap.Logger
}
func NewHealthMonitor(logger *zap.Logger) *HealthMonitor {
        return &HealthMonitor{
                startTime: time.Now(),
                checkers:  make(map[string]HealthChecker),
                logger:    logger,
        }
}
func (hm *HealthMonitor) RegisterChecker(name string, checker HealthChecker) {
        hm.checkers[name] = checker
}
func (hm *HealthMonitor) HealthCheck() *HealthStatus {
        status := &HealthStatus{
                Status:     "healthy",
                Timestamp:  time.Now().UTC(),
                Version:    "1.0.0",
                Uptime:     time.Since(hm.startTime).String(),
                Components: make(map[string]string),
                System:     hm.getSystemInfo(),
        }
        for name, checker := range hm.checkers {
                componentStatus := checker.HealthCheck()
                status.Components[name] = componentStatus.Status
                if componentStatus.Status != "healthy" {
                        status.Status = "degraded"
                }
        }
        hm.checkCriticalComponents(status)
        return status
}
func (hm *HealthMonitor) checkCriticalComponents(status *HealthStatus) {
        criticalComponents := []string{"api", "engine", "database"}
        for _, component := range criticalComponents {
                if compStatus, exists := status.Components[component]; exists {
                        if compStatus != "healthy" {
                                status.Status = "unhealthy"
                                return
                        }
                } else {
                        status.Components[component] = "unknown"
                        status.Status = "degraded"
                }
        }
}
func (hm *HealthMonitor) getSystemInfo() *SystemInfo {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        return &SystemInfo{
                GoVersion:    runtime.Version(),
                NumCPU:       runtime.NumCPU(),
                NumGoroutine: runtime.NumGoroutine(),
                Memory: &MemoryStats{
                        Alloc:      m.Alloc,
                        TotalAlloc: m.TotalAlloc,
                        Sys:        m.Sys,
                        NumGC:      m.NumGC,
                },
        }
}
func (hm *HealthMonitor) DeepHealthCheck() *DetailedHealthStatus {
        basicHealth := hm.HealthCheck()
        detailed := &DetailedHealthStatus{
                HealthStatus: *basicHealth,
                Checks:       make([]*HealthCheck, 0),
        }
        for name, checker := range hm.checkers {
                check := &HealthCheck{
                        Name:      name,
                        Status:    "healthy",
                        Timestamp: time.Now().UTC(),
                }
                componentStatus := checker.HealthCheck()
                check.Status = componentStatus.Status
                check.Details = componentStatus.Components
                if componentStatus.System != nil {
                        check.Metrics = map[string]interface{}{
                                "memory_alloc": componentStatus.System.Memory.Alloc,
                                "goroutines":   componentStatus.System.NumGoroutine,
                        }
                }
                detailed.Checks = append(detailed.Checks, check)
        }
        return detailed
}
func (s *Server) healthHandler(c *gin.Context) {
        health := s.healthMonitor.HealthCheck()
        statusCode := http.StatusOK
        if health.Status == "unhealthy" {
                statusCode = http.StatusServiceUnavailable
        } else if health.Status == "degraded" {
                statusCode = http.StatusOK
        }
        c.JSON(statusCode, health)
}
func (s *Server) deepHealthHandler(c *gin.Context) {
        detailedHealth := s.healthMonitor.DeepHealthCheck()
        statusCode := http.StatusOK
        if detailedHealth.Status == "unhealthy" {
                statusCode = http.StatusServiceUnavailable
        }
        c.JSON(statusCode, detailedHealth)
}
func (s *Server) readyHandler(c *gin.Context) {
        health := s.healthMonitor.HealthCheck()
        if health.Status == "healthy" || health.Status == "degraded" {
                c.JSON(http.StatusOK, gin.H{
                        "status":    "ready",
                        "timestamp": time.Now().UTC(),
                })
        } else {
                c.JSON(http.StatusServiceUnavailable, gin.H{
                        "status":    "not_ready",
                        "timestamp": time.Now().UTC(),
                        "reason":    "Service is not healthy",
                })
        }
}
func (s *Server) liveHandler(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
                "status":    "live",
                "timestamp": time.Now().UTC(),
        })
}
// func (s *Server) metricsHandler(c *gin.Context) { DIHAPUS karena dideklarasikan ganda (di handlers.go)
//       ...
// }
func (s *Server) setupHealthRoutes() {
        health := s.router.Group("/health")
        {
                health.GET("", s.healthHandler)
                health.GET("/deep", s.deepHealthHandler)
                health.GET("/ready", s.readyHandler)
                health.GET("/live", s.liveHandler)
                health.GET("/metrics", s.metricsHandler)
        }
}
type DetailedHealthStatus struct {
        HealthStatus
        Checks []*HealthCheck `json:"checks"`
}
type HealthCheck struct {
        Name      string                 `json:"name"`
        Status    string                 `json:"status"`
        Timestamp time.Time              `json:"timestamp"`
        Details   map[string]string      `json:"details,omitempty"`
        Metrics   map[string]interface{} `json:"metrics,omitempty"`
        Error     string                 `json:"error,omitempty"`
}
type DatabaseHealthChecker struct {
        logger *zap.Logger
}
func (dhc *DatabaseHealthChecker) HealthCheck() *HealthStatus {
        status := &HealthStatus{
                Status:    "healthy",
                Timestamp: time.Now().UTC(),
                Components: map[string]string{
                        "connection": "healthy",
                        "queries":    "healthy",
                },
        }
        return status
}
type CacheHealthChecker struct {
        logger *zap.Logger
}
func (chc *CacheHealthChecker) HealthCheck() *HealthStatus {
        status := &HealthStatus{
                Status:    "healthy",
                Timestamp: time.Now().UTC(),
                Components: map[string]string{
                        "connection": "healthy",
                        "operations": "healthy",
                },
        }
        return status
}
