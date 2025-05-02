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
	Token        string      `json:"authToken"`
	User         models.User `json:"user"`
	RefreshToken string      `json:"refreshToken"`
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

		refreshToken := utils.GenerateRandomToken(64)
		// Store refreshToken in DB (optional but recommended)
		_, err = db.Collection("refresh_tokens").InsertOne(context.TODO(), bson.M{
			"userId":    user.ID,
			"token":     refreshToken,
			"createdAt": time.Now(),
		})

		user.Password = "" // Hide password

		response := LoginResponse{
			Token:        token,
			User:         user,
			RefreshToken: refreshToken,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func LogoutHandler(db *mongo.Database) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID") // from your auth middleware
		if userID == nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "missing user ID in context")
			return
		}

		// Optionally read refresh token from header/body if needed
		// For now, we’ll remove all user refresh tokens
		_, err := db.Collection("refresh_tokens").DeleteMany(context.TODO(), bson.M{"userId": userID})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError,
				"Could not logout.",
				err.Error(),
			)
			return
		}

		w.WriteHeader(http.StatusNoContent) // 204 No Content
	}
}
