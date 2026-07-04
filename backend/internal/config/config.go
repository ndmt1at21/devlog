// Package config loads runtime configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all runtime settings for the blog backend.
type Config struct {
	Port string

	// Observability. LogLevel is debug|info|warn|error; LogFormat is "json"
	// (one object per line, for log aggregation) or "text" (human-readable dev).
	LogLevel  string
	LogFormat string

	// Storage: "memory" (zero-infra dev default) or "mysql".
	DBDriver string
	DBDSN    string

	// Google OAuth client for "Đăng nhập với Google" (embedded federated
	// login). Empty GoogleClientID disables the flow; email/password auth
	// works regardless.
	GoogleClientID     string
	GoogleClientSecret string

	// Session cookie signing secret. Also keys the embedded auth provider's
	// access tokens (domain-separated).
	SessionSecret string
	// CookieSecure controls the Secure flag (disable for plain-HTTP localhost).
	CookieSecure bool

	// AppBaseURL is the public origin of the frontend; Stripe redirects back here
	// after Checkout (e.g. http://localhost:3000).
	AppBaseURL string

	// Stripe (international card payments via Checkout). Empty keys disable the
	// card method and the modal falls back to the demo flow.
	StripeSecretKey     string
	StripeWebhookSecret string

	// MoMo (Vietnam). Empty keys disable the MoMo method (demo fallback).
	MomoPartnerCode    string
	MomoAccessKey      string
	MomoSecretKey      string
	MomoCreateEndpoint string
	MomoQueryEndpoint  string

	// Object storage for article images: any S3-compatible store (Cloudflare R2
	// in production, MinIO for local dev). Incomplete settings disable uploads.
	S3Endpoint        string // https://<account_id>.r2.cloudflarestorage.com
	S3Bucket          string
	S3Region          string // "auto" on R2
	S3AccessKeyID     string
	S3SecretAccessKey string
	// ImageBaseURL is the public origin the bucket is served from (R2 custom
	// domain behind Cloudflare's CDN, e.g. https://img.example.com). It prefixes
	// every stored image URL, and article bodies may only embed images under it.
	ImageBaseURL string
}

// Load reads configuration from the environment, applying sensible dev defaults.
func Load() (Config, error) {
	c := Config{
		Port:               env("PORT", "8080"),
		LogLevel:           strings.ToLower(env("LOG_LEVEL", "info")),
		LogFormat:          strings.ToLower(env("LOG_FORMAT", "text")),
		DBDriver:           strings.ToLower(env("DB_DRIVER", "memory")),
		DBDSN:              os.Getenv("DB_DSN"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		SessionSecret:      env("SESSION_SECRET", "dev-insecure-session-secret-change-me"),
		CookieSecure:       env("COOKIE_SECURE", "false") == "true",

		AppBaseURL: env("APP_BASE_URL", "http://localhost:3000"),

		StripeSecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),

		MomoPartnerCode:    os.Getenv("MOMO_PARTNER_CODE"),
		MomoAccessKey:      os.Getenv("MOMO_ACCESS_KEY"),
		MomoSecretKey:      os.Getenv("MOMO_SECRET_KEY"),
		MomoCreateEndpoint: env("MOMO_CREATE_ENDPOINT", "https://test-payment.momo.vn/v3/gateway/api/create"),
		MomoQueryEndpoint:  env("MOMO_QUERY_ENDPOINT", "https://test-payment.momo.vn/v3/gateway/api/query"),

		S3Endpoint:        strings.TrimRight(os.Getenv("S3_ENDPOINT"), "/"),
		S3Bucket:          os.Getenv("S3_BUCKET"),
		S3Region:          env("S3_REGION", "auto"),
		S3AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
		S3SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
		ImageBaseURL:      strings.TrimRight(os.Getenv("IMAGE_BASE_URL"), "/"),
	}

	switch c.DBDriver {
	case "memory":
	case "mysql":
		if c.DBDSN == "" {
			return c, fmt.Errorf("DB_DSN is required when DB_DRIVER=mysql")
		}
	default:
		return c, fmt.Errorf("unknown DB_DRIVER %q (want memory|mysql)", c.DBDriver)
	}
	return c, nil
}

// StripeEnabled reports whether real Stripe card payments are configured. When
// false the card method falls back to the demo flow.
func (c Config) StripeEnabled() bool { return c.StripeSecretKey != "" }

// MomoEnabled reports whether real MoMo payments are configured. When false the
// MoMo method falls back to the demo flow.
func (c Config) MomoEnabled() bool {
	return c.MomoPartnerCode != "" && c.MomoAccessKey != "" && c.MomoSecretKey != ""
}

// UploadsEnabled reports whether image uploads are fully configured. When
// false, POST /uploads answers 503 and the editor surfaces the error.
func (c Config) UploadsEnabled() bool {
	return c.S3Endpoint != "" && c.S3Bucket != "" && c.S3AccessKeyID != "" &&
		c.S3SecretAccessKey != "" && c.ImageBaseURL != ""
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
