package models

import(
	"time"
)


type Transaction struct {
	ID              int       `json:"id"`
	SubscriptionID  int       `json:"subscription_id"`
	TransactionDate time.Time `json:"transaction_date"`
	Amount          float64   `json:"amount"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
 
