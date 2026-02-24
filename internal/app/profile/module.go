package profile

import (
	"log/slog"

	"github.com/RuLap/sportmates-api/internal/app/refdata"
	"github.com/RuLap/sportmates-api/internal/pkg/storage/minio"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	repo           Repository
	service        Service
	refdataService refdata.Service
	Handler        Handler
}

func NewModule(log *slog.Logger, pool *pgxpool.Pool, minio *minio.Service, refdataService refdata.Service) *Module {
	repo := NewRepository(pool)

	service := NewService(log, minio, repo, refdataService)

	handler := NewHandler(log, service)

	return &Module{
		repo:    repo,
		service: service,
		Handler: *handler,
	}
}
