package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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
		
		// Write initial message and flush immediately
		initialResponse := fmt.Sprintf("data: %s\n\n", initialData)
		c.Write([]byte(initialResponse))
		
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
			
			// Write the initial message first
			if _, err := w.WriteString(initialResponse); err != nil {
				log.Printf("Error writing initial SSE message: %v", err)
				return
			}
			if err := w.Flush(); err != nil {
				log.Printf("Error flushing initial SSE message: %v", err)
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
						log.Printf("Error writing SSE message for user %s: %v", userID, err)
						return
					}
					if err := w.Flush(); err != nil {
						log.Printf("Error flushing SSE message for user %s: %v", userID, err)
						return
					}
					
				case <-time.After(30 * time.Second):
					// Send heartbeat every 30 seconds to keep connection alive
					heartbeat := fmt.Sprintf("data: {\"type\": \"heartbeat\", \"timestamp\": \"%s\"}\n\n", 
						time.Now().Format(time.RFC3339))
					if _, err := w.WriteString(heartbeat); err != nil {
						log.Printf("Error writing heartbeat for user %s: %v", userID, err)
						return
					}
					if err := w.Flush(); err != nil {
						log.Printf("Error flushing heartbeat for user %s: %v", userID, err)
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

// StreamNotificationsGin handles SSE connections for real-time notifications (Gin version)
func StreamNotificationsGin(db *sqlx.DB, hub *sse.SSEHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id query parameter is required"})
			return
		}

		w := c.Writer
		// r := c.Request // Remove unused variable

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("X-Accel-Buffering", "no")

		clientID := fmt.Sprintf("%s_%d", userID, time.Now().UnixNano())
		client := &sse.SSEClient{
			UserID:  userID,
			Channel: make(chan []byte, 10),
			Done:    make(chan bool),
			ID:      clientID,
		}
		hub.RegisterClient(client)

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
		initialResponse := fmt.Sprintf("data: %s\n\n", initialData)

		flusher, ok := w.(http.Flusher)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming unsupported"})
			hub.UnregisterClient(client)
			return
		}

		// Write initial message and flush
		w.Write([]byte(initialResponse))
		flusher.Flush()

		// Send existing notifications after initial message
		go func() {
			time.Sleep(100 * time.Millisecond)
			sendExistingNotifications(db, client)
		}()

		closeNotify := c.Writer.CloseNotify()
		for {
			select {
			case message, ok := <-client.Channel:
				if !ok {
					hub.UnregisterClient(client)
					return
				}
				w.Write(message)
				flusher.Flush()
			case <-time.After(30 * time.Second):
				heartbeat := fmt.Sprintf("data: {\"type\": \"heartbeat\", \"timestamp\": \"%s\"}\n\n", time.Now().Format(time.RFC3339))
				w.Write([]byte(heartbeat))
				flusher.Flush()
			case <-client.Done:
				hub.UnregisterClient(client)
				return
			case <-closeNotify:
				hub.UnregisterClient(client)
				return
			}
		}
	}
}

// SSEHelpHandlerGin provides SSE documentation/help (Gin version)
func SSEHelpHandlerGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		help := map[string]interface{}{
			"message": "This endpoint provides Server-Sent Events (SSE) for real-time notifications.",
			"usage":   "/notifications/stream?user_id={USER_ID}",
			"note":    "You must provide a valid user_id as a query parameter.",
		}
		c.JSON(200, help)
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