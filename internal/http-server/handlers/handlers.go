package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/ZiganshinDev/medods/internal/config"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Storage interface {
	Insert(ctx context.Context, ip string, refreshToken string) (primitive.ObjectID, error)
}

// TokenManager provides logic for JWT & Refresh tokens generation and parsing.
type TokenManager interface {
	NewJWT(userId string, ttl time.Duration) (string, error)
	Parse(accessToken string) (string, error)
	NewRefreshToken() (string, error)
}

// TODO
func Init(cfg *config.Config, storage Storage, tokenManager TokenManager) http.Handler {
	return nil
}
