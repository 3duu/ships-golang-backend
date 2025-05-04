package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"ships-backend/internal/models"
	"ships-backend/internal/utils"
)

type RegisterResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func generateRandomToken(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func sendVerificationEmail(toEmail, token string) {
	from := "your@email.com"
	pass := "your-password"
	host := "smtp.yourprovider.com"
	port := "587"

	localIP := utils.GetLocalIP()
	link := fmt.Sprintf("http://"+localIP+":8080/api/verify-email?token=%s", token)
	body := "Click to verify your email: " + link

	msg := "From: " + from + "\n" +
		"To: " + toEmail + "\n" +
		"Subject: Verify your email\n\n" + body

	smtp.SendMail(host+":"+port,
		smtp.PlainAuth("", from, pass, host),
		from, []string{toEmail}, []byte(msg))
}

type AuthHandler struct {
	db *mongo.Database
}

func NewAuthHandler(db *mongo.Database) *AuthHandler {
	return &AuthHandler{db: db}
}

func (h *AuthHandler) RegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		// Decode JSON manually (no c.BindJSON)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			//http.Error(w, "Invalid request", http.StatusBadRequest)

			utils.RespondWithError(w, http.StatusBadRequest,
				"Error decoding request body",
				err.Error(),
			)

			return
		}

		// Check if user already exists
		var existing models.User
		err := h.db.Collection("users").FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&existing)
		if err == nil {
			http.Error(w, "Email already registered", http.StatusBadRequest)
			return
		}

		// Hash password
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 10)

		// Create user object
		user := models.User{
			ID:            primitive.NewObjectID(),
			Name:          req.Name,
			Email:         req.Email,
			Password:      string(hashedPassword),
			CreatedAt:     time.Now(),
			EmailVerified: false,
			VerifyToken:   generateRandomToken(32),
			Location: models.Location{
				Type:        "Point",
				Coordinates: []float64{0.0, 0.0}, // default empty location
			},
		}

		// Insert into MongoDB
		_, err = h.db.Collection("users").InsertOne(context.Background(), user)
		if err != nil {

			utils.RespondWithError(w, http.StatusInternalServerError,
				"Could not create user.",
				err.Error(),
			)

			return
		}

		// Generate JWT
		token, err := utils.GenerateJWT(user.ID.Hex())
		if err != nil {

			utils.RespondWithError(w, http.StatusInternalServerError,
				"Could not create user.",
				"Token generation failed - "+err.Error(),
			)
			return
		}

		user.Password = "" // Never return password

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		err = json.NewEncoder(w).Encode(RegisterResponse{
			Token: token,
			User:  user,
		})
		if err != nil {
			log.Printf("Failed to encode response: %v", err)
		}

		go sendVerificationEmail(user.Email, user.VerifyToken)
	}
}
