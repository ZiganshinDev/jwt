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
	InsertToken(ctx context.Context, userName string, refreshToken string) error
	DeleteToken(ctx context.Context, refreshToken string) error
	SwitchToken(ctx context.Context, oldToken string, newToken string, userName string) error
	ChechInRepo(ctx context.Context, refreshToken string) bool
}

type TokenManager interface {
	NewJWT(userId string, ttl time.Duration) (string, error)
	Parse(accessToken string) (string, error)
	NewRefreshToken() (string, error)
}

type Handler struct {
	storage      Storage
	tokenManager TokenManager
	cfg          *config.Config
}

type response struct {
	Name         string `json:"user_name"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func New(cfg *config.Config, storage Storage, tokenManager TokenManager) *Handler {
	return &Handler{
		storage:      storage,
		tokenManager: tokenManager,
		cfg:          cfg,
	}
}

func (h *Handler) NewRouter() http.Handler {
	router := http.NewServeMux()

	router.Handle("/auth", http.HandlerFunc(h.authHandler()))

	router.Handle("/refresh", http.HandlerFunc(h.refreshHandler()))

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

		if err := h.storage.InsertToken(context.TODO(), userName, refreshToken); err != nil {
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

		if ok := h.storage.ChechInRepo(context.Background(), refreshToken); !ok {
			http.Error(w, "Bad Request", http.StatusBadRequest)
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

		newRefreshToken, err := h.tokenManager.NewRefreshToken()
		if err != nil {
			log.Printf("%s: %v", op, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := h.storage.SwitchToken(context.TODO(), refreshToken, newRefreshToken, userName); err != nil {
			log.Printf("%s: %v", op, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
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
