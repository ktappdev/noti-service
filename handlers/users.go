package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/ktappdev/noti-service/models"
)

// CreateUser creates a new user
func CreateUser(db *sqlx.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := new(models.User)
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
}