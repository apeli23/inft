package models

import(
	"time"
)

type Partner struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	PhoneNumber    string    `json:"phone_number"`
	BillingAddress string    `json:"billing_address"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}