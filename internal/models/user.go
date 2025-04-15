package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type ProfileUpdateRequest struct {
	Name      string   `json:"name"`
	Bio       string   `json:"bio"`
	Interests []string `json:"interests"`
	Gender    string   `json:"gender"`
	Location  Location `json:"location"`
}

type Location struct {
	Type        string    `bson:"type" json:"type"`               // Always "Point"
	Coordinates []float64 `bson:"coordinates" json:"coordinates"` // [lng, lat]
}

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password,omitempty" json:"-"`
	Bio       string             `bson:"bio,omitempty" json:"bio,omitempty"`
	Gender    string             `bson:"gender" json:"gender"`       // e.g., "male", "female", "non-binary"
	Interests []string           `bson:"interests" json:"interests"` // e.g., ["anime", "gaming", "rock"]
	Location  Location           `bson:"location" json:"location"`   // For geo queries
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}
