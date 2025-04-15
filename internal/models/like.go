package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Like struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FromUser  primitive.ObjectID `bson:"fromUser" json:"fromUser"`
	ToUser    primitive.ObjectID `bson:"toUser" json:"toUser"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}
