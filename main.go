package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

	// Initialize and start SSE hub
	sseHub = sse.NewSSEHub()
	go sseHub.Run()

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Cache-Control, Authorization, X-Requested-With",
		AllowCredentials: false,
		ExposeHeaders:    "Content-Length, Content-Type",
	}))
	// app.Use(logger.New(logger.Config{
	// 	Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	// }))

	// Routes
	app.Post("/users", handlers.CreateUser(db))
	app.Post("/notifications/product-owner", handlers.CreateProductOwnerNotification(db, sseHub))
	app.Post("/notifications/reply", handlers.CreateReplyNotification(db, sseHub))
	app.Get("/notifications/latest", handlers.GetLatestNotifications(db))
	app.Get("/notifications", handlers.GetAllNotifications(db))
	app.Get("/notifications/unread", handlers.GetAllUnreadNotifications(db))
	app.Delete("/notifications", handlers.DeleteReadNotifications(db))
	app.Post("/notifications/:id/read", handlers.MarkNotificationAsRead(db, sseHub))

	// SSE route
	app.Get("/notifications/stream", handlers.StreamNotifications(db, sseHub))
	
	// SSE documentation route
	app.Get("/sse-help", handlers.SSEHelpHandler())
	
	// Test SSE endpoint
	app.Get("/test/sse", handlers.TestSSEHandler())

	log.Printf("Server starting on port 3001...")
	log.Printf("SSE endpoint available at: /notifications/stream?user_id=YOUR_USER_ID")
	log.Printf("SSE documentation available at: /sse-help")
	log.Fatal(app.Listen(":3001"))
}
