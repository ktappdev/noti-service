// before user notifications
package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Notification struct {
	ID          string    `db:"id" json:"id"`
	ProductID   string    `db:"product_id" json:"product_id"`
	ProductName string    `db:"product_name" json:"product_name"`
	ReceiverID  string    `db:"receiver_id" json:"receiver_id"`
	BusinessID  string    `db:"business_id" json:"business_id"`
	ReviewTitle string    `db:"review_title" json:"review_title"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	FromName    string    `db:"from_name" json:"from_name"`
	FromID      string    `db:"from_id" json:"from_id"`
	Read        bool      `db:"read" json:"read"`
	CommentID   *string   `db:"comment_id" json:"comment_id"`
	ReviewID    *string   `db:"review_id" json:"review_id"`
}

type User struct {
	ID       string `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	FullName string `db:"full_name" json:"full_name"`
}

type Business struct {
	ID           string `db:"id" json:"id"`
	UserID       string `db:"user_id" json:"user_id"`
	BusinessName string `db:"business_name" json:"business_name"`
}

var db *sqlx.DB

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
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	app.Post("/users", createUser)
	app.Post("/businesses", createBusiness)
	app.Post("/notifications", createNotification)
	app.Get("/notifications/latest", getLatestNotifications)
	app.Get("/notifications", getAllNotifications)
	app.Delete("/notifications", deleteReadNotifications)

	log.Fatal(app.Listen(":3001"))
}

func createSchema() error {
	schema := `
    CREATE TABLE IF NOT EXISTS users (
        id VARCHAR(255) PRIMARY KEY,
        username VARCHAR(255) NOT NULL,
        full_name VARCHAR(255) NOT NULL
    );

    CREATE TABLE IF NOT EXISTS businesses (
        id VARCHAR(255) PRIMARY KEY,
        user_id VARCHAR(255) NOT NULL,
        business_name VARCHAR(255) NOT NULL,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS notifications (
        id VARCHAR(255) PRIMARY KEY,
        product_id VARCHAR(255) NOT NULL,
        product_name VARCHAR(255) NOT NULL,
        receiver_id VARCHAR(255) NOT NULL,
        business_id VARCHAR(255) NOT NULL,
        review_title TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        from_name VARCHAR(255) NOT NULL,
        from_id VARCHAR(255) NOT NULL,
        read BOOLEAN DEFAULT FALSE,
        comment_id VARCHAR(255),
        review_id VARCHAR(255),
        FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE CASCADE,
        FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE
    );
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

func createBusiness(c *fiber.Ctx) error {
	business := new(Business)
	if err := c.BodyParser(business); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	var exists bool
	err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", business.UserID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if !exists {
		return c.Status(400).SendString("User does not exist")
	}

	query := `INSERT INTO businesses (id, user_id, business_name) VALUES (:id, :user_id, :business_name) RETURNING id`
	rows, err := db.NamedQuery(query, business)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&business.ID)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
	}

	return c.Status(201).JSON(business)
}

func createNotification(c *fiber.Ctx) error {
	notification := new(Notification)
	if err := c.BodyParser(notification); err != nil {
		log.Println("ERROR: Most likely the product doesn't have an owner, can't set notification", notification)
		return c.Status(400).SendString(err.Error())
	}

	var exists bool
	err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", notification.ReceiverID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if !exists {
		return c.Status(400).SendString("User does not exist")
	}

	err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM businesses WHERE id = $1 AND user_id = $2)", notification.BusinessID, notification.ReceiverID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if !exists {
		return c.Status(400).SendString("Business does not exist or does not belong to the user")
	}

	query := `INSERT INTO notifications (id, product_id, product_name, receiver_id, business_id, review_title, from_name, from_id, read, comment_id, review_id) 
              VALUES (:id, :product_id, :product_name, :receiver_id, :business_id, :review_title, :from_name, :from_id, :read, :comment_id, :review_id) RETURNING id, created_at`
	rows, err := db.NamedQuery(query, notification)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&notification.ID, &notification.CreatedAt)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
	}

	return c.Status(201).JSON(notification)
}

func getLatestNotifications(c *fiber.Ctx) error {
	receiverID := c.Query("receiver_id")
	if receiverID == "" {
		return c.Status(400).SendString("receiver_id query parameter is required")
	}

	query := `SELECT * FROM notifications 
              WHERE receiver_id = $1 
              ORDER BY created_at DESC 
              LIMIT 1`
	var notification Notification
	err := db.Get(&notification, query, receiverID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return c.Status(200).SendString("No notifications found")
		}
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(notification)
}

func getAllNotifications(c *fiber.Ctx) error {
	receiverID := c.Query("receiver_id")
	if receiverID == "" {
		return c.Status(400).SendString("receiver_id query parameter is required")
	}

	query := `SELECT * FROM notifications 
              WHERE receiver_id = $1 
              ORDER BY created_at DESC`
	var notifications []Notification
	err := db.Select(&notifications, query, receiverID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(notifications)
}

func deleteReadNotifications(c *fiber.Ctx) error {
	receiverID := c.Query("receiver_id")
	if receiverID == "" {
		return c.Status(400).SendString("receiver_id query parameter is required")
	}

	query := `DELETE FROM notifications WHERE receiver_id = $1 AND read = true RETURNING *`
	var deletedNotifications []Notification
	err := db.Select(&deletedNotifications, query, receiverID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(fiber.Map{
		"deleted":       len(deletedNotifications),
		"notifications": deletedNotifications,
	})
}
