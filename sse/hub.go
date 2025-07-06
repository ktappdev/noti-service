package sse

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/ktappdev/noti-service/models"
)

// SSEClient represents a connected SSE client
type SSEClient struct {
	UserID  string
	Channel chan []byte
	Done    chan bool
	ID      string
}

// SSEHub manages SSE connections and broadcasts
type SSEHub struct {
	clients    map[string][]*SSEClient // userID -> []*SSEClient
	register   chan *SSEClient
	unregister chan *SSEClient
	broadcast  chan models.NotificationMessage
	mutex      sync.RWMutex
}

// NewSSEHub creates a new SSE hub
func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients:    make(map[string][]*SSEClient),
		register:   make(chan *SSEClient),
		unregister: make(chan *SSEClient),
		broadcast:  make(chan models.NotificationMessage),
	}
}

// Run starts the SSE hub event loop
func (h *SSEHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make([]*SSEClient, 0)
			}
			h.clients[client.UserID] = append(h.clients[client.UserID], client)
			h.mutex.Unlock()
			log.Printf("SSE client registered for user %s (total clients: %d)", client.UserID, len(h.clients[client.UserID]))

		case client := <-h.unregister:
			h.mutex.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				for i, c := range clients {
					if c.ID == client.ID {
						// Remove client from slice
						h.clients[client.UserID] = append(clients[:i], clients[i+1:]...)
						close(c.Channel)
						close(c.Done)
						break
					}
				}
				// Remove user entry if no clients left
				if len(h.clients[client.UserID]) == 0 {
					delete(h.clients, client.UserID)
				}
			}
			h.mutex.Unlock()
			log.Printf("SSE client unregistered for user %s", client.UserID)

		case message := <-h.broadcast:
			h.mutex.RLock()
			if clients, ok := h.clients[message.UserID]; ok {
				messageBytes, err := json.Marshal(message)
				if err != nil {
					log.Printf("Error marshaling SSE message: %v", err)
					h.mutex.RUnlock()
					continue
				}

				sseData := fmt.Sprintf("data: %s\n\n", messageBytes)
				for _, client := range clients {
					select {
					case client.Channel <- []byte(sseData):
					default:
						// Client channel is full, skip
						log.Printf("SSE client channel full for user %s", message.UserID)
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastToUser sends a notification to all connected clients for a specific user
func (h *SSEHub) BroadcastToUser(userID string, event string, notificationType string, notification interface{}) {
	message := models.NotificationMessage{
		UserID:       userID,
		Type:         notificationType,
		Notification: notification,
		Event:        event,
	}

	select {
	case h.broadcast <- message:
	default:
		log.Printf("SSE broadcast channel full, dropping message for user %s", userID)
	}
}

// RegisterClient registers a new SSE client
func (h *SSEHub) RegisterClient(client *SSEClient) {
	h.register <- client
}

// UnregisterClient unregisters an SSE client
func (h *SSEHub) UnregisterClient(client *SSEClient) {
	h.unregister <- client
}