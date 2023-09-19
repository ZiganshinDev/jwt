package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RefreshSessions interface {
	Insert(ctx context.Context, ip string, refreshToken string) (primitive.ObjectID, error)
	Delete(ctx context.Context, refreshToken string) error
}

type Reposiories struct {
	Refresh RefreshRepo
}

func NewRepositories(db *mongo.Database) *Reposiories {
	return &Reposiories{
		Refresh: *NewRefreshRepo(db),
	}
}
