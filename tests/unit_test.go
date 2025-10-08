package tests

import (
	"testing"
	"time"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/core"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRelativisticEngine(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	engine := core.NewRelativisticEngine(nil, nil, logger)

	t.Run("CalculatePropagationDelay", func(t *testing.T) {
		nodeA := &types.Node{
			Position: types.Position{Latitude: 40.7128, Longitude: -74.0060, Altitude: 0},
		}
		nodeB := &types.Node{
			Position: types.Position{Latitude: 34.0522, Longitude: -118.2437, Altitude: 0},
		}

		delay, err := engine.CalculatePropagationDelay(nodeA, nodeB)
		assert.NoError(t, err)
		assert.Greater(t, delay, time.Duration(0))
	})

	t.Run("ValidateTimestamp", func(t *testing.T) {
		timestamp := time.Now().UTC().Add(-time.Second)
		position := types.Position{Latitude: 40.7128, Longitude: -74.0060, Altitude: 0}

		valid, result := engine.ValidateTimestamp(nil, timestamp, position, "test-node")
		assert.True(t, valid)
		assert.Greater(t, result.Confidence, 0.0)
	})

	t.Run("CalculateInterplanetaryDelay", func(t *testing.T) {
		delay, err := engine.CalculateInterplanetaryDelay("earth", "mars")
		assert.NoError(t, err)
		assert.Greater(t, delay, time.Duration(0))
	})
}

func TestValidationEngine(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	relativisticEngine := core.NewRelativisticEngine(nil, nil, logger)
	validationEngine := core.NewValidationEngine(relativisticEngine, logger)

	t.Run("ValidateBlock", func(t *testing.T) {
		block := &types.Block{
			Hash:       "test-block",
			Timestamp:  time.Now().UTC().Add(-time.Second),
			ProposedBy: "test-node",
			NodePosition: types.Position{
				Latitude:  40.7128,
				Longitude: -74.0060,
				Altitude:  0,
			},
		}

		result, err := validationEngine.ValidateBlockTimestamp(nil, block, "validator-node")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Valid)
	})
}
