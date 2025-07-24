// Main entry for Products service
package main

import (
	"log"
	"net/http"
	"os"
	"products/internal/consul"
	"products/internal/db"
	"products/internal/handlers"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment
	env := os.Getenv("DEPLOY_ENV")
	if env == "gcp" {
		err := godotenv.Load(".env.gcp")
		if err != nil {
			log.Printf("Error loading .env.gcp: %v", err)
		}
	} else {
		err := godotenv.Load(".env.local")
		if err != nil {
			log.Printf("Error loading .env.local: %v", err)
		}
	}

	// DB setup
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/products?sslmode=disable"
	}
	sqlDB, err := db.NewDB(dbURL)
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}
	if err := sqlDB.EnsureProductsTable(); err != nil {
		log.Fatalf("Failed to create products table: %v", err)
	}
	log.Println("Connected to PostgreSQL database.")

	handler := handlers.ProductHandler{DB: sqlDB}

	// Consul registration
	consul.RegisterWithConsul("products", 8001)

	// HTTP handlers
	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetAllProducts(w, r)
		case http.MethodPost:
			handler.CreateProduct(w, r)
		case http.MethodDelete:
			handler.DeleteProducts(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
		if os.Getenv("DEPLOY_ENV") == "gcp" {
			port = "8080"
		}
	}
	log.Printf("Products service running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
