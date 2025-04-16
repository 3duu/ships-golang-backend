package handlers

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"ships-backend/internal/middlewares"
	"ships-backend/internal/models"
	"strconv"
	"strings"
	"time"
)

func (h *Handler) NearbyUsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		userID := r.Context().Value(middlewares.UserIDKey).(string)
		currentUserID, _ := primitive.ObjectIDFromHex(userID)

		// Parse query parameters
		lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
		lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
		maxDistanceKm, _ := strconv.ParseFloat(r.URL.Query().Get("maxDistanceKm"), 64)
		if maxDistanceKm == 0 {
			maxDistanceKm = 5
		}

		gender := r.URL.Query().Get("gender")
		interestsQuery := r.URL.Query().Get("interests")
		var interests []string
		if interestsQuery != "" {
			interests = strings.Split(interestsQuery, ",")
		}

		// 🧮 Paging
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		skip, _ := strconv.Atoi(r.URL.Query().Get("skip"))
		if limit == 0 {
			limit = 10
		}

		// 🔍 Get already liked or seen users
		var liked []models.Like
		likeCursor, _ := h.DB.Collection("likes").Find(ctx, bson.M{"fromUser": currentUserID})
		_ = likeCursor.All(ctx, &liked)

		seenIDs := map[string]bool{}
		for _, like := range liked {
			seenIDs[like.ToUser.Hex()] = true
		}

		// Build `$nin` array
		var seenObjectIDs []primitive.ObjectID
		for idStr := range seenIDs {
			oid, err := primitive.ObjectIDFromHex(idStr)
			if err == nil {
				seenObjectIDs = append(seenObjectIDs, oid)
			}
		}

		// 🔍 Build query
		filter := bson.M{
			"_id": bson.M{
				"$ne":  currentUserID,
				"$nin": seenObjectIDs,
			},
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

		if gender != "" {
			filter["gender"] = gender
		}

		if len(interests) > 0 {
			filter["interests"] = bson.M{"$in": interests}
		}

		cursor, err := h.DB.Collection("users").Find(ctx, filter, &options.FindOptions{
			Limit: int64Ptr(int64(limit)),
			Skip:  int64Ptr(int64(skip)),
		})
		if err != nil {
			http.Error(w, "Error querying nearby users", http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)

		var users []models.User
		if err := cursor.All(ctx, &users); err != nil {
			http.Error(w, "Error decoding users", http.StatusInternalServerError)
			return
		}

		for i := range users {
			users[i].Password = ""
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}
