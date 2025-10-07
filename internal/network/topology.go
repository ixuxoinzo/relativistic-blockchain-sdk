package network

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/utils"
)

type TopologyManager struct {
	nodes   map[string]*types.Node
	mu      sync.RWMutex
	redis   *redis.Client
	logger  *zap.Logger
	eventCh chan TopologyEvent
}

type TopologyEvent struct {
	Type      types.EventType
	Node      *types.Node
	Timestamp time.Time
}

func NewTopologyManager(redisAddr string, logger *zap.Logger) (*TopologyManager, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
		PoolSize: 100,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	tm := &TopologyManager{
		nodes:   make(map[string]*types.Node),
		redis:   rdb,
		logger:  logger,
		eventCh: make(chan TopologyEvent, 100),
	}

	if err := tm.loadNodesFromRedis(); err != nil {
		logger.Warn("Failed to load nodes from Redis", zap.Error(err))
	}

	go tm.processEvents()
	return tm, nil
}

func (tm *TopologyManager) AddNode(node *types.Node) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.nodes[node.ID]; exists {
		return fmt.Errorf("node %s already exists", node.ID)
	}

	if err := tm.validateNode(node); err != nil {
		return fmt.Errorf("invalid node data: %w", err)
	}

	node.LastSeen = time.Now().UTC()
	node.IsActive = true
	tm.nodes[node.ID] = node

	if err := tm.persistNode(node); err != nil {
		return fmt.Errorf("failed to persist node: %w", err)
	}

	tm.eventCh <- TopologyEvent{
		Type:      types.EventTypeNodeRegistered,
		Node:      node,
		Timestamp: time.Now().UTC(),
	}

	tm.logger.Info("Node added successfully",
		zap.String("node_id", node.ID),
		zap.Float64("lat", node.Position.Latitude),
		zap.Float64("lon", node.Position.Longitude),
		zap.String("region", node.Metadata.Region),
	)
	return nil
}

func (tm *TopologyManager) GetNode(nodeID string) (*types.Node, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	node, exists := tm.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}
	return node, nil
}

func (tm *TopologyManager) RemoveNode(nodeID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	node, exists := tm.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	delete(tm.nodes, nodeID)
	ctx := context.Background()
	key := fmt.Sprintf("node:%s", nodeID)
	if err := tm.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to remove node from Redis: %w", err)
	}

	tm.eventCh <- TopologyEvent{
		Type:      types.EventTypeNodeRemoved,
		Node:      node,
		Timestamp: time.Now().UTC(),
	}

	tm.logger.Info("Node removed", zap.String("node_id", nodeID))
	return nil
}

func (tm *TopologyManager) UpdateNodePosition(nodeID string, newPos types.Position) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	node, exists := tm.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	if newPos.Latitude < -90 || newPos.Latitude > 90 {
		return fmt.Errorf("invalid latitude: %f", newPos.Latitude)
	}
	if newPos.Longitude < -180 || newPos.Longitude > 180 {
		return fmt.Errorf("invalid longitude: %f", newPos.Longitude)
	}

	node.Position = newPos
	node.LastSeen = time.Now().UTC()

	if err := tm.persistNode(node); err != nil {
		return fmt.Errorf("failed to update node in Redis: %w", err)
	}

	tm.eventCh <- TopologyEvent{
		Type:      types.EventTypeNodeUpdated,
		Node:      node,
		Timestamp: time.Now().UTC(),
	}

	tm.logger.Debug("Node position updated",
		zap.String("node_id", nodeID),
		zap.Float64("new_lat", newPos.Latitude),
		zap.Float64("new_lon", newPos.Longitude),
	)
	return nil
}

