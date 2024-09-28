package reviewit

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// GetParentCommentUserID retrieves the userId of the user who made the parent comment
func GetParentCommentUserID(parentID string) (string, error) {
	dbURL := os.Getenv("REVIEWIT_DATABASE_URL")
	if dbURL == "" {
		fmt.Println("REVIEWIT_DATABASE_URL environment variable is required")
	}
	// Connect to the database
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		return "", fmt.Errorf("error connecting to database: %w", err)
	}
	defer db.Close()
	fmt.Println(db.Stats().MaxOpenConnections)

	// SQL query to get the userId directly from the Comment table
	query := `
		SELECT "userId"
		FROM "Comment"
		WHERE "id" = $1
	`

	var userID string
	err = db.Get(&userID, query, parentID)
	if err != nil {
		return "", fmt.Errorf("error querying database: %w", err)
	}
	fmt.Println(userID)

	return userID, nil
}
