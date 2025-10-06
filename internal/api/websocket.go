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
)

type WebSocketManager struct {
	upgrader  websocket.Upgrader
	clients   map[*WebSocketClient]bool
	broadcast chan []byte
	register  chan *WebSocketClient
	unregister chan *WebSocketClient
	logger    *zap.Logger
	mu        sync.RWMutex
}

type WebSocketClient struct {
	conn     *websocket.Conn
	send     chan []byte
	manager  *WebSocketManager
	userID   string
	channels map[string]bool
}

type WebSocketMessage struct {
	Type    string          `json:"type"`
	Channel string          `json:"channel,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

func NewWebSocketManager(logger *zap.Logger) *WebSocketManager {
	return &WebSocketManager{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients:    make(map[*WebSocketClient]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *WebSocketClient),
		unregister: make(chan *WebSocketClient),
		logger:     logger,
	}
}

func (wm *WebSocketManager) Start() {
	go wm.run()
}

func (wm *WebSocketManager) run() {
	for {
		select {
		case client := <-wm.register:
			wm.mu.Lock()
			wm.clients[client] = true
			wm.mu.Unlock()
			wm.logger.Info("WebSocket client connected", zap.String("user_id", client.userID))

		case client := <-wm.unregister:
			wm.mu.Lock()
			if _, ok := wm.clients[client]; ok {
				delete(wm.clients, client)
				close(client.send)
			}
			wm.mu.Unlock()
			wm.logger.Info("WebSocket client disconnected", zap.String("user_id", client.userID))

		case message := <-wm.broadcast:
			wm.mu.RLock()
			for client := range wm.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(wm.clients, client)
				}
			}
			wm.mu.RUnlock()
		}
	}
}

func (wm *WebSocketManager) Broadcast(message interface{}) error {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	wm.broadcast <- msgBytes
	return nil
}

func (wm *WebSocketManager) BroadcastToChannel(channel string, message interface{}) error {
	msg := WebSocketMessage{
		Type:    "channel_message",
		Channel: channel,
		Data:    message,
	}

	return wm.Broadcast(msg)
}

func (wm *WebSocketManager) HandleConnection(c *gin.Context) {
	conn, err := wm.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		wm.logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	client := &WebSocketClient{
		conn:     conn,
		send:     make(chan []byte, 256),
		manager:  wm,
		userID:   userID,
		channels: make(map[string]bool),
	}

	wm.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *WebSocketClient) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512 * 1024)
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
				c.manager.logger.Warn("WebSocket read error", zap.Error(err))
			}
			break
		}

		c.handleMessage(&message)
	}
}

func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
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
		c.sendPong()
	case "auth":
		c.handleAuth(message)
	default:
		c.sendError("unknown_message_type", "Unknown message type")
	}
}

func (c *WebSocketClient) handleSubscribe(message *WebSocketMessage) {
	var data struct {
		Channels []string `json:"channels"`
	}
	
	if err := json.Unmarshal(message.Data, &data); err != nil {
		c.sendError("invalid_subscription", "Invalid subscription data")
		return
	}

	for _, channel := range data.Channels {
		if c.isValidChannel(channel) {
			c.channels[channel] = true
			c.manager.logger.Debug("Client subscribed to channel", 
				zap.String("user_id", c.userID),
				zap.String("channel", channel),
			)
		}
	}

	c.sendMessage(WebSocketMessage{
		Type: "subscribed",
		Data: data,
	})
}

func (c *WebSocketClient) handleUnsubscribe(message *WebSocketMessage) {
	var data struct {
		Channels []string `json:"channels"`
	}
	
	if err := json.Unmarshal(message.Data, &data); err != nil {
		c.sendError("invalid_unsubscription", "Invalid unsubscription data")
		return
	}

	for _, channel := range data.Channels {
		delete(c.channels, channel)
		c.manager.logger.Debug("Client unsubscribed from channel", 
			zap.String("user_id", c.userID),
			zap.String("channel", channel),
		)
	}

	c.sendMessage(WebSocketMessage{
		Type: "unsubscribed",
		Data: data,
	})
}

func (c *WebSocketClient) handleAuth(message *WebSocketMessage) {
	var data struct {
		Token string `json:"token"`
	}
	
	if err := json.Unmarshal(message.Data, &data); err != nil {
		c.sendError("invalid_auth", "Invalid auth data")
		return
	}

	c.sendMessage(WebSocketMessage{
		Type: "auth_success",
	})
}

func (c *WebSocketClient) sendPong() {
	c.sendMessage(WebSocketMessage{
		Type: "pong",
	})
}

func (c *WebSocketClient) sendError(code, message string) {
	c.sendMessage(WebSocketMessage{
		Type:  "error",
		Error: message,
	})
}

func (c *WebSocketClient) sendMessage(message WebSocketMessage) {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		c.manager.logger.Error("Failed to marshal WebSocket message", zap.Error(err))
		return
	}

	select {
	case c.send <- msgBytes:
	default:
		c.manager.unregister <- c
		c.conn.Close()
	}
}

func (c *WebSocketClient) isValidChannel(channel string) bool {
	validChannels := map[string]bool{
		"nodes":      true,
		"metrics":    true,
		"alerts":     true,
		"consensus":  true,
		"network":    true,
	}
	return validChannels[channel]
}

func (s *Server) websocketHandler(c *gin.Context) {
	s.websocketManager.HandleConnection(c)
}

func (s *Server) setupWebSocketRoutes() {
	ws := s.router.Group("/ws")
	{
		ws.GET("", s.websocketHandler)
		ws.GET("/metrics", s.websocketHandler)
		ws.GET("/alerts", s.websocketHandler)
	}
}

func (s *Server) startWebSocketBroadcasts() {
	go s.broadcastMetrics()
	go s.broadcastAlerts()
	go s.broadcastNodeUpdates()
}

func (s *Server) broadcastMetrics() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		metrics := s.engine.GetNetworkMetrics()
		s.websocketManager.BroadcastToChannel("metrics", metrics)
	}
}

func (s *Server) broadcastAlerts() {
	// Implementation for broadcasting alerts
}

func (s *Server) broadcastNodeUpdates() {
	// Implementation for broadcasting node updates
}