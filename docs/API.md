```markdown
# Relativistic SDK API Documentation

## Overview
The Relativistic SDK provides APIs for relativistic blockchain consensus and network management, accounting for physical constraints like light-speed delays in distributed networks.

## Base URL
```

http://localhost:8080/api/v1

```

## Authentication
Most endpoints require JWT authentication:
```

Authorization: Bearer <jwt-token>

```

## Health Check

### GET /health
Returns service health status.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2023-01-01T00:00:00Z",
  "version": "1.0.0",
  "components": {
    "api": "healthy",
    "engine": "healthy",
    "network": "healthy",
    "database": "healthy",
    "cache": "healthy"
  }
}
```

GET /health/ready

Returns service readiness status.

Response:

```json
{
  "status": "ready",
  "timestamp": "2023-01-01T00:00:00Z"
}
```

Nodes Management

GET /nodes

Retrieve all registered nodes.

Query Parameters:

· active (boolean): Filter by active status
· region (string): Filter by region
· provider (string): Filter by cloud provider

Response:

```json
{
  "nodes": [
    {
      "id": "node-123",
      "position": {
        "latitude": 40.7128,
        "longitude": -74.0060,
        "altitude": 0
      },
      "address": "node-123.example.com:8080",
      "metadata": {
        "region": "us-east",
        "provider": "aws",
        "version": "1.0.0"
      },
      "is_active": true,
      "last_seen": "2023-01-01T00:00:00Z",
      "capabilities": ["blockchain", "consensus"]
    }
  ],
  "total_count": 1,
  "active_count": 1
}
```

GET /nodes/{nodeId}

Retrieve specific node details.

Response:

```json
{
  "node": {
    "id": "node-123",
    "position": {
      "latitude": 40.7128,
      "longitude": -74.0060,
      "altitude": 0
    },
    "address": "node-123.example.com:8080",
    "metadata": {
      "region": "us-east",
      "provider": "aws",
      "version": "1.0.0"
    },
    "is_active": true,
    "last_seen": "2023-01-01T00:00:00Z",
    "capabilities": ["blockchain", "consensus"]
  }
}
```

POST /nodes

Register a new node.

Request:

```json
{
  "id": "node-123",
  "position": {
    "latitude": 40.7128,
    "longitude": -74.0060,
    "altitude": 0
  },
  "address": "node-123.example.com:8080",
  "metadata": {
    "region": "us-east",
    "provider": "aws",
    "version": "1.0.0"
  },
  "capabilities": ["blockchain", "consensus"]
}
```

Response:

```json
{
  "success": true,
  "node_id": "node-123",
  "message": "Node registered successfully"
}
```

PUT /nodes/{nodeId}

Update node information.

Request:

```json
{
  "position": {
    "latitude": 40.7128,
    "longitude": -74.0060,
    "altitude": 100
  },
  "is_active": true
}
```

DELETE /nodes/{nodeId}

Remove a node from the network.

Calculations

POST /calculations/propagation

Calculate propagation delays between nodes.

Request:

```json
{
  "source": "node-123",
  "targets": ["node-456", "node-789"],
  "include_path": true
}
```

Response:

```json
{
  "results": {
    "node-456": {
      "source_node": "node-123",
      "target_node": "node-456",
      "theoretical_delay": "45.2ms",
      "distance_km": 3941.2,
      "path": ["node-123", "node-456"],
      "success": true,
      "error": null
    },
    "node-789": {
      "source_node": "node-123",
      "target_node": "node-789",
      "theoretical_delay": "78.5ms",
      "distance_km": 6732.8,
      "path": ["node-123", "node-789"],
      "success": true,
      "error": null
    }
  }
}
```

POST /calculations/batch-delays

Calculate delays for multiple node pairs in batch.

Request:

```json
{
  "pairs": [
    {"source": "node-123", "target": "node-456"},
    {"source": "node-123", "target": "node-789"},
    {"source": "node-456", "target": "node-789"}
  ]
}
```

Response:

```json
{
  "delays": {
    "node-123-node-456": "45.2ms",
    "node-123-node-789": "78.5ms",
    "node-456-node-789": "62.1ms"
  }
}
```

GET /calculations/interplanetary/{planetA}/{planetB}

Calculate interplanetary communication delay.

Response:

```json
{
  "planet_a": "earth",
  "planet_b": "mars",
  "current_distance_km": 225000000,
  "theoretical_delay": "12.5m",
  "speed_of_light_delay": "12.5m"
}
```

Validation

POST /validation/timestamp

Validate block timestamp considering relativistic effects.

Request:

```json
{
  "timestamp": "2023-01-01T00:00:00Z",
  "position": {
    "latitude": 34.0522,
    "longitude": -118.2437,
    "altitude": 0
  },
  "origin_node": "node-123",
  "block_hash": "0xabc123...",
  "validation_type": "strict"
}
```

Response:

