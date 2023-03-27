package main

import (
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
  
  type Subscription struct {
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
  
  type Transaction struct {
	  ID              int       `json:"id"`
	  SubscriptionID  int       `json:"subscription_id"`
	  TransactionDate time.Time `json:"transaction_date"`
	  Amount          float64   `json:"amount"`
	  Status          string    `json:"status"`
	  CreatedAt       time.Time `json:"created_at"`
	  UpdatedAt       time.Time `json:"updated_at"`
  }
 