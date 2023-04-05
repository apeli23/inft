package handlers

import (
	"encoding/json"
	"fmt"
	"infinity/database"
	"infinity/models"
	"net/http"
	"time"
	"context"
	"github.com/gorilla/mux"
)

// view all transactions
func getAllTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	// Set response header
	w.Header().Set("Content-Type", "application/json")

	// Create a new database connection
	db, err := database.Connectdb()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to database: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query the transactions table
	rows, err := db.Query("SELECT id, subscription_id, transaction_date, amount, status, created_at, updated_at FROM transactions")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error querying database: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Build an array of Transactions objects
	var transactions []models.Transactions
	for rows.Next() {
		var transaction models.Transactions
		err := rows.Scan(&transaction.ID, &transaction.SubscriptionID, &transaction.TransactionDate, &transaction.Amount, &transaction.Status, &transaction.CreatedAt, &transaction.UpdatedAt)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error scanning row: %v", err), http.StatusInternalServerError)
			return
		}
		transactions = append(transactions, transaction)
	}

	// Encode the array of Partner objects in JSON format and write it to the response
	if err := json.NewEncoder(w).Encode(transactions); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
		return
	}
}

func createTransaction(w http.ResponseWriter, r *http.Request) {
	// Parse the request body into a Transactions struct
	var transaction models.Transactions
	err := json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the transaction data
	if transaction.SubscriptionID == 0 || transaction.Amount == 0 {
		http.Error(w, "SubscriptionID and Amount are required fields", http.StatusBadRequest)
		return
	}

	// Create a new database connection
	db, err := database.Connectdb()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Insert the new transaction into the transactions table
	sqlStatement := `INSERT INTO transactions (subscription_id, transaction_date, amount, status, created_at, updated_at)
					 VALUES ($1, $2, $3, $4, $5, $6)
					 RETURNING id`
	var id int
	err = db.QueryRow(sqlStatement, transaction.SubscriptionID, transaction.TransactionDate, transaction.Amount, transaction.Status, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response status code to 201 Created and include the new transaction's ID in the response body
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}


//find a partner
// Get a single transaction by ID
func getTransaction(w http.ResponseWriter, r *http.Request) {
	// Set response header
	w.Header().Set("Content-Type", "application/json")

	// Parse transaction ID from request URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Get database connection from the pool
	db, err := database.Connectdb()

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare SQL statement
	stmt, err := db.PrepareContext(ctx, "SELECT id, subscription_id, transaction_date, amount, status, created_at, updated_at FROM transactions WHERE id=$1")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error preparing statement: %v", err), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// Execute SQL query with the prepared statement
	row := stmt.QueryRowContext(ctx, id)

	// Scan the row into a Transactions object
	var transaction models.Transactions
	err = row.Scan(&transaction.ID, &transaction.SubscriptionID, &transaction.TransactionDate, &transaction.Amount, &transaction.Status, &transaction.CreatedAt, &transaction.UpdatedAt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error scanning row: %v", err), http.StatusInternalServerError)
		return
	}
	// Encode the Transactions object in JSON format and write it to the response
	if err := json.NewEncoder(w).Encode(transaction); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
	}
}

// update transaction
func updateTransaction(w http.ResponseWriter, r *http.Request) {
	// Set response header
	w.Header().Set("Content-Type", "application/json")

	// Get the transaction ID from the request URL
	id, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, "missing transaction ID", http.StatusBadRequest)
		return
	}

	// Parse the request body into a Transactions object
	var transaction models.Transactions
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		http.Error(w, fmt.Sprintf("error decoding request body: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Get a database connection from the pool
	db, err := database.Connectdb()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Update the transaction in the database
	res, err := db.ExecContext(r.Context(), `
		UPDATE transactions 
		SET subscription_id=$1, transaction_date=$2, amount=$3, status=$4, updated_at=$5 
		WHERE id=$6`,
		transaction.SubscriptionID, transaction.TransactionDate, transaction.Amount, transaction.Status, time.Now(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("error updating transaction: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting rows affected: %v", err), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "transaction not found", http.StatusNotFound)
		return
	}

	// Fetch the updated transaction from the database
	err = db.QueryRowContext(r.Context(), `
		SELECT id, subscription_id, transaction_date, amount, status, created_at, updated_at 
		FROM transactions 
		WHERE id=$1`, id).Scan(&transaction.ID, &transaction.SubscriptionID, &transaction.TransactionDate, &transaction.Amount, &transaction.Status, &transaction.CreatedAt, &transaction.UpdatedAt)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching updated transaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Encode the Transactions object in JSON format and write it to the response
	if err := json.NewEncoder(w).Encode(transaction); err != nil {
		http.Error(w, fmt.Sprintf("error encoding response: %v", err), http.StatusInternalServerError)
	}
}

//delete subscription
// Delete a transaction
func deleteTransaction(w http.ResponseWriter, r *http.Request) {
	// Set response header
	w.Header().Set("Content-Type", "application/json")

	// Get the ID of the transaction to delete from the URL parameters
	params := mux.Vars(r)
	id := params["id"]

	// Create a new database connection
	db, err := database.Connectdb()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to database: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Delete the transaction from the database
	_, err = db.Exec("DELETE FROM transactions WHERE id=$1", id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting transaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Write a success message to the response
	fmt.Fprintf(w, "Transaction %s deleted successfully", id)
}

func TransactionsRouter() * mux.Router {
	router  := mux.NewRouter()
	
	router.HandleFunc("/transactions", getAllTransactionsHandler).Methods("GET")
	router.HandleFunc("/transactions", createTransaction).Methods("POST")
	router.HandleFunc("/transactions/{id}", getTransaction).Methods("GET")
	router.HandleFunc("/transactions/{id}", updateTransaction).Methods("PUT")
	router.HandleFunc("/transactions/{id}", deleteTransaction).Methods("DELETE")

	return router
}
