package mongodb

import (
	"context"
	"fmt"

	"github.com/ZiganshinDev/medods/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Storage struct {
	db *mongo.Database
}

func NewStorage(client *mongo.Client, database string) *Storage {
	return &Storage{db: client.Database(database)}
}

type RefreshRepo struct {
	db *mongo.Collection
}

const (
	usersCollection = "users"
	rToken          = "refresh_token"
)

func (s *Storage) NewRefreshRepo() *RefreshRepo {
	return &RefreshRepo{
		db: s.db.Collection(usersCollection),
	}
}

func (r *RefreshRepo) InsertToken(ctx context.Context, userName string, refreshToken string) error {
	const op = "storage.mongodb.Insert"

	if _, err := r.db.InsertOne(ctx, models.Users{
		Name:         userName,
		RefreshToken: refreshToken,
	}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *RefreshRepo) DeleteToken(ctx context.Context, token string) error {
	const op = "storage.mongodb.Delete"

	filter := bson.M{rToken: token}

	if _, err := r.db.DeleteOne(ctx, filter); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *RefreshRepo) SwitchToken(ctx context.Context, oldToken string, newToken string, userName string) error {
	const op = "storage.mongodb.SwitchToken"

	filter := bson.M{rToken: oldToken}
	if _, err := r.db.DeleteOne(ctx, filter); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if _, err := r.db.InsertOne(ctx, models.Users{
		Name:         userName,
		RefreshToken: newToken,
	}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *RefreshRepo) ChechInRepo(ctx context.Context, refreshToken string) bool {
	filter := bson.M{"refresh_token": refreshToken}

	var user models.Users
	err := r.db.FindOne(ctx, filter).Decode(&user)

	return err == nil
}
