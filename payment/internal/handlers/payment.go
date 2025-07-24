package handlers

import (
	"encoding/json"
	"net/http"
	"payment/internal/db"
)

type PaymentHandler struct {
	DB *db.DB
}

func (h *PaymentHandler) GetPayments(w http.ResponseWriter, r *http.Request) {
	payments, err := h.DB.GetPayments()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

func (h *PaymentHandler) DeletePayments(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IDs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid or missing IDs"))
		return
	}
	deleted, err := h.DB.DeletePayments(req.IDs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to delete payments"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"deleted": deleted})
}
