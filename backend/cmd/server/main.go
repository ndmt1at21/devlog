// Command server is the devnote blog backend: a stdlib net/http API that serves
// content, handles auth in-process (embedded IAM logic: argon2id passwords,
// signed tokens, Google login), and handles Pro/coffee payments. It boots from
// environment configuration (see .env.example) and applies sensible dev
// defaults, so `go run ./cmd/server` works with zero infrastructure.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/authlocal"
	"github.com/ndmt1at21/devlog/backend/internal/config"
	"github.com/ndmt1at21/devlog/backend/internal/handler"
	"github.com/ndmt1at21/devlog/backend/internal/payment"
	"github.com/ndmt1at21/devlog/backend/internal/platform/logger"
	"github.com/ndmt1at21/devlog/backend/internal/session"
	"github.com/ndmt1at21/devlog/backend/internal/storage"
)

func main() {
	if err := run(); err != nil {
		// slog.Default() is the app logger once run() has configured it, or the
		// stdlib default if we failed before that (e.g. bad config).
		slog.Error("server exited", "err", err.Error())
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Structured logging (log/slog). Set as the process default so any code that
	// logs without a request-scoped logger still uses the configured level/format.
	log := logger.New(os.Stdout, logger.Options{Level: cfg.LogLevel, Format: cfg.LogFormat})
	slog.SetDefault(log)

	// Root context cancelled on SIGINT/SIGTERM for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := storage.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Ping(ctx); err != nil {
		return err
	}

	api := &handler.API{Store: store, Cfg: cfg, Logger: log}

	// Auth runs in-process against the blog's own store (embedded IAM logic),
	// so it is always on; the first registered account becomes the author.
	api.Auth = authlocal.New(store, cfg.SessionSecret, authlocal.GoogleConfig{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
	})
	api.Sessions = session.New(cfg.SessionSecret, cfg.CookieSecure)
	log.Info("auth ready", "mode", "embedded", "google_login", cfg.GoogleClientID != "")

	// Real payment providers are optional; empty keys fall back to the demo flow.
	if cfg.StripeEnabled() {
		api.Stripe = payment.NewStripe(cfg.StripeSecretKey, cfg.StripeWebhookSecret)
		log.Info("payments enabled", "provider", "stripe")
	}
	if cfg.MomoEnabled() {
		api.Momo = payment.NewMomo(cfg.MomoPartnerCode, cfg.MomoAccessKey, cfg.MomoSecretKey, cfg.MomoCreateEndpoint, cfg.MomoQueryEndpoint)
		log.Info("payments enabled", "provider", "momo")
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.NewRouter(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Serve until the root context is cancelled, then drain in-flight requests.
	errc := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", srv.Addr, "driver", cfg.DBDriver)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errc <- err
		}
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		log.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
