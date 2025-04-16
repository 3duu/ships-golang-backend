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
	Type        string    `bson:"type" json:"type"`
	Coordinates []float64 `bson:"coordinates" json:"coordinates"`
}

type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `bson:"name" json:"name"`
	Email         string             `bson:"email" json:"email"`
	Password      string             `bson:"password,omitempty" json:"-"`
	Bio           string             `bson:"bio,omitempty" json:"bio,omitempty"`
	Gender        string             `bson:"gender" json:"gender"`       // e.g., "male", "female", "non-binary"
	Interests     []string           `bson:"interests" json:"interests"` // e.g., ["anime", "gaming", "rock"]
	Birth         time.Time
	Location      Location  `bson:"location" json:"location"` // For geo queries
	EmailVerified bool      `bson:"emailVerified" json:"emailVerified"`
	VerifyToken   string    `bson:"verifyToken,omitempty" json:"-"`
	CreatedAt     time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time `bson:"updatedAt" json:"updatedAt"`
}

// Hex returns the string version of the user's ObjectID
func (u *User) Hex() string {
	return u.ID.Hex()
}

type RegisterRequest struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Bio       string    `json:"bio"`
	Gender    string    `json:"gender"`
	Birth     time.Time `json:"birth"`
	Interests []string  `json:"interests"`
	Location  Location  `json:"location"`
}
