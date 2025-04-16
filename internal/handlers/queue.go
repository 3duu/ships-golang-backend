package handlers

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"ships-backend/internal/middlewares"
	"ships-backend/internal/models"
	"strconv"
	"time"
)

func (h *Handler) SwipeQueueHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		userID := r.Context().Value(middlewares.UserIDKey).(string)
		currentUserID, _ := primitive.ObjectIDFromHex(userID)

		lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
		lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
		maxDistanceKm, _ := strconv.ParseFloat(r.URL.Query().Get("maxDistanceKm"), 64)
		if maxDistanceKm == 0 {
			maxDistanceKm = 5
		}

		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		skip, _ := strconv.Atoi(r.URL.Query().Get("skip"))
		if limit == 0 {
			limit = 10
		}

		// 🧠 Get seen users
		var seen []models.Seen
		cursor, _ := h.DB.Collection("seen").Find(ctx, bson.M{"userId": currentUserID})
		_ = cursor.All(ctx, &seen)

		seenIDs := map[string]bool{}
		for _, s := range seen {
			seenIDs[s.SeenUser.Hex()] = true
		}

		var exclude []primitive.ObjectID
		for idStr := range seenIDs {
			oid, err := primitive.ObjectIDFromHex(idStr)
			if err == nil {
				exclude = append(exclude, oid)
			}
		}

		idFilter := bson.M{
			"$ne": currentUserID,
		}

		if len(exclude) > 0 {
			idFilter["$nin"] = exclude
		}

		filter := bson.M{
			"_id": idFilter,
			"location": bson.M{
				"$near": bson.M{
					"$geometry": bson.M{
						"type":        "Point",
						"coordinates": []float64{lng, lat},
					},
					"$maxDistance": maxDistanceKm * 1000,
				},
			},
		}

		opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(skip))
		result, err := h.DB.Collection("users").Find(ctx, filter, opts)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Query failed", http.StatusInternalServerError)
			return
		}

		var users []models.User
		if err := result.All(ctx, &users); err != nil {
			http.Error(w, "Failed to load users", http.StatusInternalServerError)
			return
		}

		// Auto-track as "seen"
		seenCol := h.DB.Collection("seen")
		for _, u := range users {
			seenCol.UpdateOne(ctx,
				bson.M{"userId": currentUserID, "seenUser": u.ID},
				bson.M{
					"$setOnInsert": bson.M{
						"timestamp": time.Now(),
					},
				},
				options.Update().SetUpsert(true),
			)
			u.Password = "" // sanitize
		}

		// Preload metadata
		count, _ := h.DB.Collection("users").CountDocuments(ctx, filter)
		hasMore := int64(skip+limit) < count

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"users":   users,
			"next":    skip + limit,
			"hasMore": hasMore,
		})
	}
}
