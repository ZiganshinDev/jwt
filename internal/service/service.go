package service

import (
	"context"
	"time"

	"github.com/ZiganshinDev/medods/internal/auth"
	"github.com/ZiganshinDev/medods/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type RefreshSessions interface {
	Insert(ctx context.Context, ip string, refreshToken string) (primitive.ObjectID, error)
	Delete(ctx context.Context, refreshToken string) error
}

type Services struct {
	RefreshSessions RefreshSessions
}

type Deps struct {
	Repos           *repository.Reposiories
	TokenManager    auth.TokenManager
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func NewServices(deps Deps) *Services {
	refreshService := NewRefreshService(&deps.Repos.Refresh)

	return &Services{
		RefreshSessions: refreshService,
	}
}
