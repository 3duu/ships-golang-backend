package main

import (
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"ships-backend/internal/middlewares"

	"github.com/gorilla/mux"
	"ships-backend/internal/database"
	"ships-backend/internal/handlers"
)

func main() {
	database.InitMongoDB()
	db := database.MongoDB
	log.Println("🚀 Server is running on :8080")
	http.ListenAndServe(":8080", setupRoutes(db))
}

func setupRoutes(db *mongo.Database) *mux.Router {
	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/api/login", handlers.LoginHandler(db)).Methods("POST")
	r.HandleFunc("/api/register", handlers.RegisterHandler(db)).Methods("POST")

	// Protected subrouter
	auth := r.PathPrefix("/api").Subrouter()
	auth.Use(middlewares.AuthMiddleware)

	auth.HandleFunc("/profile", handlers.GetProfileHandler(db)).Methods("GET")
	auth.HandleFunc("/profile", handlers.UpdateProfileHandler(db)).Methods("PUT")

	return r
}
