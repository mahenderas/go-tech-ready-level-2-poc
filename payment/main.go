package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"payment/internal/consul"
	"payment/internal/db"
	"payment/internal/handlers"
	"payment/internal/pubsub"

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
		dbURL = "postgres://postgres:postgres@localhost:5432/payments?sslmode=disable"
	}
	sqlDB, err := db.NewDB(dbURL)
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}
	log.Println("Connected to PostgreSQL database.")

	if err := sqlDB.EnsurePaymentsTable(); err != nil {
		log.Fatalf("Failed to create payments table: %v", err)
	}
	handler := handlers.PaymentHandler{DB: sqlDB}

	// Consul registration
	consul.RegisterWithConsul("payment", 8003)

	// Pub/Sub setup
	projectID := os.Getenv("PUBSUB_PROJECT_ID")
	if projectID == "" {
		projectID = "test-project"
	}
	ctx := context.Background()
	ps, err := pubsub.SetupPubSub(ctx, projectID, sqlDB)
	if err != nil {
		log.Fatalf("Failed to setup Pub/Sub: %v", err)
	}
	go ps.ListenForOrderEvents(ctx)

	// HTTP handlers
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	http.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetPayments(w, r)
		case http.MethodDelete:
			handler.DeletePayments(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
		if os.Getenv("DEPLOY_ENV") == "gcp" {
			port = "8080"
		}
	}
	log.Printf("Payment service HTTP server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
