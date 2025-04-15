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
	"strings"
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

	link := fmt.Sprintf("http://localhost:8080/api/verify-email?token=%s", token)
	body := "Click to verify your email: " + link

	msg := "From: " + from + "\n" +
		"To: " + toEmail + "\n" +
		"Subject: Verify your email\n\n" + body

	smtp.SendMail(host+":"+port,
		smtp.PlainAuth("", from, pass, host),
		from, []string{toEmail}, []byte(msg))
}

func RegisterHandler(db *mongo.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		req.Email = strings.ToLower(strings.TrimSpace(req.Email))
		if req.Email == "" || req.Password == "" || req.Name == "" || req.Birth == (time.Time{}) {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		users := db.Collection("users")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Check for existing email
		var existing models.User
		err := users.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existing)
		if err == nil {
			http.Error(w, "Email already registered", http.StatusConflict)
			return
		}

		// Hashed password
		hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

		newUser := models.User{
			ID:        primitive.NewObjectID(),
			Name:      req.Name,
			Email:     req.Email,
			Password:  string(hashedPwd),
			Bio:       req.Bio,
			Gender:    req.Gender,
			Interests: req.Interests,
			Location:  req.Location,
			Birth:     req.Birth,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		verifyToken := generateRandomToken(32)
		newUser.VerifyToken = verifyToken
		newUser.EmailVerified = false

		_, err = users.InsertOne(ctx, newUser)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		// Generate token
		token, err := utils.GenerateJWT(newUser.ID.Hex())
		if err != nil {
			http.Error(w, "Token generation failed", http.StatusInternalServerError)
			return
		}

		newUser.Password = ""

		err = json.NewEncoder(w).Encode(RegisterResponse{
			Token: token,
			User:  newUser,
		})
		if err != nil {
			log.Printf("Failed to encode response: %v", err)
			return
		} else {
			go sendVerificationEmail(newUser.Email, verifyToken)
			w.WriteHeader(http.StatusCreated)
		}
	}
}
