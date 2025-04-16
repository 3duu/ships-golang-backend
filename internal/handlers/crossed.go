package handlers

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"ships-backend/internal/middlewares"
	"ships-backend/internal/models"
)

func (h *Handler) GetCrossedPathsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middlewares.UserIDKey).(string)
		objID, _ := primitive.ObjectIDFromHex(userID)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 🧠 Parse query params
		since := r.URL.Query().Get("since")
		limitStr := r.URL.Query().Get("limit")

		duration := 7 * 24 * time.Hour // default 7 days
		if since != "" {
			d, err := time.ParseDuration(since)
			if err == nil {
				duration = d
			}
		}

		limit := int64(50)
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = int64(l)
			}
		}

		// Filter by user and time
		filter := bson.M{
			"$and": []bson.M{
				{
					"$or": []bson.M{
						{"user1": objID},
						{"user2": objID},
					},
				},
				{
					"timestamp": bson.M{"$gte": time.Now().Add(-duration)},
				},
			},
		}

		opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "timestamp", Value: -1}})

		cursor, err := h.DB.Collection("crossed_paths").Find(ctx, filter, opts)
		if err != nil {
			http.Error(w, "Error loading crossed paths", http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)

		var crossed []models.CrossedPath
		if err := cursor.All(ctx, &crossed); err != nil {
			http.Error(w, "Decode error", http.StatusInternalServerError)
			return
		}

		userCol := h.DB.Collection("users")
		for i := range crossed {
			var otherID primitive.ObjectID
			if crossed[i].User1 == objID {
				otherID = crossed[i].User2
			} else {
				otherID = crossed[i].User1
			}

			var other models.User
			err := userCol.FindOne(ctx, bson.M{"_id": otherID}).Decode(&other)
			if err == nil {
				other.Password = ""
				crossed[i].OtherUser = &other
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(crossed)
	}
}
