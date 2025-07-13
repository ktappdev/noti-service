package main

import "github.com/jmoiron/sqlx"

func dropTables(db *sqlx.DB) error {
	_, err := db.Exec(`
        DROP TABLE IF EXISTS notifications;
        DROP TABLE IF EXISTS product_owners_notifications;
        DROP TABLE IF EXISTS user_notifications;
        DROP TABLE IF EXISTS owners;
        DROP TABLE IF EXISTS users;
    `)
	return err
}
