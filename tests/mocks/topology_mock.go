package mocks

import (
	"sync"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type TopologyMock struct {
	nodes map[string]*types.Node
	mu    sync.RWMutex
}

func NewTopologyMock() *TopologyMock {
	return &TopologyMock{
		nodes: make(map[string]*types.Node),
	}
}

func (tm *TopologyMock) AddNode(node *types.Node) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.nodes[node.ID] = node
	return nil
}

func (tm *TopologyMock) GetNode(nodeID string) (*types.Node, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	node, exists := tm.nodes[nodeID]
	if !exists {
		return nil, types.NewError(types.ErrNotFound, "node not found")
	}
	return node, nil
}

func (tm *TopologyMock) GetAllNodes() []*types.Node {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	nodes := make([]*types.Node, 0, len(tm.nodes))
	for _, node := range tm.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

func (tm *TopologyMock) GetActiveNodes() []*types.Node {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var activeNodes []*types.Node
	for _, node := range tm.nodes {
		if node.IsActive {
			activeNodes = append(activeNodes, node)
		}
	}
	return activeNodes
}

func (tm *TopologyMock) RemoveNode(nodeID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.nodes, nodeID)
	return nil
}

func (tm *TopologyMock) UpdateNodePosition(nodeID string, position types.Position) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	node, exists := tm.nodes[nodeID]
	if !exists {
		return types.NewError(types.ErrNotFound, "node not found")
	}

	node.Position = position
	return nil
}
