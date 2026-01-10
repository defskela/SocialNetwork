package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/defskela/SocialNetwork/internal/config"
	"github.com/defskela/SocialNetwork/internal/delivery/http"
	"github.com/defskela/SocialNetwork/internal/repository"
	"github.com/defskela/SocialNetwork/internal/repository/postgres"
	"github.com/defskela/SocialNetwork/internal/service"
	"github.com/defskela/SocialNetwork/pkg/client/postgresql"
	"github.com/defskela/SocialNetwork/pkg/migrator"
)

// @title Social Network API
// @version 1.0
// @description REST API for Social Network application

// @host localhost:8080
// @BasePath /api/v1

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.MustLoad()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	migrator.MustRun(&cfg.Postgres, "migrations")

	pgClient, err := postgresql.NewClient(ctx, 3, &cfg.Postgres)
	if err != nil {
		return fmt.Errorf("failed to create postgres client: %w", err)
	}

	defer pgClient.Close()

	userRepo := postgres.NewUserRepository(pgClient)
	repos := repository.NewRepository(userRepo)
	services, err := service.NewService(repos, cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to create services: %w", err)
	}
	handlers := http.NewHandler(services)

	srv := http.NewServer(cfg, handlers.Init())

	go func() {
		if err := srv.Run(); err != nil {
			fmt.Printf("server error: %s\n", err)
		}
	}()

	fmt.Printf("SocialNetwork service started on %s\n", cfg.HTTPServer.Address)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	fmt.Println("SocialNetwork service shutting down")

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("error occurred on server shutting down: %w", err)
	}

	return nil
}
