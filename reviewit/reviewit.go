package reviewit

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// GetParentCommentUserID retrieves the userId of the user who made the parent comment
func GetParentCommentUserID(parentID string) (string, error) {
	fmt.Println("=== START: GetParentCommentUserID ===")
	fmt.Printf("Looking up user ID for parent comment ID: %s\n", parentID)

	dbURL := os.Getenv("REVIEWIT_DATABASE_URL")
	if dbURL == "" {
		fmt.Println("ERROR: REVIEWIT_DATABASE_URL environment variable is required")
		return "", fmt.Errorf("REVIEWIT_DATABASE_URL environment variable is required")
	}

	// Print connection string for debugging
	fmt.Printf("Connecting to ReviewIt database: %s\n", dbURL)

	// Connect to the database
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		fmt.Printf("ERROR connecting to ReviewIt database: %v\n", err)
		return "", fmt.Errorf("error connecting to database: %w", err)
	}
	defer db.Close()
	fmt.Printf("Database connection stats: MaxOpenConnections=%d\n", db.Stats().MaxOpenConnections)

	// SQL query to get the userId directly from the Comment table
	query := `
		SELECT "userId"
		FROM "Comment"
		WHERE "id" = $1
	`
	fmt.Printf("Executing SQL query: %s with parentID=%s\n", query, parentID)

	var userID string
	err = db.Get(&userID, query, parentID)
	if err != nil {
		fmt.Printf("ERROR querying database: %v\n", err)
		return "", fmt.Errorf("error querying database: %w", err)
	}
	fmt.Printf("Successfully retrieved userID: %s for parentID: %s\n", userID, parentID)
	fmt.Println("=== END: GetParentCommentUserID ===")

	return userID, nil
}
