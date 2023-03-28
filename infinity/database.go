package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"
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
	rows, err := db.Query("SELECT * FROM partners")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, email, phoneNumber, billingAddress string
		var createdAt, updatedAt time.Time
	
		err := rows.Scan(&id, &name, &email, &phoneNumber, &billingAddress, &createdAt, &updatedAt)
		if err != nil {
			panic(err)
		}
		// fmt.Printf("id: %d, name: %s, email: %s, phone number: %s, billing address: %s, created at: %s, updated at: %s\n", 
		// 	id, name, email, phoneNumber, billingAddress, createdAt.Format(time.RFC3339), updatedAt.Format(time.RFC3339))
	}
	return db, nil
}