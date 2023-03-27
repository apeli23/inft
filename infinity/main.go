package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"

	_ "github.com/lib/pq"
)

// RESTAPI
const (
	secretKey = "super-secret-key"
	apiKey    = "12345"
)

// Home is a handler function that writes "super secret area" to the ResponseWriter
func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "super secret area")
}

// CreateJWT generates a new JWT token with an expiration time of one hour
func CreateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(time.Hour).Unix()

	tokenStr, err := token.SignedString([]byte(secretKey))
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
			return []byte(secretKey), nil
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
	fmt.Println("opening database...")
	// DATABASE
	db, err := sql.Open("postgres", "postgres://postgres:Postgres_94@localhost/infinity?sslmode=disable")

	fmt.Println("db", db)
	if err != nil {
		fmt.Println("error...")
		panic(err)
	}
	defer db.Close()
// Postgres_94
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
		fmt.Printf("id: %d, name: %s, email: %s, phone number: %s, billing address: %s, created at: %s, updated at: %s\n", 
			id, name, email, phoneNumber, billingAddress, createdAt.Format(time.RFC3339), updatedAt.Format(time.RFC3339))
	}

	fmt.Println("opening api...")
	// Use http.NewServeMux() to create a new ServeMux and register handlers
	mux := http.NewServeMux()
	mux.Handle("/api", ValidateJWT(http.HandlerFunc(Home)))
	mux.HandleFunc("/jwt", GetJWT)

	// Use http.ListenAndServe() to start the server on port 8080
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
	
 
