package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Users struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name         string             `bson:"name"`
	RefreshToken string             `bson:"refresh_token"`
}
