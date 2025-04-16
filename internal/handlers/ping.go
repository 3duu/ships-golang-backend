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

type PingLocationRequest struct {
	Coordinates []float64 `json:"coordinates"` // [lng, lat]
}

func (h *Handler) PingLocationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middlewares.UserIDKey).(string)
		objID, _ := primitive.ObjectIDFromHex(userID)

		var req PingLocationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if len(req.Coordinates) != 2 {
			http.Error(w, "Invalid coordinates", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		now := time.Now()

		// Update user’s location and updatedAt
		_, err := h.DB.Collection("users").UpdateByID(ctx, objID, bson.M{
			"$set": bson.M{
				"location": bson.M{
					"type":        "Point",
					"coordinates": req.Coordinates,
				},
				"updatedAt": now,
			},
		})
		if err != nil {
			http.Error(w, "Failed to update location", http.StatusInternalServerError)
			return
		}

		// Find nearby users who were recently active
		nearbyCursor, err := h.DB.Collection("users").Find(ctx, bson.M{
			"_id": bson.M{"$ne": objID},
			"location": bson.M{
				"$near": bson.M{
					"$geometry": bson.M{
						"type":        "Point",
						"coordinates": req.Coordinates,
					},
					"$maxDistance": 100, // 100 meters
				},
			},
			"updatedAt": bson.M{
				"$gte": now.Add(-10 * time.Minute), // seen recently
			},
		})
		if err != nil {
			http.Error(w, "Nearby scan failed", http.StatusInternalServerError)
			return
		}

		var others []models.User
		if err := nearbyCursor.All(ctx, &others); err != nil {
			http.Error(w, "Failed to decode users", http.StatusInternalServerError)
			return
		}

		crossedCol := h.DB.Collection("crossed_paths")
		for _, other := range others {
			id1, id2 := objID, other.ID
			if id1.Hex() > id2.Hex() {
				id1, id2 = id2, id1
			}

			// Check if they already crossed recently
			exists := crossedCol.FindOne(ctx, bson.M{
				"user1":     id1,
				"user2":     id2,
				"timestamp": bson.M{"$gte": now.Add(-1 * time.Hour)},
			})
			if exists.Err() == mongo.ErrNoDocuments {
				crossed := models.CrossedPath{
					User1:     id1,
					User2:     id2,
					Timestamp: now,
					Location: models.Location{
						Type:        "Point",
						Coordinates: req.Coordinates,
					},
				}
				_, _ = crossedCol.InsertOne(ctx, crossed)
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Location updated and paths checked",
		})
	}
}
