package database

import "github.com/jmoiron/sqlx"

// CreateSchema creates the database tables and indexes
func CreateSchema(db *sqlx.DB) error {
	schema := `
    CREATE TABLE IF NOT EXISTS users (
        id VARCHAR(255) PRIMARY KEY,
        username VARCHAR(255) NOT NULL,
        full_name VARCHAR(255) NOT NULL
    );

    CREATE TABLE IF NOT EXISTS user_notifications (
        id VARCHAR(255) PRIMARY KEY,
        parent_user_id VARCHAR(255) NOT NULL,
        content TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        read BOOLEAN DEFAULT FALSE,
        notification_type VARCHAR(50) NOT NULL,
        comment_id VARCHAR(255),
        from_id VARCHAR(255),
        review_id VARCHAR(255),
        parent_id VARCHAR(255),
        from_name VARCHAR(255),
        product_id VARCHAR(255),
        FOREIGN KEY (parent_user_id) REFERENCES users(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS product_owner_notifications (
        id VARCHAR(255) PRIMARY KEY,
        owner_id VARCHAR(255) NOT NULL,
        product_id VARCHAR(255) NOT NULL,
        product_name VARCHAR(255) NOT NULL,
        business_id VARCHAR(255) NOT NULL,
        review_title TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        from_name VARCHAR(255) NOT NULL,
        from_id VARCHAR(255) NOT NULL,
        read BOOLEAN DEFAULT FALSE,
        comment_id VARCHAR(255),
        review_id VARCHAR(255),
        notification_type VARCHAR(50) NOT NULL,
        FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS like_notifications (
        id VARCHAR(255) PRIMARY KEY,
        target_user_id VARCHAR(255) NOT NULL,
        target_type VARCHAR(50) NOT NULL CHECK (target_type IN ('comment', 'review')),
        target_id VARCHAR(255) NOT NULL,
        from_id VARCHAR(255) NOT NULL,
        from_name VARCHAR(255) NOT NULL,
        product_id VARCHAR(255),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        read BOOLEAN DEFAULT FALSE,
        FOREIGN KEY (target_user_id) REFERENCES users(id) ON DELETE CASCADE,
        FOREIGN KEY (from_id) REFERENCES users(id) ON DELETE CASCADE
    );

    CREATE INDEX IF NOT EXISTS idx_user_notifications_user_id ON user_notifications(parent_user_id);
    CREATE INDEX IF NOT EXISTS idx_user_notifications_created_at ON user_notifications(created_at);
    CREATE INDEX IF NOT EXISTS idx_product_owner_notifications_owner_id ON product_owner_notifications(owner_id);
    CREATE INDEX IF NOT EXISTS idx_product_owner_notifications_created_at ON product_owner_notifications(created_at);
    CREATE INDEX IF NOT EXISTS idx_like_notifications_target_user_id ON like_notifications(target_user_id);
    CREATE INDEX IF NOT EXISTS idx_like_notifications_created_at ON like_notifications(created_at);
    `
	_, err := db.Exec(schema)
	return err
}