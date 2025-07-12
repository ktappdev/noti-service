package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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

// CreateProductOwnerNotificationGin creates a new product owner notification (Gin version)
func CreateProductOwnerNotificationGin(db *sqlx.DB, hub *sse.SSEHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("createProductOwnerNotification (Gin)")
		notification := new(models.ProductOwnerNotification)
		if err := c.ShouldBindJSON(notification); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		notification.NotificationType = "review"
		fmt.Println("this is the notification", notification)

		var exists bool
		err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", notification.OwnerID)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !exists {
			log.Println("owner does not exist")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Owner does not exist"})
			return
		}

		query := `INSERT INTO product_owner_notifications (id, owner_id, product_id, product_name, business_id, review_title, from_name, from_id, read, comment_id, review_id, notification_type)
	              VALUES (:id, :owner_id, :product_id, :product_name, :business_id, :review_title, :from_name, :from_id, :read, :comment_id, :review_id, :notification_type) RETURNING id, created_at`
		rows, err := db.NamedQuery(query, notification)
		if err != nil {
			log.Printf("Error creating notification: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&notification.ID, &notification.CreatedAt)
			if err != nil {
				log.Printf("Error scanning notification result: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		hub.BroadcastToUser(notification.OwnerID, "new_notification", "owner", notification)

		c.JSON(http.StatusCreated, notification)
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

// CreateReplyNotificationGin creates a new reply notification (Gin version)
func CreateReplyNotificationGin(db *sqlx.DB, hub *sse.SSEHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		notification := new(models.UserNotification)
		if err := c.ShouldBindJSON(notification); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		notification.NotificationType = "reply"
		log.Printf("Creating reply notification: %+v", notification)

		parentUserID, err := reviewit.GetParentCommentUserID(notification.ParentID)
		if err != nil {
			log.Printf("ERROR getting parent user ID: %v", err)
			if err.Error() == "REVIEWIT_DATABASE_URL environment variable is required" {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "ReviewIt database connection not configured. Please set REVIEWIT_DATABASE_URL environment variable."})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting parent user ID: " + err.Error()})
			return
		}
		notification.ParentUserID = parentUserID

		var exists bool
		err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", notification.ParentUserID)
		if err != nil {
			log.Printf("Error checking user existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		if !exists {
			log.Println("user don't exist")
			c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
			return
		}

		query := `INSERT INTO user_notifications (id, parent_user_id, content, read, notification_type, comment_id, from_id, review_id, parent_id, from_name, product_id)
	  VALUES (:id, :parent_user_id, :content, :read, :notification_type, :comment_id, :from_id, :review_id, :parent_id, :from_name, :product_id) RETURNING id, created_at`
		rows, err := db.NamedQuery(query, notification)
		if err != nil {
			log.Printf("Error inserting notification: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
			return
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&notification.ID, &notification.CreatedAt)
			if err != nil {
				log.Printf("Error scanning notification result: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve created notification"})
				return
			}
		}

		hub.BroadcastToUser(notification.ParentUserID, "new_notification", "user", notification)

		c.JSON(http.StatusCreated, notification)
	}
}

// CreateLikeNotification creates a new like notification
func CreateLikeNotification(db *sqlx.DB, hub *sse.SSEHub) fiber.Handler {
	return func(c *fiber.Ctx) error {
		notification := new(models.LikeNotification)
		if err := c.BodyParser(notification); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		log.Printf("Creating like notification: %+v", notification)

		// Validate target_type
		if notification.TargetType != "comment" && notification.TargetType != "review" {
			return c.Status(400).SendString("target_type must be 'comment' or 'review'")
		}

		// Get the target user ID based on target type and ID
		var targetUserID string
		var err error

		if notification.TargetType == "comment" {
			// For comments, get the user who made the comment
			targetUserID, err = reviewit.GetCommentUserID(notification.TargetID)
			if err != nil {
				log.Printf("ERROR getting comment user ID: %v", err)
				if err.Error() == "REVIEWIT_DATABASE_URL environment variable is required" {
					return c.Status(500).SendString("ReviewIt database connection not configured. Please set REVIEWIT_DATABASE_URL environment variable.")
				}
				return c.Status(500).SendString(fmt.Sprintf("Error getting comment user ID: %v", err))
			}
		} else if notification.TargetType == "review" {
			// For reviews, get the user who made the review
			targetUserID, err = reviewit.GetReviewUserID(notification.TargetID)
			if err != nil {
				log.Printf("ERROR getting review user ID: %v", err)
				if err.Error() == "REVIEWIT_DATABASE_URL environment variable is required" {
					return c.Status(500).SendString("ReviewIt database connection not configured. Please set REVIEWIT_DATABASE_URL environment variable.")
				}
				return c.Status(500).SendString(fmt.Sprintf("Error getting review user ID: %v", err))
			}
		}

		notification.TargetUserID = targetUserID

		// Check if the target user exists
		var exists bool
		err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", notification.TargetUserID)
		if err != nil {
			log.Printf("Error checking target user existence: %v", err)
			return c.Status(500).SendString("Internal server error")
		}
		if !exists {
			log.Println("target user doesn't exist")
			return c.Status(400).SendString("Target user does not exist")
		}

		// Check if the from user exists
		err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", notification.FromID)
		if err != nil {
			log.Printf("Error checking from user existence: %v", err)
			return c.Status(500).SendString("Internal server error")
		}
		if !exists {
			log.Println("from user doesn't exist")
			return c.Status(400).SendString("From user does not exist")
		}

		// Don't create notification if user likes their own content
		if notification.TargetUserID == notification.FromID {
			return c.Status(200).SendString("No notification created for self-like")
		}

		// Insert the notification
		query := `INSERT INTO like_notifications (id, target_user_id, target_type, target_id, from_id, from_name, product_id, read)
	              VALUES (gen_random_uuid(), :target_user_id, :target_type, :target_id, :from_id, :from_name, :product_id, :read) 
	              RETURNING id, created_at`
		
		rows, err := db.NamedQuery(query, notification)
		if err != nil {
			log.Printf("Error inserting like notification: %v", err)
			return c.Status(500).SendString("Failed to create like notification")
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&notification.ID, &notification.CreatedAt)
			if err != nil {
				log.Printf("Error scanning like notification result: %v", err)
				return c.Status(500).SendString("Failed to retrieve created like notification")
			}
		}

		// Broadcast to SSE clients
		hub.BroadcastToUser(notification.TargetUserID, "new_notification", "like", notification)

		return c.Status(201).JSON(notification)
	}
}

// CreateLikeNotificationGin creates a new like notification (Gin version)
func CreateLikeNotificationGin(db *sqlx.DB, hub *sse.SSEHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		notification := new(models.LikeNotification)
		if err := c.ShouldBindJSON(notification); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("Creating like notification: %+v", notification)

		if notification.TargetType != "comment" && notification.TargetType != "review" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "target_type must be 'comment' or 'review'"})
			return
		}

		var targetUserID string
		var err error

		if notification.TargetType == "comment" {
			targetUserID, err = reviewit.GetCommentUserID(notification.TargetID)
			if err != nil {
				log.Printf("ERROR getting comment user ID: %v", err)
				if err.Error() == "REVIEWIT_DATABASE_URL environment variable is required" {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "ReviewIt database connection not configured. Please set REVIEWIT_DATABASE_URL environment variable."})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting comment user ID: " + err.Error()})
				return
			}
		} else if notification.TargetType == "review" {
			targetUserID, err = reviewit.GetReviewUserID(notification.TargetID)
			if err != nil {
				log.Printf("ERROR getting review user ID: %v", err)
				if err.Error() == "REVIEWIT_DATABASE_URL environment variable is required" {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "ReviewIt database connection not configured. Please set REVIEWIT_DATABASE_URL environment variable."})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting review user ID: " + err.Error()})
				return
			}
		}

		notification.TargetUserID = targetUserID

		var exists bool
		err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", notification.TargetUserID)
		if err != nil {
			log.Printf("Error checking target user existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		if !exists {
			log.Println("target user doesn't exist")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Target user does not exist"})
			return
		}

		err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", notification.FromID)
		if err != nil {
			log.Printf("Error checking from user existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		if !exists {
			log.Println("from user doesn't exist")
			c.JSON(http.StatusBadRequest, gin.H{"error": "From user does not exist"})
			return
		}

		if notification.TargetUserID == notification.FromID {
			c.JSON(http.StatusOK, gin.H{"message": "No notification created for self-like"})
			return
		}

		query := `INSERT INTO like_notifications (id, target_user_id, target_type, target_id, from_id, from_name, product_id, read)
	              VALUES (gen_random_uuid(), :target_user_id, :target_type, :target_id, :from_id, :from_name, :product_id, :read) 
	              RETURNING id, created_at`
		rows, err := db.NamedQuery(query, notification)
		if err != nil {
			log.Printf("Error inserting like notification: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create like notification"})
			return
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&notification.ID, &notification.CreatedAt)
			if err != nil {
				log.Printf("Error scanning like notification result: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve created like notification"})
				return
			}
		}

		hub.BroadcastToUser(notification.TargetUserID, "new_notification", notification.TargetType, notification)

		c.JSON(http.StatusCreated, notification)
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

