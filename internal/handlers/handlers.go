package handlers

import (
	"go.mongodb.org/mongo-driver/mongo"
	"ships-backend/internal/ws"
)

type Handler struct {
	DB        *mongo.Database
	WSManager *ws.Manager
}

func NewHandler(db *mongo.Database, wsManager *ws.Manager) *Handler {
	return &Handler{
		DB:        db,
		WSManager: wsManager,
	}
}
