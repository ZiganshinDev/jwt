package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Users struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name         string             `bson:"name"`
	RefreshToken string             `bson:"refresh_token"`
	CreatedTime  time.Time          `bson:"created_time"`
}
