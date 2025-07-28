package db

import (
	"authentication/internal/models"
	"database/sql"
	"time"

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

// EnsureUsersTable creates the users table if it doesn't exist
func (db *DB) EnsureUsersTable() error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	)`
	_, err := db.Conn.Exec(query)
	return err
}

// CreateUser inserts a new user into the database
func (db *DB) CreateUser(u *models.User) error {
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	_, err := db.Conn.Exec("INSERT INTO users (username, password, email, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)", u.Username, u.Password, u.Email, u.CreatedAt, u.UpdatedAt)
	return err
}

// GetUserByUsername fetches a user by username
func (db *DB) GetUserByUsername(username string) (*models.User, error) {
	row := db.Conn.QueryRow("SELECT id, username, password, email, created_at, updated_at FROM users WHERE username=$1", username)
	var u models.User
	err := row.Scan(&u.ID, &u.Username, &u.Password, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateUserPassword updates a user's password and updated_at
func (db *DB) UpdateUserPassword(username, newPassword string) error {
	now := time.Now()
	_, err := db.Conn.Exec("UPDATE users SET password=$1, updated_at=$2 WHERE username=$3", newPassword, now, username)
	return err
}
