package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type UserPhoto struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"userId"`
	Data      []byte             `bson:"data"`     // raw image bytes
	MimeType  string             `bson:"mimeType"` // e.g. image/jpeg
	CreatedAt time.Time          `bson:"createdAt"`
	Order     int                `bson:"order"`
}

type UpdatePhotoOrderRequest struct {
	PhotoIDs []string `json:"photoIds"`
}
