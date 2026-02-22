package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RuLap/sportmates-api/internal/pkg/config"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	httpServer *http.Server
}

func New(r *chi.Mux, cfg config.HTTPServer) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.Address,
			Handler:      r,
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

func (s *Server) Run(ctx context.Context) error {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctxShutdown, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctxShutdown)
}
