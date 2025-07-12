package handlers

import (
	"github.com/gin-gonic/gin"
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

// CreateUserGin creates a new user (Gin version)
func CreateUserGin(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := new(models.User)
		if err := c.ShouldBindJSON(user); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		query := `INSERT INTO users (id, username, full_name) VALUES (:id, :username, :full_name) RETURNING id`
		rows, err := db.NamedQuery(query, user)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		if rows.Next() {
			err = rows.Scan(&user.ID)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(201, user)
	}
}