package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	MatchID   primitive.ObjectID `bson:"matchId"`
	FromUser  primitive.ObjectID `bson:"fromUser"`
	Content   string             `bson:"content"`
	CreatedAt time.Time          `bson:"createdAt"`
}

type ChatMessagePayload struct {
	Type     string             `json:"type"`
	MatchID  primitive.ObjectID `bson:"matchId"`
	FromUser primitive.ObjectID `bson:"fromUser"`
	Text     string             `json:"text"`
	Time     time.Time          `json:"time"`
}
