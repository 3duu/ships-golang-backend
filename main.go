package main

import (
	"log"
	"net/http"
	"ships-backend/internal/middlewares"
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

	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:19006", "http://localhost:8081"}, // ← your frontend's origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}).Handler(r)

	auth := r.PathPrefix("/api").Subrouter()

	public := r.PathPrefix("/api").Subrouter()

	// Public routes
	public.HandleFunc("/auth/login", handlers.LoginHandler(h.DB)).Methods("POST")
	//public.HandleFunc("/register", handlers.RegisterHandler(h.DB)).Methods("POST")
	public.HandleFunc("/verify-email", handlers.VerifyEmailHandler(h.DB)).Methods("GET")
	public.HandleFunc("/auth/register", handlers.RegisterHandler(h.DB)).Methods("POST")

	// Protected subrouter
	auth.Use(middlewares.AuthMiddleware)

	wsAuth := r.PathPrefix("/ws").Subrouter()
	wsAuth.Use(middlewares.AuthMiddleware)
	wsAuth.Handle("/", middlewares.AuthMiddleware(handlers.WebSocketHandler(h.WSManager)))
	wsAuth.Handle("/chat", middlewares.AuthMiddleware(h.WebSocketChatHandler())).Methods("GET")

	auth.HandleFunc("/profile", handlers.GetProfileHandler(h.DB)).Methods("GET")
	auth.HandleFunc("/profile", handlers.UpdateProfileHandler(h.DB)).Methods("PUT")
	//auth.Handle("/like/{userId}", middlewares.AuthMiddleware(handlers.LikeUserHandler(h))).Methods("POST")
	auth.Handle("/nearby-users", middlewares.AuthMiddleware(h.NearbyUsersHandler())).Methods("GET")
	auth.Handle("/queue", middlewares.AuthMiddleware(h.SwipeQueueHandler())).Methods("GET")
	auth.Handle("/swipe/{userId}", middlewares.AuthMiddleware(h.SwipeHandler())).Methods("POST")
	auth.Handle("/got-liked", middlewares.AuthMiddleware(h.GetYouGotLikedHandler())).Methods("GET")
	auth.Handle("/ping-location", middlewares.AuthMiddleware(h.PingLocationHandler())).Methods("POST")
	auth.Handle("/crossed-paths", middlewares.AuthMiddleware(h.GetCrossedPathsHandler())).Methods("GET")
	auth.Handle("/upload-photo", middlewares.AuthMiddleware(h.UploadPhotoHandler())).Methods("POST")
	auth.HandleFunc("/photo/{userId}", h.GetUserPhotoHandler()).Methods("GET")
	auth.Handle("/photo-order", middlewares.AuthMiddleware(h.UpdatePhotoOrderHandler())).Methods("PUT")
	auth.Handle("/photo/{photoId}", middlewares.AuthMiddleware(h.DeletePhotoHandler())).Methods("DELETE")
	auth.Handle("/messages/{matchId}", middlewares.AuthMiddleware(h.SendMessageHandler())).Methods("POST")
	auth.Handle("/messages/{matchId}", middlewares.AuthMiddleware(h.GetMessagesHandler())).Methods("GET")

	http.ListenAndServe(":8080", handler)

	return r
}
