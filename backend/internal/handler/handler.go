package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/ndmt1at21/devlog/backend/internal/authn"
	"github.com/ndmt1at21/devlog/backend/internal/config"
	"github.com/ndmt1at21/devlog/backend/internal/domain"
	"github.com/ndmt1at21/devlog/backend/internal/payment"
	"github.com/ndmt1at21/devlog/backend/internal/platform/id"
	"github.com/ndmt1at21/devlog/backend/internal/session"
)

// apiV1 is the version prefix every route lives under.
const apiV1 = "/api/v1"

// API holds the dependencies shared by HTTP handlers.
type API struct {
	Store    domain.Store
	Cfg      config.Config
	Auth     authn.Provider   // nil when IAM is not configured
	Sessions *session.Manager // nil when IAM is not configured
	Stripe   *payment.Stripe  // nil when Stripe is not configured
	Momo     *payment.Momo    // nil when MoMo is not configured
}

// NewRouter wires up the application's HTTP routes (all under /api/v1).
func (a *API) NewRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET "+apiV1+"/health", a.health)

	// Content
	mux.HandleFunc("GET "+apiV1+"/articles", a.listArticles)
	// Publishing requires a session + the IAM "articles:create" permission.
	mux.HandleFunc("POST "+apiV1+"/articles", a.createArticle)
	mux.HandleFunc("GET "+apiV1+"/articles/featured", a.featuredArticle)
	mux.HandleFunc("GET "+apiV1+"/categories", a.categories)
	mux.HandleFunc("GET "+apiV1+"/articles/{slug}", a.getArticle)
	mux.HandleFunc("GET "+apiV1+"/articles/{slug}/comments", a.listComments)
	mux.HandleFunc("POST "+apiV1+"/articles/{slug}/comments", a.createComment)

	// Auth (IAM BFF) + Pro
	mux.HandleFunc("POST "+apiV1+"/auth/login", a.login)
	mux.HandleFunc("POST "+apiV1+"/auth/register", a.register)
	mux.HandleFunc("POST "+apiV1+"/auth/forgot-password", a.forgotPassword)
	mux.HandleFunc("POST "+apiV1+"/auth/logout", a.logout)
	mux.HandleFunc("GET "+apiV1+"/auth/me", a.me)
	// Federated "Đăng nhập với Google" (redirect flow via IAM).
	mux.HandleFunc("GET "+apiV1+"/auth/google/login", a.googleLogin)
	mux.HandleFunc("GET "+apiV1+"/auth/google/callback", a.googleCallback)
	mux.HandleFunc("GET "+apiV1+"/pro/plans", a.proPlans)
	mux.HandleFunc("GET "+apiV1+"/me/subscription", a.getSubscription)
	mux.HandleFunc("POST "+apiV1+"/me/subscription", a.subscribe)

	// "Buy me a coffee" payments (Stripe Checkout + MoMo). Anonymous-allowed.
	mux.HandleFunc("POST "+apiV1+"/coffee/checkout", a.coffeeCheckout)
	mux.HandleFunc("GET "+apiV1+"/coffee/{id}/status", a.coffeeStatus)
	// Provider webhooks (server-to-server). These must be reachable at the
	// backend's public URL directly, not via the Next.js /api rewrite. They
	// return provider-specific bodies, not the JSON envelope.
	mux.HandleFunc("POST "+apiV1+"/webhooks/stripe", a.stripeWebhook)
	mux.HandleFunc("POST "+apiV1+"/webhooks/momo", a.momoWebhook)

	// Outer→inner: assign trace id, log, CORS, resolve the session user.
	return traceMiddleware(logging(cors(a.withSession(mux))))
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusOK, map[string]any{
		"status": "ok",
		"driver": a.Cfg.DBDriver,
		"auth":   a.Cfg.AuthEnabled(),
		"time":   time.Now().UTC(),
	})
}

// traceMiddleware attaches a unique trace id to each request (honoring an
// inbound X-Request-Id), exposing it via context and the X-Trace-Id header so
// every response — and the client — can correlate to server logs.
func traceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Request-Id")
		if traceID == "" {
			traceID = id.NewV7()
		}
		w.Header().Set("X-Trace-Id", traceID)
		next.ServeHTTP(w, r.WithContext(withTraceID(r.Context(), traceID)))
	})
}

// logging logs each request with its method, path, trace id, and duration.
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s trace=%s %s", r.Method, r.URL.Path, traceIDFrom(r.Context()), time.Since(start))
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
