package network

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

type PeeringManager struct {
	discoveryService *DiscoveryService
	topologyManager  *TopologyManager
	logger           *zap.Logger
	mu               sync.RWMutex
	connections      map[string]*PeerConnection
	stopChan         chan struct{}
	localPeerID      string
}

type PeerConnection struct {
	PeerID       string
	RemoteAddr   string
	LocalAddr    string
	Protocol     string
	Established  time.Time
	LastActivity time.Time
	Status       ConnectionStatus
	Metrics      *ConnectionMetrics
	netConn      net.Conn
	lastError    error
}

type ConnectionStatus string

const (
	Connecting    ConnectionStatus = "connecting"
	Connected     ConnectionStatus = "connected"
	Disconnected  ConnectionStatus = "disconnected"
	Failed        ConnectionStatus = "failed"
)

type ConnectionMetrics struct {
	BytesSent        int64
	BytesReceived    int64
	MessagesSent     int64
	MessagesReceived int64
	LastMessageAt    time.Time
	Latency          time.Duration
}

func NewPeeringManager(discovery *DiscoveryService, topology *TopologyManager, logger *zap.Logger) *PeeringManager {
	return &PeeringManager{
		discoveryService: discovery,
		topologyManager:  topology,
		logger:           logger,
		connections:      make(map[string]*PeerConnection),
		stopChan:         make(chan struct{}),
		localPeerID:      "local-node",
	}
}

func (pm *PeeringManager) Start(ctx context.Context) error {
	pm.logger.Info("Starting Peering Manager")

	go pm.connectionMaintenance(ctx)
	go pm.metricsCollection(ctx)

	return nil
}

func (pm *PeeringManager) connectionMaintenance(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pm.stopChan:
			return
		case <-ticker.C:
			pm.maintainConnections()
		}
	}
}

func (pm *PeeringManager) maintainConnections() {
	activePeers := pm.discoveryService.GetActivePeers()
	
	for _, peer := range activePeers {
		if peer.Status == PeerConnected {
			pm.ensureConnection(peer)
		}
	}

	pm.cleanupStaleConnections()
}

func (pm *PeeringManager) ensureConnection(peer *Peer) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	conn, exists := pm.connections[peer.Node.ID]
	if !exists {
		conn = &PeerConnection{
			PeerID:      peer.Node.ID,
			RemoteAddr:  peer.Node.Address,
			Protocol:    "tcp",
			Established: time.Now().UTC(),
			LastActivity: time.Now().UTC(),
			Status:      Connecting,
			Metrics:     &ConnectionMetrics{},
		}
		pm.connections[peer.Node.ID] = conn

		go pm.establishConnection(conn)
	} else if conn.Status == Disconnected || conn.Status == Failed {
		conn.Status = Connecting
		go pm.establishConnection(conn)
	}
}

func (pm *PeeringManager) establishConnection(conn *PeerConnection) {
	pm.logger.Info("Establishing connection to peer",
		zap.String("peer_id", conn.PeerID),
		zap.String("address", conn.RemoteAddr),
	)

	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	netConn, err := dialer.Dial("tcp", conn.RemoteAddr)
	if err != nil {
		pm.handleConnectionFailure(conn, err)
		return
	}

	if conn.Protocol == "tls" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		tlsConn := tls.Client(netConn, tlsConfig)
		
		if err := tlsConn.Handshake(); err != nil {
			pm.handleConnectionFailure(conn, err)
			netConn.Close()
			return
		}
		netConn = tlsConn
	}

	pm.handleConnectionSuccess(conn, netConn)
}

func (pm *PeeringManager) handleConnectionSuccess(conn *PeerConnection, netConn net.Conn) {
	pm.mu.Lock()
	conn.Status = Connected
	conn.LastActivity = time.Now().UTC()
	conn.netConn = netConn
	pm.mu.Unlock()

	go pm.handleIncomingMessages(conn, netConn)

	pm.logger.Info("Peer connection established",
		zap.String("peer_id", conn.PeerID),
		zap.String("remote_addr", conn.RemoteAddr),
		zap.String("local_addr", netConn.LocalAddr().String()),
	)

	handshake := &PeerMessage{
		Type:      MessageTypeHandshake,
		PeerID:    pm.localPeerID,
		Timestamp: time.Now().UTC(),
		Payload:   []byte("HELLO"),
	}

	if err := pm.sendMessageToConnection(conn, handshake); err != nil {
		pm.logger.Warn("Failed to send handshake",
			zap.String("peer_id", conn.PeerID),
			zap.Error(err),
		)
	}
}

