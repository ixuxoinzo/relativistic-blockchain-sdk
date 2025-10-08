package tests

import (
	"testing"
	"time"
	"strconv"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/core"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/network"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
	"go.uber.org/zap"
)

func BenchmarkPropagationCalculation(b *testing.B) {
	logger, _ := zap.NewProduction()
	topology := setupBenchmarkTopology(b, logger)
	engine := core.NewRelativisticEngine(topology, nil, logger)

	nodes := generateTestNodes(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.BatchCalculateDelays(nodes)
		if err != nil {
			b.Fatalf("Batch calculation failed: %v", err)
		}
	}
}

func BenchmarkTimestampValidation(b *testing.B) {
	logger, _ := zap.NewProduction()
	topology := setupBenchmarkTopology(b, logger)
	engine := core.NewRelativisticEngine(topology, nil, logger)

	block := &types.Block{
		Hash:       "benchmark-block",
		Timestamp:  time.Now().UTC(),
		ProposedBy: "node-1",
		NodePosition: types.Position{
			Latitude:  40.7128,
			Longitude: -74.0060,
			Altitude:  0,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.ValidateTimestamp(nil, block.Timestamp, block.NodePosition, "node-2")
	}
}

func setupBenchmarkTopology(b *testing.B, logger *zap.Logger) *network.TopologyManager {
	topology, err := network.NewTopologyManager("localhost:6379", logger)
	if err != nil {
		b.Fatalf("Failed to create topology manager: %v", err)
	}
	return topology
}

func generateTestNodes(count int) []*types.Node {
	nodes := make([]*types.Node, count)
	for i := 0; i < count; i++ {
		nodes[i] = &types.Node{
			ID:      "node-" + strconv.Itoa(i),
			Position: types.Position{Latitude: float64(i), Longitude: float64(i), Altitude: 0},
			Address:  "node.example.com:8080",
			Metadata: types.Metadata{Region: "test", Provider: "test", Version: "1.0.0"},
			IsActive: true,
			LastSeen: time.Now().UTC(),
		}
	}
	return nodes
}
