package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type CrossedPath struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	User1        primitive.ObjectID `bson:"user1"`
	User2        primitive.ObjectID `bson:"user2"`
	Timestamp    time.Time          `bson:"timestamp"`
	Location     Location           `bson:"location"` // where they crossed
	TimesCrossed int                `bson:"timesCrossed"`
	OtherUser    *User              `bson:"-" json:"otherUser,omitempty"`
}
