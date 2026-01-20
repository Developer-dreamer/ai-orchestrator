package websocket

import (
	"ai-orchestrator/internal/common/logger"
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

var ErrNilHub = errors.New("hub is nil")

type ConnectionHub interface {
	Add(userID string, conn *websocket.Conn) *Client
	Remove(userID string)
	GetClientByID(userID string) ([]*Client, error)
	GetAllClients() map[string][]*Client
}

var ErrNilUpgrader = errors.New("websocket upgrader is nil")

type Manager struct {
	logger   logger.Logger
	upgrader *websocket.Upgrader
	clients  *Hub
	mu       sync.RWMutex
}

func NewManager(l logger.Logger, upgrader *websocket.Upgrader, hub *Hub) (*Manager, error) {
	if l == nil {
		return nil, logger.ErrNilLogger
	}
	if upgrader == nil {
		return nil, ErrNilUpgrader
	}
	if hub == nil {
		return nil, ErrNilHub
	}

	return &Manager{
		logger:   l,
		upgrader: upgrader,
		clients:  hub,
		mu:       sync.RWMutex{},
	}, nil
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
		m.logger.ErrorContext(ctx, "No clients found for userID", "userID", userID)
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
