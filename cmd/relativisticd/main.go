package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/api"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/config"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/core"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/metrics"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/network"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Initialize metrics
	metricsCollector := metrics.NewCollector()
	go metricsCollector.Start(context.Background())

	// Initialize topology manager
	topology, err := network.NewTopologyManager(cfg.Redis.Address, logger)
	if err != nil {
		logger.Fatal("Failed to initialize topology manager", zap.Error(err))
	}

	// Initialize latency monitor
	latencyMonitor := network.NewLatencyMonitor(topology, logger)
	go latencyMonitor.StartMonitoring(context.Background())

	// Initialize relativistic engine
	relativisticEngine := core.NewRelativisticEngine(topology, latencyMonitor, logger)

	// Initialize API server
	server := api.NewServer(relativisticEngine, topology, logger)

	// Start server in goroutine
	go func() {
		logger.Info("Starting Relativistic Blockchain SDK server", 
			zap.String("address", cfg.Server.Address))
		if err := server.Start(cfg.Server.Address); err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.Shutdown(ctx)
	latencyMonitor.Stop()
	metricsCollector.Stop()

	logger.Info("Server stopped")
}
