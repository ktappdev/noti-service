package main

func dropTables() error {
	_, err := db.Exec(`
        DROP TABLE IF EXISTS notifications;
        DROP TABLE IF EXISTS owners;
        DROP TABLE IF EXISTS users;
    `)
	return err
}
