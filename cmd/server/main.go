package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/11DingKing/xinzhou-science-platforms/internal/auth"
	"github.com/11DingKing/xinzhou-science-platforms/internal/config"
	"github.com/11DingKing/xinzhou-science-platforms/internal/httpapi"
	"github.com/11DingKing/xinzhou-science-platforms/internal/platform"
	"github.com/11DingKing/xinzhou-science-platforms/internal/repository"
	"github.com/11DingKing/xinzhou-science-platforms/internal/storage"
	"github.com/11DingKing/xinzhou-science-platforms/internal/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cfg := config.FromEnv()
	db, err := storage.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database open failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	repos := repository.New(db)
	accounts := auth.NewService(repos)
	api := httpapi.New(cfg, repos, accounts)
	api.SetPlatformService(platform.NewService(db))
	jobs := worker.New(repos)
	go jobs.Run(ctx)
	server := &http.Server{Addr: cfg.ListenAddr, Handler: api.Handler(), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
		jobs.Stop()
	}()
	slog.Info("quality governance server listening", "addr", cfg.ListenAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
