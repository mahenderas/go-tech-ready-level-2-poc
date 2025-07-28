package db

import (
	"database/sql"
	"encoding/json"
	"orders/internal/models"

	"github.com/lib/pq"
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

func (db *DB) CreateOrder(order models.Order) error {
	productsJSON, err := json.Marshal(order.Products)
	if err != nil {
		return err
	}
	_, err = db.Conn.Exec("INSERT INTO orders (id, status, amount, products) VALUES ($1, $2, $3, $4)", order.ID, order.Status, order.Amount, productsJSON)
	return err
}

func (db *DB) DeleteOrders(ids []string) (int64, error) {
	query := "DELETE FROM orders WHERE id = ANY($1)"
	res, err := db.Conn.Exec(query, pq.Array(ids))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (db *DB) GetAllOrders() ([]models.Order, error) {
	rows, err := db.Conn.Query("SELECT id, status, amount, products FROM orders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dbOrders []models.Order
	for rows.Next() {
		var o models.Order
		var productsJSON []byte
		if err := rows.Scan(&o.ID, &o.Status, &o.Amount, &productsJSON); err == nil {
			// Unmarshal products
			_ = json.Unmarshal(productsJSON, &o.Products)
			dbOrders = append(dbOrders, o)
		}
	}
	return dbOrders, nil
}

// GetOrderByID retrieves an order by its ID
func (db *DB) GetOrderByID(id string) (*models.Order, error) {
	var o models.Order
	var productsJSON []byte
	err := db.Conn.QueryRow("SELECT id, status, amount, products FROM orders WHERE id = $1", id).Scan(&o.ID, &o.Status, &o.Amount, &productsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Order not found
		}
		return nil, err
	}
	// Unmarshal products
	if err := json.Unmarshal(productsJSON, &o.Products); err != nil {
		return nil, err
	}
	return &o, nil
}

func (db *DB) UpdateOrderStatus(orderID, status string) error {
	_, err := db.Conn.Exec("UPDATE orders SET status = $1 WHERE id = $2", status, orderID)
	return err
}

func (db *DB) EnsureOrdersTable() error {
	_, err := db.Conn.Exec(`
	CREATE TABLE IF NOT EXISTS orders (
		id TEXT PRIMARY KEY,
		status TEXT NOT NULL,
		amount INT NOT NULL,
		products JSONB NOT NULL
	)`)
	return err
}
