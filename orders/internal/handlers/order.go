package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"orders/internal/db"
	"orders/internal/models"

	"github.com/google/uuid"
)

type OrderHandler struct {
	DB *db.DB
}

func (h *OrderHandler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.DB.GetAllOrders()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// CreateOrder handles POST /orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) (*models.Order, error) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil, nil
	}
	var req struct {
		Products []models.OrderProduct `json:"products"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}
	amount := 0
	for _, p := range req.Products {
		amount += p.Price
	}
	order := models.Order{
		ID:       generateOrderID(),
		Status:   "created",
		Products: req.Products,
		Amount:   amount,
	}
	if err := h.DB.CreateOrder(order); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
	return &order, nil
}

// DeleteOrder handles DELETE /orders/delete with array of ids
func (h *OrderHandler) DeleteOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IDs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid or missing IDs"))
		return
	}
	rowsAffected, err := h.DB.DeleteOrders(req.IDs)
	if err != nil {
		log.Printf("DeleteOrders error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Failed to delete orders: %v", err)))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"deleted": rowsAffected})
}

// generateOrderID returns a new UUID string
func generateOrderID() string {
	return uuid.NewString()
}