// GetLatestNotificationsGin gets the latest notifications for a user (Gin version)
func GetLatestNotificationsGin(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			c.JSON(400, gin.H{"error": "user_id is required"})
			return
		}

		// Use the correct notification model for your schema
		var notifications []models.UserNotification
		query := `SELECT * FROM user_notifications WHERE parent_user_id = $1 ORDER BY created_at DESC LIMIT 10`
		err := db.Select(&notifications, query, userID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, notifications)
	}
}

// GetAllNotificationsGin gets all notifications for a user (Gin version)
func GetAllNotificationsGin(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			c.JSON(400, gin.H{"error": "user_id is required"})
			return
		}

		var notifications []models.UserNotification
		query := `SELECT * FROM user_notifications WHERE parent_user_id = $1 ORDER BY created_at DESC`
		err := db.Select(&notifications, query, userID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, notifications)
	}
}

// GetAllUnreadNotificationsGin gets all unread notifications for a user (Gin version)
func GetAllUnreadNotificationsGin(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			c.JSON(400, gin.H{"error": "user_id is required"})
			return
		}

		var notifications []models.UserNotification
		query := `SELECT * FROM user_notifications WHERE parent_user_id = $1 AND read = false ORDER BY created_at DESC`
		err := db.Select(&notifications, query, userID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, notifications)
	}
}

// DeleteReadNotificationsGin deletes all read notifications for a user (Gin version)
func DeleteReadNotificationsGin(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			c.JSON(400, gin.H{"error": "user_id is required"})
			return
		}

		query := `DELETE FROM user_notifications WHERE parent_user_id = $1 AND read = true`
		_, err := db.Exec(query, userID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.Status(204)
	}
}

// MarkNotificationAsReadGin marks a notification as read (Gin version)
func MarkNotificationAsReadGin(db *sqlx.DB, hub *sse.SSEHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(400, gin.H{"error": "id is required"})
			return
		}

		query := `UPDATE user_notifications SET read = true WHERE id = $1 RETURNING parent_user_id`
		var userID string
		err := db.Get(&userID, query, id)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		hub.BroadcastToUser(userID, "notification_read", "notification", gin.H{"id": id})
		c.Status(204)
	}
}