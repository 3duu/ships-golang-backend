package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
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

	link := fmt.Sprintf("http://localhost:8080/api/verify-email?token=%s", token)
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

func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Check for existing user
	var existing models.User
	err := h.db.Collection("users").FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&existing)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	user := models.User{
		ID:        primitive.NewObjectID(),
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
	}

	verifyToken := generateRandomToken(32)
	user.VerifyToken = verifyToken
	user.EmailVerified = false

	_, err = db.Collection("users").InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	//token, _ := utils.GenerateJWT(user.ID.String()) // <- generate token as done in login

	//c.JSON(http.StatusOK, gin.H{"token": token, "userId": user.ID.Hex()})

	// Generate token
	token, err := utils.GenerateJWT(user.ID.Hex())
	if err != nil {
		http.Error(w, "Token generation failed", http.StatusInternalServerError)
		return
	}

	user.Password = ""

	err = json.NewEncoder(w).Encode(RegisterResponse{
		Token: token,
		User:  user,
	})
	if err != nil {
		log.Printf("Failed to encode response: %v", err)
		return
	} else {
		go sendVerificationEmail(user.Email, verifyToken)
		w.WriteHeader(http.StatusCreated)
	}

}
