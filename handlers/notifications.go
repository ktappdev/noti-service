package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/ktappdev/noti-service/models"
	"github.com/ktappdev/noti-service/reviewit"
	"github.com/ktappdev/noti-service/sse"
)

// parseCommaSeparatedString converts a comma-separated string to a Go slice
func parseCommaSeparatedString(str string) []string {
	if str == "" {
		return []string{}
	}
	
	return strings.Split(str, ",")
}

// CreateProductOwnerNotification creates a new product owner notification
func CreateProductOwnerNotification(db *sqlx.DB, hub *sse.SSEHub) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fmt.Println("createProductOwnerNotification")
		notification := new(models.ProductOwnerNotification)
		if err := c.BodyParser(notification); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// Set the notification type
		notification.NotificationType = "owner_review"
		
		// Set target_type and target_url
		targetType := "review"
		notification.TargetType = &targetType
		if notification.ReviewID != "" {
			targetURL := "/review/" + notification.ReviewID
			notification.TargetURL = &targetURL
		}
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

		query := `INSERT INTO product_owner_notifications (id, owner_id, product_id, product_name, business_id, review_title, from_name, from_id, read, comment_id, review_id, notification_type, target_type, target_url)
	              VALUES (:id, :owner_id, :product_id, :product_name, :business_id, :review_title, :from_name, :from_id, :read, :comment_id, :review_id, :notification_type, :target_type, :target_url) RETURNING id, created_at`
		rows, err := db.NamedQuery(query, notification)
		if err != nil {
			log.Printf("Error creating product owner notification for owner %s: %v", notification.OwnerID, err)
			return c.Status(500).SendString("Failed to create notification")
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

// CreateCommentNotification creates a new comment notification (for comments on reviews)
func CreateCommentNotification(db *sqlx.DB, hub *sse.SSEHub) fiber.Handler {
	return func(c *fiber.Ctx) error {
		notification := new(models.UserNotification)
		if err := c.BodyParser(notification); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// Set the notification type
		notification.NotificationType = "reply_review"
		
		// Set target_type and target_url
		targetType := "review"
		notification.TargetType = &targetType
		if notification.ReviewID != "" {
			targetURL := "/review/" + notification.ReviewID
			notification.TargetURL = &targetURL
		}
		log.Printf("Creating comment notification: %+v", notification)

		// For comments on reviews, target_user_id should be provided directly
		if notification.ParentUserID == "" {
			return c.Status(400).SendString("target_user_id (parent_user_id) is required for comment notifications")
		}

		// Check if the target user exists
		var exists bool
		err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", notification.ParentUserID)
		if err != nil {
			log.Printf("Error checking target user existence: %v", err)
			return c.Status(500).SendString("Internal server error")
		}
		if !exists {
			log.Printf("Target user %s does not exist", notification.ParentUserID)
			return c.Status(400).SendString("Target user does not exist. Please ensure the user is created in the notification service first.")
		}

		// Check if the from user exists
		err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", notification.FromID)
		if err != nil {
			log.Printf("Error checking from user existence: %v", err)
			return c.Status(500).SendString("Internal server error")
		}
		if !exists {
			log.Printf("From user %s does not exist", notification.FromID)
			return c.Status(400).SendString("From user does not exist. Please ensure the user is created in the notification service first.")
		}

		// Insert the notification
		query := `INSERT INTO user_notifications (id, parent_user_id, content, read, notification_type, comment_id, from_id, review_id, parent_id, from_name, product_id, target_type, target_url)
	  VALUES (:id, :parent_user_id, :content, :read, :notification_type, :comment_id, :from_id, :review_id, :parent_id, :from_name, :product_id, :target_type, :target_url) RETURNING id, created_at`
		rows, err := db.NamedQuery(query, notification)
		if err != nil {
			log.Printf("Error inserting comment notification for user %s: %v", notification.ParentUserID, err)
			return c.Status(500).SendString("Failed to create comment notification")
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&notification.ID, &notification.CreatedAt)
			if err != nil {
				log.Printf("Error scanning comment notification result: %v", err)
				return c.Status(500).SendString("Failed to retrieve created comment notification")
			}
		}

		// Broadcast to SSE clients
		hub.BroadcastToUser(notification.ParentUserID, "new_notification", "user", notification)

		return c.Status(201).JSON(notification)
	}
}

