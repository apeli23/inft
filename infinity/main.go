package main

import (
	"fmt"
	"log"
	"net/http"
	"infinity/handlers"
	"github.com/gorilla/mux"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func init() {
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }
}
 
func main() {
	 
	router := mux.NewRouter()

// Use http.NewServeMux() to create a new ServeMux and register handlers
	router.Handle("/api", handlers.ValidateJWT(http.HandlerFunc(handlers.Home)))
	router.HandleFunc("/jwt", handlers.GetJWT)
	router.PathPrefix("/partners").Handler(handlers.PartnersRouter())
	router.PathPrefix("/subscriptions").Handler(handlers.SubscriptionsRouter())

// Use http.ListenAndServe() to start the server on port 8080
	if err := http.ListenAndServe(":8080", router); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}


