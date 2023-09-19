package service

import (
	"context"

	"github.com/ZiganshinDev/medods/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefreshService struct {
	repo repository.RefreshSessions
}

func NewRefreshService(repo repository.RefreshSessions) *RefreshService {
	return &RefreshService{repo: repo}
}

func (s *RefreshService) Insert(ctx context.Context, ip string, refreshToken string) (primitive.ObjectID, error) {
	return s.repo.Insert(ctx, ip, refreshToken)
}

func (s *RefreshService) Delete(ctx context.Context, refreshToken string) error {
	return s.repo.Delete(ctx, refreshToken)
}
