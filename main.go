package main

import (
	"fmt"
	"log"
	"net/http"
	"ships-backend/internal/middlewares"
	"ships-backend/internal/utils"
	"ships-backend/internal/ws"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
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
	setupRoutes(handler)
}

func setupRoutes(h *handlers.Handler) *mux.Router {
	r := mux.NewRouter()

	// CORS wrapper
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:19006", "http://localhost:8081"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}).Handler(r)

	// 🔓 Public routes
	public := r.PathPrefix("/api/public").Subrouter()
	public.HandleFunc("/auth/login", handlers.LoginHandler(h.DB)).Methods("POST")
	public.HandleFunc("/auth/register", handlers.NewAuthHandler(h.DB).RegisterHandler()).Methods("POST")
	public.HandleFunc("/verify-email", handlers.VerifyEmailHandler(h.DB)).Methods("GET")

	// 🔐 Authenticated routes
	auth := r.PathPrefix("/api/auth").Subrouter()
	auth.Use(middlewares.AuthMiddleware)
	auth.HandleFunc("/profile", handlers.GetProfileHandler(h.DB)).Methods("GET")
	auth.HandleFunc("/profile", handlers.UpdateProfileHandler(h.DB)).Methods("PUT")
	auth.HandleFunc("/logout", handlers.LogoutHandler(h.DB)).Methods("POST")

	// Other protected routes...
	auth.Handle("/nearby-users", h.NearbyUsersHandler()).Methods("GET")
	auth.Handle("/queue", h.SwipeQueueHandler()).Methods("GET")
	auth.Handle("/swipe/{userId}", h.SwipeHandler()).Methods("POST")

	// WebSocket routes
	ws := r.PathPrefix("/ws").Subrouter()
	ws.Use(middlewares.AuthMiddleware)
	ws.Handle("/", handlers.WebSocketHandler(h.WSManager)).Methods("GET")
	ws.Handle("/chat", h.WebSocketChatHandler()).Methods("GET")

	localIP := utils.GetLocalIP()
	port := ":8080"

	fmt.Printf("🚀 Server running at http://%s%s\n", localIP, port)

	http.ListenAndServe("0.0.0.0"+port, handler)
	/*err := http.ListenAndServe("0.0.0.0"+port, handler)
	if err != nil
		log.Println(err.Error())*/

	return r
}
