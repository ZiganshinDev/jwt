package handler

import (
	"context"
	"fmt"
	"log"
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
	ChechUserToken(ctx context.Context, refreshToken string, userName string, ttl time.Duration) bool
	GetCreatedTime(ctx context.Context, refreshToken string, userName string) (time.Time, error)
}

type TokenManager interface {
	NewJWT(userId string, ttl time.Duration) (string, error)
	Parse(accessToken string) (string, error)
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
	const op = "http-server.hanlder.authHandler"

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userName := r.Header.Get(name)
		if userName == "" {
			http.Error(w, fmt.Sprintf("Header '%v' is missing", name), http.StatusBadRequest)
			return
		}

		accessToken, err := h.tokenManager.NewJWT(userName, h.cfg.AccessTokenTTL)
		if err != nil {
			log.Printf("%s: %v", op, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		refreshToken, err := h.tokenManager.NewRefreshToken()
		if err != nil {
			log.Printf("%s: %v", op, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		count, err := h.storage.CountTokens(context.TODO(), userName)
		if err != nil {
			log.Printf("%s: %v", op, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if count > 4 {
			if err := h.storage.DeleteTokensByName(context.TODO(), userName); err != nil {
				log.Printf("%s: %v", op, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		if err := h.storage.InsertToken(context.TODO(), userName, refreshToken, time.Now()); err != nil {
			log.Printf("%s: %v", op, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		setJWTToken(w, accessToken, h.cfg.AccessTokenTTL)

		response := response{
			Name:         userName,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}

		renderJSON(w, response)
	}
}

func (h *Handler) refreshHandler() http.HandlerFunc {
	const op = "http-server.hanlder.refreshHandler"

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		refreshToken := r.Header.Get(token)
		if refreshToken == "" {
			http.Error(w, fmt.Sprintf("Header '%v' is missing", token), http.StatusBadRequest)
			return
		}

		userName := r.Header.Get(name)
		if userName == "" {
			http.Error(w, fmt.Sprintf("Header '%v' is missing", name), http.StatusBadRequest)
			return
		}

		if ok := h.storage.ChechUserToken(context.Background(), refreshToken, userName, h.cfg.JWT.RefreshTokenTTL); !ok {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		accessToken, err := h.tokenManager.NewJWT(userName, h.cfg.AccessTokenTTL)
		if err != nil {
			log.Printf("%s: %v", op, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		newRefreshToken, err := h.tokenManager.NewRefreshToken()
		if err != nil {
			log.Printf("%s: %v", op, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		createdTime, err := h.storage.GetCreatedTime(context.TODO(), refreshToken, userName)
		if err != nil {
			log.Printf("%s: %v", op, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if createdTime.Add(h.cfg.JWT.RefreshTokenTTL).Before(time.Now()) {
			if err := h.storage.DeleteToken(context.TODO(), refreshToken); err != nil {
				log.Printf("%s: %v", op, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return

		} else {
			if err := h.storage.SwitchToken(context.TODO(), refreshToken, newRefreshToken, userName, time.Now()); err != nil {
				log.Printf("%s: %v", op, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		setJWTToken(w, accessToken, h.cfg.AccessTokenTTL)

		response := response{
			Name:         userName,
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken,
		}

		renderJSON(w, response)
	}
}
