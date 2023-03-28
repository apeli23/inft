package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func init() {
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }
}

// RESTAPI
 

// Home is a handler function that writes "super secret area" to the ResponseWriter
func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "super secret area")
}

// CreateJWT generates a new JWT token with an expiration time of one hour
func CreateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour).Unix()

	tokenStr, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

// ValidateJWT is a middleware function that validates the JWT token in the "Token" header
// If the token is valid, the next handler is called. Otherwise, a 401 Unauthorized response is returned
func ValidateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Token")

		// If there's no token in the request header, return a 401 Unauthorized response
		if tokenStr == "" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "not authorized")
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			// Make sure the token's signing method is HMAC
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(os.Getenv("SECRET_KEY")), nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "not authorized: %v", err)
			return
		}

		if !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "not authorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}
var apiKey = os.Getenv("API_KEY")

// GetJWT is a handler function that generates a JWT token and returns it in the ResponseWriter
func GetJWT(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("Access")

	// If the "Access" header doesn't contain the API key, return a 401 Unauthorized response
	if apiKey != apiKey {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "not authorized")
		return
	}

	token, err := CreateJWT()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error creating token: %v", err)
		return
	}

	fmt.Fprint(w, token)
}

func main() {
		// DATABASE
	fmt.Println("opening database...")
	db, err := connectdb()
	if err != nil {
		fmt.Printf("failed to connect to database: %v\n", err)
		return
	}
	defer db.Close() 
	 
	router := mux.NewRouter()

// endpoints for partners
	router.HandleFunc("/api/partners", getAllPartnersHandler).Methods("GET")
	router.HandleFunc("/api/partners", createPartner).Methods("POST")
	router.HandleFunc("/api/partners/{id}", getPartner).Methods("GET")
	router.HandleFunc("/api/partners/{id}", updatePartner).Methods("PUT")
	router.HandleFunc("/api/partners/{id}", deletePartner).Methods("DELETE")



	// Use http.NewServeMux() to create a new ServeMux and register handlers
	router.Handle("/api", ValidateJWT(http.HandlerFunc(Home)))
	router.HandleFunc("/jwt", GetJWT)

	// Use http.ListenAndServe() to start the server on port 8080
	if err := http.ListenAndServe(":8080", router); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}


// view all partners
func getAllPartnersHandler(w http.ResponseWriter, r *http.Request) {
	// Set response header
	w.Header().Set("Content-Type", "application/json")

	// Create a new database connection
	db, err := connectdb()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to database: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query the partners table
	rows, err := db.Query("SELECT id, name, email, phone_number, billing_address, created_at, updated_at FROM partners")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error querying database: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Build an array of Partner objects
	var partners []Partner
	for rows.Next() {
		var partner Partner
		err := rows.Scan(&partner.ID, &partner.Name, &partner.Email, &partner.PhoneNumber, &partner.BillingAddress, &partner.CreatedAt, &partner.UpdatedAt)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error scanning row: %v", err), http.StatusInternalServerError)
			return
		}
		partners = append(partners, partner)
	}

	// Encode the array of Partner objects in JSON format and write it to the response
	if err := json.NewEncoder(w).Encode(partners); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
		return
	}
}

//create a partmer
func createPartner(w http.ResponseWriter, r *http.Request) {
	// Parse the request body into a Partner struct
	var partner Partner
	err := json.NewDecoder(r.Body).Decode(&partner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the partner data
	if partner.Name == "" || partner.Email == "" {
		http.Error(w, "Name and Email are required fields", http.StatusBadRequest)
		return
	}

	// Create a new database connection
	db, err := connectdb()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Insert the new partner into the partners table
	sqlStatement := `INSERT INTO partners (name, email, phone_number, billing_address, created_at, updated_at)
					 VALUES ($1, $2, $3, $4, $5, $6)
					 RETURNING id`
	var id int
	err = db.QueryRow(sqlStatement, partner.Name, partner.Email, partner.PhoneNumber, partner.BillingAddress, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response status code to 201 Created and include the new partner's ID in the response body
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

//find a partner
func getPartner(w http.ResponseWriter, r *http.Request) {
	// Set response header
	w.Header().Set("Content-Type", "application/json")

	// Parse partner ID from request URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Get database connection from the pool
	db, err := connectdb()

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare SQL statement
	stmt, err := db.PrepareContext(ctx, "SELECT id, name, email, phone_number, billing_address, created_at, updated_at FROM partners WHERE id=$1")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error preparing statement: %v", err), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// Execute SQL query with the prepared statement
	row := stmt.QueryRowContext(ctx, id)

	// Scan the row into a Partner object
	var partner Partner
	err = row.Scan(&partner.ID, &partner.Name, &partner.Email, &partner.PhoneNumber, &partner.BillingAddress, &partner.CreatedAt, &partner.UpdatedAt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error scanning row: %v", err), http.StatusInternalServerError)
		return
	}

	// Encode the Partner object in JSON format and write it to the response
	if err := json.NewEncoder(w).Encode(partner); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
	}
}

// update a partner
func updatePartner(w http.ResponseWriter, r *http.Request) {
	// Set response header
	w.Header().Set("Content-Type", "application/json")

	// Get the partner ID from the request URL
	id, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, "missing partner ID", http.StatusBadRequest)
		return
	}

	// Parse the request body into a Partner object
	var partner Partner
	if err := json.NewDecoder(r.Body).Decode(&partner); err != nil {
		http.Error(w, fmt.Sprintf("error decoding request body: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	db, err := connectdb()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()
	
	// Update the partner in the database
	res, err := db.ExecContext(r.Context(), `
		UPDATE partners 
		SET name=$1, email=$2, phone_number=$3, billing_address=$4, updated_at=$5 
		WHERE id=$6`,
		partner.Name, partner.Email, partner.PhoneNumber, partner.BillingAddress, time.Now(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("error updating partner: %v", err), http.StatusInternalServerError)
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
		SELECT id, name, email, phone_number, billing_address, created_at, updated_at 
		FROM partners 
		WHERE id=$1`, id).Scan(&partner.ID, &partner.Name, &partner.Email, &partner.PhoneNumber, &partner.BillingAddress, &partner.CreatedAt, &partner.UpdatedAt)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching updated partner: %v", err), http.StatusInternalServerError)
		return
	}

	// Encode the Partner object in JSON format and write it to the response
	if err := json.NewEncoder(w).Encode(partner); err != nil {
		http.Error(w, fmt.Sprintf("error encoding response: %v", err), http.StatusInternalServerError)
	}
}

//delete a partner
func deletePartner(w http.ResponseWriter, r *http.Request) {
	// Set response header
	w.Header().Set("Content-Type", "application/json")

	// Get the ID of the partner to delete from the URL parameters
	params := mux.Vars(r)
	id := params["id"]

	// Create a new database connection
	db, err := connectdb()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error connecting to database: %v", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Delete the partner from the database
	_, err = db.Exec("DELETE FROM partners WHERE id=$1", id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error deleting partner: %v", err), http.StatusInternalServerError)
		return
	}

	// Write a success message to the response
	fmt.Fprintf(w, "Partner %s deleted successfully", id)
}

