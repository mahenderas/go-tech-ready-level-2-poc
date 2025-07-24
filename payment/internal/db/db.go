package db

import (
	"database/sql"
	"payment/internal/models"

	_ "github.com/lib/pq"
)

type DB struct {
	Conn *sql.DB
}

func NewDB(dbURL string) (*DB, error) {
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	return &DB{Conn: conn}, nil
}

// Add method to ensure Payments table exists
func (db *DB) EnsurePaymentsTable() error {
	_, err := db.Conn.Exec(`CREATE TABLE IF NOT EXISTS payments (
		transaction_id TEXT PRIMARY KEY,
		order_id TEXT,
		status TEXT,
		amount INT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}

func (db *DB) GetPayments() ([]models.Payment, error) {
	rows, err := db.Conn.Query("SELECT transaction_id, order_id, status, amount, created_at FROM payments ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var payments []models.Payment
	for rows.Next() {
		var p models.Payment
		if err := rows.Scan(&p.TransactionID, &p.OrderID, &p.Status, &p.Amount, &p.CreatedAt); err == nil {
			payments = append(payments, p)
		}
	}
	return payments, nil
}

func (db *DB) DeletePayments(ids []string) (int64, error) {
	query := "DELETE FROM payments WHERE transaction_id = ANY($1)"
	res, err := db.Conn.Exec(query, ids)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (db *DB) InsertPayment(p models.Payment) error {
	_, err := db.Conn.Exec(
		"INSERT INTO payments (transaction_id, order_id, status, amount, created_at) VALUES ($1, $2, $3, $4, $5)",
		p.TransactionID, p.OrderID, p.Status, p.Amount, p.CreatedAt,
	)
	return err
}
