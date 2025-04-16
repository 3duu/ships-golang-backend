package handlers

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"net/http"
	"ships-backend/internal/middlewares"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"ships-backend/internal/models"
)

func (h *Handler) UploadPhotoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middlewares.UserIDKey).(string)
		objID, _ := primitive.ObjectIDFromHex(userID)

		photoCol := h.DB.Collection("user_photos")
		count, err := photoCol.CountDocuments(r.Context(), bson.M{"userId": objID})
		if err != nil {
			http.Error(w, "Could not verify photo count", http.StatusInternalServerError)
			return
		}

		if count >= 6 {
			http.Error(w, "Maximum of 6 photos allowed", http.StatusForbidden)
			return
		}

		file, header, err := r.FormFile("photo")
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		mime := header.Header.Get("Content-Type")
		if mime != "image/jpeg" && mime != "image/png" {
			http.Error(w, "Only JPEG/PNG allowed", http.StatusUnsupportedMediaType)
			return
		}

		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read image data", http.StatusInternalServerError)
			return
		}

		photo := models.UserPhoto{
			UserID:    objID,
			Data:      data,
			MimeType:  mime,
			Order:     int(count),
			CreatedAt: time.Now(),
		}

		_, err = h.DB.Collection("user_photos").InsertOne(r.Context(), photo)
		if err != nil {
			http.Error(w, "Failed to store photo", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message":"Photo uploaded"}`))
	}
}

func (h *Handler) GetUserPhotosHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := mux.Vars(r)["userId"]
		objID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := h.DB.Collection("user_photos").Find(
			ctx,
			bson.M{"userId": objID},
			options.Find().SetSort(bson.D{{Key: "order", Value: 1}}),
		)
		if err != nil {
			http.Error(w, "Error loading photos", http.StatusInternalServerError)
			return
		}

		var photos []models.UserPhoto
		if err := cursor.All(ctx, &photos); err != nil {
			http.Error(w, "Decode error", http.StatusInternalServerError)
			return
		}

		// Optional: hide binary data if you’re just listing references or IDs
		for i := range photos {
			photos[i].Data = nil
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(photos)
	}
}

func (h *Handler) GetUserPhotoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID := vars["userId"]
		objID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		var photo models.UserPhoto
		err = h.DB.Collection("user_photos").FindOne(r.Context(), bson.M{"userId": objID}).Decode(&photo)
		if err != nil {
			http.Error(w, "Photo not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", photo.MimeType)
		w.Write(photo.Data)
	}
}

func (h *Handler) UpdatePhotoOrderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middlewares.UserIDKey).(string)
		userObjID, _ := primitive.ObjectIDFromHex(userID)

		var req struct {
			PhotoIDs []string `json:"photoIds"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.PhotoIDs) == 0 {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		photoCol := h.DB.Collection("user_photos")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		for order, photoID := range req.PhotoIDs {
			objID, err := primitive.ObjectIDFromHex(photoID)
			if err != nil {
				continue
			}

			// Ensure photo belongs to the user
			filter := bson.M{"_id": objID, "userId": userObjID}
			update := bson.M{"$set": bson.M{"order": order}}

			_, _ = photoCol.UpdateOne(ctx, filter, update)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Photo order updated",
		})
	}
}

func (h *Handler) DeletePhotoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(middlewares.UserIDKey).(string)
		userObjID, _ := primitive.ObjectIDFromHex(userID)

		photoID := mux.Vars(r)["photoId"]
		objID, err := primitive.ObjectIDFromHex(photoID)
		if err != nil {
			http.Error(w, "Invalid photo ID", http.StatusBadRequest)
			return
		}

		photoCol := h.DB.Collection("user_photos")

		// Check ownership
		var photo models.UserPhoto
		err = photoCol.FindOne(r.Context(), bson.M{"_id": objID}).Decode(&photo)
		if err != nil {
			http.Error(w, "Photo not found", http.StatusNotFound)
			return
		}

		if photo.UserID != userObjID {
			http.Error(w, "Not authorized to delete this photo", http.StatusForbidden)
			return
		}

		// Delete the photo
		_, err = photoCol.DeleteOne(r.Context(), bson.M{"_id": objID})
		if err != nil {
			http.Error(w, "Failed to delete photo", http.StatusInternalServerError)
			return
		}

		// Fetch remaining photos, sorted by order
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := photoCol.Find(ctx, bson.M{"userId": userObjID}, options.Find().SetSort(bson.M{"order": 1}))
		if err != nil {
			http.Error(w, "Failed to reorder photos", http.StatusInternalServerError)
			return
		}

		var photos []models.UserPhoto
		if err := cursor.All(ctx, &photos); err != nil {
			http.Error(w, "Error loading photos", http.StatusInternalServerError)
			return
		}

		// Reassign order starting from 0
		for i, p := range photos {
			if p.Order != i {
				_, _ = photoCol.UpdateOne(ctx,
					bson.M{"_id": p.ID},
					bson.M{"$set": bson.M{"order": i}},
				)
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Photo deleted and reordered",
		})
	}
}
