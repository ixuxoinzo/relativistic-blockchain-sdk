```markdown
# Examples

## Basic Setup

### Node Registration
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
    "github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/api"
)

func main() {
    client := api.NewClient("http://localhost:8080", "your-api-key")
    
    node := &types.NodeRegistrationRequest{
        ID: "node-nyc-001",
        Position: types.Position{
            Latitude:  40.7128,
            Longitude: -74.0060,
            Altitude:  0,
        },
        Address: "node-nyc-001.example.com:8080",
        Metadata: types.Metadata{
            Region:   "us-east",
            Provider: "aws",
            Version:  "1.0.0",
        },
        Capabilities: []string{"blockchain", "consensus", "validation"},
    }

    err := client.RegisterNode(context.Background(), node)
    if err != nil {
        log.Fatalf("Failed to register node: %v", err)
    }
    
    fmt.Println("Node registered successfully")
}
```

Propagation Calculation

```go
func calculatePropagationDelays() {
    client := api.NewClient("http://localhost:8080", "your-api-key")
    
    request := &types.PropagationRequest{
        Source:  "node-nyc-001",
        Targets: []string{"node-lax-001", "node-lon-001"},
        IncludePath: true,
    }

    results, err := client.CalculatePropagation(context.Background(), request)
    if err != nil {
        log.Fatalf("Failed to calculate propagation: %v", err)
    }

    for target, result := range results {
        fmt.Printf("To %s: %v delay, %.2f km\n", 
            target, result.TheoreticalDelay, result.Distance)
    }
}
```

Advanced Usage

Batch Validation

```go
func validateMultipleBlocks() {
    client := api.NewClient("http://localhost:8080", "your-api-key")
    
    items := []types.ValidatableItem{
        {
            Type: types.ItemTypeBlock,
            Block: &types.Block{
                Hash:      "0xabc123def456...",
                Timestamp: time.Now().UTC().Add(-2 * time.Second),
                ProposedBy: "node-nyc-001",
                NodePosition: types.Position{
                    Latitude:  40.7128,
                    Longitude: -74.0060,
                    Altitude:  0,
                },
                Data: []byte("block data"),
            },
        },
        {
            Type: types.ItemTypeTransaction,
            Transaction: &types.Transaction{
                Hash:      "0x789ghi012jkl...",
                Timestamp: time.Now().UTC().Add(-1 * time.Second),
                NodePosition: types.Position{
                    Latitude:  34.0522,
                    Longitude: -118.2437,
                    Altitude:  0,
                },
                Data: []byte("transaction data"),
            },
        },
    }

    request := &types.BatchValidationRequest{
        Items:      items,
        OriginNode: "validator-node-001",
    }

    results, err := client.BatchValidate(context.Background(), request)
    if err != nil {
        log.Fatalf("Batch validation failed: %v", err)
    }

    for hash, result := range results {
        status := "VALID"
        if !result.Valid {
            status = "INVALID"
        }
        fmt.Printf("%s: %s (confidence: %.2f)\n", hash, status, result.Confidence)
    }
}
```

Real-time Monitoring

```go
func monitorNetwork() {
    client := api.NewClient("http://localhost:8080", "your-api-key")
    
    ctx := context.Background()
    
    metrics, err := client.GetMetrics(ctx)
    if err != nil {
        log.Fatalf("Failed to get metrics: %v", err)
    }

    fmt.Printf("Total Nodes: %d\n", metrics.TotalNodes)
    fmt.Printf("Active Nodes: %d\n", metrics.ActiveNodes)
    fmt.Printf("Network Coverage: %.2f%%\n", metrics.NetworkCoverage*100)
    fmt.Printf("Avg Propagation Delay: %v\n", metrics.AveragePropagationDelay)

    historical, err := client.GetHistoricalMetrics(ctx, "24h", "propagation_delay")
    if err != nil {
        log.Fatalf("Failed to get historical metrics: %v", err)
    }

    for _, point := range historical.Data {
        fmt.Printf("%s: %.2f ms\n", point.Timestamp.Format("15:04"), point.Value)
    }
}
```

WebSocket Integration

JavaScript Client