// CreateReplyNotification creates a new reply notification (for replies to comments)
func CreateReplyNotification(db *sqlx.DB, hub *sse.SSEHub) fiber.Handler {
	return func(c *fiber.Ctx) error {
		notification := new(models.UserNotification)
		if err := c.BodyParser(notification); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// Set the notification type
		notification.NotificationType = "reply_comment"
		
		// Set target_type and target_url
		targetType := "comment"
		notification.TargetType = &targetType
		if notification.ReviewID != "" {
			if notification.CommentID != nil && *notification.CommentID != "" {
				targetURL := "/review/" + notification.ReviewID + "?cid=" + *notification.CommentID
				notification.TargetURL = &targetURL
			} else {
				targetURL := "/review/" + notification.ReviewID
				notification.TargetURL = &targetURL
			}
		}
		log.Printf("Creating reply notification: %+v", notification)

		// For replies, parent_user_id should be provided directly by the frontend
		// If not provided, try to look it up from the parent comment
		if notification.ParentUserID == "" {
			if notification.ParentID == "" {
				return c.Status(400).SendString("Either parent_user_id or parent_id (comment ID) is required for reply notifications")
			}

			// ParentID should be a comment ID, look up the user who made that comment
			parentUserID, err := reviewit.GetParentCommentUserID(notification.ParentID)
			if err != nil {
				log.Printf("ERROR getting parent user ID from comment %s: %v", notification.ParentID, err)
				// Check if it's a missing environment variable error
				if err.Error() == "REVIEWIT_DATABASE_URL environment variable is required" {
					return c.Status(500).SendString("ReviewIt database connection not configured. Please set REVIEWIT_DATABASE_URL environment variable.")
				}
				return c.Status(500).SendString(fmt.Sprintf("Error getting parent user ID from comment: %v", err))
			}
			notification.ParentUserID = parentUserID
			log.Printf("Looked up user ID from comment %s: %s", notification.ParentID, parentUserID)
		}

		// Check if the target user exists
		var exists bool
		err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", notification.ParentUserID)
		if err != nil {
			log.Printf("Error checking target user existence: %v", err)
			return c.Status(500).SendString("Internal server error")
		}
		if !exists {
			log.Printf("Target user %s does not exist", notification.ParentUserID)
			return c.Status(400).SendString("Target user does not exist. Please ensure the user is created in the notification service first.")
		}

		// Check if the from user exists
		err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", notification.FromID)
		if err != nil {
			log.Printf("Error checking from user existence: %v", err)
			return c.Status(500).SendString("Internal server error")
		}
		if !exists {
			log.Printf("From user %s does not exist", notification.FromID)
			return c.Status(400).SendString("From user does not exist. Please ensure the user is created in the notification service first.")
		}

		// Insert the notification
		query := `INSERT INTO user_notifications (id, parent_user_id, content, read, notification_type, comment_id, from_id, review_id, parent_id, from_name, product_id, target_type, target_url)
	  VALUES (:id, :parent_user_id, :content, :read, :notification_type, :comment_id, :from_id, :review_id, :parent_id, :from_name, :product_id, :target_type, :target_url) RETURNING id, created_at`
		rows, err := db.NamedQuery(query, notification)
		if err != nil {
			log.Printf("Error inserting reply notification for user %s: %v", notification.ParentUserID, err)
			return c.Status(500).SendString("Failed to create reply notification")
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&notification.ID, &notification.CreatedAt)
			if err != nil {
				log.Printf("Error scanning reply notification result: %v", err)
				return c.Status(500).SendString("Failed to retrieve created reply notification")
			}
		}

		// Broadcast to SSE clients
		hub.BroadcastToUser(notification.ParentUserID, "new_notification", "user", notification)

		return c.Status(201).JSON(notification)
	}
}

