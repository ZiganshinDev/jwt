package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Auth struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	RefreshToken string
	Ip           string
}
