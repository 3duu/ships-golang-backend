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
			h.WSManager.SendTo(fromID.Hex(), models.ChatMessagePayload{
				Type:     "alert",
				Text:     "🎉 It's a match!",
				FromUser: match.User1,
				Time:     time.Now(),
			})
			h.WSManager.SendTo(toID.Hex(), models.ChatMessagePayload{
				Type:     "alert",
				Text:     "🎉 You got a match!",
				FromUser: match.User2,
				Time:     time.Now(),
			})

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

func (h *Handler) GetYouGotLikedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		userID := r.Context().Value(middlewares.UserIDKey).(string)
		currentUserID, _ := primitive.ObjectIDFromHex(userID)

		// Step 1: Get all swipes where someone liked me
		swipeCol := h.DB.Collection("swipes")
		cursor, err := swipeCol.Find(ctx, bson.M{
			"toUser": currentUserID,
			"action": bson.M{"$in": []string{"like", "superlike"}},
		})
		if err != nil {
			http.Error(w, "Failed to fetch likes", http.StatusInternalServerError)
			return
		}

		var likes []models.Swipe
		if err := cursor.All(ctx, &likes); err != nil {
			http.Error(w, "Failed to decode likes", http.StatusInternalServerError)
			return
		}

		// Step 2: Get users I already swiped on
		var seenUsers []models.Swipe
		seenCursor, _ := swipeCol.Find(ctx, bson.M{"fromUser": currentUserID})
		_ = seenCursor.All(ctx, &seenUsers)

		seenMap := make(map[string]bool)
		for _, s := range seenUsers {
			seenMap[s.ToUser.Hex()] = true
		}

		// Step 3: Filter likes where I haven’t swiped back
		var likersToShow []primitive.ObjectID
		for _, like := range likes {
			if !seenMap[like.FromUser.Hex()] {
				likersToShow = append(likersToShow, like.FromUser)
			}
		}

		// Step 4: Load user profiles
		userCol := h.DB.Collection("users")
		userCursor, err := userCol.Find(ctx, bson.M{
			"_id": bson.M{"$in": likersToShow},
		})
		if err != nil {
			http.Error(w, "Failed to load users", http.StatusInternalServerError)
			return
		}

		var users []models.User
		_ = userCursor.All(ctx, &users)

		for i := range users {
			users[i].Password = ""
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}
