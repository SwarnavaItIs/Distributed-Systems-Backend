package ws

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 512
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

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))

	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

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
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	defer c.manager.Unregister(c.ID)

	for {
		select {
		case message := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				fmt.Println("websocket write failed:", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Println("websocket ping failed:", err)
				return
			}
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
		c.Conn.Close()
	})
}
