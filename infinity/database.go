package main

import (
	"fmt"
	"database/sql"
	"os"
)


// Database Connection
func connectdb() (*sql.DB, error) {
	// Open a connection to the database
	db, err := sql.Open("postgres", os.Getenv("DB_CONN_STR"))
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}

	// Ping the database to verify that the connection is working
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return db, nil
}