package refdata

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	locationRepo LocationRepository
	sportRepo    SportRepository
	Service      Service
	Handler      Handler
}

func NewModule(log *slog.Logger, pool *pgxpool.Pool) *Module {
	locationRepo := NewLocationRepository(pool)
	sportRepo := NewSportRepository(pool)

	service := NewService(log, locationRepo, sportRepo)

	handler := NewHandler(log, service)

	return &Module{
		locationRepo: locationRepo,
		sportRepo:    sportRepo,
		Service:      service,
		Handler:      *handler,
	}
}
