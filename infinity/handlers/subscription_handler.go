package handlers

import (
	"encoding/json"
	"fmt"
	"time"
	"context"
	"infinity/database"
	"infinity/models"
	"net/http"
	"github.com/gorilla/mux"
)

// view all subscriptions
func getAllSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	// Set response header
	w.Header().Set("Content-Type", "application/json")

	// Create a new database connection
	db, err := database.Connectdb()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to database: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query the subscriptions table
	rows, err := db.Query("SELECT id, customer_msisdn, subscription_date, status, billing_amount, billing_cycle, start_date, end_date, created_at, updated_at FROM subscriptions")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error querying database: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Build an array of Subscription objects
	var subscriptions []models.Subscriptions
	for rows.Next() {
		var subscription models.Subscriptions
		err := rows.Scan(&subscription.ID, &subscription.CustomerMSISDN, &subscription.SubscriptionDate, &subscription.Status, &subscription.BillingAmount, &subscription.BillingCycle, &subscription.StartDate, &subscription.EndDate, &subscription.CreatedAt, &subscription.UpdatedAt)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error scanning row: %v", err), http.StatusInternalServerError)
			return
		}
		subscriptions = append(subscriptions, subscription)
	}

	// Encode the array of Partner objects in JSON format and write it to the response
	if err := json.NewEncoder(w).Encode(subscriptions); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
		return
	}
}

// view a subscription 

func getSubscription(w http.ResponseWriter, r *http.Request) {
	 // Set response header
	w.Header().Set("Content-Type", "application/json")

	// Parse partner ID from request URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Get database connection from the pool
	db, err := database.Connectdb()

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare SQL statement
	stmt, err := db.PrepareContext(ctx, "SELECT id, customer_msisdn, subscription_date, status, billing_amount, billing_cycle, start_date, end_date, created_at, updated_at FROM subscriptions WHERE id=$1")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error preparing statement: %v", err), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// Execute SQL query with the prepared statement
	row := stmt.QueryRowContext(ctx, id)

	// Scan the row into a Subscription object
	var subscription models.Subscriptions
	err = row.Scan(&subscription.ID, &subscription.CustomerMSISDN, &subscription.SubscriptionDate, &subscription.Status, &subscription.BillingAmount, &subscription.BillingCycle, &subscription.StartDate, &subscription.EndDate, &subscription.CreatedAt, &subscription.UpdatedAt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error scanning row: %v", err), http.StatusInternalServerError)
		return
	}

	// Encode the Partner object in JSON format and write it to the response
	if err := json.NewEncoder(w).Encode(subscription); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
	}
}

//create a subscription
func createSubscription(w http.ResponseWriter, r *http.Request) {
	// Parse the request body into a Partner struct
	var subscription models.Subscriptions
	err := json.NewDecoder(r.Body).Decode(&subscription)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	 // Validate the subscription data
	 if subscription.PartnerID == 0 || subscription.CustomerMSISDN == "" || subscription.SubscriptionDate.IsZero() || subscription.StartDate.IsZero() || subscription.EndDate.IsZero() {
        http.Error(w, "PartnerID, CustomerMSISDN, SubscriptionDate, StartDate, and EndDate are required fields", http.StatusBadRequest)
        return
    }

	// Create a new database connection
	db, err := database.Connectdb()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Insert the new subscription into the subscriptions table
	sqlStatement := `INSERT INTO subscriptions (id,partner_id, customer_msisdn, subscription_date, status, billing_amount, billing_cycle, start_date, end_date, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
					RETURNING id`
	var id int
	err = db.QueryRow(sqlStatement, subscription.PartnerID, subscription.CustomerMSISDN, subscription.SubscriptionDate, subscription.Status, subscription.BillingAmount, subscription.BillingCycle, subscription.StartDate, subscription.EndDate, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response status code to 201 Created and include the new partner's ID in the response body
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

// update subscription
func updateSubscription(w http.ResponseWriter, r *http.Request) {
	// Set response header
	w.Header().Set("Content-Type", "application/json")

	// Get the partner ID from the request URL
	id, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, "missing subscription ID", http.StatusBadRequest)
		return
	}

	// Parse the request body into a Partner object
	var subscription models.Subscriptions
	if err := json.NewDecoder(r.Body).Decode(&subscription); err != nil {
		http.Error(w, fmt.Sprintf("error decoding request body: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	db, err := database.Connectdb()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()
	
	// Update the subscription in the database
	res, err := db.ExecContext(r.Context(), `
		UPDATE subscriptions 
		SET partner_id=$1, customer_msisdn=$2, subscription_date=$3, status=$4, billing_amount=$5, billing_cycle=$6, start_date=$7, end_date=$8, updated_at=$9 
		WHERE id=$10`,
		subscription.PartnerID, subscription.CustomerMSISDN, subscription.SubscriptionDate, subscription.Status, subscription.BillingAmount, subscription.BillingCycle, subscription.StartDate, subscription.EndDate, time.Now(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("error updating subscription: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting rows affected: %v", err), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "partner not found", http.StatusNotFound)
		return
	}

	// Fetch the updated partner from the database
	err = db.QueryRowContext(r.Context(), `
		SELECT id, customer_msisdn, subscription_date, status, billing_amount, billing_cycle, start_date, end_date, created_at, updated_at
		FROM subscriptions 
		WHERE id=$1`, id).Scan(&subscription.ID, &subscription.CustomerMSISDN, &subscription.SubscriptionDate, &subscription.Status, &subscription.BillingAmount, &subscription.BillingCycle, &subscription.StartDate, &subscription.EndDate, &subscription.CreatedAt, &subscription.UpdatedAt)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching updated subscription: %v", err), http.StatusInternalServerError)
		return
	}

	// Encode the Partner object in JSON format and write it to the response
	if err := json.NewEncoder(w).Encode(subscription); err != nil {
		http.Error(w, fmt.Sprintf("error encoding response: %v", err), http.StatusInternalServerError)
	}
}

//delete subscription
func deleteSubscription(w http.ResponseWriter, r *http.Request) {
	// Set response header
	w.Header().Set("Content-Type", "application/json")

	// Get the ID of the subscription to delete from the URL parameters
	params := mux.Vars(r)
	id := params["id"]

	// Create a new database connection
	db, err := database.Connectdb()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to database: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()
	
	// Delete the subscription from the database
	_, err = db.Exec("DELETE FROM subscriptions WHERE id=$1", id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting subscription: %v", err), http.StatusInternalServerError)
		return
	}

	// Write a success message to the response
	fmt.Fprintf(w, "Subscription %s deleted successfully", id)
}


func SubscriptionsRouter() * mux.Router {
	router  := mux.NewRouter()

	// endpoints for subscriptions
	router.HandleFunc("/subscriptions", getAllSubscriptionsHandler).Methods("GET")
	router.HandleFunc("/subscriptions/{id}", getSubscription).Methods("GET")
	router.HandleFunc("/subscriptions", createSubscription).Methods("POST")
	router.HandleFunc("/subscriptions/{id}", updateSubscription).Methods("PUT")
	router.HandleFunc("/subscriptions/{id}", deleteSubscription).Methods("DELETE")
	
	return router
}
