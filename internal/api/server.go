package api
import (
        "context"
        "fmt"
        "net/http"
        "time"
        "github.com/gin-gonic/gin"
        "go.uber.org/zap"
        "github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/config"
        "github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/consensus"
        "github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/core"
        "github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/network"
        "github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/security"
        //"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type Server struct {
        engine           *core.Engine // Diubah dari *core.RelativisticEngine
        topologyManager  *network.TopologyManager
        timingManager    *consensus.TimingManager
        securityValidator *security.SecurityValidator
        healthMonitor    HealthChecker
        startTime        time.Time
        logger           *zap.Logger
        router           *gin.Engine
        httpServer       *http.Server
        config           *config.Config
	websocketManager *WebSocketManager

}
func NewServer(engine *core.Engine, topology *network.TopologyManager, timing *consensus.TimingManager, securityValidator *security.SecurityValidator, logger *zap.Logger, healthMonitor HealthChecker, websocketManager *WebSocketManager) *Server {
        server := &Server{
                engine:           engine,
                topologyManager:  topology,
                timingManager:    timing,
                securityValidator: securityValidator,
                healthMonitor:    healthMonitor,
                logger:           logger,
                startTime:        time.Now(),
		websocketManager: websocketManager,

        }
        server.setupRouter()
        return server
}
func (s *Server) setupRouter() {
        if s.config.Server.Environment == "production" {
                gin.SetMode(gin.ReleaseMode)
        }
        s.router = gin.New()
        s.router.Use(gin.Recovery())
        s.router.Use(s.loggingMiddleware())
        s.router.Use(s.securityMiddleware())
        s.router.Use(s.corsMiddleware())
        s.setupRoutes()
}
func (s *Server) Start(address string) error {
        s.httpServer = &http.Server{
                Addr:         address,
                Handler:      s.router,
                ReadTimeout:  15 * time.Second,
                WriteTimeout: 15 * time.Second,
                IdleTimeout:  60 * time.Second,
        }
        s.logger.Info("Starting HTTP server", zap.String("address", address))
        if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
                return fmt.Errorf("failed to start server: %w", err)
        }
        return nil
}
func (s *Server) Shutdown(ctx context.Context) error {
        s.logger.Info("Shutting down HTTP server")
        if s.httpServer != nil {
                if err := s.httpServer.Shutdown(ctx); err != nil {
                        return fmt.Errorf("failed to shutdown server: %w", err)
                }
        }
        s.logger.Info("HTTP server shutdown completed")
        return nil
}
func (s *Server) loggingMiddleware() gin.HandlerFunc {
        return func(c *gin.Context) {
                start := time.Now()
                path := c.Request.URL.Path
                query := c.Request.URL.RawQuery
                c.Next()
                end := time.Now()
                latency := end.Sub(start)
                fields := []zap.Field{
                        zap.Int("status", c.Writer.Status()),
                        zap.String("method", c.Request.Method),
                        zap.String("path", path),
                        zap.String("query", query),
                        zap.String("ip", c.ClientIP()),
                        zap.String("user-agent", c.Request.UserAgent()),
                        zap.Duration("latency", latency),
                }
                if len(c.Errors) > 0 {
                        s.logger.Error(c.Errors.String(), fields...)
                } else {
                        s.logger.Info("HTTP request", fields...)
                }
        }
}
func (s *Server) securityMiddleware() gin.HandlerFunc {
        return func(c *gin.Context) {
                c.Header("X-Content-Type-Options", "nosniff")
                c.Header("X-Frame-Options", "DENY")
                c.Header("X-XSS-Protection", "1; mode=block")
                c.Header("Strict-Transport-Security", "max-age=31536000")
                c.Next()
        }
}
func (s *Server) corsMiddleware() gin.HandlerFunc {
        return func(c *gin.Context) {
                c.Header("Access-Control-Allow-Origin", "*")
                c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
                if c.Request.Method == "OPTIONS" {
                        c.AbortWithStatus(204)
                        return
                }
                c.Next()
        }
}
func (s *Server) rootHandler(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
                "name":    "Relativistic Blockchain SDK",
                "version": "1.0.0",
                "status":  "running",
        })
}
