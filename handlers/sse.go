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

		// Set SSE headers for proper streaming
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")
		c.Set("X-Accel-Buffering", "no")  // Disable proxy buffering
		// Don't set Content-Length - let it stream
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

		// Prepare initial connection message to be sent inside StreamWriter
		initialMsg := models.NotificationMessage{
			UserID: userID,
			Type:   "connection",
			Event:  "stream_established",
			Notification: map[string]string{
				"status": "connected",
				"time":   time.Now().Format(time.RFC3339),
			},
		}
		initialData, _ := json.Marshal(initialMsg)
		initialResponse := fmt.Sprintf("data: %s\n\n", initialData)
		
		// Use the working streaming approach without problematic channels
		c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("SSE StreamWriter panic recovered: %v", r)
				}
				// Ensure client is unregistered on exit
				hub.UnregisterClient(client)
			}()
			
			// Validate writer
			if w == nil {
				log.Printf("SSE StreamWriter received nil buffer for user %s", userID)
				return
			}
			
			// Write the initial message first to establish the connection
			if _, err := w.WriteString(initialResponse); err != nil {
				// Connection failed immediately, exit gracefully
				return
			}
			if err := w.Flush(); err != nil {
				// Connection failed immediately, exit gracefully
				return
			}
			
			
			// Send existing notifications after initial message
			go func() {
				time.Sleep(100 * time.Millisecond)
				sendExistingNotifications(db, client)
			}()
			
			// Use polling approach instead of select with channels
			for {
				// Check for new messages with timeout
				select {
				case message, ok := <-client.Channel:
					if !ok {
						return
					}
					if _, err := w.Write(message); err != nil {
						// Connection is closed, exit gracefully
						return
					}
					if err := w.Flush(); err != nil {
						// Connection is closed, exit gracefully
						return
					}
					
				case <-time.After(30 * time.Second):
					// Send heartbeat every 30 seconds to keep connection alive
					heartbeat := fmt.Sprintf("data: {\"type\": \"heartbeat\", \"timestamp\": \"%s\"}\n\n", 
						time.Now().Format(time.RFC3339))
					if _, err := w.WriteString(heartbeat); err != nil {
						// Connection is closed, exit gracefully without logging error
						return
					}
					if err := w.Flush(); err != nil {
						// Connection is closed, exit gracefully without logging error
						return
					}
				}
				
				// Check if client is done (non-blocking)
				select {
				case <-client.Done:
					return
				default:
					// Continue loop
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

	// Get unread like notifications
	likeQuery := `SELECT * FROM like_notifications
                  WHERE target_user_id = $1 AND read = false
                  ORDER BY created_at DESC`
	var likeNotifications []models.LikeNotification
	err = db.Select(&likeNotifications, likeQuery, client.UserID)
	if err != nil {
		log.Printf("Error fetching like notifications for SSE: %v", err)
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

	// Send like notifications
	for _, notification := range likeNotifications {
		message := models.NotificationMessage{
			UserID:       client.UserID,
			Type:         "like",
			Event:        "existing_notification",
			Notification: notification,
		}
		messageBytes, _ := json.Marshal(message)
		sseData := fmt.Sprintf("data: %s\n\n", messageBytes)

		select {
		case client.Channel <- []byte(sseData):
		default:
			log.Printf("Failed to send existing like notification to client %s", client.ID)
		}
	}
}