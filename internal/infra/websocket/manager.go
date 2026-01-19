package websocket

import (
	"ai-orchestrator/internal/common"
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

type ConnectionHub interface {
	Add(userID string, conn *websocket.Conn) *Client
	Remove(userID string)
	GetClientByID(userID string) ([]*Client, error)
	GetAllClients() map[string][]*Client
}

type Manager struct {
	logger   common.Logger
	upgrader *websocket.Upgrader
	clients  *Hub
	mu       sync.RWMutex
}

func NewManager(logger common.Logger, upgrader *websocket.Upgrader, hub *Hub) *Manager {
	return &Manager{
		logger:   logger,
		upgrader: upgrader,
		clients:  hub,
		mu:       sync.RWMutex{},
	}
}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "Missing userID", http.StatusUnauthorized)
		return
	}

	// TODO Check if user exists in db

	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.logger.Error("Upgrade failed", "error", err)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	client := m.clients.Add(userID, conn)
	go client.writePump()
}

func (m *Manager) SendToClient(ctx context.Context, userID string, data json.RawMessage) error {
	m.mu.RLock()
	clients, err := m.clients.GetClientByID(userID)
	m.mu.RUnlock()
	if err != nil {
		m.logger.ErrorContext(ctx, "Not clients found for userID", "userID", userID)
		return err
	}

	for _, client := range clients {
		err = client.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			m.logger.ErrorContext(ctx, "Error sending to client", "error", err, "userID", userID)
			return err
		}
	}

	return nil
}
