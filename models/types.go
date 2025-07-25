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

// LikeNotification represents a like notification
type LikeNotification struct {
	ID           string    `db:"id" json:"id"`
	TargetUserID string    `db:"target_user_id" json:"target_user_id"` // User who owns the liked content
	TargetType   string    `db:"target_type" json:"target_type"`       // "comment" or "review"
	TargetID     string    `db:"target_id" json:"target_id"`           // ID of the liked content
	FromID       string    `db:"from_id" json:"from_id"`               // User who liked
	FromName     string    `db:"from_name" json:"from_name"`           // Name of user who liked
	ProductID    string    `db:"product_id" json:"product_id"`         // Product context
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	Read         bool      `db:"read" json:"read"`
}

// SystemNotification represents a system/admin notification
type SystemNotification struct {
	ID              string    `db:"id" json:"id"`
	TargetUserIDs   string    `db:"target_user_ids" json:"-"` // Comma-separated user IDs
	TargetUserIDsArray []string `json:"target_user_ids"`      // JSON representation
	Title           string    `db:"title" json:"title"`
	Message         string    `db:"message" json:"message"`
	CtaURL          *string   `db:"cta_url" json:"cta_url"`           // Optional call-to-action URL
	Icon            *string   `db:"icon" json:"icon"`                 // Optional icon hint (info/success/warning/error)
	Read            bool      `db:"read" json:"read"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	NotificationType string   `db:"notification_type" json:"notification_type"` // Always "system"
}

// NotificationMessage represents a message sent through SSE
type NotificationMessage struct {
	UserID       string      `json:"user_id"`
	Type         string      `json:"type"` // "user", "owner", "like", or "system"
	Notification interface{} `json:"notification"`
	Event        string      `json:"event"` // "new_notification", "notification_read", etc.
}