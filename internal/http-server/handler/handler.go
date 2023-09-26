package handler

import (
	"fmt"
	"net/http"

	"github.com/ZiganshinDev/medods/internal/config"
)

type Auth interface {
	GetRefreshToken(userName string) (string, error)
	GetAccessToken(userName string) (string, error)
	ValidToken(refreshToken string, userName string) bool
	CheckCountTokensByUser(userName string) error
	InsertToken(refreshToken string, userName string) error
	SwitchToken(newToken string, userName string) error
}

type Logger func(http.Handler) http.Handler

type Handler struct {
	cfg    *config.Config
	auth   Auth
	logger Logger
}

type response struct {
	Name         string `json:"user_name"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func New(cfg *config.Config, auth Auth, logger Logger) *Handler {
	return &Handler{
		cfg:    cfg,
		auth:   auth,
		logger: logger,
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

		if err := h.auth.CheckCountTokensByUser(userName); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		refreshToken, err := h.auth.GetRefreshToken(userName)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := h.auth.InsertToken(refreshToken, userName); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		accessToken, err := h.auth.GetAccessToken(userName)
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

		if err := renderJSON(w, response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
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

		if ok := h.auth.ValidToken(refreshTokenFromHeader, userName); !ok {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		newRefreshToken, err := h.auth.GetRefreshToken(userName)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := h.auth.SwitchToken(newRefreshToken, userName); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		accessToken, err := h.auth.GetAccessToken(userName)
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

		if err := renderJSON(w, response); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}
