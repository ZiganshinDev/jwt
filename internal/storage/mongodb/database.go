package mongodb

import "go.mongodb.org/mongo-driver/mongo"

type Storage struct {
	db *mongo.Database
}

func NewStorage(client *mongo.Client, database string) *Storage {
	return &Storage{db: client.Database(database)}
}
