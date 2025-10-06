package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/core"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/internal/network"
	"github.com/ixuxoinzo/relativistic-blockchain-sdk/pkg/types"
)

type WebSocketManager struct {
	upgrader   websocket.Upgrader
	clients    map[*WebSocketClient]bool
	broadcast  chan WebSocketMessage
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	logger     *zap.Logger
	mu         sync.RWMutex
	engine     *core.RelativisticEngine
	topology   *network.TopologyManager
}

type WebSocketClient struct {
	conn     *websocket.Conn
	send     chan WebSocketMessage
	manager  *WebSocketManager
	userID   string
	channels map[string]bool
	sessionID string
}

type WebSocketMessage struct {
	Type      string          `json:"type"`
	Channel   string          `json:"channel,omitempty"`
	Data      interface{}     `json:"data,omitempty"`
	Error     string          `json:"error,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	RequestID string          `json:"request_id,omitempty"`
}

func NewWebSocketManager(engine *core.RelativisticEngine, topology *network.TopologyManager, logger *zap.Logger) *WebSocketManager {
	return &WebSocketManager{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // In production, implement proper origin checking
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients:    make(map[*WebSocketClient]bool),
		broadcast:  make(chan WebSocketMessage, 256),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		logger:     logger,
		engine:     engine,
		topology:   topology,
	}
}

func (wm *WebSocketManager) Start() {
	go wm.run()
	go wm.startBroadcasts()
}

func (wm *WebSocketManager) run() {
	for {
		select {
		case client := <-wm.register:
			wm.mu.Lock()
			wm.clients[client] = true
			wm.mu.Unlock()
			wm.logger.Info("WebSocket client connected", 
				zap.String("user_id", client.userID),
				zap.String("session_id", client.sessionID),
			)

			// Send welcome message
			client.send <- WebSocketMessage{
				Type:      "connected",
				Timestamp: time.Now().UTC(),
				Data: map[string]interface{}{
					"session_id": client.sessionID,
					"user_id":    client.userID,
					"channels":   client.getAvailableChannels(),
				},
			}

		case client := <-wm.unregister:
			wm.mu.Lock()
			if _, ok := wm.clients[client]; ok {
				delete(wm.clients, client)
				close(client.send)
			}
			wm.mu.Unlock()
			wm.logger.Info("WebSocket client disconnected", 
				zap.String("user_id", client.userID),
				zap.String("session_id", client.sessionID),
			)

		case message := <-wm.broadcast:
			wm.broadcastToClients(message)
		}
	}
}

func (wm *WebSocketManager) broadcastToClients(message WebSocketMessage) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	for client := range wm.clients {
		// Check if client is subscribed to the message channel
		if message.Channel == "" || client.channels[message.Channel] {
			select {
			case client.send <- message:
			default:
				// Client buffer full, disconnect
				close(client.send)
				delete(wm.clients, client)
			}
		}
	}
}

func (wm *WebSocketManager) Broadcast(message WebSocketMessage) {
	wm.broadcast <- message
}

func (wm *WebSocketManager) BroadcastToChannel(channel string, messageType string, data interface{}) {
	wm.Broadcast(WebSocketMessage{
		Type:      messageType,
		Channel:   channel,
		Data:      data,
		Timestamp: time.Now().UTC(),
	})
}

func (wm *WebSocketManager) HandleConnection(c *gin.Context) {
	conn, err := wm.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		wm.logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		userID = "anonymous_" + generateSessionID()
	}

	sessionID := generateSessionID()

	client := &WebSocketClient{
		conn:      conn,
		send:      make(chan WebSocketMessage, 256),
		manager:   wm,
		userID:    userID,
		sessionID: sessionID,
		channels:  make(map[string]bool),
	}

	// Subscribe to default channels
	client.channels["system"] = true
	client.channels["notifications"] = true

	wm.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *WebSocketClient) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512 * 1024) // 512KB
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message WebSocketMessage
		err := c.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.manager.logger.Warn("WebSocket read error", 
					zap.String("user_id", c.userID),
					zap.Error(err),
				)
			}
			break
		}

		c.handleMessage(&message)
	}
}

func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(30 * time.Second) // Ping interval
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteJSON(message)
			if err != nil {
				c.manager.logger.Warn("WebSocket write error", 
					zap.String("user_id", c.userID),
					zap.Error(err),
				)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *WebSocketClient) handleMessage(message *WebSocketMessage) {
	switch message.Type {
	case "subscribe":
		c.handleSubscribe(message)
	case "unsubscribe":
		c.handleUnsubscribe(message)
	case "ping":
		c.handlePing(message)
	case "calculate_propagation":
		c.handleCalculatePropagation(message)
	case "validate_timestamp":
		c.handleValidateTimestamp(message)
	case "get_metrics":
		c.handleGetMetrics(message)
	case "auth":
		c.handleAuth(message)
	default:
		c.sendError("unknown_message_type", "Unknown message type: "+message.Type, message.RequestID)
	}
}

func (c *WebSocketClient) handleSubscribe(message *WebSocketMessage) {
	var data struct {
		Channels []string `json:"channels"`
	}
	
	if err := json.Unmarshal(message.Data.([]byte), &data); err != nil {
		c.sendError("invalid_subscription", "Invalid subscription data", message.RequestID)
		return
	}

	subscribed := make([]string, 0)
	for _, channel := range data.Channels {
		if c.isValidChannel(channel) {
			c.channels[channel] = true
			subscribed = append(subscribed, channel)
			c.manager.logger.Debug("Client subscribed to channel", 
				zap.String("user_id", c.userID),
				zap.String("channel", channel),
			)
		}
	}

	c.sendMessage(WebSocketMessage{
		Type:      "subscribed",
		RequestID: message.RequestID,
		Timestamp: time.Now().UTC(),
		Data: map[string]interface{}{
			"channels": subscribed,
		},
	})
}

func (c *WebSocketClient) handleUnsubscribe(message *WebSocketMessage) {
	var data struct {
		Channels []string `json:"channels"`
	}
	
	if err := json.Unmarshal(message.Data.([]byte), &data); err != nil {
		c.sendError("invalid_unsubscription", "Invalid unsubscription data", message.RequestID)
		return
	}

	unsubscribed := make([]string, 0)
	for _, channel := range data.Channels {
		if c.channels[channel] {
			delete(c.channels, channel)
			unsubscribed = append(unsubscribed, channel)
			c.manager.logger.Debug("Client unsubscribed from channel", 
				zap.String("user_id", c.userID),
				zap.String("channel", channel),
			)
		}
	}

	c.sendMessage(WebSocketMessage{
		Type:      "unsubscribed",
		RequestID: message.RequestID,
		Timestamp: time.Now().UTC(),
		Data: map[string]interface{}{
			"channels": unsubscribed,
		},
	})
}

func (c *WebSocketClient) handlePing(message *WebSocketMessage) {
	c.sendMessage(WebSocketMessage{
		Type:      "pong",
		RequestID: message.RequestID,
		Timestamp: time.Now().UTC(),
		Data: map[string]interface{}{
			"server_time": time.Now().UTC(),
		},
	})
}

func (c *WebSocketClient) handleCalculatePropagation(message *WebSocketMessage) {
	var data struct {
		Source  string   `json:"source"`
		Targets []string `json:"targets"`
	}
	
	if err := json.Unmarshal(message.Data.([]byte), &data); err != nil {
		c.sendError("invalid_calculation_request", "Invalid calculation data", message.RequestID)
		return
	}

	// Perform calculation in goroutine to avoid blocking
	go func() {
		results, err := c.manager.engine.CalculatePropagationPath(data.Source, data.Targets)
		if err != nil {
			c.sendError("calculation_failed", err.Error(), message.RequestID)
			return
		}

		c.sendMessage(WebSocketMessage{
			Type:      "calculation_result",
			RequestID: message.RequestID,
			Timestamp: time.Now().UTC(),
			Data: map[string]interface{}{
				"source":  data.Source,
				"targets": data.Targets,
				"results": results,
			},
		})
	}()
}

func (c *WebSocketClient) handleValidateTimestamp(message *WebSocketMessage) {
	var data struct {
		Timestamp  time.Time     `json:"timestamp"`
		Position   types.Position `json:"position"`
		OriginNode string        `json:"origin_node"`
	}
	
	if err := json.Unmarshal(message.Data.([]byte), &data); err != nil {
		c.sendError("invalid_validation_request", "Invalid validation data", message.RequestID)
		return
	}

	go func() {
		valid, result := c.manager.engine.ValidateTimestamp(context.Background(), data.Timestamp, data.Position, data.OriginNode)

		c.sendMessage(WebSocketMessage{
			Type:      "validation_result",
			RequestID: message.RequestID,
			Timestamp: time.Now().UTC(),
			Data: map[string]interface{}{
				"valid":      valid,
				"confidence": result.Confidence,
				"reason":     result.Reason,
				"expected_delay": result.ExpectedDelay.String(),
				"actual_diff":   result.ActualDiff.String(),
			},
		})
	}()
}

func (c *WebSocketClient) handleGetMetrics(message *WebSocketMessage) {
	metrics := c.manager.engine.GetNetworkMetrics()
	
	c.sendMessage(WebSocketMessage{
		Type:      "metrics",
		RequestID: message.RequestID,
		Timestamp: time.Now().UTC(),
		Data:      metrics,
	})
}

func (c *WebSocketClient) handleAuth(message *WebSocketMessage) {
	var data struct {
		Token string `json:"token"`
	}
	
	if err := json.Unmarshal(message.Data.([]byte), &data); err != nil {
		c.sendError("invalid_auth", "Invalid auth data", message.RequestID)
		return
	}

	// In production, validate the token properly
	c.userID = "authenticated_user" // Set from token claims
	c.sendMessage(WebSocketMessage{
		Type:      "auth_success",
		RequestID: message.RequestID,
		Timestamp: time.Now().UTC(),
		Data: map[string]interface{}{
			"user_id": c.userID,
		},
	})
}

func (c *WebSocketClient) sendError(code, message, requestID string) {
	c.sendMessage(WebSocketMessage{
		Type:      "error",
		RequestID: requestID,
		Timestamp: time.Now().UTC(),
		Error:     message,
		Data: map[string]interface{}{
			"code": code,
		},
	})
}

func (c *WebSocketClient) sendMessage(message WebSocketMessage) {
	select {
	case c.send <- message:
	default:
		// Client buffer full, disconnect
		c.manager.unregister <- c
		c.conn.Close()
	}
}

func (c *WebSocketClient) isValidChannel(channel string) bool {
	validChannels := map[string]bool{
		"system":        true,
		"notifications": true,
		"metrics":       true,
		"nodes":         true,
		"alerts":        true,
		"consensus":     true,
		"network":       true,
		"calculations":  true,
		"validations":   true,
	}
	return validChannels[channel]
}

func (c *WebSocketClient) getAvailableChannels() []string {
	return []string{
		"system",
		"notifications", 
		"metrics",
		"nodes",
		"alerts",
		"consensus",
		"network",
		"calculations",
		"validations",
	}
}

func (wm *WebSocketManager) startBroadcasts() {
	go wm.broadcastMetrics()
	go wm.broadcastNodeUpdates()
	go wm.broadcastAlerts()
	go wm.broadcastSystemStatus()
}

func (wm *WebSocketManager) broadcastMetrics() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		metrics := wm.engine.GetNetworkMetrics()
		wm.BroadcastToChannel("metrics", "metrics_update", metrics)
	}
}

func (wm *WebSocketManager) broadcastNodeUpdates() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		nodes := wm.topology.GetAllNodes()
		activeNodes := wm.topology.GetActiveNodes()
		
		wm.BroadcastToChannel("nodes", "nodes_update", map[string]interface{}{
			"total_nodes":  len(nodes),
			"active_nodes": len(activeNodes),
			"nodes":        nodes,
		})
	}
}

func (wm *WebSocketManager) broadcastAlerts() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// In production, get real alerts from monitoring system
		alerts := []map[string]interface{}{
			{
				"id":        "alert_1",
				"type":      "info",
				"message":   "System operating normally",
				"timestamp": time.Now().UTC(),
			},
		}
		
		wm.BroadcastToChannel("alerts", "alerts_update", alerts)
	}
}

func (wm *WebSocketManager) broadcastSystemStatus() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		status := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
			"uptime":    time.Since(wm.startTime).String(),
		}
		
		wm.BroadcastToChannel("system", "system_status", status)
	}
}

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// Server integration methods
func (s *Server) websocketHandler(c *gin.Context) {
	s.websocketManager.HandleConnection(c)
}

func (s *Server) setupWebSocketRoutes() {
	ws := s.router.Group("/ws")
	{
		ws.GET("", s.websocketHandler)
	}
}

func (s *Server) initializeWebSocketManager() {
	s.websocketManager = NewWebSocketManager(s.engine, s.topologyManager, s.logger)
	s.websocketManager.Start()
}