package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type SwipeAction string

const (
	LikeSwipe      SwipeAction = "like"
	DislikeSwipe   SwipeAction = "dislike"
	SuperLikeSwipe SwipeAction = "superlike"
)

type Swipe struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	FromUser   primitive.ObjectID `bson:"fromUser"`
	ToUser     primitive.ObjectID `bson:"toUser"`
	Action     SwipeAction        `bson:"action"`
	CreatedAt  time.Time          `bson:"createdAt"`
	ValidUntil time.Time          `bson:"validUntil"`
	Source     string             `bson:"source"` // e.g., "queue", "recommendation", "search"
}
