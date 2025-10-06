package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

func (s *Server) metricsHandler(c *gin.Context) {
	metrics := s.engine.GetNetworkMetrics()
	c.JSON(http.StatusOK, metrics)
}

func (s *Server) networkStatusHandler(c *gin.Context) {
	nodes := s.topologyManager.GetAllNodes()
	
	status := gin.H{
		"total_nodes":   len(nodes),
		"active_nodes":  len(s.topologyManager.GetActiveNodes()),
		"regions":       s.getRegionDistribution(nodes),
		"timestamp":     time.Now().UTC(),
	}

	c.JSON(http.StatusOK, status)
}

func (s *Server) registerNodeHandler(c *gin.Context) {
	var request types.NodeRegistrationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		s.logger.Error("Invalid node registration request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	node := &types.Node{
		ID:       request.ID,
		Position: request.Position,
		Address:  request.Address,
		Metadata: types.Metadata{
			Region:      request.Region,
			Provider:    request.Provider,
			Version:     request.Version,
			Capabilities: request.Capabilities,
		},
		IsActive: true,
		LastSeen: time.Now().UTC(),
	}

	if err := s.topologyManager.AddNode(node); err != nil {
		s.logger.Error("Failed to register node", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Node registered successfully",
		"node_id": node.ID,
	})
}

func (s *Server) getNodesHandler(c *gin.Context) {
	nodes := s.topologyManager.GetAllNodes()
	c.JSON(http.StatusOK, nodes)
}

func (s *Server) getNodeHandler(c *gin.Context) {
	nodeID := c.Param("id")
	
	node, err := s.topologyManager.GetNode(nodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	c.JSON(http.StatusOK, node)
}

func (s *Server) updateNodePositionHandler(c *gin.Context) {
	nodeID := c.Param("id")
	
	var request struct {
		Position types.Position `json:"position"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := s.topologyManager.UpdateNodePosition(nodeID, request.Position); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Node position updated"})
}

func (s *Server) deleteNodeHandler(c *gin.Context) {
	nodeID := c.Param("id")
	
	if err := s.topologyManager.RemoveNode(nodeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Node removed"})
}

func (s *Server) calculatePropagationHandler(c *gin.Context) {
	var request struct {
		Source  string   `json:"source"`
		Targets []string `json:"targets"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	results, err := s.engine.CalculatePropagationPath(request.Source, request.Targets)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (s *Server) calculateInterplanetaryHandler(c *gin.Context) {
	var request struct {
		PlanetA string `json:"planet_a"`
		PlanetB string `json:"planet_b"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	delay, err := s.engine.CalculateInterplanetaryDelay(request.PlanetA, request.PlanetB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"planet_a": request.PlanetA,
		"planet_b": request.PlanetB,
		"delay":    delay.String(),
		"delay_ms": delay.Milliseconds(),
	})
}

func (s *Server) validateTimestampHandler(c *gin.Context) {
	var request struct {
		Timestamp   time.Time     `json:"timestamp"`
		Position    types.Position `json:"position"`
		OriginNode  string        `json:"origin_node"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	valid, result := s.engine.ValidateTimestamp(c.Request.Context(), request.Timestamp, request.Position, request.OriginNode)

	response := gin.H{
		"valid":       valid,
		"confidence":  result.Confidence,
		"reason":      result.Reason,
		"expected_delay": result.ExpectedDelay.String(),
		"actual_diff":   result.ActualDiff.String(),
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) getConsensusTimingHandler(c *gin.Context) {
	validatorNodes := c.QueryArray("validators")
	if len(validatorNodes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validators parameter required"})
		return
	}

	timing, err := s.timingManager.CalculateConsensusTiming(validatorNodes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, timing)
}

func (s *Server) getOffsetsHandler(c *gin.Context) {
	nodeID := c.Query("node_id")
	
	if nodeID != "" {
		offset, err := s.timingManager.GetNodeOffset(nodeID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Offset not found"})
			return
		}
		c.JSON(http.StatusOK, offset)
		return
	}

	offsets := s.timingManager.GetAllOffsets()
	c.JSON(http.StatusOK, offsets)
}

func (s *Server) validateConsensusHandler(c *gin.Context) {
	var request struct {
		Block      *types.Block `json:"block"`
		Validators []string     `json:"validators"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := s.timingManager.ValidateBlockConsensus(request.Block, request.Validators)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (s *Server) getRegionDistribution(nodes []*types.Node) map[string]int {
	distribution := make(map[string]int)
	for _, node := range nodes {
		distribution[node.Metadata.Region]++
	}
	return distribution
}