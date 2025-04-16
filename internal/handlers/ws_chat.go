package handlers

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"ships-backend/internal/middlewares"
	"ships-backend/internal/models"
	"time"
)

func (h *Handler) WebSocketChatHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.Context().Value(middlewares.UserIDKey).(string)
		userObjID, _ := primitive.ObjectIDFromHex(userIDStr)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
			return
		}

		h.WSManager.AddClient(userIDStr, conn)
		defer h.WSManager.RemoveClient(userIDStr)

		for {
			_, rawMsg, err := conn.ReadMessage()
			if err != nil {
				break
			}

			var payload map[string]interface{}
			if err := json.Unmarshal(rawMsg, &payload); err != nil {
				continue
			}

			msgType, _ := payload["type"].(string)

			switch msgType {
			case "typing":
				matchIDStr, ok := payload["matchId"].(string)
				if !ok || matchIDStr == "" {
					continue
				}

				matchID, err := primitive.ObjectIDFromHex(matchIDStr)
				if err != nil {
					continue
				}

				// Check match
				var match struct {
					User1 primitive.ObjectID `bson:"user1"`
					User2 primitive.ObjectID `bson:"user2"`
				}
				err = h.DB.Collection("matches").FindOne(r.Context(), bson.M{
					"_id": matchID,
					"$or": []bson.M{
						{"user1": userObjID},
						{"user2": userObjID},
					},
				}).Decode(&match)
				if err != nil {
					continue
				}

				// Determine recipient
				toUser := match.User1
				if toUser == userObjID {
					toUser = match.User2
				}

				h.WSManager.SendTo(toUser.Hex(), /*map[string]interface{}{
						"type":    "typing",
						"from":    userIDStr,
						"matchId": matchID.Hex(),
						"sentAt":  time.Now(),
					}*/

					models.ChatMessagePayload{
						Type:     msgType,
						FromUser: userObjID,
						MatchID:  matchID,
						Time:     time.Now(),
					})
			}
		}
	}
}