```javascript
class RelativisticClient {
    constructor(baseUrl) {
        this.baseUrl = baseUrl;
        this.ws = null;
        this.subscriptions = new Set();
    }

    connect() {
        this.ws = new WebSocket(`${this.baseUrl}/ws`);
        
        this.ws.onopen = () => {
            console.log('Connected to Relativistic SDK');
            this.resubscribe();
        };

        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
        };

        this.ws.onclose = () => {
            console.log('Disconnected from Relativistic SDK');
            setTimeout(() => this.connect(), 5000);
        };
    }

    subscribe(channel) {
        this.subscriptions.add(channel);
        
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({
                type: 'subscribe',
                channels: [channel]
            }));
        }
    }

    resubscribe() {
        if (this.subscriptions.size > 0) {
            this.ws.send(JSON.stringify({
                type: 'subscribe', 
                channels: Array.from(this.subscriptions)
            }));
        }
    }

    handleMessage(message) {
        switch (message.type) {
            case 'metrics_update':
                this.onMetricsUpdate(message.data);
                break;
            case 'nodes_update':
                this.onNodesUpdate(message.data);
                break;
            case 'alerts_update':
                this.onAlertsUpdate(message.data);
                break;
        }
    }

    onMetricsUpdate(data) {
        console.log('Metrics Update:', data);
        document.getElementById('total-nodes').textContent = data.total_nodes;
        document.getElementById('active-nodes').textContent = data.active_nodes;
        document.getElementById('network-coverage').textContent = 
            (data.network_coverage * 100).toFixed(1) + '%';
    }

    onNodesUpdate(data) {
        console.log('Nodes Update:', data);
        if (data.action === 'node_joined') {
            this.showNotification(`Node ${data.node_id} joined the network`);
        } else if (data.action === 'node_left') {
            this.showNotification(`Node ${data.node_id} left the network`);
        }
    }

    onAlertsUpdate(data) {
        console.log('Alert:', data);
        this.showAlert(data.level, data.message);
    }

    showNotification(message) {
        const notification = document.createElement('div');
        notification.className = 'notification';
        notification.textContent = message;
        document.body.appendChild(notification);
        
        setTimeout(() => notification.remove(), 3000);
    }

    showAlert(level, message) {
        const alert = document.createElement('div');
        alert.className = `alert alert-${level}`;
        alert.textContent = message;
        document.body.appendChild(alert);
        
        setTimeout(() => alert.remove(), 5000);
    }
}

const client = new RelativisticClient('ws://localhost:8080');
client.connect();
client.subscribe('metrics');
client.subscribe('nodes'); 
client.subscribe('alerts');
```

Go WebSocket Client

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/gorilla/websocket"
)

type WSClient struct {
    conn *websocket.Conn
}

func NewWSClient(url string) (*WSClient, error) {
    conn, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        return nil, err
    }
    
    return &WSClient{conn: conn}, nil
}

func (c *WSClient) Subscribe(channels []string) error {
    msg := map[string]interface{}{
        "type":     "subscribe",
        "channels": channels,
    }
    
    return c.conn.WriteJSON(msg)
}

func (c *WSClient) Listen(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            _, message, err := c.conn.ReadMessage()
            if err != nil {
                log.Printf("Read error: %v", err)
                time.Sleep(5 * time.Second)
                continue
            }
            
            var data map[string]interface{}
            if err := json.Unmarshal(message, &data); err != nil {
                log.Printf("Parse error: %v", err)
                continue
            }
            
            c.handleMessage(data)
        }
    }
}

func (c *WSClient) handleMessage(data map[string]interface{}) {
    msgType, _ := data["type"].(string)
    
    switch msgType {
    case "metrics_update":
        c.handleMetricsUpdate(data["data"].(map[string]interface{}))
    case "nodes_update":
        c.handleNodesUpdate(data["data"].(map[string]interface{}))
    case "alerts_update":
        c.handleAlertsUpdate(data["data"].(map[string]interface{}))
    }
}

func (c *WSClient) handleMetricsUpdate(data map[string]interface{}) {
    fmt.Printf("Metrics - Total Nodes: %.0f, Active: %.0f, Coverage: %.2f%%\n",
        data["total_nodes"].(float64),
        data["active_nodes"].(float64),
        data["network_coverage"].(float64)*100,
    )
}

func (c *WSClient) handleNodesUpdate(data map[string]interface{}) {
    action, _ := data["action"].(string)
    nodeID, _ := data["node_id"].(string)
    
    fmt.Printf("Node %s: %s\n", nodeID, action)
}

func (c *WSClient) handleAlertsUpdate(data map[string]interface{}) {
    level, _ := data["level"].(string)
    message, _ := data["message"].(string)
    
    fmt.Printf("ALERT [%s]: %s\n", level, message)
}

func main() {
    client, err := NewWSClient("ws://localhost:8080/ws")
    if err != nil {
        log.Fatal(err)
    }
    defer client.conn.Close()
    
    err = client.Subscribe([]string{"metrics", "nodes", "alerts"})
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    client.Listen(ctx)
}
```

Integration with Blockchain

Block Validation Middleware

```go
package main

