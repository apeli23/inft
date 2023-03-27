package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"log"
	"github.com/dgrijalva/jwt-go"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
    "os"
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

	fmt.Println("db", db)
	if err != nil {
		fmt.Printf("failed to connect to database: %v\n", err)
		return
	}
	defer db.Close()
	 

		// REST API

	// Use http.NewServeMux() to create a new ServeMux and register handlers
	mux := http.NewServeMux()
	mux.Handle("/api", ValidateJWT(http.HandlerFunc(Home)))
	mux.HandleFunc("/jwt", GetJWT)

	// Use http.ListenAndServe() to start the server on port 8080
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
	
 
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
