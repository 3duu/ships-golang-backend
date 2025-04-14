package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"ships-backend/internal/database"
	"ships-backend/internal/handlers"
)

func main() {
	database.InitMongoDB()
	db := database.MongoDB

	r := mux.NewRouter()
	r.HandleFunc("/api/login", handlers.LoginHandler(db)).Methods("POST")

	log.Println("🚀 Server is running on :8080")
	http.ListenAndServe(":8080", r)
}