import (
    "context"
    "time"

    "github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/core"
    "github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type BlockValidator struct {
    engine *core.RelativisticEngine
}

func NewBlockValidator(engine *core.RelativisticEngine) *BlockValidator {
    return &BlockValidator{engine: engine}
}

func (v *BlockValidator) ValidateBlock(block *types.Block, validatorNodeID string) (*types.ValidationResult, error) {
    result, err := v.engine.ValidateBlockTimestamp(context.Background(), block, validatorNodeID)
    if err != nil {
        return nil, err
    }
    
    if !result.Valid {
        return result, fmt.Errorf("block timestamp validation failed: %s", result.Reason)
    }
    
    return result, nil
}

func (v *BlockValidator) ShouldAcceptBlock(block *types.Block, localNodeID string) bool {
    result, err := v.ValidateBlock(block, localNodeID)
    if err != nil {
        return false
    }
    
    return result.Valid && result.Confidence > 0.8
}
```

Consensus Integration

```go
package main

import (
    "sync"
    "time"

    "github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/core"
    "github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type RelativisticConsensus struct {
    engine      *core.RelativisticEngine
    validators  map[string]*types.Node
    mu          sync.RWMutex
}

func NewRelativisticConsensus(engine *core.RelativisticEngine) *RelativisticConsensus {
    return &RelativisticConsensus{
        engine:     engine,
        validators: make(map[string]*types.Node),
    }
}

func (rc *RelativisticConsensus) AddValidator(node *types.Node) {
    rc.mu.Lock()
    defer rc.mu.Unlock()
    
    rc.validators[node.ID] = node
}

func (rc *RelativisticConsensus) ValidateProposal(block *types.Block) (bool, map[string]*types.ValidationResult) {
    rc.mu.RLock()
    defer rc.mu.RUnlock()
    
    results := make(map[string]*types.ValidationResult)
    validCount := 0
    
    for validatorID := range rc.validators {
        result, err := rc.engine.ValidateBlockTimestamp(context.Background(), block, validatorID)
        if err != nil {
            continue
        }
        
        results[validatorID] = result
        if result.Valid && result.Confidence > 0.7 {
            validCount++
        }
    }
    
    totalValidators := len(rc.validators)
    approvalRatio := float64(validCount) / float64(totalValidators)
    
    return approvalRatio > 0.67, results
}

func (rc *RelativisticConsensus) CalculateOptimalBlockTime() time.Duration {
    rc.mu.RLock()
    defer rc.mu.RUnlock()
    
    var maxDelay time.Duration
    nodes := make([]*types.Node, 0, len(rc.validators))
    
    for _, node := range rc.validators {
        nodes = append(nodes, node)
    }
    
    delays, err := rc.engine.BatchCalculateDelays(nodes)
    if err != nil {
        return 10 * time.Second
    }
    
    for _, delay := range delays {
        if delay > maxDelay {
            maxDelay = delay
        }
    }
    
    optimalTime := maxDelay * 3
    if optimalTime < 5*time.Second {
        optimalTime = 5 * time.Second
    }
    
    return optimalTime
}
```

Custom Metrics Collection

```go
package main

import (
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/metrics"
)

type CustomMetrics struct {
    blocksValidated    prometheus.Counter
    validationLatency  prometheus.Histogram
    propagationErrors  prometheus.CounterVec
}

func NewCustomMetrics() *CustomMetrics {
    return &CustomMetrics{
        blocksValidated: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "relativistic_blocks_validated_total",
            Help: "Total number of blocks validated",
        }),
        validationLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name:    "relativistic_validation_latency_seconds",
            Help:    "Validation latency distribution",
            Buckets: prometheus.DefBuckets,
        }),
        propagationErrors: *prometheus.NewCounterVec(prometheus.CounterOpts{
            Name: "relativistic_propagation_errors_total",
            Help: "Total propagation errors by type",
        }, []string{"error_type"}),
    }
}

func (cm *CustomMetrics) Register() {
    prometheus.MustRegister(cm.blocksValidated)
    prometheus.MustRegister(cm.validationLatency)
    prometheus.MustRegister(cm.propagationErrors)
}

func (cm *CustomMetrics) RecordValidation(start time.Time, success bool) {
    cm.blocksValidated.Inc()
    cm.validationLatency.Observe(time.Since(start).Seconds())
    
    if !success {
        cm.propagationErrors.WithLabelValues("validation_failed").Inc()
    }
}
```

These examples demonstrate various ways to integrate the Relativistic SDK into your blockchain applications, from basic node registration to advanced consensus mechanisms and real-time monitoring.

```
