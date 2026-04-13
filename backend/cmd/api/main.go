package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/auth"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/config"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/db"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/router"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/storage"
)

func main() {
	_ = godotenv.Load()

	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}
	if cfg.Env == "production" {
		log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("connect postgres")
	}
	defer pool.Close()

	if err := db.RunMigrations(cfg.DBURL, "internal/db/migrations"); err != nil {
		log.Fatal().Err(err).Msg("run migrations")
	}

	mc, err := storage.New(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("connect minio")
	}

	jwtMgr := auth.NewManager(cfg.JWTSecret, cfg.HRPin)

	r := router.New(cfg, pool, mc, jwtMgr)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.Port).Str("env", cfg.Env).Msg("api listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelShutdown()
	_ = srv.Shutdown(shutdownCtx)
}
