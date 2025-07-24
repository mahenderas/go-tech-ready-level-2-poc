package models

import "time"

type Payment struct {
	TransactionID string    `json:"transaction_id"`
	OrderID       string    `json:"order_id"`
	Status        string    `json:"status"`
	Amount        int       `json:"amount"`
	CreatedAt     time.Time `json:"created_at"`
}
