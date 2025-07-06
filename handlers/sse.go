package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/ktappdev/noti-service/models"
	"github.com/ktappdev/noti-service/sse"
	"github.com/valyala/fasthttp"
)

// StreamNotifications handles SSE connections for real-time notifications
func StreamNotifications(db *sqlx.DB, hub *sse.SSEHub) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Query("user_id")
		if userID == "" {
			return c.Status(400).SendString("user_id query parameter is required")
		}

		// Set SSE headers
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		// CORS headers are handled by the main middleware, don't override here

		// Generate unique client ID
		clientID := fmt.Sprintf("%s_%d", userID, time.Now().UnixNano())

		// Create SSE client
		client := &sse.SSEClient{
			UserID:  userID,
			Channel: make(chan []byte, 10),
			Done:    make(chan bool),
			ID:      clientID,
		}

		// Register client
		hub.RegisterClient(client)

		// Send initial connection message immediately to establish stream
		initialMsg := models.NotificationMessage{
			UserID: userID,
			Type:   "system",
			Event:  "connected",
			Notification: map[string]string{
				"message": "Connected to notification stream",
				"time":    time.Now().Format(time.RFC3339),
			},
		}
		initialData, _ := json.Marshal(initialMsg)
		c.Write([]byte(fmt.Sprintf("data: %s\n\n", initialData)))

		// Send existing unread notifications on connection
		go func() {
			time.Sleep(100 * time.Millisecond) // Small delay to ensure connection is established
			sendExistingNotifications(db, client)
		}()

		// Handle client disconnect
		defer func() {
			hub.UnregisterClient(client)
		}()

		// Keep connection alive and send notifications
		c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("SSE StreamWriter panic recovered: %v", r)
				}
			}()
			
			for {
				select {
				case message, ok := <-client.Channel:
					if !ok {
						// Channel closed, exit gracefully
						return
					}
					if _, err := w.Write(message); err != nil {
						log.Printf("Error writing SSE message: %v", err)
						return
					}
					if err := w.Flush(); err != nil {
						log.Printf("Error flushing SSE message: %v", err)
						return
					}
				case <-client.Done:
					return
				case <-c.Context().Done():
					return
				}
			}
		}))

		return nil
	}
}

func sendExistingNotifications(db *sqlx.DB, client *sse.SSEClient) {
	// Get unread user notifications
	userQuery := `SELECT * FROM user_notifications
                  WHERE parent_user_id = $1 AND read = false
                  ORDER BY created_at DESC`
	var userNotifications []models.UserNotification
	err := db.Select(&userNotifications, userQuery, client.UserID)
	if err != nil {
		log.Printf("Error fetching user notifications for SSE: %v", err)
	}

	// Get unread owner notifications
	ownerQuery := `SELECT * FROM product_owner_notifications
                   WHERE owner_id = $1 AND read = false
                   ORDER BY created_at DESC`
	var ownerNotifications []models.ProductOwnerNotification
	err = db.Select(&ownerNotifications, ownerQuery, client.UserID)
	if err != nil {
		log.Printf("Error fetching owner notifications for SSE: %v", err)
	}

	// Send user notifications
	for _, notification := range userNotifications {
		message := models.NotificationMessage{
			UserID:       client.UserID,
			Type:         "user",
			Event:        "existing_notification",
			Notification: notification,
		}
		messageBytes, _ := json.Marshal(message)
		sseData := fmt.Sprintf("data: %s\n\n", messageBytes)

		select {
		case client.Channel <- []byte(sseData):
		default:
			log.Printf("Failed to send existing user notification to client %s", client.ID)
		}
	}

	// Send owner notifications
	for _, notification := range ownerNotifications {
		message := models.NotificationMessage{
			UserID:       client.UserID,
			Type:         "owner",
			Event:        "existing_notification",
			Notification: notification,
		}
		messageBytes, _ := json.Marshal(message)
		sseData := fmt.Sprintf("data: %s\n\n", messageBytes)

		select {
		case client.Channel <- []byte(sseData):
		default:
			log.Printf("Failed to send existing owner notification to client %s", client.ID)
		}
	}
}