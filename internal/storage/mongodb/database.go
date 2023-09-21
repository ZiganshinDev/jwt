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
	name            = "name"
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

func (r *RefreshRepo) DeleteToken(ctx context.Context, refreshToken string) error {
	const op = "storage.mongodb.Delete"

	filter := bson.M{rToken: refreshToken}

	if _, err := r.db.DeleteOne(ctx, filter); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *RefreshRepo) DeleteTokensByName(ctx context.Context, userName string) error {
	const op = "storage.mongodb.Delete"

	filter := bson.M{name: userName}

	if _, err := r.db.DeleteMany(ctx, filter); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *RefreshRepo) SwitchToken(ctx context.Context, oldRefreshToken string, newRefreshToken string, userName string) error {
	const op = "storage.mongodb.SwitchToken"

	filter := bson.M{rToken: oldRefreshToken}
	if _, err := r.db.DeleteOne(ctx, filter); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if _, err := r.db.InsertOne(ctx, models.Users{
		Name:         userName,
		RefreshToken: newRefreshToken,
	}); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *RefreshRepo) Count(ctx context.Context, userName string) (int64, error) {
	const op = "storage.mongodb.Count"

	filter := bson.M{name: userName}

	count, err := r.db.CountDocuments(context.Background(), filter)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return count, nil
}

func (r *RefreshRepo) ChechInRepo(ctx context.Context, refreshToken string, userName string) bool {
	filter := bson.M{rToken: refreshToken, name: userName}

	var user models.Users
	err := r.db.FindOne(ctx, filter).Decode(&user)

	return err == nil
}
