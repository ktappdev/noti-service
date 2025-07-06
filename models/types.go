package models

import "time"

// UserNotification represents a notification for a regular user
type UserNotification struct {
	ID               string    `db:"id" json:"id"`
	ParentUserID     string    `db:"parent_user_id" json:"parent_user_id"`
	Content          string    `db:"content" json:"content"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	Read             bool      `db:"read" json:"read"`
	NotificationType string    `db:"notification_type" json:"notification_type"`
	CommentID        string    `db:"comment_id" json:"comment_id"`
	ReviewID         string    `db:"review_id" json:"review_id"`
	FromID           string    `db:"from_id" json:"from_id"`
	ParentID         string    `db:"parent_id" json:"parent_id"`
	FromName         string    `db:"from_name" json:"from_name"`
	ProductID        string    `db:"product_id" json:"product_id"`
}

// ProductOwnerNotification represents a notification for a product owner
type ProductOwnerNotification struct {
	ID               string    `db:"id" json:"id"`
	OwnerID          string    `db:"owner_id" json:"owner_id"`
	ProductID        string    `db:"product_id" json:"product_id"`
	ProductName      string    `db:"product_name" json:"product_name"`
	BusinessID       string    `db:"business_id" json:"business_id"`
	ReviewTitle      string    `db:"review_title" json:"review_title"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	FromName         string    `db:"from_name" json:"from_name"`
	FromID           string    `db:"from_id" json:"from_id"`
	Read             bool      `db:"read" json:"read"`
	CommentID        *string   `db:"comment_id" json:"comment_id"`
	ReviewID         *string   `db:"review_id" json:"review_id"`
	NotificationType string    `db:"notification_type" json:"notification_type"`
}

// User represents a user in the system
type User struct {
	ID       string `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	FullName string `db:"full_name" json:"full_name"`
}

// NotificationMessage represents a message sent through SSE
type NotificationMessage struct {
	UserID       string      `json:"user_id"`
	Type         string      `json:"type"` // "user" or "owner"
	Notification interface{} `json:"notification"`
	Event        string      `json:"event"` // "new_notification", "notification_read", etc.
}