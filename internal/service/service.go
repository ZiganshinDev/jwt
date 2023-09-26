package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ZiganshinDev/medods/internal/config"
)

type Storage interface {
	InsertToken(ctx context.Context, userName string, refreshToken string, timeNow time.Time) error
	DeleteToken(ctx context.Context, refreshToken string) error
	DeleteTokensByUser(ctx context.Context, userName string) error
	SwitchToken(ctx context.Context, oldRefreshToken string, newRefreshToken string, userName string, timeNow time.Time) error
	CountTokens(ctx context.Context, userName string) (int64, error)
	GetCreatedTime(ctx context.Context, refreshToken string, userName string) (time.Time, error)
	GetTokenByUser(ctx context.Context, userName string) (string, error)
}

type TokenManager interface {
	NewJWT(userId string, ttl time.Duration) (string, error)
	NewRefreshToken() (string, error)
	HashToken(token string) ([]byte, error)
	CompareTokens(providedToken string, hashedToken []byte) bool
}

type Service struct {
	cfg          *config.Config
	storage      Storage
	tokenManager TokenManager
}

func New(cfg *config.Config, storage Storage, tokenManager TokenManager) (*Service, error) {
	return &Service{
		cfg:          cfg,
		storage:      storage,
		tokenManager: tokenManager}, nil
}

func (s *Service) GetRefreshToken(userName string) (string, error) {
	const op = "service.GetRefreshToken"

	refreshToken, err := s.tokenManager.NewRefreshToken()
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return refreshToken, nil
}

func (s *Service) GetAccessToken(userName string) (string, error) {
	const op = "service.GetAccessToken"

	accessToken, err := s.tokenManager.NewJWT(userName, s.cfg.AccessTokenTTL)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, nil
}

func (s *Service) ValidToken(tokenFromHeader string, userName string) bool {
	tokenFromDB, err := s.getTokenFromDB(userName)
	if err != nil {
		return false
	}

	if ok := s.tokenManager.CompareTokens(tokenFromHeader, []byte(tokenFromDB)); !ok {
		return false
	}

	if ok, err := s.checkTokenTtl(tokenFromDB, userName, time.Now()); err != nil || !ok {
		return false
	}

	return true
}

func (s *Service) CheckCountTokensByUser(userName string) error {
	const op = "service.CheckCountTokensByUser"

	count, err := s.storage.CountTokens(context.TODO(), userName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if count > 0 {
		if err := s.storage.DeleteTokensByUser(context.TODO(), userName); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *Service) InsertToken(refreshToken string, userName string) error {
	const op = "service.InsertToken"

	hashedToken, err := s.tokenManager.HashToken(refreshToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.storage.InsertToken(context.TODO(), userName, string(hashedToken), time.Now()); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) SwitchToken(newToken string, userName string) error {
	const op = "service.switchToken"

	oldToken, err := s.getTokenFromDB(userName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	hashedToken, err := s.tokenManager.HashToken(newToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.storage.SwitchToken(context.TODO(), oldToken, string(hashedToken), userName, time.Now()); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) getTokenFromDB(userName string) (string, error) {
	const op = "service.getTokenFromDB"

	refreshTokenFromDB, err := s.storage.GetTokenByUser(context.Background(), userName)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return refreshTokenFromDB, nil
}

func (s *Service) checkTokenTtl(tokenFromDB string, userName string, time time.Time) (bool, error) {
	const op = "service.checkTokenTtl"

	createdTime, err := s.storage.GetCreatedTime(context.TODO(), tokenFromDB, userName)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if createdTime.Add(s.cfg.JWT.RefreshTokenTTL).Before(time) {
		if err := s.storage.DeleteToken(context.TODO(), tokenFromDB); err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}

		return false, nil
	}

	return true, nil
}
