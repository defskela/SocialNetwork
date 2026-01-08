package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"social-network/internal/config"
	"social-network/internal/delivery/http"
	"social-network/internal/repository"
	"social-network/internal/repository/postgres"
	"social-network/internal/service"
	"social-network/pkg/client/postgresql"
)

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

	pgClient, err := postgresql.NewClient(ctx, 3, cfg.Postgres)
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
		return fmt.Errorf("error occured on server shutting down: %w", err)
	}

	return nil
}
