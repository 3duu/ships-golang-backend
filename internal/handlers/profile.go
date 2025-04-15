package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"ships-backend/internal/middlewares"
	"ships-backend/internal/models"
)

func GetProfileHandler(db *mongo.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middlewares.UserIDKey).(string)
		objID, _ := primitive.ObjectIDFromHex(userID)

		var user models.User
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := db.Collection("users").FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		user.Password = ""
		json.NewEncoder(w).Encode(user)
	}
}

func UpdateProfileHandler(db *mongo.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middlewares.UserIDKey).(string)
		objID, _ := primitive.ObjectIDFromHex(userID)

		var update models.ProfileUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		// Email is not updated here on purpose
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := db.Collection("users").UpdateByID(ctx, objID, bson.M{
			"$set": bson.M{
				"name":      update.Name,
				"bio":       update.Bio,
				"gender":    update.Gender,
				"interests": update.Interests,
				"location":  update.Location,
				"updatedAt": time.Now(),
			},
		})

		if err != nil {
			http.Error(w, "Failed to update profile", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully."})
	}
}
