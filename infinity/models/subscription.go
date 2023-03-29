package models

import(
	"time"
)

type Subscriptions struct {
	ID               int       `json:"id"`
	PartnerID        int       `json:"partner_id"`
	CustomerMSISDN   string    `json:"customer_msisdn"`
	SubscriptionDate time.Time `json:"subscription_date"`
	Status           string    `json:"status"`
	BillingAmount    float64   `json:"billing_amount"`
	BillingCycle     string    `json:"billing_cycle"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}