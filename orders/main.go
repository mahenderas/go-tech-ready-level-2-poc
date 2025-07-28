package main

import (
	"context"
	"log"
	"net/http"
	"orders/internal/consul"
	"orders/internal/db"
	"orders/internal/handlers"
	"orders/internal/pubsub"
	"orders/internal/middleware"
	"os"

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
		dbURL = "postgres://postgres:postgres@localhost:5432/orders?sslmode=disable"
	}
	sqlDB, err := db.NewDB(dbURL)
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}
	if err := sqlDB.EnsureOrdersTable(); err != nil {
		log.Fatalf("Failed to create orders table: %v", err)
	}
	log.Println("Connected to PostgreSQL database.")

	handler := handlers.OrderHandler{DB: sqlDB}

	// Consul registration
	consul.RegisterWithConsul("orders", 8002)

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

	go ps.ListenForPaymentEvents(ctx)

	// HTTP handlers
	http.Handle("/orders", middleware.JwtTokenValidation(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetAllOrders(w, r)
		case http.MethodPost:
			if order, err := handler.CreateOrder(w, r); err == nil {
				go pubsub.PublishOrderEvent(ctx, ps, *order)
			}
		case http.MethodDelete:
			handler.DeleteOrders(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	// Add Handler for individual order /orders/{id}
	http.Handle("/orders/{id}", middleware.JwtTokenValidation(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetOrderByID(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
		if os.Getenv("DEPLOY_ENV") == "gcp" {
			port = "8080"
		}
	}
	log.Printf("Orders service running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
