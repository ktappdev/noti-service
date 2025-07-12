package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/ktappdev/noti-service/database"
	"github.com/ktappdev/noti-service/handlers"
	"github.com/ktappdev/noti-service/sse"
	_ "github.com/lib/pq"
)

var db *sqlx.DB
var sseHub *sse.SSEHub

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

	if err := database.CreateSchema(db); err != nil {
		log.Fatal(err)
	}

	sseHub = sse.NewSSEHub()
	go sseHub.Run()

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Cache-Control, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Routes
	r.POST("/users", handlers.CreateUserGin(db))
	r.POST("/notifications/product-owner", handlers.CreateProductOwnerNotificationGin(db, sseHub))
	r.POST("/notifications/reply", handlers.CreateReplyNotificationGin(db, sseHub))
	r.POST("/notifications/like", handlers.CreateLikeNotificationGin(db, sseHub))
	r.GET("/notifications/latest", handlers.GetLatestNotificationsGin(db))
	r.GET("/notifications", handlers.GetAllNotificationsGin(db))
	r.GET("/notifications/unread", handlers.GetAllUnreadNotificationsGin(db))
	r.DELETE("/notifications", handlers.DeleteReadNotificationsGin(db))
	r.PUT("/notifications/:id/read", handlers.MarkNotificationAsReadGin(db, sseHub))

	// SSE route
	r.GET("/notifications/stream", handlers.StreamNotificationsGin(db, sseHub))
	// SSE documentation route
	r.GET("/sse-help", handlers.SSEHelpHandlerGin())
	// Test SSE endpoint
	r.GET("/test/sse", handlers.TestSSEHandlerGin())

	log.Printf("Server starting on port 3001...")
	r.Run(":3001")
}
