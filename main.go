package main

import (
	"log"
	"net/http"
	"ships-backend/internal/middlewares"
	"ships-backend/internal/ws"

	"github.com/gorilla/mux"
	"ships-backend/internal/database"
	"ships-backend/internal/handlers"
)

func main() {
	database.InitMongoDB()
	db := database.MongoDB
	wsManager := ws.NewManager()
	handler := handlers.NewHandler(db, wsManager)
	database.EnsureIndexes(db)
	log.Println("🚀 Server is running on :8080")
	http.ListenAndServe(":8080", setupRoutes(handler))
}

func setupRoutes(h *handlers.Handler) *mux.Router {
	r := mux.NewRouter()
	auth := r.PathPrefix("/api").Subrouter()

	public := r.PathPrefix("/api").Subrouter()

	// Public routes
	public.HandleFunc("/login", handlers.LoginHandler(h.DB)).Methods("POST")
	public.HandleFunc("/register", handlers.RegisterHandler(h.DB)).Methods("POST")
	public.HandleFunc("/verify-email", handlers.VerifyEmailHandler(h.DB)).Methods("GET")

	// Protected subrouter
	auth.Use(middlewares.AuthMiddleware)

	r.Handle("/ws", middlewares.AuthMiddleware(handlers.WebSocketHandler(h.WSManager)))
	auth.HandleFunc("/profile", handlers.GetProfileHandler(h.DB)).Methods("GET")
	auth.HandleFunc("/profile", handlers.UpdateProfileHandler(h.DB)).Methods("PUT")
	auth.Handle("/like/{userId}", middlewares.AuthMiddleware(handlers.LikeUserHandler(h))).Methods("POST")
	auth.Handle("/nearby-users", middlewares.AuthMiddleware(h.NearbyUsersHandler())).Methods("GET")
	auth.Handle("/queue", middlewares.AuthMiddleware(h.SwipeQueueHandler())).Methods("GET")
	auth.Handle("/swipe/{userId}", middlewares.AuthMiddleware(h.SwipeHandler())).Methods("POST")

	return r
}
