package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIEndpoint struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
	AuthRequired bool   `json:"auth_required"`
	AdminRequired bool  `json:"admin_required"`
}

type APIDocumentation struct {
	Title       string        `json:"title"`
	Version     string        `json:"version"`
	Description string        `json:"description"`
	Endpoints   []APIEndpoint `json:"endpoints"`
}

func (s *Server) docsHandler(c *gin.Context) {
	docs := &APIDocumentation{
		Title:       "Relativistic Blockchain SDK API",
		Version:     "1.0.0",
		Description: "API for relativistic blockchain consensus and network management",
		Endpoints:   s.getEndpoints(),
	}

	c.JSON(http.StatusOK, docs)
}

func (s *Server) getEndpoints() []APIEndpoint {
	return []APIEndpoint{
		{
			Method:      "GET",
			Path:        "/api/v1/health",
			Description: "Get service health status",
			AuthRequired: false,
			AdminRequired: false,
		},
		{
			Method:      "GET",
			Path:        "/api/v1/metrics",
			Description: "Get system metrics",
			AuthRequired: false,
			AdminRequired: false,
		},
		{
			Method:      "GET",
			Path:        "/api/v1/status",
			Description: "Get service status",
			AuthRequired: false,
			AdminRequired: false,
		},
		{
			Method:      "GET",
			Path:        "/api/v1/nodes",
			Description: "Get all nodes",
			AuthRequired: false,
			AdminRequired: false,
		},
		{
			Method:      "POST",
			Path:        "/api/v1/nodes",
			Description: "Register a new node",
			AuthRequired: true,
			AdminRequired: false,
		},
		{
			Method:      "GET",
			Path:        "/api/v1/nodes/:id",
			Description: "Get node by ID",
			AuthRequired: false,
			AdminRequired: false,
		},
		{
			Method:      "PUT",
			Path:        "/api/v1/nodes/:id/position",
			Description: "Update node position",
			AuthRequired: true,
			AdminRequired: false,
		},
		{
			Method:      "DELETE",
			Path:        "/api/v1/nodes/:id",
			Description: "Delete a node",
			AuthRequired: true,
			AdminRequired: true,
		},
		{
			Method:      "POST",
			Path:        "/api/v1/calculations/propagation",
			Description: "Calculate propagation delays",
			AuthRequired: false,
			AdminRequired: false,
		},
		{
			Method:      "POST",
			Path:        "/api/v1/calculations/interplanetary",
			Description: "Calculate interplanetary delays",
			AuthRequired: false,
			AdminRequired: false,
		},
		{
			Method:      "POST",
			Path:        "/api/v1/validation/timestamp",
			Description: "Validate timestamp",
			AuthRequired: false,
			AdminRequired: false,
		},
		{
			Method:      "GET",
			Path:        "/api/v1/consensus/timing",
			Description: "Get consensus timing",
			AuthRequired: false,
			AdminRequired: false,
		},
		{
			Method:      "GET",
			Path:        "/api/v1/consensus/offsets",
			Description: "Get node offsets",
			AuthRequired: false,
			AdminRequired: false,
		},
		{
			Method:      "POST",
			Path:        "/api/v1/consensus/validate",
			Description: "Validate consensus",
			AuthRequired: false,
			AdminRequired: false,
		},
		{
			Method:      "GET",
			Path:        "/api/v1/network/topology",
			Description: "Get network topology",
			AuthRequired: false,
			AdminRequired: false,
		},
		{
			Method:      "GET",
			Path:        "/api/v1/network/latency",
			Description: "Get latency measurements",
			AuthRequired: false,
			AdminRequired: false,
		},
		{
			Method:      "GET",
			Path:        "/api/v1/network/alerts",
			Description: "Get network alerts",
			AuthRequired: false,
			AdminRequired: false,
		},
		{
			Method:      "GET",
			Path:        "/api/v1/admin/stats",
			Description: "Get admin statistics",
			AuthRequired: true,
			AdminRequired: true,
		},
		{
			Method:      "POST",
			Path:        "/api/v1/admin/cache/clear",
			Description: "Clear system cache",
			AuthRequired: true,
			AdminRequired: true,
		},
		{
			Method:      "GET",
			Path:        "/ws",
			Description: "WebSocket connection for real-time updates",
			AuthRequired: false,
			AdminRequired: false,
		},
	}
}

func (s *Server) swaggerHandler(c *gin.Context) {
	swaggerJSON := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Relativistic Blockchain SDK API",
			"version": "1.0.0",
			"description": "API for relativistic blockchain consensus and network management"
		},
		"servers": [
			{
				"url": "http://localhost:8080",
				"description": "Development server"
			}
		],
		"paths": {
			"/api/v1/health": {
				"get": {
					"summary": "Get service health",
					"responses": {
						"200": {
							"description": "Service health status"
						}
					}
				}
			}
		}
	}`

	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, swaggerJSON)
}

type ExampleRequest struct {
	Description string      `json:"description"`
	Request     interface{} `json:"request"`
	Response    interface{} `json:"response"`
}

func (s *Server) examplesHandler(c *gin.Context) {
	examples := map[string][]ExampleRequest{
		"node_registration": {
			{
				Description: "Register a new node",
				Request: map[string]interface{}{
					"id": "node-123",
					"position": map[string]float64{
						"latitude":  40.7128,
						"longitude": -74.0060,
						"altitude":  0,
					},
					"address": "node-123.example.com:8080",
					"region":  "us-east",
					"provider": "aws",
					"version": "1.0.0",
					"capabilities": []string{"blockchain", "consensus"},
				},
				Response: map[string]interface{}{
					"success": true,
					"message": "Node registered successfully",
					"node_id": "node-123",
				},
			},
		},
		"propagation_calculation": {
			{
				Description: "Calculate propagation delay between nodes",
				Request: map[string]interface{}{
					"source": "node-123",
					"targets": []string{"node-456", "node-789"},
				},
				Response: map[string]interface{}{
					"node-456": map[string]interface{}{
						"source_node": "node-123",
						"target_node": "node-456",
						"theoretical_delay": "45ms",
						"distance": 350.5,
						"success": true,
					},
					"node-789": map[string]interface{}{
						"source_node": "node-123",
						"target_node": "node-789",
						"theoretical_delay": "120ms",
						"distance": 1200.8,
						"success": true,
					},
				},
			},
		},
		"timestamp_validation": {
			{
				Description: "Validate block timestamp",
				Request: map[string]interface{}{
					"timestamp": "2023-01-01T00:00:00Z",
					"position": map[string]float64{
						"latitude":  34.0522,
						"longitude": -118.2437,
						"altitude":  0,
					},
					"origin_node": "node-123",
				},
				Response: map[string]interface{}{
					"valid": true,
					"confidence": 0.95,
					"reason": "Timestamp is within acceptable range",
					"expected_delay": "45ms",
					"actual_diff": "30ms",
				},
			},
		},
	}

	c.JSON(http.StatusOK, examples)
}

func (s *Server) setupDocumentationRoutes() {
	docs := s.router.Group("/docs")
	{
		docs.GET("", s.docsHandler)
		docs.GET("/swagger.json", s.swaggerHandler)
		docs.GET("/examples", s.examplesHandler)
	}
}