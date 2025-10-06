package tests

import (
	"sync"
	"testing"
	"time"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/core"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/network"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
	"go.uber.org/zap"
)

func TestLoad(t *testing.T) {
	logger, _ := zap.NewProduction()
	topology := setupLoadTestTopology(t, logger)
	engine := core.NewRelativisticEngine(topology, nil, logger)

	concurrentUsers := 100
	requestsPerUser := 100

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			for j := 0; j < requestsPerUser; j++ {
				performLoadTestRequest(t, engine, userID, j)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	totalRequests := concurrentUsers * requestsPerUser
	throughput := float64(totalRequests) / duration.Seconds()

	t.Logf("Load test completed:")
	t.Logf("  Concurrent users: %d", concurrentUsers)
	t.Logf("  Requests per user: %d", requestsPerUser)
	t.Logf("  Total requests: %d", totalRequests)
	t.Logf("  Total duration: %v", duration)
	t.Logf("  Throughput: %.2f requests/second", throughput)

	if throughput < 100 {
		t.Errorf("Throughput too low: %.2f requests/second", throughput)
	}
}

func performLoadTestRequest(t *testing.T, engine *core.RelativisticEngine, userID, requestID int) {
	block := &types.Block{
		Hash:      string(userID) + "-" + string(requestID),
		Timestamp: time.Now().UTC(),
		ProposedBy: "node-" + string(userID),
		NodePosition: types.Position{
			Latitude:  float64(userID),
			Longitude: float64(requestID),
			Altitude:  0,
		},
	}

	valid, result := engine.ValidateTimestamp(nil, block.Timestamp, block.NodePosition, "node-0")
	if !valid {
		t.Errorf("Validation failed for request %d-%d: %v", userID, requestID, result.Reason)
	}
}

func setupLoadTestTopology(t *testing.T, logger *zap.Logger) *network.TopologyManager {
	topology, err := network.NewTopologyManager("localhost:6379", logger)
	if err != nil {
		t.Fatalf("Failed to create topology manager: %v", err)
	}

	for i := 0; i < 10; i++ {
		node := &types.Node{
			ID:       "node-" + string(i),
			Position: types.Position{Latitude: float64(i), Longitude: float64(i), Altitude: 0},
			Address:  "node.example.com:8080",
			Metadata: types.Metadata{Region: "test", Provider: "test", Version: "1.0.0"},
			IsActive: true,
			LastSeen: time.Now().UTC(),
		}
		topology.AddNode(node)
	}

	return topology
}