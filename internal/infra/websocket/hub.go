package websocket

import (
	"errors"
	"github.com/gorilla/websocket"
)

var (
	ErrClientNotFound = errors.New("client not found")
)

type Hub struct {
	// UserID -> Connection list
	// "uuid-123" -> [conn1 (phone), conn2 (laptop)]
	clients map[string][]*Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string][]*Client),
	}
}

func (hub *Hub) Add(userID string, conn *websocket.Conn) *Client {
	client := NewClient(conn)
	hub.clients[userID] = append(hub.clients[userID], client)

	return client
}

func (hub *Hub) Remove(userID string) {
	delete(hub.clients, userID)
}

func (hub *Hub) GetClientByID(userId string) ([]*Client, error) {
	clients := hub.clients[userId]
	if clients == nil {
		return nil, ErrClientNotFound
	}

	return clients, nil
}

func (hub *Hub) GetAllClients() map[string][]*Client {
	return hub.clients
}
