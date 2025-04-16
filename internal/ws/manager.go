package ws

import (
	"github.com/gorilla/websocket"
	"sync"
)

type Client struct {
	Conn   *websocket.Conn
	UserID string
}

type Manager struct {
	clients map[string]*Client
	lock    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*Client),
	}
}

func (m *Manager) AddClient(userID string, conn *websocket.Conn) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.clients[userID] = &Client{Conn: conn, UserID: userID}
}

func (m *Manager) RemoveClient(userID string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.clients, userID)
}

func (m *Manager) SendTo(userID, message string) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if client, ok := m.clients[userID]; ok {
		client.Conn.WriteJSON(map[string]string{
			"type": "match",
			"msg":  message,
		})
	}
}
