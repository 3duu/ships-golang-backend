package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"ships-backend/internal/middlewares"
	"ships-backend/internal/models"
)

type SwipeRequest struct {
	Action string `json:"action"`
	Source string `json:"source"`
}

func (h *Handler) SwipeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		fromUserID := r.Context().Value(middlewares.UserIDKey).(string)
		toUserID := mux.Vars(r)["userId"]

		fromID, err := primitive.ObjectIDFromHex(fromUserID)
		toID, err2 := primitive.ObjectIDFromHex(toUserID)
		if err != nil || err2 != nil || fromID == toID {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		var req SwipeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		action := models.SwipeAction(req.Action)
		if action != models.LikeSwipe && action != models.DislikeSwipe && action != models.SuperLikeSwipe {
			http.Error(w, "Invalid swipe action", http.StatusBadRequest)
			return
		}

		swipe := models.Swipe{
			FromUser:   fromID,
			ToUser:     toID,
			Action:     action,
			Source:     req.Source,
			CreatedAt:  time.Now(),
			ValidUntil: time.Now().Add(24 * time.Hour),
		}

		swipes := h.DB.Collection("swipes")
		_, err = swipes.InsertOne(ctx, swipe)
		if err != nil {
			http.Error(w, "Failed to record swipe", http.StatusInternalServerError)
			return
		}

		// Check for mutual match (like or superlike)
		var reverse models.Swipe
		err = swipes.FindOne(ctx, bson.M{
			"fromUser": toID,
			"toUser":   fromID,
			"action": bson.M{
				"$in": []models.SwipeAction{models.LikeSwipe, models.SuperLikeSwipe},
			},
		}).Decode(&reverse)

		if err == nil && (action == models.LikeSwipe || action == models.SuperLikeSwipe) {
			// Mutual match!
			match := models.NewMatch(fromID, toID)
			_, _ = h.DB.Collection("matches").InsertOne(ctx, match)

			// Notify both users via WebSocket (if connected)
			h.WSManager.SendTo(fromID.Hex(), "🎉 It's a match!")
			h.WSManager.SendTo(toID.Hex(), "🎉 You got a match!")

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]any{
				"match": true,
			})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{
			"match": false,
		})
	}
}
