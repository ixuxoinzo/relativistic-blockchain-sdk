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
func TestPerformance(t *testing.T) {
        logger, _ := zap.NewProduction()
        topology := setupPerformanceTopology(t, logger)
        engine := core.NewRelativisticEngine(topology, nil, logger)
        tests := []struct {
                name     string
                testFunc func(*testing.T)
        }{
                {"PropagationCalculation", func(t *testing.T) { testPropagationPerformance(t, engine) }},
                {"TimestampValidation", func(t *testing.T) { testValidationPerformance(t, engine) }},
                {"BatchOperations", func(t *testing.T) { testBatchPerformance(t, engine) }},
        }
        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        start := time.Now()
                        tt.testFunc(t)
                        duration := time.Since(start)
                        if duration > 5*time.Second {
                                t.Errorf("Performance test %s took too long: %v", tt.name, duration)
                        }
                })
        }
}
func testPropagationPerformance(t *testing.T, engine *core.RelativisticEngine) {
        nodes := generatePerformanceNodes(50)
        for i := 0; i < 1000; i++ {
                _, err := engine.CalculatePropagationDelay(nodes[0], nodes[1])
                if err != nil {
                        t.Fatalf("Propagation calculation failed: %v", err)
                }
        }
}
func testValidationPerformance(t *testing.T, engine *core.RelativisticEngine) {
        block := &types.Block{
                Hash:       "performance-block",
                Timestamp:  time.Now().UTC(),
                ProposedBy: "node-0",
                NodePosition: types.Position{
                        Latitude:  40.7128,
                        Longitude: -74.0060,
                        Altitude:  0,
                },
        }
        for i := 0; i < 1000; i++ {
                _, _ = engine.ValidateTimestamp(nil, block.Timestamp, block.NodePosition, "node-1")
        }
}
func testBatchPerformance(t *testing.T, engine *core.RelativisticEngine) {
        nodes := generatePerformanceNodes(100)
        _, err := engine.BatchCalculateDelays(nodes)
        if err != nil {
                t.Fatalf("Batch calculation failed: %v", err)
        }
}
func setupPerformanceTopology(t *testing.T, logger *zap.Logger) *network.TopologyManager {
        topology, err := network.NewTopologyManager("localhost:6379", logger)
        if err != nil {
                t.Fatalf("Failed to create topology manager: %v", err)
        }
        return topology
}
func generatePerformanceNodes(count int) []*types.Node {
        nodes := make([]*types.Node, count)
        for i := 0; i < count; i++ {
                nodes[i] = &types.Node{
                        ID:       "node-" + strconv.Itoa(i),
                        Position: types.Position{Latitude: float64(i), Longitude: float64(i), Altitude: 0},
                        Address:  "node.example.com:8080",
                        Metadata: types.Metadata{Region: "test", Provider: "test", Version: "1.0.0"},
                        IsActive: true,
                        LastSeen: time.Now().UTC(),
                }
        }
        return nodes
}
