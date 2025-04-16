package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Seen struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"userId"`   // viewer
	SeenUser  primitive.ObjectID `bson:"seenUser"` // who they saw
	Timestamp time.Time          `bson:"timestamp"`
}