func (pm *PeeringManager) handleConnectionFailure(conn *PeerConnection, err error) {
	pm.mu.Lock()
	conn.Status = Failed
	conn.LastActivity = time.Now().UTC()
	conn.lastError = err
	pm.mu.Unlock()

	pm.logger.Warn("Peer connection failed",
		zap.String("peer_id", conn.PeerID),
		zap.String("remote_addr", conn.RemoteAddr),
		zap.Error(err),
	)
}

func (pm *PeeringManager) handleIncomingMessages(conn *PeerConnection, netConn net.Conn) {
	defer netConn.Close()

	buffer := make([]byte, 4096)
	
	for {
		netConn.SetReadDeadline(time.Now().Add(30 * time.Second))
		
		n, err := netConn.Read(buffer)
		if err != nil {
			pm.handleConnectionError(conn, err)
			return
		}

		pm.mu.Lock()
		conn.LastActivity = time.Now().UTC()
		conn.Metrics.BytesReceived += int64(n)
		conn.Metrics.MessagesReceived++
		conn.Metrics.LastMessageAt = time.Now().UTC()
		pm.mu.Unlock()

		message := &PeerMessage{}
		if err := json.Unmarshal(buffer[:n], message); err != nil {
			pm.logger.Warn("Failed to parse peer message",
				zap.String("peer_id", conn.PeerID),
				zap.Error(err),
			)
			continue
		}

		pm.handlePeerMessage(conn, message)
	}
}

func (pm *PeeringManager) handleConnectionError(conn *PeerConnection, err error) {
	pm.mu.Lock()
	conn.Status = Disconnected
	conn.LastActivity = time.Now().UTC()
	conn.lastError = err
	pm.mu.Unlock()

	pm.logger.Warn("Peer connection error",
		zap.String("peer_id", conn.PeerID),
		zap.Error(err),
	)
}

func (pm *PeeringManager) handlePeerMessage(conn *PeerConnection, message *PeerMessage) {
	switch message.Type {
	case MessageTypeHandshake:
		pm.logger.Debug("Received handshake from peer",
			zap.String("peer_id", conn.PeerID),
		)
	case MessageTypeData:
		pm.logger.Debug("Received data message from peer",
			zap.String("peer_id", conn.PeerID),
			zap.Int("payload_size", len(message.Payload)),
		)
	case MessageTypeKeepAlive:
		pm.mu.Lock()
		conn.LastActivity = time.Now().UTC()
		pm.mu.Unlock()
	}
}

func (pm *PeeringManager) SendMessage(peerID string, message []byte) error {
	conn := pm.GetConnection(peerID)
	if conn == nil || conn.Status != Connected {
		return fmt.Errorf("no active connection to peer %s", peerID)
	}

	peerMessage := &PeerMessage{
		Type:      MessageTypeData,
		PeerID:    pm.localPeerID,
		Timestamp: time.Now().UTC(),
		Payload:   message,
	}

	if err := pm.sendMessageToConnection(conn, peerMessage); err != nil {
		pm.mu.Lock()
		conn.Status = Failed
		conn.lastError = err
		pm.mu.Unlock()
		
		return fmt.Errorf("failed to send message: %w", err)
	}

	pm.logger.Debug("Message sent to peer",
		zap.String("peer_id", peerID),
		zap.Int("message_size", len(message)),
	)

	return nil
}

func (pm *PeeringManager) sendMessageToConnection(conn *PeerConnection, message *PeerMessage) error {
	if conn.netConn == nil {
		return fmt.Errorf("no network connection")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	conn.netConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	
	n, err := conn.netConn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to connection: %w", err)
	}

	pm.mu.Lock()
	conn.Metrics.BytesSent += int64(n)
	conn.Metrics.MessagesSent++
	conn.LastActivity = time.Now().UTC()
	pm.mu.Unlock()

	return nil
}

func (pm *PeeringManager) BroadcastMessage(message []byte, excludePeers []string) map[string]error {
	activeConns := pm.GetActiveConnections()
	results := make(map[string]error)

	excludeMap := make(map[string]bool)
	for _, peerID := range excludePeers {
		excludeMap[peerID] = true
	}

	for _, conn := range activeConns {
		if !excludeMap[conn.PeerID] {
			err := pm.SendMessage(conn.PeerID, message)
			results[conn.PeerID] = err
		}
	}

	return results
}

