package main

import (
	"context"
	"time"

	mail_services "github.com/RuLap/sportmates-api/internal/app/mail/services"
	"github.com/RuLap/sportmates-api/internal/app/profile"
	"github.com/RuLap/sportmates-api/internal/app/refdata"
	"github.com/RuLap/sportmates-api/internal/app/user"
	"github.com/RuLap/sportmates-api/internal/pkg/config"
	"github.com/RuLap/sportmates-api/internal/pkg/http"
	"github.com/RuLap/sportmates-api/internal/pkg/jwthelper"
	"github.com/RuLap/sportmates-api/internal/pkg/logger"
	"github.com/RuLap/sportmates-api/internal/pkg/middleware"
	"github.com/RuLap/sportmates-api/internal/pkg/rabbitmq"
	"github.com/RuLap/sportmates-api/internal/pkg/redis"
	"github.com/RuLap/sportmates-api/internal/pkg/server"
	postgres "github.com/RuLap/sportmates-api/internal/pkg/storage"
	"github.com/RuLap/sportmates-api/internal/pkg/storage/minio"
	validation "github.com/RuLap/sportmates-api/internal/pkg/validator"
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	//Config-----------------------------------------------------------------------------------------------------------

	cfg := config.MustLoad()

	logger := logger.New(logger.Config{
		Level:   cfg.Env,
		LokiURL: cfg.Log.LokiURL,
		Labels:  cfg.Log.LokiLabels,
	})

	validation.Init()

	//Additional Services-----------------------------------------------------------------------------------------------
	redisClient, err := redis.NewClient(cfg.Redis, logger)
	if err != nil {
		logger.Error("failed to connect to Redis", "error", err)
		return
	}
	logger.Info("init redis client successfully")

	redisService := redis.NewService(redisClient)
	logger.Info("init redis service successfully")

	mqClient, err := rabbitmq.NewClient(&cfg.RabbitMQ, logger)
	if err != nil {
		logger.Error("failed to connect to RabbitMQ", "error", err)
		return
	}
	defer mqClient.Close()
	logger.Info("init rabbitmq client successfully")

	mqService := rabbitmq.NewService(mqClient)
	logger.Info("init rabbitmq service successfully")

	storage, err := postgres.InitDB(cfg.PostgresConnString)
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		return
	}

	jwtHelper, err := jwthelper.NewJwtHelper(cfg.JWT.Secret)
	if err != nil {
		logger.Error("failed to create JWT helper", "error", err)
		return
	}

	minioClient, err := minio.New(&cfg.MinioConfig)
	if err != nil {
		logger.Error("failed to init MinIO: %w", err)
	}

	minioService := minio.NewService(minioClient)

	//Modules----------------------------------------------------------------------------------------------------------

	authModule := user.NewModule(logger, storage.Database(), jwtHelper, redisService, mqService)
	refdataModule := refdata.NewModule(logger, storage.Database())
	profileModule := profile.NewModule(logger, storage.Database(), minioService, refdataModule.Service)

	var mailService *mail_services.MailService
	if mqService != nil {
		mailService = mail_services.NewMailService(
			logger,
			mqService,
			&cfg.SMTP,
		)

		go func() {
			logger.Info("starting mail service consumer")
			if err := mailService.StartConsumer(context.Background()); err != nil {
				logger.Error("mail service consumer failed", "error", err)
			}
		}()
	} else {
		logger.Warn("mail service not started - RabbitMQ not available")
	}
	logger.Info("Init mail service successfully")

	//Router-----------------------------------------------------------------------------------------------------------

	router := chi.NewRouter()

	router.Use(chi_middleware.RequestID)
	router.Use(chi_middleware.RealIP)
	router.Use(http.RequestLogger(logger))
	router.Use(http.Recover(logger))
	router.Use(chi_middleware.Timeout(60 * time.Second))

	router.Route("/users", func(r chi.Router) {
		r.Post("/register", authModule.Handler.Register)
		r.Post("/login", authModule.Handler.Login)
		r.Post("/refresh", authModule.Handler.RefreshTokens)

		r.Route("/email", func(r chi.Router) {
			r.Post("/confirm", authModule.Handler.ConfirmEmail)
			r.With(middleware.AuthMiddleware(jwtHelper)).
				Post("/send-confirmation", authModule.Handler.SendConfirmationLink)
			r.With(middleware.AuthMiddleware(jwtHelper)).
				Get("/confirmed", authModule.Handler.CheckEmailConfirmed)
		})

		r.With(middleware.AuthMiddleware(jwtHelper)).Post("/logout", authModule.Handler.Logout)
	})

	router.Route("/profiles", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtHelper))

		r.Get("/{id}", profileModule.Handler.GetUserByID)
		r.Get("/avatar/upload-url", profileModule.Handler.GetAvatarUploadURL)
		r.Post("/avatar", profileModule.Handler.ConfirmAvatarUpload)
	})

	//Server-----------------------------------------------------------------------------------------------------------

	srv := server.New(router, cfg.HTTPServer)
	logger.Info("starting", "address", cfg.HTTPServer.Address)

	if err := srv.Run(context.Background()); err != nil {
		logger.Error("server stopped with error", "error", err)
	}
}
