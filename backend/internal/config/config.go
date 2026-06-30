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

	// Storage: "memory" (zero-infra dev default) or "mysql".
	DBDriver string
	DBDSN    string

	// IAM (OAuth2/OIDC) integration. Empty IAMIssuer disables real auth.
	IAMIssuer       string // tenant protocol base, e.g. http://localhost:8080/t/devnote
	IAMTenantID     string
	IAMClientID     string
	IAMClientSecret string

	// Session cookie signing secret.
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
}

// Load reads configuration from the environment, applying sensible dev defaults.
func Load() (Config, error) {
	c := Config{
		Port:            env("PORT", "8080"),
		DBDriver:        strings.ToLower(env("DB_DRIVER", "memory")),
		DBDSN:           os.Getenv("DB_DSN"),
		IAMIssuer:       os.Getenv("IAM_ISSUER_URL"),
		IAMTenantID:     os.Getenv("IAM_TENANT_ID"),
		IAMClientID:     os.Getenv("IAM_CLIENT_ID"),
		IAMClientSecret: os.Getenv("IAM_CLIENT_SECRET"),
		SessionSecret:   env("SESSION_SECRET", "dev-insecure-session-secret-change-me"),
		CookieSecure:    env("COOKIE_SECURE", "false") == "true",

		AppBaseURL: env("APP_BASE_URL", "http://localhost:3000"),

		StripeSecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),

		MomoPartnerCode:    os.Getenv("MOMO_PARTNER_CODE"),
		MomoAccessKey:      os.Getenv("MOMO_ACCESS_KEY"),
		MomoSecretKey:      os.Getenv("MOMO_SECRET_KEY"),
		MomoCreateEndpoint: env("MOMO_CREATE_ENDPOINT", "https://test-payment.momo.vn/v3/gateway/api/create"),
		MomoQueryEndpoint:  env("MOMO_QUERY_ENDPOINT", "https://test-payment.momo.vn/v3/gateway/api/query"),
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

// AuthEnabled reports whether IAM integration is configured.
func (c Config) AuthEnabled() bool {
	return c.IAMIssuer != "" && c.IAMClientID != ""
}

// StripeEnabled reports whether real Stripe card payments are configured. When
// false the card method falls back to the demo flow.
func (c Config) StripeEnabled() bool { return c.StripeSecretKey != "" }

// MomoEnabled reports whether real MoMo payments are configured. When false the
// MoMo method falls back to the demo flow.
func (c Config) MomoEnabled() bool {
	return c.MomoPartnerCode != "" && c.MomoAccessKey != "" && c.MomoSecretKey != ""
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
