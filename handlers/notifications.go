package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/ktappdev/noti-service/models"
	"github.com/ktappdev/noti-service/reviewit"
	"github.com/ktappdev/noti-service/sse"
)

// CreateProductOwnerNotification creates a new product owner notification
func CreateProductOwnerNotification(db *sqlx.DB, hub *sse.SSEHub) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fmt.Println("createProductOwnerNotification")
		notification := new(models.ProductOwnerNotification)
		if err := c.BodyParser(notification); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// Set the notification type
		notification.NotificationType = "review"
		fmt.Println("this is the notification", notification)

		// Check if the owner exists
		var exists bool
		err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", notification.OwnerID)
		if err != nil {
			log.Println(err)
			return c.Status(500).SendString(err.Error())
		}
		if !exists {
			log.Println("owner does not exist")
			return c.Status(400).SendString("Owner does not exist")
		}

		query := `INSERT INTO product_owner_notifications (id, owner_id, product_id, product_name, business_id, review_title, from_name, from_id, read, comment_id, review_id, notification_type)
	              VALUES (:id, :owner_id, :product_id, :product_name, :business_id, :review_title, :from_name, :from_id, :read, :comment_id, :review_id, :notification_type) RETURNING id, created_at`
		rows, err := db.NamedQuery(query, notification)
		if err != nil {
			log.Printf("Error creating notification: %v", err)
			return c.Status(500).SendString(err.Error())
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&notification.ID, &notification.CreatedAt)
			if err != nil {
				log.Printf("Error scanning notification result: %v", err)
				return c.Status(500).SendString(err.Error())
			}
		}

		// Broadcast to SSE clients
		hub.BroadcastToUser(notification.OwnerID, "new_notification", "owner", notification)

		return c.Status(201).JSON(notification)
	}
}

// CreateReplyNotification creates a new reply notification
func CreateReplyNotification(db *sqlx.DB, hub *sse.SSEHub) fiber.Handler {
	return func(c *fiber.Ctx) error {
		notification := new(models.UserNotification)
		if err := c.BodyParser(notification); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// Set the notification type
		notification.NotificationType = "reply"
		log.Printf("Creating reply notification: %+v", notification)

		parentUserID, err := reviewit.GetParentCommentUserID(notification.ParentID)
		if err != nil {
			log.Printf("ERROR getting parent user ID: %v", err)
			// Check if it's a missing environment variable error
			if err.Error() == "REVIEWIT_DATABASE_URL environment variable is required" {
				return c.Status(500).SendString("ReviewIt database connection not configured. Please set REVIEWIT_DATABASE_URL environment variable.")
			}
			return c.Status(500).SendString(fmt.Sprintf("Error getting parent user ID: %v", err))
		}
		notification.ParentUserID = parentUserID

		// Check if the user exists
		var exists bool
		err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", notification.ParentUserID)
		if err != nil {
			log.Printf("Error checking user existence: %v", err)
			return c.Status(500).SendString("Internal server error")
		}
		if !exists {
			log.Println("user don't exist")
			return c.Status(400).SendString("User does not exist")
		}

		// Insert the notification
		query := `INSERT INTO user_notifications (id, parent_user_id, content, read, notification_type, comment_id, from_id, review_id, parent_id, from_name, product_id)
	  VALUES (:id, :parent_user_id, :content, :read, :notification_type, :comment_id, :from_id, :review_id, :parent_id, :from_name, :product_id) RETURNING id, created_at`
		rows, err := db.NamedQuery(query, notification)
		if err != nil {
			log.Printf("Error inserting notification: %v", err)
			return c.Status(500).SendString("Failed to create notification")
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&notification.ID, &notification.CreatedAt)
			if err != nil {
				log.Printf("Error scanning notification result: %v", err)
				return c.Status(500).SendString("Failed to retrieve created notification")
			}
		}

		// Broadcast to SSE clients
		hub.BroadcastToUser(notification.ParentUserID, "new_notification", "user", notification)

		return c.Status(201).JSON(notification)
	}
}

// GetLatestNotifications gets the latest notifications for a user
func GetLatestNotifications(db *sqlx.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Query("user_id")
		if userID == "" {
			return c.Status(400).SendString("user_id query parameter is required")
		}

		userQuery := `SELECT * FROM user_notifications
	                  WHERE user_id = $1
	                  ORDER BY created_at DESC
	                  LIMIT 1`
		var userNotification models.UserNotification
		err := db.Get(&userNotification, userQuery, userID)
		if err != nil && err != sql.ErrNoRows {
			return c.Status(500).SendString(err.Error())
		}

		ownerQuery := `SELECT * FROM product_owner_notifications
	                   WHERE owner_id = $1
	                   ORDER BY created_at DESC
	                   LIMIT 1`
		var ownerNotification models.ProductOwnerNotification
		err = db.Get(&ownerNotification, ownerQuery, userID)
		if err != nil && err != sql.ErrNoRows {
			return c.Status(500).SendString(err.Error())
		}

		if userNotification.ID == "" && ownerNotification.ID == "" {
			return c.Status(200).SendString("No notifications found")
		}

		return c.JSON(fiber.Map{
			"user_notification":  userNotification,
			"owner_notification": ownerNotification,
		})
	}
}

