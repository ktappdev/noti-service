package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	// "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/ktappdev/noti-service/reviewit"
	_ "github.com/lib/pq"
)

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

type User struct {
	ID       string `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	FullName string `db:"full_name" json:"full_name"`
}

var db *sqlx.DB

func createSchema() error {
	schema := `
    CREATE TABLE IF NOT EXISTS users (
        id VARCHAR(255) PRIMARY KEY,
        username VARCHAR(255) NOT NULL,
        full_name VARCHAR(255) NOT NULL
    );

    CREATE TABLE IF NOT EXISTS user_notifications (
        id VARCHAR(255) PRIMARY KEY,
        parent_user_id VARCHAR(255) NOT NULL,
        content TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        read BOOLEAN DEFAULT FALSE,
        notification_type VARCHAR(50) NOT NULL,
        comment_id VARCHAR(255),
        from_id VARCHAR(255),
        review_id VARCHAR(255),
        parent_id VARCHAR(255),
        from_name VARCHAR(255),
        product_id VARCHAR(255),
        FOREIGN KEY (parent_user_id) REFERENCES users(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS product_owner_notifications (
        id VARCHAR(255) PRIMARY KEY,
        owner_id VARCHAR(255) NOT NULL,
        product_id VARCHAR(255) NOT NULL,
        product_name VARCHAR(255) NOT NULL,
        business_id VARCHAR(255) NOT NULL,
        review_title TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        from_name VARCHAR(255) NOT NULL,
        from_id VARCHAR(255) NOT NULL,
        read BOOLEAN DEFAULT FALSE,
        comment_id VARCHAR(255),
        review_id VARCHAR(255),
        notification_type VARCHAR(50) NOT NULL,
        FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
    );

    CREATE INDEX IF NOT EXISTS idx_user_notifications_user_id ON user_notifications(parent_user_id);
    CREATE INDEX IF NOT EXISTS idx_user_notifications_created_at ON user_notifications(created_at);
    CREATE INDEX IF NOT EXISTS idx_product_owner_notifications_owner_id ON product_owner_notifications(owner_id);
    CREATE INDEX IF NOT EXISTS idx_product_owner_notifications_created_at ON product_owner_notifications(created_at);
    `
	_, err := db.Exec(schema)
	return err
}

func createUser(c *fiber.Ctx) error {
	user := new(User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	query := `INSERT INTO users (id, username, full_name) VALUES (:id, :username, :full_name) RETURNING id`
	rows, err := db.NamedQuery(query, user)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&user.ID)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
	}

	return c.Status(201).JSON(user)
}

func createProductOwnerNotification(c *fiber.Ctx) error {
	fmt.Println("createProductOwnerNotification")
	notification := new(ProductOwnerNotification)
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

	return c.Status(201).JSON(notification)
}

func createReplyNotification(c *fiber.Ctx) error {
	// Read the raw body
	// rawBody := c.Body()
	// log.Printf("Raw request body: %s", rawBody)

	notification := new(UserNotification)
	if err := c.BodyParser(notification); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	// Set the notification type
	notification.NotificationType = "reply"
	log.Printf("Creating reply notification: %+v", notification)

	parentUserID, err := reviewit.GetParentCommentUserID(notification.ParentID)
	if err != nil {
		log.Printf("Error getting parent user ID: %v", err)
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

	return c.Status(201).JSON(notification)
}

func getLatestNotifications(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(400).SendString("user_id query parameter is required")
	}

	userQuery := `SELECT * FROM user_notifications
                  WHERE user_id = $1
                  ORDER BY created_at DESC
                  LIMIT 1`
	var userNotification UserNotification
	err := db.Get(&userNotification, userQuery, userID)
	if err != nil && err != sql.ErrNoRows {
		return c.Status(500).SendString(err.Error())
	}

	ownerQuery := `SELECT * FROM product_owner_notifications
                   WHERE owner_id = $1
                   ORDER BY created_at DESC
                   LIMIT 1`
	var ownerNotification ProductOwnerNotification
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

func getAllNotifications(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(400).SendString("user_id query parameter is required")
	}

	userQuery := `SELECT * FROM user_notifications
                  WHERE parent_user_id = $1
                  ORDER BY created_at DESC`
	var userNotifications []UserNotification
	err := db.Select(&userNotifications, userQuery, userID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	ownerQuery := `SELECT * FROM product_owner_notifications
                   WHERE owner_id = $1
                   ORDER BY created_at DESC`
	var ownerNotifications []ProductOwnerNotification
	err = db.Select(&ownerNotifications, ownerQuery, userID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(fiber.Map{
		"user_notifications":  userNotifications,
		"owner_notifications": ownerNotifications,
	})
}

func getAllUnreadNotifications(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(400).SendString("user_id query parameter is required")
	}

	userQuery := `SELECT * FROM user_notifications
                  WHERE parent_user_id = $1 AND read = false
                  ORDER BY created_at DESC`
	var userNotifications []UserNotification
	err := db.Select(&userNotifications, userQuery, userID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	ownerQuery := `SELECT * FROM product_owner_notifications
                   WHERE owner_id = $1 AND read = false
                   ORDER BY created_at DESC`
	var ownerNotifications []ProductOwnerNotification
	err = db.Select(&ownerNotifications, ownerQuery, userID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(fiber.Map{
		"user_notifications":  userNotifications,
		"owner_notifications": ownerNotifications,
	})
}

func deleteReadNotifications(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(400).SendString("user_id query parameter is required")
	}

	userQuery := `DELETE FROM user_notifications WHERE user_id = $1 AND read = true RETURNING *`
	var deletedUserNotifications []UserNotification
	err := db.Select(&deletedUserNotifications, userQuery, userID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	ownerQuery := `DELETE FROM product_owner_notifications WHERE owner_id = $1 AND read = true RETURNING *`
	var deletedOwnerNotifications []ProductOwnerNotification
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

func markNotificationAsRead(c *fiber.Ctx) error {
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

	return c.SendString("Notification marked as read")
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	var err error
	db, err = sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if err := createSchema(); err != nil {
		log.Fatal(err)
	}

	app := fiber.New()
	// app.Use(cors.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
	// app.Use(logger.New(logger.Config{
	// 	Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	// }))

	app.Post("/users", createUser)
	app.Post("/notifications/product-owner", createProductOwnerNotification)
	app.Post("/notifications/reply", createReplyNotification)
	app.Get("/notifications/latest", getLatestNotifications)
	app.Get("/notifications", getAllNotifications)
	app.Delete("/notifications", deleteReadNotifications)
	app.Post("/notifications/:id/read", markNotificationAsRead)

	log.Fatal(app.Listen(":3001"))
}
