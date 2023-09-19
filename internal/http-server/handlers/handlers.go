package handlers

import (
	"net/http"

	"github.com/ZiganshinDev/medods/internal/auth"
	"github.com/ZiganshinDev/medods/internal/config"
	"github.com/ZiganshinDev/medods/internal/service"
)

type Handler struct {
	services     *service.Services
	tokenManager auth.TokenManager
}

func NewHandler(services *service.Services, tokenManager auth.TokenManager) *Handler {
	return &Handler{
		services:     services,
		tokenManager: tokenManager,
	}
}

func (h *Handler) Init(cfg *config.Config) http.Handler {
	return nil
}
