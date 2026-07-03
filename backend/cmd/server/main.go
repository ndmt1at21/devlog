// Command server is the devnote blog backend: a stdlib net/http API that serves
// content, handles auth in-process (embedded IAM logic: argon2id passwords,
// signed tokens, Google login), and handles Pro/coffee payments. It boots from
// environment configuration (see .env.example) and applies sensible dev
// defaults, so `go run ./cmd/server` works with zero infrastructure.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/authlocal"
	"github.com/ndmt1at21/devlog/backend/internal/config"
	"github.com/ndmt1at21/devlog/backend/internal/handler"
	"github.com/ndmt1at21/devlog/backend/internal/payment"
	"github.com/ndmt1at21/devlog/backend/internal/session"
	"github.com/ndmt1at21/devlog/backend/internal/storage"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

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

	api := &handler.API{Store: store, Cfg: cfg}

	// Auth runs in-process against the blog's own store (embedded IAM logic),
	// so it is always on; the first registered account becomes the author.
	api.Auth = authlocal.New(store, cfg.SessionSecret, authlocal.GoogleConfig{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
	})
	api.Sessions = session.New(cfg.SessionSecret, cfg.CookieSecure)
	if cfg.GoogleClientID != "" {
		log.Printf("auth: embedded (google login enabled)")
	} else {
		log.Printf("auth: embedded (google login disabled; set GOOGLE_CLIENT_ID to enable)")
	}

	// Real payment providers are optional; empty keys fall back to the demo flow.
	if cfg.StripeEnabled() {
		api.Stripe = payment.NewStripe(cfg.StripeSecretKey, cfg.StripeWebhookSecret)
		log.Printf("payments: Stripe enabled")
	}
	if cfg.MomoEnabled() {
		api.Momo = payment.NewMomo(cfg.MomoPartnerCode, cfg.MomoAccessKey, cfg.MomoSecretKey, cfg.MomoCreateEndpoint, cfg.MomoQueryEndpoint)
		log.Printf("payments: MoMo enabled")
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.NewRouter(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Serve until the root context is cancelled, then drain in-flight requests.
	errc := make(chan error, 1)
	go func() {
		log.Printf("listening on :%s (driver=%s)", cfg.Port, cfg.DBDriver)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errc <- err
		}
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		log.Printf("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
