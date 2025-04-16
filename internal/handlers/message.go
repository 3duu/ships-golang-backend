package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"ships-backend/internal/middlewares"
	"ships-backend/internal/models"
	"time"
)

func (h *Handler) SendMessageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		matchIDStr := mux.Vars(r)["matchId"]
		matchID, err := primitive.ObjectIDFromHex(matchIDStr)
		if err != nil {
			http.Error(w, "Invalid match ID", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value(middlewares.UserIDKey).(string)
		senderID, _ := primitive.ObjectIDFromHex(userID)

		// Verify user is part of this match
		matchCol := h.DB.Collection("matches")
		var match models.Match
		err = matchCol.FindOne(r.Context(), bson.M{
			"_id": matchID,
			"$or": []bson.M{
				{"user1": senderID},
				{"user2": senderID},
			},
		}).Decode(&match)
		if err != nil {
			http.Error(w, "Not authorized to message in this match", http.StatusForbidden)
			return
		}

		var req struct {
			Content string `json:"content"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
			http.Error(w, "Invalid message content", http.StatusBadRequest)
			return
		}

		msg := models.Message{
			MatchID:   matchID,
			FromUser:  senderID,
			Content:   req.Content,
			CreatedAt: time.Now(),
		}

		_, err = h.DB.Collection("messages").InsertOne(r.Context(), msg)
		if err != nil {
			http.Error(w, "Failed to send message", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Message sent",
		})

		// Determine recipient
		receiverID := match.User1
		if match.User1 == senderID {
			receiverID = match.User2
		}

		// Notify via WebSocket if receiver is online
		h.WSManager.SendTo(receiverID.Hex(), models.ChatMessagePayload{
			Type:     "message",
			MatchID:  matchID,
			Text:     req.Content,
			FromUser: match.User1,
			Time:     time.Now(),
		})

	}
}

func (h *Handler) GetMessagesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		matchIDStr := mux.Vars(r)["matchId"]
		matchID, err := primitive.ObjectIDFromHex(matchIDStr)
		if err != nil {
			http.Error(w, "Invalid match ID", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value(middlewares.UserIDKey).(string)
		senderID, _ := primitive.ObjectIDFromHex(userID)

		matchCol := h.DB.Collection("matches")
		var match models.Match
		err = matchCol.FindOne(r.Context(), bson.M{
			"_id": matchID,
			"$or": []bson.M{
				{"user1": senderID},
				{"user2": senderID},
			},
		}).Decode(&match)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		cursor, err := h.DB.Collection("messages").Find(
			r.Context(),
			bson.M{"matchId": matchID},
			options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}),
		)
		if err != nil {
			http.Error(w, "Failed to load messages", http.StatusInternalServerError)
			return
		}

		var messages []models.Message
		if err := cursor.All(r.Context(), &messages); err != nil {
			http.Error(w, "Decode error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	}
}
