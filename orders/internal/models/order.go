package models

type OrderProduct struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type Order struct {
	ID       string         `json:"id"`
	Status   string         `json:"status"`
	Products []OrderProduct `json:"products"`
	Amount   int            `json:"amount"`
}
