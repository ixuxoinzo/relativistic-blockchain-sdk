package api

import (
	"github.com/gin-gonic/gin"
)

func (s *Server) setupRoutes() {
	api := s.router.Group("/api/v1")
	
	api.Use(s.rateLimitMiddleware())
	api.Use(s.timeoutMiddleware(30 * time.Second))
	api.Use(s.requestSizeMiddleware(10 * 1024 * 1024))

	api.GET("/health", s.healthHandler)
	api.GET("/metrics", s.metricsHandler)
	api.GET("/status", s.statusHandler)

	nodes := api.Group("/nodes")
	{
		nodes.GET("", s.getNodesHandler)
		nodes.GET("/:id", s.validateNodeMiddleware(), s.getNodeHandler)
		nodes.POST("", s.authMiddleware(), s.registerNodeHandler)
		nodes.PUT("/:id/position", s.authMiddleware(), s.validateNodeMiddleware(), s.updateNodePositionHandler)
		nodes.DELETE("/:id", s.authMiddleware(), s.adminOnlyMiddleware(), s.validateNodeMiddleware(), s.deleteNodeHandler)
	}

	calculations := api.Group("/calculations")
	{
		calculations.POST("/propagation", s.calculatePropagationHandler)
		calculations.POST("/interplanetary", s.calculateInterplanetaryHandler)
		calculations.POST("/batch", s.batchCalculationHandler)
	}

	validation := api.Group("/validation")
	{
		validation.POST("/timestamp", s.validateTimestampHandler)
		validation.POST("/block", s.validateBlockHandler)
		validation.POST("/batch", s.batchValidationHandler)
	}

	consensus := api.Group("/consensus")
	{
		consensus.GET("/timing", s.getConsensusTimingHandler)
		consensus.GET("/offsets", s.getOffsetsHandler)
		consensus.POST("/validate", s.validateConsensusHandler)
		consensus.GET("/health", s.consensusHealthHandler)
	}

	network := api.Group("/network")
	{
		network.GET("/topology", s.getTopologyHandler)
		network.GET("/latency", s.getLatencyHandler)
		network.GET("/peers", s.getPeersHandler)
		network.GET("/alerts", s.getAlertsHandler)
	}

	admin := api.Group("/admin")
	admin.Use(s.authMiddleware())
	admin.Use(s.adminOnlyMiddleware())
	{
		admin.GET("/stats", s.adminStatsHandler)
		admin.POST("/cache/clear", s.clearCacheHandler)
		admin.GET("/logs", s.getLogsHandler)
		admin.POST("/maintenance", s.maintenanceHandler)
	}

	ws := api.Group("/ws")
	{
		ws.GET("", s.websocketHandler)
	}

	s.router.Static("/docs", "./docs")
	s.router.Static("/static", "./static")
	
	s.router.GET("/", s.rootHandler)
	s.router.NoRoute(s.notFoundHandler)
}

func (s *Server) statusHandler(c *gin.Context) {
	status := gin.H{
		"status":    "operational",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC(),
		"services": gin.H{
			"api":       "running",
			"engine":    "running",
			"network":   "running",
			"consensus": "running",
		},
	}

	c.JSON(http.StatusOK, status)
}

func (s *Server) batchCalculationHandler(c *gin.Context) {
	var request struct {
		Nodes []string `json:"nodes"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	nodes := make([]*types.Node, len(request.Nodes))
	for i, nodeID := range request.Nodes {
		node, err := s.topologyManager.GetNode(nodeID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Node not found: %s", nodeID)})
			return
		}
		nodes[i] = node
	}

	results, err := s.engine.BatchCalculateDelays(nodes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (s *Server) validateBlockHandler(c *gin.Context) {
	var request struct {
		Block      *types.Block `json:"block"`
		OriginNode string       `json:"origin_node"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := s.engine.ValidateBlockTimestamp(c.Request.Context(), request.Block, request.OriginNode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (s *Server) batchValidationHandler(c *gin.Context) {
	var request struct {
		Items      []*types.ValidatableItem `json:"items"`
		OriginNode string                   `json:"origin_node"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	results, err := s.engine.BatchValidateTimestamps(c.Request.Context(), request.Items, request.OriginNode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (s *Server) consensusHealthHandler(c *gin.Context) {
	validatorNodes := c.QueryArray("validators")
	if len(validatorNodes) == 0 {
		nodes := s.topologyManager.GetAllNodes()
		validatorNodes = make([]string, len(nodes))
		for i, node := range nodes {
			validatorNodes[i] = node.ID
		}
	}

	health := s.timingManager.CheckConsensusHealth(validatorNodes)
	c.JSON(http.StatusOK, health)
}

func (s *Server) getTopologyHandler(c *gin.Context) {
	graph := s.topologyManager.GetTopologyGraph()
	c.JSON(http.StatusOK, graph)
}

func (s *Server) getLatencyHandler(c *gin.Context) {
	measurements := s.topologyManager.GetAllLatencyMeasurements()
	c.JSON(http.StatusOK, measurements)
}

func (s *Server) getPeersHandler(c *gin.Context) {
	peers := s.topologyManager.GetAllPeers()
	c.JSON(http.StatusOK, peers)
}

func (s *Server) getAlertsHandler(c *gin.Context) {
	includeAcknowledged := c.Query("include_acknowledged") == "true"
	alerts := s.topologyManager.GetAlerts(includeAcknowledged)
	c.JSON(http.StatusOK, alerts)
}

func (s *Server) adminStatsHandler(c *gin.Context) {
	stats := gin.H{
		"engine_metrics": s.engine.GetEngineMetrics(),
		"network_metrics": s.topologyManager.GetNetworkMetrics(),
		"consensus_stats": s.timingManager.GetConsensusStats(),
		"timestamp": time.Now().UTC(),
	}

	c.JSON(http.StatusOK, stats)
}

func (s *Server) clearCacheHandler(c *gin.Context) {
	s.engine.ClearCache()
	s.timingManager.ClearCache()
	
	c.JSON(http.StatusOK, gin.H{"message": "Cache cleared successfully"})
}

func (s *Server) getLogsHandler(c *gin.Context) {
	level := c.Query("level")
	lines := c.DefaultQuery("lines", "100")
	
	logs := s.getLogEntries(level, lines)
	c.JSON(http.StatusOK, logs)
}

func (s *Server) maintenanceHandler(c *gin.Context) {
	var request struct {
		Action string `json:"action"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	switch request.Action {
	case "recalculate_offsets":
		s.timingManager.RecalculateAllOffsets()
		c.JSON(http.StatusOK, gin.H{"message": "Offsets recalculated"})
	case "sync_nodes":
		s.timingManager.SyncAllNodes()
		c.JSON(http.StatusOK, gin.H{"message": "Nodes synchronized"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown action"})
	}
}

func (s *Server) notFoundHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error": "Endpoint not found",
		"path":  c.Request.URL.Path,
	})
}

func (s *Server) getLogEntries(level string, lines string) []string {
	return []string{"Log retrieval not implemented"}
}