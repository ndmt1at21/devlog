package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/authn"
	"github.com/ndmt1at21/devlog/backend/internal/config"
	"github.com/ndmt1at21/devlog/backend/internal/domain"
	"github.com/ndmt1at21/devlog/backend/internal/payment"
	"github.com/ndmt1at21/devlog/backend/internal/session"
)

// API holds the dependencies shared by HTTP handlers.
type API struct {
	Store    domain.Store
	Cfg      config.Config
	Auth     authn.Provider   // nil when IAM is not configured
	Sessions *session.Manager // nil when IAM is not configured
	Stripe   *payment.Stripe  // nil when Stripe is not configured
	Momo     *payment.Momo    // nil when MoMo is not configured
}

// NewRouter wires up the application's HTTP routes.
func (a *API) NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", a.health)

	// Content
	mux.HandleFunc("GET /api/articles", a.listArticles)
	mux.HandleFunc("GET /api/articles/featured", a.featuredArticle)
	mux.HandleFunc("GET /api/categories", a.categories)
	mux.HandleFunc("GET /api/articles/{slug}", a.getArticle)
	mux.HandleFunc("GET /api/articles/{slug}/comments", a.listComments)
	mux.HandleFunc("POST /api/articles/{slug}/comments", a.createComment)

	// Auth (IAM BFF) + Pro
	mux.HandleFunc("POST /api/auth/login", a.login)
	mux.HandleFunc("POST /api/auth/register", a.register)
	mux.HandleFunc("POST /api/auth/forgot-password", a.forgotPassword)
	mux.HandleFunc("POST /api/auth/logout", a.logout)
	mux.HandleFunc("GET /api/auth/me", a.me)
	mux.HandleFunc("GET /api/pro/plans", a.proPlans)
	mux.HandleFunc("GET /api/me/subscription", a.getSubscription)
	mux.HandleFunc("POST /api/me/subscription", a.subscribe)

	// "Buy me a coffee" payments (Stripe Checkout + MoMo). Anonymous-allowed.
	mux.HandleFunc("POST /api/coffee/checkout", a.coffeeCheckout)
	mux.HandleFunc("GET /api/coffee/{id}/status", a.coffeeStatus)
	// Provider webhooks (server-to-server). These must be reachable at the
	// backend's public URL directly, not via the Next.js /api rewrite.
	mux.HandleFunc("POST /api/webhooks/stripe", a.stripeWebhook)
	mux.HandleFunc("POST /api/webhooks/momo", a.momoWebhook)

	// Resolve the session user for every request.
	return logging(cors(a.withSession(mux)))
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"driver": a.Cfg.DBDriver,
		"auth":   a.Cfg.AuthEnabled(),
		"time":   time.Now().UTC(),
	})
}

// logging logs each request with its method, path, and duration.
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// cors allows the Next.js frontend to call the API during development.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