```json
{
  "valid": true,
  "confidence": 0.95,
  "calculated_delay": "15.2ms",
  "allowed_time_window": "25ms",
  "actual_difference": "12.3ms",
  "reason": "timestamp_within_acceptable_range",
  "suggested_adjustment": "2.7ms"
}
```

POST /validation/batch

Validate multiple timestamps in batch.

Request:

```json
{
  "items": [
    {
      "timestamp": "2023-01-01T00:00:00Z",
      "position": {
        "latitude": 40.7128,
        "longitude": -74.0060,
        "altitude": 0
      },
      "origin_node": "node-123",
      "block_hash": "0xabc123..."
    },
    {
      "timestamp": "2023-01-01T00:00:01Z",
      "position": {
        "latitude": 34.0522,
        "longitude": -118.2437,
        "altitude": 0
      },
      "origin_node": "node-456",
      "block_hash": "0xdef456..."
    }
  ]
}
```

Response:

```json
{
  "results": {
    "0xabc123...": {
      "valid": true,
      "confidence": 0.95,
      "calculated_delay": "15.2ms",
      "reason": "timestamp_within_acceptable_range"
    },
    "0xdef456...": {
      "valid": false,
      "confidence": 0.45,
      "calculated_delay": "18.7ms",
      "reason": "timestamp_exceeds_maximum_delay"
    }
  }
}
```

Metrics and Monitoring

GET /metrics

Get system metrics and statistics.

Response:

```json
{
  "metrics": {
    "total_nodes": 150,
    "active_nodes": 142,
    "network_coverage": 0.95,
    "average_propagation_delay": "45.2ms",
    "validation_success_rate": 0.98,
    "system_uptime": "99.5%",
    "requests_processed": 125000
  },
  "timestamp": "2023-01-01T00:00:00Z"
}
```

GET /metrics/historical

Get historical metrics data.

Query Parameters:

· period (string): 1h, 24h, 7d, 30d
· metric (string): Specific metric to retrieve

Response:

```json
{
  "metric": "propagation_delay",
  "period": "24h",
  "data": [
    {
      "timestamp": "2023-01-01T00:00:00Z",
      "value": 45.2
    },
    {
      "timestamp": "2023-01-01T01:00:00Z",
      "value": 43.8
    }
  ]
}
```

WebSocket API

GET /ws

Real-time updates for metrics, nodes, and alerts.

Connection:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
```

Message Types:

Subscribe to Channels

```json
{
  "type": "subscribe",
  "channels": ["metrics", "nodes", "alerts"]
}
```

Metrics Update

```json
{
  "type": "metrics_update",
  "channel": "metrics",
  "data": {
    "total_nodes": 150,
    "active_nodes": 142,
    "network_coverage": 0.95
  },
  "timestamp": "2023-01-01T00:00:00Z"
}
```

Nodes Update

```json
{
  "type": "nodes_update",
  "channel": "nodes",
  "data": {
    "action": "node_joined",
    "node_id": "node-123",
    "position": {
      "latitude": 40.7128,
      "longitude": -74.0060
    },
    "timestamp": "2023-01-01T00:00:00Z"
  }
}
```

Alerts Update

```json
{
  "type": "alerts_update",
  "channel": "alerts",
  "data": {
    "level": "warning",
    "message": "High propagation delay detected",
    "node_id": "node-456",
    "metric": "propagation_delay",
    "value": 120.5,
    "threshold": 100.0,
    "timestamp": "2023-01-01T00:00:00Z"
  }
}
```

Error Responses

All endpoints may return the following error structure:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid timestamp format",
    "details": {
      "field": "timestamp",
      "issue": "must be ISO 8601 format"
    },
    "timestamp": "2023-01-01T00:00:00Z"
  }
}
```

Common Error Codes:

· AUTH_REQUIRED: Authentication required
· INVALID_TOKEN: Invalid JWT token
· NODE_NOT_FOUND: Node not found
· VALIDATION_ERROR: Request validation failed
· CALCULATION_ERROR: Propagation calculation failed
· RATE_LIMITED: Too many requests
· INTERNAL_ERROR: Internal server error

Rate Limiting

· General endpoints: 1000 requests per hour
· Calculation endpoints: 100 requests per minute
· WebSocket connections: 10 concurrent connections per IP

Rate limit headers are included in responses:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

Pagination

List endpoints support pagination:

Query Parameters:

· page (number): Page number (default: 1)
· limit (number): Items per page (default: 50, max: 1000)

Response Headers:

```
X-Total-Count: 1500
X-Page-Count: 30
X-Current-Page: 1
```

SDK Client Libraries

Official client libraries available for:

· Go: github.com/ixuxoinzo/relativistic-sdk-go
· JavaScript: @relativistic/sdk
· Python: relativistic-sdk

Example usage with Go client:

```go
import "github.com/ixuxoinzo/relativistic-sdk-go"

client := relativistic.NewClient("http://localhost:8080")
nodes, err := client.ListNodes(context.Background())
```

Changelog

v1.0.0 (2024-01-01)

· Initial release
· Basic node management
· Propagation calculations
· Timestamp validation
· Real-time WebSocket API

```
