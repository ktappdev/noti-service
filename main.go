package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Notification struct {
	ID          string    `json:"id"`
	ProductID   string    `json:"product_id"`
	ProductName string    `json:"product_name"`
	UserID      string    `json:"user_id"`
	BusinessID  string    `json:"business_id"`
	ReviewTitle string    `json:"review_title"`
	CreatedAt   time.Time `json:"created_at"`
	FromName    string    `json:"from_name"`
	FromID      string    `json:"from_id"`
	Read        bool      `json:"read"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

type Business struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	BusinessName string `json:"business_name"`
}

var db *sql.DB

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	var err error
	db, err = sql.Open("postgres", connStr)
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
        user_id VARCHAR(255) NOT NULL,
        business_id VARCHAR(255) NOT NULL,
        review_title TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        from_name VARCHAR(255) NOT NULL,
        from_id VARCHAR(255) NOT NULL,
        read BOOLEAN DEFAULT FALSE,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
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

	query := `INSERT INTO users (id, username, full_name) VALUES ($1, $2, $3) RETURNING id`
	err := db.QueryRow(query, user.ID, user.Username, user.FullName).Scan(&user.ID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Status(201).JSON(user)
}

func createBusiness(c *fiber.Ctx) error {
	business := new(Business)
	if err := c.BodyParser(business); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", business.UserID).Scan(&exists)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if !exists {
		return c.Status(400).SendString("User does not exist")
	}

	query := `INSERT INTO businesses (id, user_id, business_name) VALUES ($1, $2, $3) RETURNING id`
	err = db.QueryRow(query, business.ID, business.UserID, business.BusinessName).Scan(&business.ID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Status(201).JSON(business)
}

func createNotification(c *fiber.Ctx) error {
	notification := new(Notification)
	if err := c.BodyParser(notification); err != nil {
		log.Println("ERROR: Mostlikely the product don't have an owner, can't set notification", notification)
		return c.Status(400).SendString(err.Error())
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", notification.UserID).Scan(&exists)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if !exists {
		return c.Status(400).SendString("User does not exist")
	}

	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM businesses WHERE id = $1 AND user_id = $2)", notification.BusinessID, notification.UserID).Scan(&exists)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if !exists {
		return c.Status(400).SendString("Business does not exist or does not belong to the user")
	}

	query := `INSERT INTO notifications (id, product_id, product_name, user_id, business_id, review_title, from_name, from_id, read) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at`
	err = db.QueryRow(query, notification.ID, notification.ProductID, notification.ProductName, notification.UserID, notification.BusinessID,
		notification.ReviewTitle, notification.FromName, notification.FromID, notification.Read).Scan(&notification.ID, &notification.CreatedAt)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Status(201).JSON(notification)
}

func getLatestNotifications(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(400).SendString("user_id query parameter is required")
	}

	query := `SELECT id, product_id, product_name, user_id, business_id, review_title, created_at, from_name, from_id, read 
              FROM notifications 
              WHERE user_id = $1 
              ORDER BY created_at DESC 
              LIMIT 1`
	var notification Notification
	err := db.QueryRow(query, userID).Scan(
		&notification.ID,
		&notification.ProductID,
		&notification.ProductName,
		&notification.UserID,
		&notification.BusinessID,
		&notification.ReviewTitle,
		&notification.CreatedAt,
		&notification.FromName,
		&notification.FromID,
		&notification.Read,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(200).SendString("No notifications found")
		}
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(notification)
}

func getAllNotifications(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(400).SendString("user_id query parameter is required")
	}

	query := `SELECT id, product_id, product_name, user_id, business_id, review_title, created_at, from_name, from_id, read 
              FROM notifications 
              WHERE user_id = $1 
              ORDER BY created_at DESC`
	rows, err := db.Query(query, userID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var notification Notification
		if err := rows.Scan(
			&notification.ID,
			&notification.ProductID,
			&notification.ProductName,
			&notification.UserID,
			&notification.BusinessID,
			&notification.ReviewTitle,
			&notification.CreatedAt,
			&notification.FromName,
			&notification.FromID,
			&notification.Read,
		); err != nil {
			return c.Status(500).SendString(err.Error())
		}
		notifications = append(notifications, notification)
	}

	return c.JSON(notifications)
}

func deleteReadNotifications(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(400).SendString("user_id query parameter is required")
	}

	query := `DELETE FROM notifications WHERE user_id = $1 AND read = true RETURNING id, product_id, product_name, user_id, business_id, review_title, created_at, from_name, from_id, read`
	rows, err := db.Query(query, userID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()

	var deletedNotifications []Notification
	for rows.Next() {
		var notification Notification
		if err := rows.Scan(
			&notification.ID,
			&notification.ProductID,
			&notification.ProductName,
			&notification.UserID,
			&notification.BusinessID,
			&notification.ReviewTitle,
			&notification.CreatedAt,
			&notification.FromName,
			&notification.FromID,
			&notification.Read,
		); err != nil {
			return c.Status(500).SendString(err.Error())
		}
		deletedNotifications = append(deletedNotifications, notification)
	}

	return c.JSON(fiber.Map{
		"deleted":       len(deletedNotifications),
		"notifications": deletedNotifications,
	})
}
