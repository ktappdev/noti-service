package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/ktappdev/noti-service/database"
	"github.com/ktappdev/noti-service/handlers"
	"github.com/ktappdev/noti-service/sse"
	_ "github.com/lib/pq"
)

// DatabaseManager manages database connections with proper pooling
type DatabaseManager struct {
	NotificationDB *sqlx.DB
	ReviewitDB     *sqlx.DB
}

// NewDatabaseManager creates a new database manager with proper connection pooling
func NewDatabaseManager() (*DatabaseManager, error) {
	// Load notification database connection
	notiConnStr := os.Getenv("DATABASE_URL")
	if notiConnStr == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	notiDB, err := sqlx.Connect("postgres", notiConnStr)
	if err != nil {
		return nil, err
	}

	// Configure notification DB connection pool
	notiDB.SetMaxOpenConns(25)                 // Maximum number of open connections
	notiDB.SetMaxIdleConns(5)                  // Maximum number of idle connections
	notiDB.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime of a connection

	// Load ReviewIt database connection
	reviewitConnStr := os.Getenv("REVIEWIT_DATABASE_URL")
	if reviewitConnStr == "" {
		log.Fatal("REVIEWIT_DATABASE_URL environment variable is required")
	}

	reviewitDB, err := sqlx.Connect("postgres", reviewitConnStr)
	if err != nil {
		notiDB.Close() // Clean up notification DB if ReviewIt DB fails
		return nil, err
	}

	// Configure ReviewIt DB connection pool
	reviewitDB.SetMaxOpenConns(15)                 // Slightly smaller pool for external DB
	reviewitDB.SetMaxIdleConns(3)                  // Fewer idle connections for external DB
	reviewitDB.SetConnMaxLifetime(3 * time.Minute) // Shorter lifetime for external DB

	return &DatabaseManager{
		NotificationDB: notiDB,
		ReviewitDB:     reviewitDB,
	}, nil
}

// Close closes all database connections
func (dm *DatabaseManager) Close() error {
	var err1, err2 error
	if dm.NotificationDB != nil {
		err1 = dm.NotificationDB.Close()
	}
	if dm.ReviewitDB != nil {
		err2 = dm.ReviewitDB.Close()
	}

	// Return the first error encountered
	if err1 != nil {
		return err1
	}
	return err2
}

// LogConnectionStats logs current connection pool statistics
func (dm *DatabaseManager) LogConnectionStats() {
	if dm.NotificationDB != nil {
		stats := dm.NotificationDB.Stats()
		log.Printf("NotificationDB Stats - Open: %d, InUse: %d, Idle: %d",
			stats.OpenConnections, stats.InUse, stats.Idle)
	}
	if dm.ReviewitDB != nil {
		stats := dm.ReviewitDB.Stats()
		log.Printf("ReviewitDB Stats - Open: %d, InUse: %d, Idle: %d",
			stats.OpenConnections, stats.InUse, stats.Idle)
	}
}

var dbManager *DatabaseManager
var sseHub *sse.SSEHub

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database manager with connection pooling
	var err error
	dbManager, err = NewDatabaseManager()
	if err != nil {
		log.Fatal("Failed to initialize database manager:", err)
	}

	defer func() {
		if err := dbManager.Close(); err != nil {
			log.Printf("Error closing database connections: %v", err)
		}
	}()

	// Create schema using the notification database
	if err := database.CreateSchema(dbManager.NotificationDB); err != nil {
		log.Fatal(err)
	}

	// Initialize and start SSE hub
	sseHub = sse.NewSSEHub()
	go sseHub.Run()

	// Start periodic connection stats logging (every 5 minutes)
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				dbManager.LogConnectionStats()
			}
		}
	}()

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
	app.Post("/users", handlers.CreateUser(dbManager.NotificationDB))
	app.Post("/notifications/product-owner", handlers.CreateProductOwnerNotification(dbManager.NotificationDB, sseHub))
	app.Post("/notifications/reply", handlers.CreateReplyNotification(dbManager.NotificationDB, dbManager.ReviewitDB, sseHub))
	app.Post("/notifications/like", handlers.CreateLikeNotification(dbManager.NotificationDB, dbManager.ReviewitDB, sseHub))
	app.Get("/notifications/latest", handlers.GetLatestNotifications(dbManager.NotificationDB))
	app.Get("/notifications", handlers.GetAllNotifications(dbManager.NotificationDB))
	app.Get("/notifications/unread", handlers.GetAllUnreadNotifications(dbManager.NotificationDB))
	app.Delete("/notifications", handlers.DeleteReadNotifications(dbManager.NotificationDB))
	app.Put("/notifications/:id/read", handlers.MarkNotificationAsRead(dbManager.NotificationDB, sseHub))

	// SSE route
	app.Get("/notifications/stream", handlers.StreamNotifications(dbManager.NotificationDB, sseHub))

	// SSE documentation route
	app.Get("/sse-help", handlers.SSEHelpHandler())

	// Test SSE endpoint
	app.Get("/test/sse", handlers.TestSSEHandler())

	log.Printf("Server starting on port 3001...")
	log.Fatal(app.Listen(":3001"))
}
