package reviewit

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// GetParentCommentUserID retrieves the userId of the user who made the parent comment
func GetParentCommentUserID(db *sqlx.DB, parentID string) (string, error) {
	// SQL query to get the userId directly from the Comment table
	query := `
		SELECT "userId"
		FROM "Comment"
		WHERE "id" = $1
	`

	var userID string
	err := db.Get(&userID, query, parentID)
	if err != nil {
		return "", fmt.Errorf("error querying database: %w", err)
	}

	return userID, nil
}

// GetCommentUserID retrieves the userId of the user who made the comment
func GetCommentUserID(db *sqlx.DB, commentID string) (string, error) {
	// SQL query to get the userId directly from the Comment table
	query := `
		SELECT "userId"
		FROM "Comment"
		WHERE "id" = $1
	`

	var userID string
	err := db.Get(&userID, query, commentID)
	if err != nil {
		return "", fmt.Errorf("error querying database: %w", err)
	}

	return userID, nil
}

// GetReviewUserID retrieves the userId of the user who made the review
func GetReviewUserID(db *sqlx.DB, reviewID string) (string, error) {
	// SQL query to get the userId directly from the Review table
	query := `
		SELECT "userId"
		FROM "Review"
		WHERE "id" = $1
	`

	var userID string
	err := db.Get(&userID, query, reviewID)
	if err != nil {
		return "", fmt.Errorf("error querying database: %w", err)
	}

	return userID, nil
}
