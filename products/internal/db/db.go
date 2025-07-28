package db

import (
	"database/sql"
	"products/internal/models"

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

func (db *DB) CreateProduct(product models.Product) error {
	_, err := db.Conn.Exec("INSERT INTO products (id, name, price) VALUES ($1, $2, $3)", product.ID, product.Name, product.Price)
	return err
}

func (db *DB) DeleteProducts(ids []string) (int64, error) {
	query := "DELETE FROM products WHERE id = ANY($1)"
	res, err := db.Conn.Exec(query, pq.Array(ids))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (db *DB) GetAllProducts() ([]models.Product, error) {
	rows, err := db.Conn.Query("SELECT id, name, price FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err == nil {
			products = append(products, p)
		}
	}
	return products, nil
}

func (db *DB) GetProductByID(id string) (*models.Product, error) {
	row := db.Conn.QueryRow("SELECT id, name, price FROM products WHERE id = $1", id)
	var p models.Product
	if err := row.Scan(&p.ID, &p.Name, &p.Price); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Product not found
		}
		return nil, err
	}
	return &p, nil
}

func (db *DB) EnsureProductsTable() error {
	_, err := db.Conn.Exec(`
	CREATE TABLE IF NOT EXISTS products (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		price INT NOT NULL
	)`)
	return err
}
