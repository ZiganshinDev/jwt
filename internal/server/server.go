package server

import (
	"context"
	"net/http"

	"github.com/ZiganshinDev/medods/internal/config"
)

type Server struct {
	httpServer *http.Server
}

func New(cfg *config.Config, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.HTTPServer.Address,
			Handler:      handler,
			ReadTimeout:  cfg.HTTPServer.Timeout,
			WriteTimeout: cfg.HTTPServer.Timeout,
			IdleTimeout:  cfg.HTTPServer.IdleTimeout,
		},
	}
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
