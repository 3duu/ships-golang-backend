package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AuthSession struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	Token     string             `bson:"token" json:"token"`
	ExpiresAt time.Time          `bson:"expiresAt" json:"expiresAt"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UserAgent string             `bson:"userAgent" json:"userAgent"`
	IP        string             `bson:"ip" json:"ip"`
}