func (tm *TopologyManager) GetAllNodes() []*types.Node {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	nodes := make([]*types.Node, 0, len(tm.nodes))
	for _, node := range tm.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

func (tm *TopologyManager) GetNodesByRegion(region string) []*types.Node {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var nodes []*types.Node
	for _, node := range tm.nodes {
		if node.Metadata.Region == region {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (tm *TopologyManager) GetActiveNodes() []*types.Node {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var nodes []*types.Node
	for _, node := range tm.nodes {
		if node.IsActive {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (tm *TopologyManager) loadNodesFromRedis() error {
	ctx := context.Background()
	keys, err := tm.redis.Keys(ctx, "node:*").Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		data, err := tm.redis.HGetAll(ctx, key).Result()
		if err != nil {
			tm.logger.Warn("Failed to load node data", zap.String("key", key), zap.Error(err))
			continue
		}
		node, err := tm.unmarshalNode(data)
		if err != nil {
			tm.logger.Warn("Failed to unmarshal node", zap.String("key", key), zap.Error(err))
			continue
		}
		tm.nodes[node.ID] = node
	}
	tm.logger.Info("Loaded nodes from Redis", zap.Int("count", len(keys)))
	return nil
}

func (tm *TopologyManager) persistNode(node *types.Node) error {
	ctx := context.Background()
	key := fmt.Sprintf("node:%s", node.ID)

	data := map[string]interface{}{
		"id":        node.ID,
		"lat":       node.Position.Latitude,
		"lon":       node.Position.Longitude,
		"alt":       node.Position.Altitude,
		"address":   node.Address,
		"last_seen": node.LastSeen.Format(time.RFC3339),
		"is_active": node.IsActive,
		"region":    node.Metadata.Region,
		"provider":  node.Metadata.Provider,
		"version":   node.Metadata.Version,
	}

	if len(node.Metadata.Capabilities) > 0 {
		capabilitiesJSON, err := json.Marshal(node.Metadata.Capabilities)
		if err == nil {
			data["capabilities"] = string(capabilitiesJSON)
		}
	}
	return tm.redis.HSet(ctx, key, data).Err()
}

func (tm *TopologyManager) unmarshalNode(data map[string]string) (*types.Node, error) {
	node := &types.Node{
		ID: data["id"],
		Position: types.Position{
			Latitude:  utils.ParseFloat(data["lat"]),
			Longitude: utils.ParseFloat(data["lon"]),
			Altitude:  utils.ParseFloat(data["alt"]),
		},
		Address: data["address"],
		Metadata: types.Metadata{
			Region:   data["region"],
			Provider: data["provider"],
			Version:  data["version"],
		},
		IsActive: data["is_active"] == "true",
	}

	if lastSeen, err := time.Parse(time.RFC3339, data["last_seen"]); err == nil {
		node.LastSeen = lastSeen
	}

	if capabilitiesJSON, exists := data["capabilities"]; exists {
		var capabilities []string
		if err := json.Unmarshal([]byte(capabilitiesJSON), &capabilities); err == nil {
			node.Metadata.Capabilities = capabilities
		}
	}
	return node, nil
}

func (tm *TopologyManager) validateNode(node *types.Node) error {
	if node.ID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if node.Position.Latitude < -90 || node.Position.Latitude > 90 {
		return fmt.Errorf("invalid latitude: %f", node.Position.Latitude)
	}
	if node.Position.Longitude < -180 || node.Position.Longitude > 180 {
		return fmt.Errorf("invalid longitude: %f", node.Position.Longitude)
	}
	if node.Position.Altitude < 0 {
		return fmt.Errorf("invalid altitude: %f", node.Position.Altitude)
	}
	return nil
}

func (tm *TopologyManager) processEvents() {
	for event := range tm.eventCh {
		tm.logger.Debug("Topology event processed",
			zap.String("event_type", string(event.Type)),
			zap.String("node_id", event.Node.ID),
			zap.Time("timestamp", event.Timestamp),
		)
	}
}

func (tm *TopologyManager) GetEventChannel() <-chan TopologyEvent {
	return tm.eventCh
}

func (tm *TopologyManager) Close() {
	close(tm.eventCh)
	tm.redis.Close()
}