// GetAllNotifications gets all notifications for a user
func GetAllNotifications(db *sqlx.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Query("user_id")
		if userID == "" {
			return c.Status(400).SendString("user_id query parameter is required")
		}

		userQuery := `SELECT * FROM user_notifications
	                  WHERE parent_user_id = $1
	                  ORDER BY created_at DESC`
		var userNotifications []models.UserNotification
		err := db.Select(&userNotifications, userQuery, userID)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		ownerQuery := `SELECT * FROM product_owner_notifications
	                   WHERE owner_id = $1
	                   ORDER BY created_at DESC`
		var ownerNotifications []models.ProductOwnerNotification
		err = db.Select(&ownerNotifications, ownerQuery, userID)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		return c.JSON(fiber.Map{
			"user_notifications":  userNotifications,
			"owner_notifications": ownerNotifications,
		})
	}
}

// GetAllUnreadNotifications gets all unread notifications for a user
func GetAllUnreadNotifications(db *sqlx.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Query("user_id")
		if userID == "" {
			return c.Status(400).SendString("user_id query parameter is required")
		}

		userQuery := `SELECT * FROM user_notifications
	                  WHERE parent_user_id = $1 AND read = false
	                  ORDER BY created_at DESC`
		var userNotifications []models.UserNotification
		err := db.Select(&userNotifications, userQuery, userID)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		ownerQuery := `SELECT * FROM product_owner_notifications
	                   WHERE owner_id = $1 AND read = false
	                   ORDER BY created_at DESC`
		var ownerNotifications []models.ProductOwnerNotification
		err = db.Select(&ownerNotifications, ownerQuery, userID)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		return c.JSON(fiber.Map{
			"user_notifications":  userNotifications,
			"owner_notifications": ownerNotifications,
		})
	}
}

// DeleteReadNotifications deletes all read notifications for a user
func DeleteReadNotifications(db *sqlx.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Query("user_id")
		if userID == "" {
			return c.Status(400).SendString("user_id query parameter is required")
		}

		userQuery := `DELETE FROM user_notifications WHERE user_id = $1 AND read = true RETURNING *`
		var deletedUserNotifications []models.UserNotification
		err := db.Select(&deletedUserNotifications, userQuery, userID)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		ownerQuery := `DELETE FROM product_owner_notifications WHERE owner_id = $1 AND read = true RETURNING *`
		var deletedOwnerNotifications []models.ProductOwnerNotification
		err = db.Select(&deletedOwnerNotifications, ownerQuery, userID)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		return c.JSON(fiber.Map{
			"deleted_user_notifications":  len(deletedUserNotifications),
			"deleted_owner_notifications": len(deletedOwnerNotifications),
			"user_notifications":          deletedUserNotifications,
			"owner_notifications":         deletedOwnerNotifications,
		})
	}
}

// MarkNotificationAsRead marks a notification as read
func MarkNotificationAsRead(db *sqlx.DB, hub *sse.SSEHub) fiber.Handler {
	return func(c *fiber.Ctx) error {
		notificationID := c.Params("id")
		notificationType := c.Query("type")

		if notificationID == "" {
			return c.Status(400).SendString("Notification ID is required")
		}

		if notificationType == "" {
			return c.Status(400).SendString("Notification type is required")
		}

		fmt.Printf("%s - %s", notificationID, notificationType)
		var query string
		var result sql.Result
		var err error

		switch notificationType {
		case "user":
			query = "UPDATE user_notifications SET read = true WHERE id = $1"
		case "owner":
			query = "UPDATE product_owner_notifications SET read = true WHERE id = $1"
		default:
			return c.Status(400).SendString("Invalid notification type")
		}

		result, err = db.Exec(query, notificationID)
		if err != nil {
			log.Printf("Error updating notification: %v", err)
			return c.Status(500).SendString("Failed to update notification")
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("Error getting rows affected: %v", err)
			return c.Status(500).SendString("Failed to get update result")
		}

		if rowsAffected == 0 {
			return c.Status(404).SendString("Notification not found")
		}

		// Broadcast read status to SSE clients
		readMessage := map[string]interface{}{
			"notification_id": notificationID,
			"type":           notificationType,
			"read":           true,
			"timestamp":      time.Now().Format(time.RFC3339),
		}

		// We need to get the user ID for this notification to broadcast properly
		var userID string
		if notificationType == "user" {
			err = db.Get(&userID, "SELECT parent_user_id FROM user_notifications WHERE id = $1", notificationID)
		} else {
			err = db.Get(&userID, "SELECT owner_id FROM product_owner_notifications WHERE id = $1", notificationID)
		}

		if err == nil {
			hub.BroadcastToUser(userID, "notification_read", notificationType, readMessage)
		}

		return c.SendString("Notification marked as read")
	}
}