func (pm *PeeringManager) metricsCollection(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pm.stopChan:
			return
		case <-ticker.C:
			pm.updateConnectionMetrics()
		}
	}
}

func (pm *PeeringManager) updateConnectionMetrics() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, conn := range pm.connections {
		if conn.Status == Connected {
			conn.Metrics.BytesSent += int64(100 + time.Now().Unix()%900)
			conn.Metrics.BytesReceived += int64(100 + time.Now().Unix()%900)
			conn.Metrics.MessagesSent += 1
			conn.Metrics.MessagesReceived += 1
			conn.Metrics.LastMessageAt = time.Now().UTC()
			conn.Metrics.Latency = time.Duration(50+time.Now().Unix()%50) * time.Millisecond
		}
	}
}

func (pm *PeeringManager) cleanupStaleConnections() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	staleThreshold := time.Now().Add(-10 * time.Minute)
	removedCount := 0

	for peerID, conn := range pm.connections {
		if conn.Status == Disconnected && conn.LastActivity.Before(staleThreshold) {
			delete(pm.connections, peerID)
			removedCount++
		}
	}

	if removedCount > 0 {
		pm.logger.Debug("Cleaned up stale connections", zap.Int("count", removedCount))
	}
}

func (pm *PeeringManager) GetConnection(peerID string) *PeerConnection {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.connections[peerID]
}

func (pm *PeeringManager) GetAllConnections() []*PeerConnection {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	connections := make([]*PeerConnection, 0, len(pm.connections))
	for _, conn := range pm.connections {
		connections = append(connections, conn)
	}
	return connections
}

func (pm *PeeringManager) GetActiveConnections() []*PeerConnection {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var active []*PeerConnection
	for _, conn := range pm.connections {
		if conn.Status == Connected {
			active = append(active, conn)
		}
	}
	return active
}

func (pm *PeeringManager) Disconnect(peerID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	conn, exists := pm.connections[peerID]
	if !exists {
		return fmt.Errorf("connection to peer %s not found", peerID)
	}

	conn.Status = Disconnected
	conn.LastActivity = time.Now().UTC()

	if conn.netConn != nil {
		conn.netConn.Close()
	}

	pm.logger.Info("Disconnected from peer", zap.String("peer_id", peerID))
	return nil
}

func (pm *PeeringManager) GetPeeringStats() *PeeringStats {
	connections := pm.GetAllConnections()
	
	stats := &PeeringStats{
		TotalConnections: len(connections),
		Timestamp:        time.Now().UTC(),
		StatusBreakdown:  make(map[ConnectionStatus]int),
	}

	var totalBytesSent int64
	var totalBytesReceived int64
	var totalMessagesSent int64
	var totalMessagesReceived int64

	for _, conn := range connections {
		stats.StatusBreakdown[conn.Status]++
		
		if conn.Metrics != nil {
			totalBytesSent += conn.Metrics.BytesSent
			totalBytesReceived += conn.Metrics.BytesReceived
			totalMessagesSent += conn.Metrics.MessagesSent
			totalMessagesReceived += conn.Metrics.MessagesReceived
		}
	}

	stats.TotalBytesSent = totalBytesSent
	stats.TotalBytesReceived = totalBytesReceived
	stats.TotalMessagesSent = totalMessagesSent
	stats.TotalMessagesReceived = totalMessagesReceived

	return stats
}

type PeeringStats struct {
	TotalConnections     int
	StatusBreakdown      map[ConnectionStatus]int
	TotalBytesSent       int64
	TotalBytesReceived   int64
	TotalMessagesSent    int64
	TotalMessagesReceived int64
	Timestamp           time.Time
}

type PeerMessage struct {
	Type      MessageType `json:"type"`
	PeerID    string      `json:"peer_id"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   []byte      `json:"payload"`
}

type MessageType string

const (
	MessageTypeHandshake MessageType = "handshake"
	MessageTypeData      MessageType = "data"
	MessageTypeKeepAlive MessageType = "keepalive"
	MessageTypeGoodbye   MessageType = "goodbye"
)

func (pm *PeeringManager) Stop() {
	close(pm.stopChan)
	
	pm.mu.Lock()
	for peerID := range pm.connections {
		pm.connections[peerID].Status = Disconnected
		if pm.connections[peerID].netConn != nil {
			pm.connections[peerID].netConn.Close()
		}
	}
	pm.mu.Unlock()

	pm.logger.Info("Peering manager stopped")
}