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
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/consensus"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/metrics"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/network"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/security"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting Relativistic Blockchain SDK")

	topology, err := network.NewTopologyManager(cfg.Redis.Address, logger)
	if err != nil {
		logger.Fatal("Failed to initialize topology manager", zap.Error(err))
	}

	latencyMonitor := network.NewLatencyMonitor(topology, logger)
	relativisticEngine := core.NewRelativisticEngine(topology, latencyMonitor, logger)
	timingManager := consensus.NewTimingManager(topology, logger)
	securityValidator := security.NewValidator(cfg.Security, logger)
	metricsCollector := metrics.NewCollector()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go latencyMonitor.StartMonitoring(ctx)
	go metricsCollector.Start(ctx)
	go securityValidator.StartCleanup(ctx)

	server := api.NewServer(relativisticEngine, topology, timingManager, securityValidator, logger)

	go func() {
		logger.Info("Starting API server", zap.String("address", cfg.Server.Address))
		if err := server.Start(cfg.Server.Address); err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	
	sig := <-quit
	logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	logger.Info("Initiating graceful shutdown")

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	latencyMonitor.Stop()
	metricsCollector.Stop()
	securityValidator.Stop()
	topology.Close()

	logger.Info("Relativistic Blockchain SDK stopped gracefully")
}