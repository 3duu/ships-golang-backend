package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Match struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	User1     primitive.ObjectID `bson:"user1"`
	User2     primitive.ObjectID `bson:"user2"`
	CreatedAt time.Time          `bson:"createdAt"`
}

func NewMatch(userA, userB primitive.ObjectID) Match {
	user1, user2 := userA, userB
	if user1.Hex() > user2.Hex() {
		user1, user2 = user2, user1
	}

	return Match{
		User1:     user1,
		User2:     user2,
		CreatedAt: time.Now(),
	}
}
