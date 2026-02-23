package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/life-journaling/core/internal/adapter/email"
	"github.com/life-journaling/core/internal/adapter/llm"
	"github.com/life-journaling/core/internal/adapter/postgres"
	"github.com/life-journaling/core/internal/config"
	"github.com/life-journaling/core/internal/handler"
	"github.com/life-journaling/core/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := postgres.NewPool(ctx, cfg.DB.DSN())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Repositories
	userRepo := postgres.NewUserRepository(pool)
	memoryRepo := postgres.NewMemoryRepository(pool)
	engagementRepo := postgres.NewEngagementLogRepository(pool)
	portraitRepo := postgres.NewPortraitRepository(pool)

	// External adapters
	emailProvider := email.NewResendProvider(cfg.Resend.APIKey, cfg.Resend.FromEmail)
	llmProvider := llm.NewGeminiProvider(cfg.Gemini.APIKey)

	// Use cases
	userSvc := usecase.NewUserService(userRepo)
	memorySvc := usecase.NewMemoryService(memoryRepo)
	portraitSvc := usecase.NewPortraitService(portraitRepo)
	engagementSvc := usecase.NewEngagementService(
		userRepo, engagementRepo, emailProvider,
	)
	ingestionSvc := usecase.NewIngestionService(
		engagementRepo, memoryRepo, llmProvider,
	)

	// Router
	router := handler.NewRouter(handler.RouterDeps{
		Config:            cfg,
		UserService:       userSvc,
		MemoryService:     memorySvc,
		PortraitService:   portraitSvc,
		EngagementService: engagementSvc,
		IngestionService:  ingestionSvc,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "port", cfg.App.Port, "env", cfg.App.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("server shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}
