```markdown  
# Relativistic Blockchain SDK 
test123
A high-performance SDK for relativistic blockchain consensus, accounting for physical constraints like light-speed delays in distributed networks.

## Features

- 🌐 **Network Topology Management** - Real-time node discovery and positioning
- ⚡ **Propagation Delay Calculation** - Light-speed delay computations between nodes  
- ⏱️ **Relativistic Timestamp Validation** - Time validation considering physical constraints
- 📊 **Real-time Monitoring** - Comprehensive metrics and analytics
- 🔒 **Enterprise Security** - JWT authentication, rate limiting, and audit logging
- 🚀 **High Performance** - Optimized for thousands of concurrent operations
- ☁️ **Cloud Native** - Kubernetes-ready with Helm charts

## Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 14+
- Redis 7+

### Installation

```bash
# Clone repository
git clone https://github.com/ixuxoinzo/relativistic-blockchain-sdk
cd relativistic-blockchain-sdk

# Setup environment
./scripts/setup.sh

# Start services
docker-compose -f deployments/docker-compose.yml up -d

# Verify installation
curl http://localhost:8080/api/v1/health
```

Using Mock Data (Development)

```bash
# Run with mock data mode
RELATIVISTIC_USE_MOCKS=true ./bin/relativisticd

# Or use demo script
./demo/mock-demo.sh
```

API Examples

Register Node

```bash
curl -X POST http://localhost:8080/api/v1/nodes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-jwt-token" \
  -d '{
    "id": "node-nyc-001",
    "position": {
      "latitude": 40.7128,
      "longitude": -74.0060,
      "altitude": 0
    },
    "address": "node-nyc-001.example.com:8080",
    "metadata": {
      "region": "us-east",
      "provider": "aws",
      "version": "1.0.0"
    },
    "capabilities": ["blockchain", "consensus"]
  }'
```

Calculate Propagation

```bash
curl -X POST http://localhost:8080/api/v1/calculations/propagation \
  -H "Content-Type: application/json" \
  -d '{
    "source": "node-nyc-001",
    "targets": ["node-lax-001", "node-lon-001"]
  }'
```

Validate Timestamp

```bash
curl -X POST http://localhost:8080/api/v1/validation/timestamp \
  -H "Content-Type: application/json" \
  -d '{
    "timestamp": "2024-01-01T00:00:00Z",
    "position": {
      "latitude": 34.0522,
      "longitude": -118.2437,
      "altitude": 0
    },
    "origin_node": "node-nyc-001",
    "block_hash": "0xabc123..."
  }'
```

Development with Mock Data

For testing without external dependencies:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/core"
    "github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
    "github.com/ixuxoinzo/relativistic-blockchain-sdk/tests/mocks"
)

func main() {
    // Use mock engine for testing
    mockEngine := mocks.NewEngineMock()
    
    // Test propagation calculation
    nodeA := &types.Node{ID: "node-1", Position: types.Position{Latitude: 40.7128, Longitude: -74.0060}}
    nodeB := &types.Node{ID: "node-2", Position: types.Position{Latitude: 34.0522, Longitude: -118.2437}}
    
    delay, err := mockEngine.CalculatePropagationDelay(nodeA, nodeB)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Propagation delay: %v\n", delay)
}
```

Project Structure

```
relativistic-blockchain-sdk/
├── cmd/relativisticd/          # Main application entry point
├── internal/                   # Private application code
│   ├── core/                   # Core relativistic engine
│   ├── network/                # Network topology management
│   ├── consensus/              # Consensus algorithms
│   ├── api/                    # HTTP API layer
│   ├── metrics/                # Monitoring and metrics
│   ├── config/                 # Configuration management
│   └── security/               # Security utilities
├── pkg/                        # Public libraries
│   ├── types/                  # Shared data types
│   ├── relativistic/           # Physics calculations
│   └── utils/                  # Utility functions
├── deployments/                # Deployment configurations
├── scripts/                    # Utility scripts
├── tests/                      # Test suites
├── docs/                       # Documentation
└── configs/                    # Configuration files
```

Monitoring

Access monitoring dashboards:

· Grafana: http://localhost:3000 (admin/admin)
· Prometheus: http://localhost:9090
· API Docs: http://localhost:8080/docs

Deployment

Docker

```bash
docker-compose -f deployments/docker-compose.yml up -d
```

Kubernetes

```bash
kubectl apply -f deployments/kubernetes/
```

Production

```bash
./scripts/deploy.sh production latest
```

Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: go test ./...
5. Submit a pull request

License

Apache 2.0 - See LICENSE file for details.

Support

🐛 Issues: GitHub Issues 
```
