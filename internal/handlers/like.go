package handlers

/*func LikeUserHandler(h *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fromUserID := r.Context().Value(middlewares.UserIDKey).(string)
		toUserID := mux.Vars(r)["userId"]

		if fromUserID == toUserID {
			http.Error(w, "You cannot like yourself", http.StatusBadRequest)
			return
		}

		fromID, err := primitive.ObjectIDFromHex(fromUserID)
		toID, err2 := primitive.ObjectIDFromHex(toUserID)
		if err != nil || err2 != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		likes := h.DB.Collection("likes")
		matches := h.DB.Collection("matches")

		// Check if the like already exists
		existing := likes.FindOne(ctx, bson.M{
			"fromUser": fromID,
			"toUser":   toID,
		})
		if existing.Err() == nil {
			http.Error(w, "You already liked this user", http.StatusConflict)
			return
		}

		// Create new Like
		_, err = likes.InsertOne(ctx, models.Like{
			FromUser:  fromID,
			ToUser:    toID,
			CreatedAt: time.Now(),
		})
		if err != nil {
			http.Error(w, "Failed to like user", http.StatusInternalServerError)
			return
		}

		// Check if mutual like exists
		var reverseLike models.Like
		err = likes.FindOne(ctx, bson.M{
			"fromUser": toID,
			"toUser":   fromID,
		}).Decode(&reverseLike)

		if err == nil {
			// Create a match
			_, err = matches.InsertOne(ctx, models.NewMatch(fromID, toID))

			if err != nil {
				http.Error(w, "Failed to create match", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"match": true}`))

			// Notify both users (if connected)
			h.WSManager.SendTo(fromID.Hex(), "🎉 It's a match with someone!")
			h.WSManager.SendTo(toID.Hex(), "🎉 It's a match with someone!")

			return
		}

		// Not a match (just a like)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"match": false}`))
	}
}*/
