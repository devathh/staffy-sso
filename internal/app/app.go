package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	staffy "github.com/devathh/staffy-proto/gen/go"
	"github.com/devathh/staffy-sso/internal/application/services"
	"github.com/devathh/staffy-sso/internal/infrastructure/cache/redis"
	"github.com/devathh/staffy-sso/internal/infrastructure/config"
	"github.com/devathh/staffy-sso/internal/infrastructure/persistence/postgres"
	"github.com/devathh/staffy-sso/internal/infrastructure/server"
	"github.com/devathh/staffy-sso/internal/infrastructure/server/handlers"
	"github.com/devathh/staffy-sso/internal/lib/jwt"
	"github.com/devathh/staffy-sso/pkg/log"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

type App struct {
	log    *slog.Logger
	server *server.Server
}

func (a *App) Start() error {
	a.log.Info("server is running")
	return a.server.Start()
}

func (a *App) Shutdown(ctx context.Context) error {
	a.log.Info("server is shutting down")
	return a.server.Shutdown(ctx)
}

// SetupApp returns App, CleanUp() n' err
func SetupApp() (*App, func(), error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, nil, fmt.Errorf("failed to load .env: %w", err)
	}

	cfg, err := config.Load(os.Getenv("APP_CONFIG_PATH"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	logHandler, err := log.SetupHandler(os.Stdout, cfg.App.Env)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init log's handler: %w", err)
	}
	log := slog.New(logHandler)

	log.Info("config is uploaded", slog.Any("server", cfg.Server), slog.Any("service", cfg.App))

	jwtGenerator := jwt.NewJWT(cfg)

	db, err := postgres.ConnectToDB(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	if err := postgres.SyncDB(db); err != nil {
		return nil, nil, fmt.Errorf("failed to sync db: %w", err)
	}

	redisClient, err := redis.ConnectToRedis(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	ur, err := postgres.NewUserRepository(db)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init user's repository: %w", err)
	}

	uc, err := redis.NewUserCache(cfg, redisClient)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init user's cache: %w", err)
	}

	service := services.NewSSOService(cfg, log, ur, uc, jwtGenerator)
	handler := handlers.NewHandler(service)
	grpcServer := grpc.NewServer()
	staffy.RegisterSSOServer(grpcServer, handler)

	server, err := server.NewServer(cfg, grpcServer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init server: %w", err)
	}

	log.Info("all components are loaded")

	cleanup := func() {
		if err := redis.Close(redisClient); err != nil {
			log.Warn("failed to close connection with redis", slog.String("error", err.Error()))
		} else {
			log.Info("redis connection was closed")
		}
	}

	return &App{
		server: server,
		log:    log,
	}, cleanup, nil
}
