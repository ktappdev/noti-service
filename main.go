package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Notification struct {
	ID         string    `db:"id" json:"id"`
	UserID     string    `db:"user_id" json:"user_id"`
	BusinessID string    `db:"business_id" json:"business_id"`
	Message    string    `db:"message" json:"message"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	From       string    `db:"from" json:"from"`
	Read       bool      `db:"read" json:"read"`
}

type User struct {
	ID       string `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Email    string `db:"email" json:"email"`
}

type Business struct {
	ID      string `db:"id" json:"id"`
	UserID  string `db:"user_id" json:"user_id"`
	Name    string `db:"name" json:"name"`
	Address string `db:"address" json:"address"`
}

var db *sql.DB

func main() {
	var err error
	// Load .env file
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = createSchema()
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	app.Post("/create-user", createUser)
	app.Post("/create-business", createBusiness)
	app.Post("/create-notification", createNotification)
	app.Get("/notifications/latest", getLatestNotifications)
	app.Get("/notifications", getAllNotifications)
	app.Delete("/notifications", deleteReadNotifications)

	log.Fatal(app.Listen(":3000"))
}

func createSchema() error {
	schema := `
    CREATE TABLE IF NOT EXISTS users (
        id VARCHAR(255) PRIMARY KEY,
        username VARCHAR(255) NOT NULL,
        email VARCHAR(255) NOT NULL UNIQUE
    );

    CREATE TABLE IF NOT EXISTS owners (
        id VARCHAR(255) PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        email VARCHAR(255) NOT NULL UNIQUE
    );

    CREATE TABLE IF NOT EXISTS businesses (
        id VARCHAR(255) PRIMARY KEY,
        user_id VARCHAR(255) NOT NULL,
        name VARCHAR(255) NOT NULL,
        address TEXT,
        FOREIGN KEY (user_id) REFERENCES owners(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS notifications (
        id VARCHAR(255) PRIMARY KEY,
        user_id VARCHAR(255) NOT NULL,
        business_id VARCHAR(255) NOT NULL,
        message TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        "from" VARCHAR(255) NOT NULL,
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

	query := `INSERT INTO users (id, username, email) VALUES ($1, $2, $3) RETURNING id`
	err := db.QueryRow(query, user.ID, user.Username, user.Email).Scan(&user.ID)
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

	// Check if the owner exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM owners WHERE id = $1)", business.UserID).Scan(&exists)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if !exists {
		return c.Status(400).SendString("Owner does not exist")
	}

	query := `INSERT INTO businesses (id, user_id, name, address) VALUES ($1, $2, $3, $4) RETURNING id`
	err = db.QueryRow(query, business.ID, business.UserID, business.Name, business.Address).Scan(&business.ID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Status(201).JSON(business)
}

func createNotification(c *fiber.Ctx) error {
	notification := new(Notification)
	if err := c.BodyParser(notification); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	// Check if the user exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", notification.UserID).Scan(&exists)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if !exists {
		return c.Status(400).SendString("User does not exist")
	}

	// Check if the business exists
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM businesses WHERE id = $1)", notification.BusinessID).Scan(&exists)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	if !exists {
		return c.Status(400).SendString("Business does not exist")
	}

	query := `INSERT INTO notifications (id, user_id, business_id, message, "from", read) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	err = db.QueryRow(query, notification.ID, notification.UserID, notification.BusinessID, notification.Message, notification.From, notification.Read).Scan(&notification.ID, &notification.CreatedAt)
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

	query := `SELECT * FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`
	var notification Notification
	err := db.QueryRow(query, userID).Scan(&notification.ID, &notification.UserID, &notification.BusinessID, &notification.Message, &notification.CreatedAt, &notification.From, &notification.Read)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).SendString("No notifications found")
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

	query := `SELECT * FROM notifications WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := db.Query(query, userID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var notification Notification
		if err := rows.Scan(&notification.ID, &notification.UserID, &notification.BusinessID, &notification.Message, &notification.CreatedAt, &notification.From, &notification.Read); err != nil {
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

	query := `DELETE FROM notifications WHERE user_id = $1 AND read = true`
	result, err := db.Exec(query, userID)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.JSON(fiber.Map{"deleted": rowsAffected})
}
