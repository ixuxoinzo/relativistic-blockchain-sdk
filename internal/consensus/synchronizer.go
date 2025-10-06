package consensus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Synchronizer struct {
	offsetManager *OffsetManager
	logger        *zap.Logger
	mu            sync.RWMutex
	syncStatus    map[string]*SyncStatus
	stopChan      chan struct{}
}

type SyncStatus struct {
	NodeID         string        `json:"node_id"`
	LastSync       time.Time     `json:"last_sync"`
	SyncCount      int           `json:"sync_count"`
	LastOffset     time.Duration `json:"last_offset"`
	AverageOffset  time.Duration `json:"average_offset"`
	Status         string        `json:"status"`
	LastError      string        `json:"last_error,omitempty"`
}

func NewSynchronizer(offsetManager *OffsetManager, logger *zap.Logger) *Synchronizer {
	return &Synchronizer{
		offsetManager: offsetManager,
		logger:        logger,
		syncStatus:    make(map[string]*SyncStatus),
		stopChan:      make(chan struct{}),
	}
}

func (s *Synchronizer) Start(ctx context.Context) error {
	s.logger.Info("Starting Synchronizer")

	go s.backgroundSync(ctx)

	s.logger.Info("Synchronizer started successfully")
	return nil
}

func (s *Synchronizer) Stop() {
	close(s.stopChan)
	s.logger.Info("Synchronizer stopped")
}

func (s *Synchronizer) backgroundSync(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.SyncAllNodes()
		}
	}
}

func (s *Synchronizer) SyncAllNodes() {
	s.logger.Debug("Starting synchronization of all nodes")

	nodes := s.getAllNodeIDs()
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make(map[string]error)

	for _, nodeID := range nodes {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			err := s.SyncNode(id)
			
			mu.Lock()
			results[id] = err
			mu.Unlock()
		}(nodeID)
	}

	wg.Wait()

	successCount := 0
	for _, err := range results {
		if err == nil {
			successCount++
		}
	}

	s.logger.Info("Batch synchronization completed",
		zap.Int("total_nodes", len(nodes)),
		zap.Int("successful", successCount),
		zap.Int("failed", len(nodes)-successCount),
	)
}

func (s *Synchronizer) SyncNode(nodeID string) error {
	s.mu.Lock()
	status, exists := s.syncStatus[nodeID]
	if !exists {
		status = &SyncStatus{
			NodeID: nodeID,
			Status: "pending",
		}
		s.syncStatus[nodeID] = status
	}
	s.mu.Unlock()

	allNodes := s.getAllNodeIDs()
	offset, err := s.offsetManager.CalculateNodeOffset(nodeID, allNodes)
	if err != nil {
		s.updateSyncStatus(nodeID, "failed", 0, err.Error())
		return fmt.Errorf("failed to calculate offset for node %s: %w", nodeID, err)
	}

	s.updateSyncStatus(nodeID, "synced", offset.Offset, "")

	s.logger.Debug("Node synchronization completed",
		zap.String("node_id", nodeID),
		zap.Duration("offset", offset.Offset),
		zap.Float64("confidence", offset.Confidence),
	)

	return nil
}

func (s *Synchronizer) updateSyncStatus(nodeID string, status string, offset time.Duration, errorMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	syncStatus, exists := s.syncStatus[nodeID]
	if !exists {
		syncStatus = &SyncStatus{
			NodeID: nodeID,
		}
		s.syncStatus[nodeID] = syncStatus
	}

	syncStatus.LastSync = time.Now().UTC()
	syncStatus.Status = status
	syncStatus.LastError = errorMsg

	if offset != 0 {
		syncStatus.LastOffset = offset
		syncStatus.SyncCount++
		
		if syncStatus.SyncCount == 1 {
			syncStatus.AverageOffset = offset
		} else {
			syncStatus.AverageOffset = time.Duration(
				(float64(syncStatus.AverageOffset)*float64(syncStatus.SyncCount-1) + float64(offset)) / float64(syncStatus.SyncCount),
			)
		}
	}
}

func (s *Synchronizer) getAllNodeIDs() []string {
	return []string{"node1", "node2", "node3"}
}

func (s *Synchronizer) GetSyncStatus(nodeID string) *SyncStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.syncStatus[nodeID]
}

func (s *Synchronizer) GetAllSyncStatus() map[string]*SyncStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := make(map[string]*SyncStatus)
	for k, v := range s.syncStatus {
		status[k] = v
	}
	return status
}

func (s *Synchronizer) GetSyncStats() *SyncStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &SyncStats{
		TotalNodes:    len(s.syncStatus),
		Timestamp:     time.Now().UTC(),
		StatusCounts:  make(map[string]int),
	}

	var totalOffset time.Duration
	nodeCount := 0

	for _, status := range s.syncStatus {
		stats.StatusCounts[status.Status]++

		if status.Status == "synced" {
			totalOffset += status.AverageOffset
			nodeCount++
		}
	}

	if nodeCount > 0 {
		stats.AverageOffset = totalOffset / time.Duration(nodeCount)
	}

	stats.SyncedNodes = stats.StatusCounts["synced"]
	stats.FailedNodes = stats.StatusCounts["failed"]
	stats.PendingNodes = stats.StatusCounts["pending"]

	return stats
}

type SyncStats struct {
	TotalNodes    int                    `json:"total_nodes"`
	SyncedNodes   int                    `json:"synced_nodes"`
	FailedNodes   int                    `json:"failed_nodes"`
	PendingNodes  int                    `json:"pending_nodes"`
	AverageOffset time.Duration          `json:"average_offset"`
	StatusCounts  map[string]int         `json:"status_counts"`
	Timestamp     time.Time              `json:"timestamp"`
}

func (s *Synchronizer) CleanupStaleStatus() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	staleThreshold := time.Now().Add(-24 * time.Hour)
	removedCount := 0

	for nodeID, status := range s.syncStatus {
		if status.LastSync.Before(staleThreshold) {
			delete(s.syncStatus, nodeID)
			removedCount++
		}
	}

	if removedCount > 0 {
		s.logger.Info("Cleaned up stale sync status",
			zap.Int("count", removedCount),
		)
	}

	return removedCount
}