// CreateSystemNotification creates a new system notification
func CreateSystemNotification(db *sqlx.DB, hub *sse.SSEHub) fiber.Handler {
	return func(c *fiber.Ctx) error {
		notification := new(models.SystemNotification)
		if err := c.BodyParser(notification); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// Set the notification type
		notification.NotificationType = "system"
		log.Printf("Creating system notification: %+v", notification)

		// Validate required fields
		if notification.Title == "" {
			return c.Status(400).SendString("title is required for system notifications")
		}
		if notification.Message == "" {
			return c.Status(400).SendString("message is required for system notifications")
		}

		// If target_user_ids is empty or nil, this is a broadcast to all users
		isBroadcast := len(notification.TargetUserIDsArray) == 0

		if !isBroadcast {
			// Check if all target users exist
			for _, userID := range notification.TargetUserIDsArray {
				var exists bool
				err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID)
				if err != nil {
					log.Printf("Error checking user existence for %s: %v", userID, err)
					return c.Status(500).SendString("Internal server error")
				}
				if !exists {
					log.Printf("Target user %s does not exist", userID)
					return c.Status(400).SendString(fmt.Sprintf("Target user %s does not exist. Please ensure the user is created in the notification service first.", userID))
				}
			}
		}

		// Convert target_user_ids to comma-separated string
		var targetUserIDsString string
		if isBroadcast {
			targetUserIDsString = "" // Empty string means broadcast to all
		} else {
			targetUserIDsString = strings.Join(notification.TargetUserIDsArray, ",")
		}

		// Insert the notification
		query := `INSERT INTO system_notifications (id, target_user_ids, title, message, cta_url, icon, read, notification_type)
	  VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at`
		
		err := db.QueryRow(query, 
			notification.ID, 
			targetUserIDsString, 
			notification.Title, 
			notification.Message, 
			notification.CtaURL, 
			notification.Icon, 
			notification.Read, 
			notification.NotificationType,
		).Scan(&notification.ID, &notification.CreatedAt)
		
		if err != nil {
			log.Printf("Error inserting system notification (broadcast: %v, targets: %d): %v", isBroadcast, len(notification.TargetUserIDsArray), err)
			return c.Status(500).SendString("Failed to create system notification")
		}

		// Set the array field for JSON response
		notification.TargetUserIDsArray = notification.TargetUserIDsArray

		// Broadcast to SSE clients
		if isBroadcast {
			// Broadcast to all connected users
			hub.BroadcastToAll("new_notification", "system", notification)
			log.Printf("Broadcasting system notification to all users")
		} else {
			// Send to specific users
			for _, userID := range notification.TargetUserIDsArray {
				hub.BroadcastToUser(userID, "new_notification", "system", notification)
				log.Printf("Sending system notification to user: %s", userID)
			}
		}

		return c.Status(201).JSON(notification)
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
			// Check if TargetID is already a user ID or a comment ID
			if len(notification.TargetID) > 5 && notification.TargetID[:5] == "user_" {
				// TargetID is already a user ID, use it directly
				targetUserID = notification.TargetID
				log.Printf("Using TargetID as user ID directly for comment: %s", targetUserID)
			} else {
				// TargetID is a comment ID, look up the user who made that comment
				targetUserID, err = reviewit.GetCommentUserID(notification.TargetID)
				if err != nil {
					log.Printf("ERROR getting comment user ID: %v", err)
					if err.Error() == "REVIEWIT_DATABASE_URL environment variable is required" {
						return c.Status(500).SendString("ReviewIt database connection not configured. Please set REVIEWIT_DATABASE_URL environment variable.")
					}
					return c.Status(500).SendString(fmt.Sprintf("Error getting comment user ID: %v", err))
				}
				log.Printf("Looked up user ID from comment: %s", targetUserID)
			}
		} else if notification.TargetType == "review" {
			// Check if TargetID is already a user ID or a review ID
			if len(notification.TargetID) > 5 && notification.TargetID[:5] == "user_" {
				// TargetID is already a user ID, use it directly
				targetUserID = notification.TargetID
				log.Printf("Using TargetID as user ID directly for review: %s", targetUserID)
			} else {
				// TargetID is a review ID, look up the user who made that review
				targetUserID, err = reviewit.GetReviewUserID(notification.TargetID)
				if err != nil {
					log.Printf("ERROR getting review user ID: %v", err)
					if err.Error() == "REVIEWIT_DATABASE_URL environment variable is required" {
						return c.Status(500).SendString("ReviewIt database connection not configured. Please set REVIEWIT_DATABASE_URL environment variable.")
					}
					return c.Status(500).SendString(fmt.Sprintf("Error getting review user ID: %v", err))
				}
				log.Printf("Looked up user ID from review: %s", targetUserID)
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

		// Get review_id and comment_id based on target_type
		var reviewID, commentID string
		if notification.TargetType == "comment" {
			// For comment likes, target_id is the comment ID
			commentID = notification.TargetID
			// Get the review_id that this comment belongs to
			var err error
			reviewID, err = reviewit.GetCommentReviewID(notification.TargetID)
			if err != nil {
				log.Printf("ERROR getting review ID for comment %s: %v", notification.TargetID, err)
				// If we can't get review ID, we'll leave it empty but still create the notification
				reviewID = ""
			}
		} else if notification.TargetType == "review" {
			// For review likes, target_id is the review ID
			reviewID = notification.TargetID
			commentID = "" // No comment ID for review likes
		}

		notification.ReviewID = reviewID
		if commentID != "" {
			notification.CommentID = &commentID
		}

		// Set notification type and target_url based on target_type
		if notification.TargetType == "comment" {
			notification.NotificationType = "like_comment"
			if reviewID != "" && commentID != "" {
				targetURL := "/review/" + reviewID + "?cid=" + commentID
				notification.TargetURL = &targetURL
			}
		} else if notification.TargetType == "review" {
			notification.NotificationType = "like_review"
			if reviewID != "" {
				targetURL := "/review/" + reviewID
				notification.TargetURL = &targetURL
			}
		}

		// Ensure from_name is populated if missing
		if notification.FromName == "" {
			var err error
			notification.FromName, err = reviewit.GetUserFullName(notification.FromID)
			if err != nil {
				log.Printf("ERROR getting user full name for %s: %v", notification.FromID, err)
				// If we can't get the name, use a fallback
				notification.FromName = "Someone"
			}
		}

		// Insert the notification
		query := `INSERT INTO like_notifications (id, target_user_id, target_type, target_id, from_id, from_name, product_id, review_id, comment_id, read, notification_type, target_url)
	              VALUES (gen_random_uuid(), :target_user_id, :target_type, :target_id, :from_id, :from_name, :product_id, :review_id, :comment_id, :read, :notification_type, :target_url) 
	              RETURNING id, created_at`
		
		rows, err := db.NamedQuery(query, notification)
		if err != nil {
			log.Printf("Error inserting like notification for user %s (target: %s): %v", notification.TargetUserID, notification.TargetType, err)
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

		likeQuery := `SELECT * FROM like_notifications
	                  WHERE target_user_id = $1
	                  ORDER BY created_at DESC`
		var likeNotifications []models.LikeNotification
		err = db.Select(&likeNotifications, likeQuery, userID)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		// Get system notifications (broadcast to all or specifically targeted to this user)
		systemQuery := `SELECT id, COALESCE(target_user_ids, '') as target_user_ids, title, message, cta_url, icon, read, created_at, notification_type 
	                    FROM system_notifications
	                    WHERE target_user_ids = '' OR target_user_ids IS NULL OR target_user_ids LIKE '%' || $1 || '%'
	                    ORDER BY created_at DESC`
		var systemNotifications []models.SystemNotification
		err = db.Select(&systemNotifications, systemQuery, userID)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		// Convert PostgreSQL arrays to Go slices for JSON response
		for i := range systemNotifications {
			systemNotifications[i].TargetUserIDsArray = parseCommaSeparatedString(systemNotifications[i].TargetUserIDs)
		}

		return c.JSON(fiber.Map{
			"user_notifications":   userNotifications,
			"owner_notifications":  ownerNotifications,
			"like_notifications":   likeNotifications,
			"system_notifications": systemNotifications,
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

		likeQuery := `SELECT * FROM like_notifications
	                  WHERE target_user_id = $1 AND read = false
	                  ORDER BY created_at DESC`
		var likeNotifications []models.LikeNotification
		err = db.Select(&likeNotifications, likeQuery, userID)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		// Get unread system notifications (broadcast to all or specifically targeted to this user)
		systemQuery := `SELECT id, COALESCE(target_user_ids, '') as target_user_ids, title, message, cta_url, icon, read, created_at, notification_type 
	                    FROM system_notifications
	                    WHERE (target_user_ids = '' OR target_user_ids IS NULL OR target_user_ids LIKE '%' || $1 || '%') AND read = false
	                    ORDER BY created_at DESC`
		var systemNotifications []models.SystemNotification
		err = db.Select(&systemNotifications, systemQuery, userID)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		// Convert PostgreSQL arrays to Go slices for JSON response
		for i := range systemNotifications {
			systemNotifications[i].TargetUserIDsArray = parseCommaSeparatedString(systemNotifications[i].TargetUserIDs)
		}

		return c.JSON(fiber.Map{
			"user_notifications":   userNotifications,
			"owner_notifications":  ownerNotifications,
			"like_notifications":   likeNotifications,
			"system_notifications": systemNotifications,
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

		userQuery := `DELETE FROM user_notifications WHERE parent_user_id = $1 AND read = true RETURNING *`
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

		likeQuery := `DELETE FROM like_notifications WHERE target_user_id = $1 AND read = true RETURNING *`
		var deletedLikeNotifications []models.LikeNotification
		err = db.Select(&deletedLikeNotifications, likeQuery, userID)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		return c.JSON(fiber.Map{
			"deleted_user_notifications":  len(deletedUserNotifications),
			"deleted_owner_notifications": len(deletedOwnerNotifications),
			"deleted_like_notifications":  len(deletedLikeNotifications),
			"user_notifications":          deletedUserNotifications,
			"owner_notifications":         deletedOwnerNotifications,
			"like_notifications":          deletedLikeNotifications,
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
		case "like":
			query = "UPDATE like_notifications SET read = true WHERE id = $1"
		case "system":
			query = "UPDATE system_notifications SET read = true WHERE id = $1"
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
			log.Printf("Attempted to mark non-existent notification as read: ID=%s, Type=%s", notificationID, notificationType)
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
		} else if notificationType == "owner" {
			err = db.Get(&userID, "SELECT owner_id FROM product_owner_notifications WHERE id = $1", notificationID)
		} else if notificationType == "like" {
			err = db.Get(&userID, "SELECT target_user_id FROM like_notifications WHERE id = $1", notificationID)
		} else if notificationType == "system" {
			// For system notifications, we need to broadcast to all affected users
			// For now, we'll skip the individual user broadcast since system notifications
			// can target multiple users or be broadcasts
			userID = "" // Will skip the broadcast below
		}

		if err == nil {
			hub.BroadcastToUser(userID, "notification_read", notificationType, readMessage)
		}

		return c.SendString("Notification marked as read")
	}
}

// MarkAllNotificationsAsRead marks all unread notifications as read for a user
func MarkAllNotificationsAsRead(db *sqlx.DB, hub *sse.SSEHub) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Query("user_id")
		if userID == "" {
			return c.Status(400).SendString("user_id query parameter is required")
		}

		notificationType := c.Query("type") // optional: "user", "owner", "like", "system", or empty for all

		var userUpdated, ownerUpdated, likeUpdated, systemUpdated int64

		// Update user notifications
		if notificationType == "" || notificationType == "user" {
			result, err := db.Exec("UPDATE user_notifications SET read = true WHERE parent_user_id = $1 AND read = false", userID)
			if err != nil {
				log.Printf("Error updating user notifications: %v", err)
				return c.Status(500).SendString("Failed to update user notifications")
			}
			userUpdated, _ = result.RowsAffected()
		}

		// Update owner notifications
		if notificationType == "" || notificationType == "owner" {
			result, err := db.Exec("UPDATE product_owner_notifications SET read = true WHERE owner_id = $1 AND read = false", userID)
			if err != nil {
				log.Printf("Error updating owner notifications: %v", err)
				return c.Status(500).SendString("Failed to update owner notifications")
			}
			ownerUpdated, _ = result.RowsAffected()
		}

		// Update like notifications
		if notificationType == "" || notificationType == "like" {
			result, err := db.Exec("UPDATE like_notifications SET read = true WHERE target_user_id = $1 AND read = false", userID)
			if err != nil {
				log.Printf("Error updating like notifications: %v", err)
				return c.Status(500).SendString("Failed to update like notifications")
			}
			likeUpdated, _ = result.RowsAffected()
		}

		// Update system notifications (only those targeted to this user or broadcast to all)
		if notificationType == "" || notificationType == "system" {
			result, err := db.Exec(`UPDATE system_notifications SET read = true 
				WHERE (target_user_ids = '' OR target_user_ids IS NULL OR target_user_ids LIKE '%' || $1 || '%') 
				AND read = false`, userID)
			if err != nil {
				log.Printf("Error updating system notifications: %v", err)
				return c.Status(500).SendString("Failed to update system notifications")
			}
			systemUpdated, _ = result.RowsAffected()
		}

		totalUpdated := userUpdated + ownerUpdated + likeUpdated + systemUpdated

		// Broadcast read status to SSE clients for each type that was updated
		readMessage := map[string]interface{}{
			"user_id":     userID,
			"read":        true,
			"timestamp":   time.Now().Format(time.RFC3339),
			"bulk_update": true,
		}

		if userUpdated > 0 {
			readMessage["type"] = "user"
			readMessage["count"] = userUpdated
			hub.BroadcastToUser(userID, "notifications_bulk_read", "user", readMessage)
		}

		if ownerUpdated > 0 {
			readMessage["type"] = "owner"
			readMessage["count"] = ownerUpdated
			hub.BroadcastToUser(userID, "notifications_bulk_read", "owner", readMessage)
		}

		if likeUpdated > 0 {
			readMessage["type"] = "like"
			readMessage["count"] = likeUpdated
			hub.BroadcastToUser(userID, "notifications_bulk_read", "like", readMessage)
		}

		if systemUpdated > 0 {
			readMessage["type"] = "system"
			readMessage["count"] = systemUpdated
			hub.BroadcastToUser(userID, "notifications_bulk_read", "system", readMessage)
		}

		return c.JSON(fiber.Map{
			"message":                      "All notifications marked as read",
			"user_notifications_updated":   userUpdated,
			"owner_notifications_updated":  ownerUpdated,
			"like_notifications_updated":   likeUpdated,
			"system_notifications_updated": systemUpdated,
			"total_updated":                totalUpdated,
			"success":                      true,
		})
	}
}