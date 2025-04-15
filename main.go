package main

import (
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

	r := mux.NewRouter()
	r.HandleFunc("/api/login", handlers.LoginHandler(db)).Methods("POST")
	r.Handle("/api/profile", middlewares.AuthMiddleware(handlers.GetProfileHandler(db))).Methods("GET")
	r.Handle("/api/profile", middlewares.AuthMiddleware(handlers.UpdateProfileHandler(db))).Methods("PUT")

	log.Println("🚀 Server is running on :8080")
	http.ListenAndServe(":8080", r)
}
