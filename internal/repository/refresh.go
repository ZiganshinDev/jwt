package repository

import (
	"context"
	"fmt"

	"github.com/ZiganshinDev/medods/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	refreshSession = "refresh_session"
)

type RefreshRepo struct {
	db *mongo.Collection
}

func NewRefreshRepo(db *mongo.Database) *RefreshRepo {
	return &RefreshRepo{
		db: db.Collection(refreshSession),
	}
}

func (r *RefreshRepo) Insert(ctx context.Context, ip string, refreshToken string) (primitive.ObjectID, error) {
	const op = "storage.mongodb.Insert"

	res, err := r.db.InsertOne(ctx, models.Auth{
		RefreshToken: refreshToken,
		Ip:           ip,
	})

	return res.InsertedID.(primitive.ObjectID), fmt.Errorf("%s: %w", op, err)
}

func (r *RefreshRepo) Delete(ctx context.Context, refreshToken string) error {
	const op = "storage.mongodb.Delete"

	_, err := r.db.DeleteOne(ctx, models.Auth{
		RefreshToken: refreshToken,
	})

	return fmt.Errorf("%s: %w", op, err)
}
