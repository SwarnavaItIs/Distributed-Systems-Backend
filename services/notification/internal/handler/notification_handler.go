package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	notificationws "github.com/swarnava/dmb/services/notification/internal/ws"
)

type NotificationHandler struct {
	manager *notificationws.Manager
}

func NewNotificationHandler(manager *notificationws.Manager) *NotificationHandler {
	return &NotificationHandler{
		manager: manager,
	}
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

	writeJSON(w, http.StatusOK, map[string]any{
		"status":            "ok",
		"service":           "notification",
		"connected_clients": h.manager.Count(),
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

	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())

	client := notificationws.NewClient(clientID, conn, h.manager)

	h.manager.Register(client)

	go client.WritePump()

	client.SendJSON(map[string]any{
		"type":      "connection.established",
		"client_id": clientID,
		"message":   "Connected to DMB Notification Service",
		"timestamp": time.Now().Format(time.RFC3339),
	})

	client.ReadPump()
}

func (h *NotificationHandler) BroadcastHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Message == "" {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}

	h.manager.BroadcastJSON(map[string]any{
		"type":      "manual.broadcast",
		"message":   req.Message,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"status":            "sent",
		"connected_clients": h.manager.Count(),
	})
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
