package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"ships-backend/internal/models"
	"ships-backend/internal/utils"
)

// Request payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Response payload
type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// LoginHandler handles POST /api/login
func LoginHandler(db *mongo.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		var user models.User
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		users := db.Collection("users")
		err := users.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized,
				"Invalid email or password",
				err.Error(),
			)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			//http.Error(w, "Invalid email or password", http.StatusUnauthorized)

			utils.RespondWithError(w, http.StatusUnauthorized,
				"Invalid email or password",
				err.Error(),
			)

			return
		}

		token, err := utils.GenerateJWT(user.ID.Hex())
		if err != nil {
			log.Println(err)
			//http.Error(w, "Failed to generate token", http.StatusInternalServerError)

			utils.RespondWithError(w, http.StatusInternalServerError,
				"Error generating token",
				err.Error(),
			)

			return
		}

		user.Password = "" // Hide password

		response := LoginResponse{
			Token: token,
			User:  user,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
