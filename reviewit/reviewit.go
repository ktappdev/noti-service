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
		return "", fmt.Errorf("REVIEWIT_DATABASE_URL environment variable is required")
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

// GetCommentUserID retrieves the userId of the user who made the comment
func GetCommentUserID(commentID string) (string, error) {
	dbURL := os.Getenv("REVIEWIT_DATABASE_URL")
	if dbURL == "" {
		return "", fmt.Errorf("REVIEWIT_DATABASE_URL environment variable is required")
	}
	// Connect to the database
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		return "", fmt.Errorf("error connecting to database: %w", err)
	}
	defer db.Close()

	// SQL query to get the userId directly from the Comment table
	query := `
		SELECT "userId"
		FROM "Comment"
		WHERE "id" = $1
	`

	var userID string
	err = db.Get(&userID, query, commentID)
	if err != nil {
		return "", fmt.Errorf("error querying database: %w", err)
	}

	return userID, nil
}

// GetReviewUserID retrieves the userId of the user who made the review
func GetReviewUserID(reviewID string) (string, error) {
	dbURL := os.Getenv("REVIEWIT_DATABASE_URL")
	if dbURL == "" {
		return "", fmt.Errorf("REVIEWIT_DATABASE_URL environment variable is required")
	}
	// Connect to the database
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		return "", fmt.Errorf("error connecting to database: %w", err)
	}
	defer db.Close()

	// SQL query to get the userId directly from the Review table
	query := `
		SELECT "userId"
		FROM "Review"
		WHERE "id" = $1
	`

	var userID string
	err = db.Get(&userID, query, reviewID)
	if err != nil {
		return "", fmt.Errorf("error querying database: %w", err)
	}

	return userID, nil
}

// GetCommentReviewID retrieves the reviewId that a comment belongs to
func GetCommentReviewID(commentID string) (string, error) {
	dbURL := os.Getenv("REVIEWIT_DATABASE_URL")
	if dbURL == "" {
		return "", fmt.Errorf("REVIEWIT_DATABASE_URL environment variable is required")
	}
	// Connect to the database
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		return "", fmt.Errorf("error connecting to database: %w", err)
	}
	defer db.Close()

	// SQL query to get the reviewId from the Comment table
	query := `
		SELECT "reviewId"
		FROM "Comment"
		WHERE "id" = $1
	`

	var reviewID string
	err = db.Get(&reviewID, query, commentID)
	if err != nil {
		return "", fmt.Errorf("error querying database: %w", err)
	}

	return reviewID, nil
}

// GetUserFullName retrieves the full name of a user by their ID
func GetUserFullName(userID string) (string, error) {
	dbURL := os.Getenv("REVIEWIT_DATABASE_URL")
	if dbURL == "" {
		return "", fmt.Errorf("REVIEWIT_DATABASE_URL environment variable is required")
	}
	// Connect to the database
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		return "", fmt.Errorf("error connecting to database: %w", err)
	}
	defer db.Close()

	// SQL query to get the user's full name from the User table
	query := `
		SELECT "fullName"
		FROM "User"
		WHERE "id" = $1
	`

	var fullName string
	err = db.Get(&fullName, query, userID)
	if err != nil {
		return "", fmt.Errorf("error querying database: %w", err)
	}

	return fullName, nil
}
