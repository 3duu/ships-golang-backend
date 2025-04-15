package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Match struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	User1     primitive.ObjectID `bson:"user1" json:"user1"`
	User2     primitive.ObjectID `bson:"user2" json:"user2"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}
