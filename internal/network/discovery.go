package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type DiscoveryService struct {
	topologyManager *TopologyManager
	logger          *zap.Logger
	mu              sync.RWMutex
	peers           map[string]*Peer
	stopChan        chan struct{}
}

type Peer struct {
	Node         *types.Node
	LastSeen     time.Time
	Status       PeerStatus
	Capabilities []string
	Version      string
}

type PeerStatus string

const (
	PeerConnected    PeerStatus = "connected"
	PeerDisconnected PeerStatus = "disconnected"
	PeerPending      PeerStatus = "pending"
)

func NewDiscoveryService(topology *TopologyManager, logger *zap.Logger) *DiscoveryService {
	return &DiscoveryService{
		topologyManager: topology,
		logger:          logger,
		peers:           make(map[string]*Peer),
		stopChan:        make(chan struct{}),
	}
}

func (ds *DiscoveryService) StartDiscovery(ctx context.Context) error {
	ds.logger.Info("Starting network discovery service")

	go ds.peerDiscovery(ctx)
	go ds.healthCheck(ctx)
	go ds.cleanupStalePeers(ctx)

	return nil
}

func (ds *DiscoveryService) peerDiscovery(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ds.stopChan:
			return
		case <-ticker.C:
			ds.discoverNewPeers()
		}
	}
}

func (ds *DiscoveryService) discoverNewPeers() {
	ds.logger.Debug("Discovering new peers")

	ds.discoverFromDNSeeds()
	ds.discoverFromKnownPeers()
	ds.discoverFromBootstrapNodes()
}

func (ds *DiscoveryService) discoverFromDNSeeds() {
	dnsSeeds := []string{
		"seed1.relativistic-sdk.com",
		"seed2.relativistic-sdk.com",
		"seed3.relativistic-sdk.com",
	}

	for _, seed := range dnsSeeds {
		addrs, err := net.LookupHost(seed)
		if err != nil {
			ds.logger.Debug("DNS seed lookup failed",
				zap.String("seed", seed),
				zap.Error(err),
			)
			continue
		}

		for _, addr := range addrs {
			peer := &Peer{
				Node: &types.Node{
					ID: fmt.Sprintf("dns-peer-%s", addr),
					Position: types.Position{
						Latitude:  0,
						Longitude: 0,
						Altitude:  0,
					},
					Address: fmt.Sprintf("%s:8080", addr),
					Metadata: types.Metadata{
						Region:       "unknown",
						Provider:     "dns-seed",
						Version:      "1.0.0",
						Capabilities: []string{"blockchain", "consensus"},
					},
					IsActive: true,
					LastSeen: time.Now().UTC(),
				},
				LastSeen:     time.Now().UTC(),
				Status:       PeerPending,
				Capabilities: []string{"blockchain", "consensus"},
				Version:      "1.0.0",
			}

			ds.AddPeer(peer)
		}
	}
}

func (ds *DiscoveryService) discoverFromKnownPeers() {
	knownPeers := ds.GetActivePeers()

	for _, peer := range knownPeers {
		peerList, err := ds.queryPeerForPeers(peer)
		if err != nil {
			ds.logger.Debug("Failed to query peer for peers",
				zap.String("peer_id", peer.Node.ID),
				zap.Error(err),
			)
			continue
		}

		for _, newPeer := range peerList {
			ds.AddPeer(newPeer)
		}
	}
}

