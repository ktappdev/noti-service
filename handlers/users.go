package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/ktappdev/noti-service/models"
)

// CreateUser creates a new user (idempotent - handles duplicates gracefully)
func CreateUser(db *sqlx.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := new(models.User)
		if err := c.BodyParser(user); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// Validate required field
		if user.ID == "" {
			return c.Status(400).SendString("User ID is required")
		}

		// Provide defaults for nullable fields to prevent DB errors
		if user.Username == "" {
			user.Username = "user_" + user.ID // Default username based on ID
		}
		if user.FullName == "" {
			user.FullName = "User " + user.ID // Default full name based on ID
		}

		// Use PostgreSQL's UPSERT (INSERT ... ON CONFLICT) for true idempotency
		query := `
			INSERT INTO users (id, username, full_name) 
			VALUES ($1, $2, $3) 
			ON CONFLICT ON CONSTRAINT users_pkey DO UPDATE SET 
				username = COALESCE(NULLIF(EXCLUDED.username, ''), users.username),
				full_name = COALESCE(NULLIF(EXCLUDED.full_name, ''), users.full_name)
			RETURNING id, username, full_name`
		
		var resultUser models.User
		err := db.QueryRow(query, user.ID, user.Username, user.FullName).Scan(
			&resultUser.ID, 
			&resultUser.Username, 
			&resultUser.FullName,
		)
		if err != nil {
			return c.Status(500).SendString("Database error: " + err.Error())
		}

		// Check if this was an insert (new user) or update (existing user)
		var exists bool
		err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", user.ID)
		if err != nil {
			// If we can't determine, assume it was created (safer for logging)
			return c.Status(201).JSON(resultUser)
		}

		// Return 200 for existing user (idempotent), 201 for new user
		// Since we used ON CONFLICT, we can't easily distinguish, so return 200 (idempotent behavior)
		return c.Status(200).JSON(resultUser)
	}
}