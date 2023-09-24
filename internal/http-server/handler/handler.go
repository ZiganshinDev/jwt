package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ZiganshinDev/medods/internal/config"
)

type Storage interface {
	InsertToken(ctx context.Context, userName string, refreshToken string, timeNow time.Time) error
	DeleteToken(ctx context.Context, refreshToken string) error
	DeleteTokensByName(ctx context.Context, userName string) error
	SwitchToken(ctx context.Context, oldRefreshToken string, newRefreshToken string, userName string, timeNow time.Time) error
	CountTokens(ctx context.Context, userName string) (int64, error)
	GetCreatedTime(ctx context.Context, refreshToken string, userName string) (time.Time, error)
	GetTokenByUser(ctx context.Context, userName string) (string, error)
}

type TokenManager interface {
	NewJWT(userId string, ttl time.Duration) (string, error)
	NewRefreshToken() (string, error)
}

type Logger func(http.Handler) http.Handler

type Handler struct {
	cfg          *config.Config
	storage      Storage
	tokenManager TokenManager
	logger       Logger
}

type response struct {
	Name         string `json:"user_name"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func New(cfg *config.Config, storage Storage, tokenManager TokenManager, logger Logger) *Handler {
	return &Handler{
		cfg:          cfg,
		storage:      storage,
		tokenManager: tokenManager,
		logger:       logger,
	}
}

func (h *Handler) NewRouter() http.Handler {
	router := http.NewServeMux()

	authHandlerWithLogger := h.logger(h.authHandler())
	router.Handle("/auth", authHandlerWithLogger)

	refreshHandlerWithLogger := h.logger(h.refreshHandler())
	router.Handle("/refresh", refreshHandlerWithLogger)

	return router
}

const (
	name  = "Name"
	token = "Token"
)

func (h *Handler) authHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userName, err := getHeader(r, name)
		if err != nil {
			http.Error(w, fmt.Sprintf("Header '%v' is missing", name), http.StatusBadRequest)
			return
		}

		if err := h.countTokensByUser(userName); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		refreshToken, err := h.createRefreshToken()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := h.insertToken(refreshToken, userName); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		accessToken, err := h.createJWTToken(userName)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		setCookies(w, refreshToken, accessToken, h.cfg.RefreshTokenTTL, h.cfg.AccessTokenTTL)

		response := response{
			Name:         userName,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}

		renderJSON(w, response)
	}
}

func (h *Handler) refreshHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		refreshTokenFromHeader, err := getHeader(r, token)
		if err != nil {
			http.Error(w, fmt.Sprintf("Header '%v' is missing", token), http.StatusBadRequest)
			return
		}

		userName, err := getHeader(r, name)
		if err != nil {
			http.Error(w, fmt.Sprintf("Header '%v' is missing", name), http.StatusBadRequest)
			return
		}

		if ok := h.isTokenCorrectly(refreshTokenFromHeader, userName); !ok {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		newRefreshToken, err := h.createRefreshToken()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := h.switchToken(newRefreshToken, userName); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		accessToken, err := h.createJWTToken(userName)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		setCookies(w, newRefreshToken, accessToken, h.cfg.RefreshTokenTTL, h.cfg.AccessTokenTTL)

		response := response{
			Name:         userName,
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken,
		}

		renderJSON(w, response)
	}
}

func (h *Handler) countTokensByUser(userName string) error {
	const op = "http-server.handler.countTokens"

	count, err := h.storage.CountTokens(context.TODO(), userName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if count > 0 {
		if err := h.storage.DeleteTokensByName(context.TODO(), userName); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (h *Handler) checkTokenTtl(tokenFromDB string, userName string) (bool, error) {
	const op = "http-server.handler.checkTokenTtl"

	createdTime, err := h.storage.GetCreatedTime(context.TODO(), tokenFromDB, userName)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if createdTime.Add(h.cfg.JWT.RefreshTokenTTL).Before(time.Now()) {
		if err := h.storage.DeleteToken(context.TODO(), tokenFromDB); err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}

		return false, nil
	}

	return true, nil
}

func (h *Handler) createJWTToken(userName string) (string, error) {
	const op = "http-server.handler.createJWTToken"

	accessToken, err := h.tokenManager.NewJWT(userName, h.cfg.AccessTokenTTL)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, nil
}

func (h *Handler) createRefreshToken() (string, error) {
	const op = "http-server.handler.createRefreshToken"

	refreshToken, err := h.tokenManager.NewRefreshToken()
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return refreshToken, nil
}

func (h *Handler) insertToken(refreshToken string, userName string) error {
	const op = "http-server.handler.insertToken"

	hashedToken, err := hashToken(refreshToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := h.storage.InsertToken(context.TODO(), userName, string(hashedToken), time.Now()); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (h *Handler) getTokenFromDB(userName string) (string, error) {
	const op = "http-server.handler.getTokenFromDB"

	refreshTokenFromDB, err := h.storage.GetTokenByUser(context.Background(), userName)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return refreshTokenFromDB, nil
}

func (h *Handler) isTokenCorrectly(tokenFromHeader string, userName string) bool {
	tokenFromDB, err := h.getTokenFromDB(userName)
	if err != nil {
		return false
	}

	if ok := compareTokens(tokenFromHeader, []byte(tokenFromDB)); !ok {
		return false
	}

	if ok, err := h.checkTokenTtl(tokenFromDB, userName); err != nil || !ok {
		return false
	}

	return true
}

func (h *Handler) switchToken(newToken string, userName string) error {
	const op = "http-server.handler.switchToken"

	oldToken, err := h.getTokenFromDB(userName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	hashedToken, err := hashToken(newToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := h.storage.SwitchToken(context.TODO(), oldToken, string(hashedToken), userName, time.Now()); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