func (ds *DiscoveryService) queryPeerForPeers(peer *Peer) ([]*Peer, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("http://%s/api/v1/peers", peer.Node.Address)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var peerResponse struct {
		Peers []*Peer `json:"peers"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&peerResponse); err != nil {
		return nil, err
	}

	return peerResponse.Peers, nil
}

func (ds *DiscoveryService) discoverFromBootstrapNodes() {
	bootstrapNodes := []string{
		"bootstrap1.relativistic-sdk.com:8080",
		"bootstrap2.relativistic-sdk.com:8080",
		"bootstrap3.relativistic-sdk.com:8080",
	}

	for i, addr := range bootstrapNodes {
		peer := &Peer{
			Node: &types.Node{
				ID: fmt.Sprintf("bootstrap-%d", i+1),
				Position: types.Position{
					Latitude:  0,
					Longitude: 0,
					Altitude:  0,
				},
				Address: addr,
				Metadata: types.Metadata{
					Region:       "global",
					Provider:     "bootstrap",
					Version:      "1.0.0",
					Capabilities: []string{"blockchain", "consensus", "bootstrap"},
				},
				IsActive: true,
				LastSeen: time.Now().UTC(),
			},
			LastSeen:     time.Now().UTC(),
			Status:       PeerConnected,
			Capabilities: []string{"blockchain", "consensus", "bootstrap"},
			Version:      "1.0.0",
		}

		ds.AddPeer(peer)
	}
}

func (ds *DiscoveryService) healthCheck(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ds.stopChan:
			return
		case <-ticker.C:
			ds.checkPeerHealth()
		}
	}
}

func (ds *DiscoveryService) checkPeerHealth() {
	ds.mu.RLock()
	peers := make([]*Peer, 0, len(ds.peers))
	for _, peer := range ds.peers {
		peers = append(peers, peer)
	}
	ds.mu.RUnlock()

	for _, peer := range peers {
		isHealthy := ds.pingPeer(peer)

		if isHealthy {
			ds.updatePeerStatus(peer.Node.ID, PeerConnected)
			peer.LastSeen = time.Now().UTC()
		} else {
			ds.updatePeerStatus(peer.Node.ID, PeerDisconnected)
		}
	}
}

func (ds *DiscoveryService) pingPeer(peer *Peer) bool {
	client := &http.Client{Timeout: 5 * time.Second}

	url := fmt.Sprintf("http://%s/health", peer.Node.Address)
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func (ds *DiscoveryService) cleanupStalePeers(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ds.stopChan:
			return
		case <-ticker.C:
			ds.removeStalePeers()
		}
	}
}

func (ds *DiscoveryService) removeStalePeers() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	staleThreshold := time.Now().Add(-10 * time.Minute)
	removedCount := 0

	for peerID, peer := range ds.peers {
		if peer.LastSeen.Before(staleThreshold) && peer.Status == PeerDisconnected {
			delete(ds.peers, peerID)
			removedCount++
		}
	}

	if removedCount > 0 {
		ds.logger.Info("Removed stale peers", zap.Int("count", removedCount))
	}
}

func (ds *DiscoveryService) AddPeer(peer *Peer) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	existingNode, err := ds.topologyManager.GetNode(peer.Node.ID)
	if err != nil {
		if err := ds.topologyManager.AddNode(peer.Node); err != nil {
			ds.logger.Warn("Failed to add peer to topology",
				zap.String("peer_id", peer.Node.ID),
				zap.Error(err),
			)
		}
	} else {
		existingNode.LastSeen = time.Now().UTC()
		existingNode.IsActive = true
	}

	ds.peers[peer.Node.ID] = peer

	ds.logger.Debug("Peer added to discovery service",
		zap.String("peer_id", peer.Node.ID),
		zap.String("status", string(peer.Status)),
	)
}

func (ds *DiscoveryService) RemovePeer(peerID string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	delete(ds.peers, peerID)

	if node, err := ds.topologyManager.GetNode(peerID); err == nil {
		node.IsActive = false
	}

	ds.logger.Debug("Peer removed from discovery service", zap.String("peer_id", peerID))
}

func (ds *DiscoveryService) GetPeer(peerID string) *Peer {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.peers[peerID]
}

func (ds *DiscoveryService) GetActivePeers() []*Peer {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var activePeers []*Peer
	for _, peer := range ds.peers {
		if peer.Status == PeerConnected {
			activePeers = append(activePeers, peer)
		}
	}
	return activePeers
}

func (ds *DiscoveryService) GetAllPeers() []*Peer {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	peers := make([]*Peer, 0, len(ds.peers))
	for _, peer := range ds.peers {
		peers = append(peers, peer)
	}
	return peers
}

func (ds *DiscoveryService) updatePeerStatus(peerID string, status PeerStatus) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if peer, exists := ds.peers[peerID]; exists {
		oldStatus := peer.Status
		peer.Status = status

		if oldStatus != status {
			ds.logger.Debug("Peer status updated",
				zap.String("peer_id", peerID),
				zap.String("old_status", string(oldStatus)),
				zap.String("new_status", string(status)),
			)
		}
	}
}

func (ds *DiscoveryService) GetDiscoveryStats() *DiscoveryStats {
	peers := ds.GetAllPeers()

	stats := &DiscoveryStats{
		TotalPeers:      len(peers),
		Timestamp:       time.Now().UTC(),
		StatusBreakdown: make(map[PeerStatus]int),
		RegionBreakdown: make(map[string]int),
	}

	for _, peer := range peers {
		stats.StatusBreakdown[peer.Status]++
		stats.RegionBreakdown[peer.Node.Metadata.Region]++
	}

	return stats
}

type DiscoveryStats struct {
	TotalPeers      int
	StatusBreakdown map[PeerStatus]int
	RegionBreakdown map[string]int
	Timestamp       time.Time
}

func (ds *DiscoveryService) Stop() {
	close(ds.stopChan)
	ds.logger.Info("Discovery service stopped")
}
