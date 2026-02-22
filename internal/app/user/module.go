package user

import (
	"log/slog"

	"github.com/RuLap/sportmates-api/internal/pkg/jwthelper"
	"github.com/RuLap/sportmates-api/internal/pkg/rabbitmq"
	"github.com/RuLap/sportmates-api/internal/pkg/redis"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	Repo    Repository
	Service Service
	Handler Handler
}

func NewModule(
	log *slog.Logger,
	pool *pgxpool.Pool,
	jwtHelper *jwthelper.JWTHelper,
	redis *redis.Service,
	rabbitmq *rabbitmq.Service,
) *Module {
	repo := NewRepository(pool)
	service := NewService(log, jwtHelper, redis, rabbitmq, repo)
	handler := NewHandler(service)

	return &Module{
		Repo:    repo,
		Service: service,
		Handler: *handler,
	}
}
