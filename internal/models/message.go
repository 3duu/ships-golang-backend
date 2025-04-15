package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Message struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MatchID  primitive.ObjectID `bson:"matchId" json:"matchId"`
	SenderID primitive.ObjectID `bson:"senderId" json:"senderId"`
	Text     string             `bson:"text" json:"text"`
	SentAt   time.Time          `bson:"sentAt" json:"sentAt"`
}
