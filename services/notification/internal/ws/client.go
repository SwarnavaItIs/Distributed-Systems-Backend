package ws

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID      string
	Conn    *websocket.Conn
	Send    chan []byte
	manager *Manager
	once    sync.Once
}

func NewClient(id string, conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		ID:      id,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		manager: manager,
	}
}

func (c *Client) ReadPump() {
	defer c.manager.Unregister(c.ID)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			fmt.Println("websocket read failed:", err)
			return
		}

		c.manager.BroadcastJSON(map[string]any{
			"type":      "client.message",
			"client_id": c.ID,
			"message":   string(message),
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

func (c *Client) WritePump() {
	for message := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			fmt.Println("websocket write failed:", err)
			return
		}
	}
}

func (c *Client) SendJSON(payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("failed to marshal websocket message:", err)
		return
	}

	select {
	case c.Send <- data:
	default:
		c.manager.Unregister(c.ID)
	}
}

func (c *Client) Close() {
	c.once.Do(func() {
		close(c.Send)
		c.Conn.Close()
	})
}