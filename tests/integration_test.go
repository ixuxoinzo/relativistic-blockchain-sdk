package tests
import (
        "testing"
        "time"
        "github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/core"
        "github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/network"
        "github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
        "github.com/stretchr/testify/assert"
        "go.uber.org/zap"
)
func TestIntegration(t *testing.T) {
        logger, _ := zap.NewDevelopment()
        topology := setupTestTopology(t, logger)
        
        latencyMonitor := network.NewLatencyMonitor(topology, logger)
        engine := core.NewEngine(topology, latencyMonitor, logger)

        t.Run("NodeRegistrationAndPropagation", func(t *testing.T) {
                node1 := &types.Node{
                        ID: "node-1",
                        Position: types.Position{
                                Latitude:  40.7128,
                                Longitude: -74.0060,
                                Altitude:  0,
                        },
                        Address: "node1.example.com:8080",
                        Metadata: types.Metadata{
                                Region:   "us-east",
                                Provider: "aws",
                                Version:  "1.0.0",
                        },
                        IsActive: true,
                        LastSeen: time.Now().UTC(),
                }
                node2 := &types.Node{
                        ID: "node-2",
                        Position: types.Position{
                                Latitude:  34.0522,
                                Longitude: -118.2437,
                                Altitude:  0,
                        },
                        Address: "node2.example.com:8080",
                        Metadata: types.Metadata{
                                Region:   "us-west",
                                Provider: "aws",
                                Version:  "1.0.0",
                        },
                        IsActive: true,
                        LastSeen: time.Now().UTC(),
                }
                err := topology.AddNode(node1)
                assert.NoError(t, err)
                err = topology.AddNode(node2)
                assert.NoError(t, err)
                results, err := engine.CalculatePropagationPath("node-1", []string{"node-2"})
                assert.NoError(t, err)
                assert.NotNil(t, results["node-2"])
                assert.True(t, results["node-2"].Success)
                assert.Greater(t, results["node-2"].TheoreticalDelay, time.Duration(0))
        })
        t.Run("TimestampValidation", func(t *testing.T) {
                block := &types.Block{
                        Hash:       "test-block",
                        Timestamp:  time.Now().UTC().Add(-time.Second),
                        ProposedBy: "node-1",
                        NodePosition: types.Position{
                                Latitude:  40.7128,
                                Longitude: -74.0060,
                                Altitude:  0,
                        },
                }
                valid, result := engine.ValidateTimestamp(nil, block.Timestamp, block.NodePosition, "node-2")
                assert.True(t, valid)
                assert.Greater(t, result.Confidence, 0.5)
        })
}
func setupTestTopology(t *testing.T, logger *zap.Logger) *network.TopologyManager {
        topology, err := network.NewTopologyManager("localhost:6379", logger)
        if err != nil {
                t.Fatalf("Failed to create topology manager: %v", err)
        }
        return topology
}
