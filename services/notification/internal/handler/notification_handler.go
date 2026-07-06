package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type NotificationHandler struct{}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *NotificationHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "notification",
	})
}

func (h *NotificationHandler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("websocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	welcomeMessage := map[string]any{
		"type":      "connection.established",
		"message":   "Connected to DMB Notification Service",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if err := conn.WriteJSON(welcomeMessage); err != nil {
		fmt.Println("failed to write welcome message:", err)
		return
	}

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("websocket client disconnected:", err)
			return
		}

		echoMessage := map[string]any{
			"type":      "echo",
			"message":   string(message),
			"timestamp": time.Now().Format(time.RFC3339),
		}

		if err := conn.WriteJSON(echoMessage); err != nil {
			fmt.Println("failed to write echo message:", err)
			return
		}

		if messageType == websocket.CloseMessage {
			return
		}
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{
		"error": message,
	})
}