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
	Done    chan struct{}
	manager *Manager
	once    sync.Once
}

func NewClient(
	id string,
	conn *websocket.Conn,
	manager *Manager,
) *Client {
	return &Client{
		ID:      id,
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Done:    make(chan struct{}),
		manager: manager,
	}
}

func (c *Client) ReadPump() {
	defer c.manager.Unregister(c.ID)

	c.Conn.SetReadLimit(maxMessageSize)

	if err := c.Conn.SetReadDeadline(
		time.Now().Add(pongWait),
	); err != nil {
		fmt.Println("failed to set read deadline:", err)
		return
	}

	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(
			time.Now().Add(pongWait),
		)
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			fmt.Println("websocket read stopped:", err)
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
		case <-c.Done:
			return

		case message := <-c.Send:
			if err := c.Conn.SetWriteDeadline(
				time.Now().Add(writeWait),
			); err != nil {
				return
			}

			if err := c.Conn.WriteMessage(
				websocket.TextMessage,
				message,
			); err != nil {
				fmt.Println("websocket write failed:", err)
				return
			}

		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(
				time.Now().Add(writeWait),
			); err != nil {
				return
			}

			if err := c.Conn.WriteMessage(
				websocket.PingMessage,
				nil,
			); err != nil {
				fmt.Println("websocket ping failed:", err)
				return
			}
		}
	}
}

func (c *Client) SendJSON(payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(
			"failed to marshal websocket message:",
			err,
		)
		return
	}

	select {
	case <-c.Done:
		return

	default:
	}

	select {
	case <-c.Done:
		return

	case c.Send <- data:

	default:
		c.manager.Unregister(c.ID)
	}
}

func (c *Client) Close() {
	c.once.Do(func() {
		close(c.Done)

		deadline := time.Now().Add(writeWait)

		err := c.Conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(
				websocket.CloseNormalClosure,
				"server shutting down",
			),
			deadline,
		)
		if err != nil {
			fmt.Println(
				"failed to send websocket close frame:",
				err,
			)
		}

		if err := c.Conn.Close(); err != nil {
			fmt.Println(
				"failed to close websocket connection:",
				err,
			)
		}
	})
}
