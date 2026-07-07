package ws

import (
	"encoding/json"
	"fmt"
	"sync"
)

type Manager struct {
	clients sync.Map
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Register(client *Client) {
	m.clients.Store(client.ID, client)
	fmt.Println("websocket client connected:", client.ID)
}

func (m *Manager) Unregister(clientID string) {
	value, ok := m.clients.LoadAndDelete(clientID)
	if !ok {
		return
	}

	client, ok := value.(*Client)
	if !ok {
		return
	}

	client.Close()
	fmt.Println("websocket client disconnected:", clientID)
}

func (m *Manager) BroadcastJSON(payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("failed to marshal broadcast payload:", err)
		return
	}

	m.clients.Range(func(key, value any) bool {
		client, ok := value.(*Client)
		if !ok {
			return true
		}

		select {
		case client.Send <- data:
		default:
			m.Unregister(client.ID)
		}

		return true
	})
}

func (m *Manager) Count() int {
	count := 0

	m.clients.Range(func(_, _ any) bool {
		count++
		return true
	})

	return count
}
