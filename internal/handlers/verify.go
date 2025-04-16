package handlers

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"ships-backend/internal/models"
	"time"
)

/*Use a library like gomail for HTML emails

Automatically login user after verification

Expire verification tokens after X hours*/

func VerifyEmailHandler(db *mongo.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Missing token", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user models.User
		err := db.Collection("users").FindOne(ctx, bson.M{"verifyToken": token}).Decode(&user)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusNotFound)
			return
		}

		_, err = db.Collection("users").UpdateByID(ctx, user.ID, bson.M{
			"$set":   bson.M{"emailVerified": true},
			"$unset": bson.M{"verifyToken": ""},
		})
		if err != nil {
			http.Error(w, "Failed to verify email", http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, "✅ Email verified successfully!")
	}
}
