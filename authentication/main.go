package main

import (
	"authentication/internal/consul"
	"authentication/internal/db"
	"authentication/internal/handlers"
	"authentication/internal/middleware"
	"log"
	"net/http"
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
		dbURL = "postgres://postgres:postgres@localhost:5432/authdb?sslmode=disable"
	}
	sqlDB, err := db.NewDB(dbURL)
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}
	if err := sqlDB.EnsureUsersTable(); err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}
	log.Println("Connected to PostgreSQL database.")

	// Consul registration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8004"
		if os.Getenv("DEPLOY_ENV") == "gcp" {
			port = "8080"
		}
	}
	consul.RegisterWithConsul("authentication", 8004)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authentication service is running"))
	})
	authHandler := &handlers.AuthHandler{DB: sqlDB}

	// Public endpoints
	mux.HandleFunc("/register", authHandler.RegisterHandler)
	mux.HandleFunc("/login", authHandler.LoginHandler)
	mux.HandleFunc("/reset-password", authHandler.ResetPasswordHandler)

	// Protected endpoint (JWT required)
	mux.Handle("/update-password", middleware.JwtTokenValidation(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("username").(string)
		authHandler.UpdatePasswordHandler(w, r, username)
	})))

	log.Printf("Authentication service running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
